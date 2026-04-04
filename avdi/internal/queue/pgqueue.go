package queue

import (
	"context"
	"database/sql"
	"errors"

	"diag-system/internal/storage"
)

var ErrNoTask = errors.New("no task available")

type PGQueue struct {
	db *sql.DB
}

func New(db *sql.DB) *PGQueue {
	return &PGQueue{db: db}
}

func (q *PGQueue) EnqueueTask(ctx context.Context, agentID int64, checkType, payload string, maxRetries int) error {
	_, err := q.db.ExecContext(ctx, `
		INSERT INTO tasks (agent_id, check_type, payload, status, retry_count, max_retries)
		VALUES ($1, $2, $3, 'pending', 0, $4)
	`, agentID, checkType, payload, maxRetries)
	return err
}

func (q *PGQueue) GetNextTask(ctx context.Context, agentID int64) (*storage.Task, error) {
	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	row := tx.QueryRowContext(ctx, `
		WITH picked AS (
			SELECT id
			FROM tasks
			WHERE agent_id = $1 AND status = 'pending'
			ORDER BY created_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		UPDATE tasks t
		SET status = 'running', started_at = NOW()
		FROM picked
		WHERE t.id = picked.id
		RETURNING t.id, t.agent_id, t.check_type, t.payload, t.status, t.retry_count, t.max_retries, t.created_at, t.started_at, t.finished_at
	`, agentID)

	var t storage.Task
	if err := row.Scan(
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoTask
		}
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &t, nil
}

func (q *PGQueue) CompleteTask(ctx context.Context, taskID int64, exitCode int, resultJSON, stdout, stderr, logs string) error {
	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	finalStatus := "done"
	if exitCode != 0 {
		finalStatus = "failed"
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE tasks
		SET status = $2, finished_at = NOW()
		WHERE id = $1
	`, taskID, finalStatus)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO task_results (task_id, exit_code, result_json, stdout, stderr, logs)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, taskID, exitCode, resultJSON, stdout, stderr, logs)
	if err != nil {
		return err
	}

	if exitCode != 0 {
		_, err = tx.ExecContext(ctx, `
			UPDATE tasks
			SET
				status = CASE
					WHEN retry_count + 1 <= max_retries THEN 'pending'
					ELSE 'failed'
				END,
				retry_count = retry_count + 1,
				started_at = CASE
					WHEN retry_count + 1 <= max_retries THEN NULL
					ELSE started_at
				END,
				finished_at = CASE
					WHEN retry_count + 1 <= max_retries THEN NULL
					ELSE NOW()
				END
			WHERE id = $1 AND retry_count < max_retries
		`, taskID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}