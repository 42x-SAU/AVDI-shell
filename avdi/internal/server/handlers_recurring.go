package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type recurringJobRow struct {
	ID               int64      `json:"id"`
	AgentID          int64      `json:"agent_id"`
	CheckType        string     `json:"check_type"`
	Payload          string     `json:"payload"`
	MaxRetries       int        `json:"max_retries"`
	IntervalSeconds  int        `json:"interval_seconds"`
	Enabled          bool       `json:"enabled"`
	NextRunAt        time.Time  `json:"next_run_at"`
	LastRunAt        *time.Time `json:"last_run_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

type recurringCreateRequest struct {
	AgentID           int64  `json:"agent_id"`
	CheckType         string `json:"check_type"`
	Payload           string `json:"payload"`
	MaxRetries        int    `json:"max_retries"`
	IntervalSeconds   int    `json:"interval_seconds"`
	Enabled           *bool  `json:"enabled,omitempty"`
	StartInSeconds    int    `json:"start_in_seconds"` // delay before first run (default 0)
}

func (s *Server) handleListRecurring(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rows, err := s.db.QueryContext(r.Context(), `
		SELECT id, agent_id, check_type, payload, max_retries, interval_seconds, enabled, next_run_at, last_run_at, created_at
		FROM recurring_jobs
		ORDER BY id ASC
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []recurringJobRow
	for rows.Next() {
		var j recurringJobRow
		if err := rows.Scan(&j.ID, &j.AgentID, &j.CheckType, &j.Payload, &j.MaxRetries, &j.IntervalSeconds, &j.Enabled, &j.NextRunAt, &j.LastRunAt, &j.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, j)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []recurringJobRow{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleCreateRecurring(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req recurringCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.AgentID < 1 || req.CheckType == "" {
		http.Error(w, "agent_id and check_type required", http.StatusBadRequest)
		return
	}
	if req.IntervalSeconds < 10 || req.IntervalSeconds > 86400 {
		http.Error(w, "interval_seconds must be between 10 and 86400", http.StatusBadRequest)
		return
	}
	if req.MaxRetries < 0 {
		req.MaxRetries = 0
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	first := time.Now().Add(time.Duration(req.StartInSeconds) * time.Second)
	if req.StartInSeconds < 0 {
		first = time.Now()
	}

	var id int64
	err := s.db.QueryRowContext(r.Context(), `
		INSERT INTO recurring_jobs (agent_id, check_type, payload, max_retries, interval_seconds, enabled, next_run_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, req.AgentID, req.CheckType, req.Payload, req.MaxRetries, req.IntervalSeconds, enabled, first).Scan(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func (s *Server) handleDeleteRecurring(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	res, err := s.db.ExecContext(r.Context(), `DELETE FROM recurring_jobs WHERE id = $1`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

type recurringPatchRequest struct {
	Enabled *bool `json:"enabled,omitempty"`
}

func (s *Server) handlePatchRecurring(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req recurringPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Enabled == nil {
		http.Error(w, "enabled field required", http.StatusBadRequest)
		return
	}
	res, err := s.db.ExecContext(r.Context(), `UPDATE recurring_jobs SET enabled = $2 WHERE id = $1`, id, *req.Enabled)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"enabled": *req.Enabled})
}

func (s *Server) runRecurringScheduler(ctx context.Context) {
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s.tickRecurring(ctx)
		}
	}
}

func (s *Server) tickRecurring(ctx context.Context) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, agent_id, check_type, payload, max_retries, interval_seconds
		FROM recurring_jobs
		WHERE enabled AND next_run_at <= NOW()
		ORDER BY next_run_at ASC
		LIMIT 100
	`)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id, agentID int64
		var checkType, payload string
		var maxRetries, intervalSec int
		if err := rows.Scan(&id, &agentID, &checkType, &payload, &maxRetries, &intervalSec); err != nil {
			continue
		}
		if err := s.queue.EnqueueTask(ctx, agentID, checkType, payload, maxRetries); err != nil {
			continue
		}
		_, _ = s.db.ExecContext(ctx, `
			UPDATE recurring_jobs
			SET last_run_at = NOW(),
			    next_run_at = NOW() + ($1 * interval '1 second')
			WHERE id = $2
		`, intervalSec, id)
	}
}
