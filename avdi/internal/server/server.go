package server

import (
    "database/sql"
    "net/http"
    "os"

    "diag-system/internal/queue"
    "diag-system/internal/storage"
)

type Server struct {
    db    *sql.DB
    queue *queue.PGQueue
    mux   *http.ServeMux
    addr  string
}

func New() (*Server, error) {
    db, err := storage.OpenPostgres()
    if err != nil {
        return nil, err
    }
    if err := db.Ping(); err != nil {
        return nil, err
    }

    s := &Server{
        db:    db,
        queue: queue.New(db),
        mux:   http.NewServeMux(),
        addr:  getenv("SERVER_ADDR", ":8080"),
    }
    s.routes()
    return s, nil
}

func (s *Server) routes() {
    s.mux.HandleFunc("GET /health", s.handleHealth)
    s.mux.HandleFunc("POST /agents/register", s.handleRegisterAgent)
    s.mux.HandleFunc("POST /agents/heartbeat", s.withAgentAuth(s.handleHeartbeat))
    s.mux.HandleFunc("POST /tasks", s.handleCreateTask)
    s.mux.HandleFunc("GET /agents/tasks/next", s.withAgentAuth(s.handleNextTask))
    s.mux.HandleFunc("POST /agents/tasks/result", s.withAgentAuth(s.handleSubmitResult))
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