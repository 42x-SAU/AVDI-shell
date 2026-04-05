package server

import (
	"net/http"
	"strings"
)

type statsResponse struct {
	AgentsTotal     int64 `json:"agents_total"`
	TasksPending    int64 `json:"tasks_pending"`
	TasksRunning    int64 `json:"tasks_running"`
	TasksDone       int64 `json:"tasks_done"`
	TasksFailed     int64 `json:"tasks_failed"`
	TasksTotal      int64 `json:"tasks_total"`
	ResultsTotal    int64 `json:"results_total"`
	RecurringActive int64 `json:"recurring_enabled"`
	RecurringTotal  int64 `json:"recurring_total"`
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var out statsResponse
	err := s.db.QueryRowContext(r.Context(), `
		SELECT
			(SELECT COUNT(*) FROM agents),
			(SELECT COUNT(*) FROM tasks WHERE status = 'pending'),
			(SELECT COUNT(*) FROM tasks WHERE status = 'running'),
			(SELECT COUNT(*) FROM tasks WHERE status = 'done'),
			(SELECT COUNT(*) FROM tasks WHERE status = 'failed'),
			(SELECT COUNT(*) FROM tasks),
			(SELECT COUNT(*) FROM task_results)
	`).Scan(
		&out.AgentsTotal,
		&out.TasksPending,
		&out.TasksRunning,
		&out.TasksDone,
		&out.TasksFailed,
		&out.TasksTotal,
		&out.ResultsTotal,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM recurring_jobs WHERE enabled`).Scan(&out.RecurringActive); err != nil {
		if strings.Contains(err.Error(), "recurring_jobs") && strings.Contains(err.Error(), "does not exist") {
			out.RecurringActive = 0
			out.RecurringTotal = 0
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if err := s.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM recurring_jobs`).Scan(&out.RecurringTotal); err != nil {
			out.RecurringTotal = 0
		}
	}

	writeJSON(w, http.StatusOK, out)
}
