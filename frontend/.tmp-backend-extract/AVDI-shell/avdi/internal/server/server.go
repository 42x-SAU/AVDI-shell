package server

import (
    "context"
    "database/sql"
    "net/http"
    "os"
    "time"

    "diag-system/internal/queue"
    "diag-system/internal/redis"
    "diag-system/internal/storage"
)

type Server struct {
    db    *sql.DB
    queue *queue.PGQueue
    mux   *http.ServeMux
    addr  string
    hub   *Hub
}

func New() (*Server, error) {
    db, err := storage.OpenPostgres()
    if err != nil {
        return nil, err
    }
    if err := db.Ping(); err != nil {
        return nil, err
    }

    // Создаём Redis клиент для лимитера, WebSocket и повторных попыток
    redisClient, err := redis.NewClientFromEnv()
    var limiter *queue.Limiter
    if err != nil {
        // Если Redis недоступен, логируем ошибку, но продолжаем без лимитера и повторных попыток
        // В production следует решить, нужно ли падать или работать без ограничений
        // Пока просто оставляем limiter = nil
    } else {
        limiter = queue.NewLimiter(redisClient, 0) // 0 означает использовать значение из окружения
    }

    // Создаём очередь с лимитером
    q := queue.NewWithLimiter(db, limiter)

    var retryManager *queue.RetryManager
    if redisClient != nil {
        // Создаём менеджер повторных попыток с очередью
        retryManager = queue.NewRetryManager(redisClient, q)
        q.SetRetryManager(retryManager)
    }

    s := &Server{
        db:    db,
        queue: q,
        mux:   http.NewServeMux(),
        addr:  getenv("SERVER_ADDR", ":8080"),
        hub:   NewHub(),
    }
    // Запускаем фоновый воркер для обработки отложенных задач
    if retryManager != nil {
        go retryManager.StartWorker(context.Background(), 30*time.Second)
    }

    s.routes()
    // Запускаем хаб в горутине
    go s.hub.Run()
    return s, nil
}

func (s *Server) routes() {
    s.mux.HandleFunc("GET /health", s.handleHealth)
    s.mux.HandleFunc("GET /agents", s.handleListAgents)
    s.mux.HandleFunc("POST /agents/register", s.handleRegisterAgent)
    s.mux.HandleFunc("POST /agents/heartbeat", s.withAgentAuth(s.handleHeartbeat))
    s.mux.HandleFunc("GET /tasks", s.handleListTasks)
    s.mux.HandleFunc("POST /tasks", s.handleCreateTask)
    s.mux.HandleFunc("GET /results", s.handleListResults)
    s.mux.HandleFunc("GET /agents/tasks/next", s.withAgentAuth(s.handleNextTask))
    s.mux.HandleFunc("POST /agents/tasks/result", s.withAgentAuth(s.handleSubmitResult))
    // Ручной перезапуск задачи
    s.mux.HandleFunc("POST /tasks/{id}/retry", s.handleRetryTask)
    // WebSocket endpoint
    s.mux.HandleFunc("GET /ws", s.handleWebSocket)
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
    // Проверяем, включен ли WebSocket (опционально)
    if enabled := os.Getenv("WEBSOCKET_ENABLED"); enabled == "false" {
        http.Error(w, "WebSocket is disabled", http.StatusServiceUnavailable)
        return
    }
    ServeWebSocket(s.hub, w, r)
}

func (s *Server) Run() error {
    return http.ListenAndServe(s.addr, s.mux)
}

func getenv(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}