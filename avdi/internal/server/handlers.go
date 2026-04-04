package server

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"diag-system/internal/queue"
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
	var req heartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	_, err := s.db.ExecContext(r.Context(), `
		UPDATE agents SET last_heartbeat = NOW() WHERE id = $1
	`, req.AgentID)
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
	var req resultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := s.queue.CompleteTask(
		r.Context(),
		req.TaskID,
		req.ExitCode,
		req.ResultJSON,
		req.Stdout,
		req.Stderr,
		req.Logs,
	); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "accepted"})
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

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

var _ = sql.ErrNoRows