package server

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"diag-system/internal/queue"
	"diag-system/internal/storage"
)

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleRegisterAgent(w http.ResponseWriter, r *http.Request) {
	var req registerAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	token, err := randomToken(32)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var id int64
	err = s.db.QueryRowContext(r.Context(), `
		INSERT INTO agents (name, token, last_heartbeat)
		VALUES ($1, $2, NOW())
		RETURNING id
	`, req.Name, token).Scan(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, registerAgentResponse{
		AgentID: id,
		Token:   token,
	})
}

func (s *Server) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	agentID, err := agentIDFromHeader(r)
	if err != nil {
		http.Error(w, "invalid x-agent-id", http.StatusBadRequest)
		return
	}

	_, err = s.db.ExecContext(r.Context(), `
		UPDATE agents SET last_heartbeat = NOW() WHERE id = $1
	`, agentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.AgentID == 0 || req.CheckType == "" {
		http.Error(w, "agent_id and check_type required", http.StatusBadRequest)
		return
	}

	if req.MaxRetries < 0 {
		req.MaxRetries = 0
	}

	if err := s.queue.EnqueueTask(r.Context(), req.AgentID, req.CheckType, req.Payload, req.MaxRetries); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"status": "queued"})
}

func (s *Server) handleNextTask(w http.ResponseWriter, r *http.Request) {
	agentID, err := agentIDFromHeader(r)
	if err != nil {
		http.Error(w, "invalid x-agent-id", http.StatusBadRequest)
		return
	}

	task, err := s.queue.GetNextTask(r.Context(), agentID)
	if err != nil {
		if errors.Is(err, queue.ErrNoTask) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func (s *Server) handleSubmitResult(w http.ResponseWriter, r *http.Request) {
	agentID, err := agentIDFromHeader(r)
	if err != nil {
		http.Error(w, "invalid x-agent-id", http.StatusBadRequest)
		return
	}

	var req resultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := s.queue.CompleteTask(
		r.Context(),
		agentID,
		req.TaskID,
		req.ExitCode,
		req.ResultJSON,
		req.Stdout,
		req.Stderr,
		req.Logs,
	); err != nil {
		if errors.Is(err, queue.ErrTaskNotFound) {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "accepted"})
}

func (s *Server) handleListAgents(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.QueryContext(r.Context(), `
		SELECT id, name, last_heartbeat, created_at
		FROM agents
		ORDER BY id ASC
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []storage.Agent
	for rows.Next() {
		var a storage.Agent
		if err := rows.Scan(&a.ID, &a.Name, &a.LastHeartbeat, &a.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, a)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []storage.Agent{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.QueryContext(r.Context(), `
		SELECT id, agent_id, check_type, payload, status, retry_count, max_retries, created_at, started_at, finished_at
		FROM tasks
		ORDER BY created_at DESC
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []storage.Task
	for rows.Next() {
		var t storage.Task
		if err := rows.Scan(
			&t.ID,
			&t.AgentID,
			&t.CheckType,
			&t.Payload,
			&t.Status,
			&t.RetryCount,
			&t.MaxRetries,
			&t.CreatedAt,
			&t.StartedAt,
			&t.FinishedAt,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, t)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []storage.Task{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleListResults(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	var limit int
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		if n > 10000 {
			n = 10000
		}
		limit = n
	}

	var resultID int64
	if v := q.Get("result_id"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n < 1 {
			http.Error(w, "invalid result_id", http.StatusBadRequest)
			return
		}
		resultID = n
	}

	var taskID int64
	if v := q.Get("task_id"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n < 1 {
			http.Error(w, "invalid task_id", http.StatusBadRequest)
			return
		}
		taskID = n
	}

	var exitCode *int
	if v := q.Get("exit_code"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			http.Error(w, "invalid exit_code", http.StatusBadRequest)
			return
		}
		exitCode = &n
	}

	query := `SELECT id, task_id, exit_code, result_json::text, stdout, stderr, logs, created_at FROM task_results WHERE 1=1`
	var args []interface{}
	arg := 1
	if resultID > 0 {
		query += fmt.Sprintf(" AND id = $%d", arg)
		args = append(args, resultID)
		arg++
	}
	if taskID > 0 {
		query += fmt.Sprintf(" AND task_id = $%d", arg)
		args = append(args, taskID)
		arg++
	}
	if exitCode != nil {
		query += fmt.Sprintf(" AND exit_code = $%d", arg)
		args = append(args, *exitCode)
		arg++
	}
	query += ` ORDER BY created_at DESC`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", arg)
		args = append(args, limit)
	}

	rows, err := s.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []storage.TaskResult
	for rows.Next() {
		var tr storage.TaskResult
		if err := rows.Scan(
			&tr.ID,
			&tr.TaskID,
			&tr.ExitCode,
			&tr.ResultJSON,
			&tr.Stdout,
			&tr.Stderr,
			&tr.Logs,
			&tr.CreatedAt,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, tr)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []storage.TaskResult{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) withAgentAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		agentID, err := agentIDFromHeader(r)
		if err != nil {
			http.Error(w, "missing x-agent-id", http.StatusUnauthorized)
			return
		}

		token := r.Header.Get("X-Agent-Token")
		if token == "" {
			http.Error(w, "missing x-agent-token", http.StatusUnauthorized)
			return
		}

		ok, err := s.validateAgent(r.Context(), agentID, token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

func (s *Server) validateAgent(ctx context.Context, agentID int64, token string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM agents WHERE id = $1 AND token = $2
		)
	`, agentID, token).Scan(&exists)
	return exists, err
}

func agentIDFromHeader(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.Header.Get("X-Agent-ID"), 10, 64)
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *Server) handleRetryTask(w http.ResponseWriter, r *http.Request) {
	// Извлекаем ID задачи из пути
	idStr := r.PathValue("id")
	taskID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid task ID", http.StatusBadRequest)
		return
	}

	// Проверяем существование задачи и её текущий статус
	var status string
	var agentID int64
	err = s.db.QueryRowContext(r.Context(), `
		SELECT status, agent_id FROM tasks WHERE id = $1
	`, taskID).Scan(&status, &agentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "task not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Разрешаем перезапуск только для завершённых (failed или done) задач
	if status != "failed" && status != "done" {
		http.Error(w, "task cannot be retried (must be failed or done)", http.StatusConflict)
		return
	}

	// Обновляем задачу: статус pending, сбрасываем started_at и finished_at
	_, err = s.db.ExecContext(r.Context(), `
		UPDATE tasks
		SET status = 'pending', started_at = NULL, finished_at = NULL
		WHERE id = $1
	`, taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "task scheduled for retry",
		"task_id": idStr,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

var _ = sql.ErrNoRows