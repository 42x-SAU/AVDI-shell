package queue

import (
	"context"
	"database/sql"
	"errors"
 go build ./cmd/serve
	"diag-system/internal/storage"
)

var (
	ErrNoTask       = errors.New("no task available")
	ErrTaskNotFound = errors.New("task not found or not owned by this agent")
	ErrRateLimited  = errors.New("agent has reached concurrent task limit")
)

type PGQueue struct {
	db      *sql.DB
	limiter *Limiter
}

func New(db *sql.DB) *PGQueue {
	return &PGQueue{db: db}
}

// NewWithLimiter создаёт очередь с лимитером для ограничения одновременных задач.
func NewWithLimiter(db *sql.DB, limiter *Limiter) *PGQueue {
	return &PGQueue{db: db, limiter: limiter}
}

func (q *PGQueue) EnqueueTask(ctx context.Context, agentID int64, checkType, payload string, maxRetries int) error {
	_, err := q.db.ExecContext(ctx, `
		INSERT INTO tasks (agent_id, check_type, payload, status, retry_count, max_retries)
		VALUES ($1, $2, $3, 'pending', 0, $4)
	`, agentID, checkType, payload, maxRetries)
	return err
}

func (q *PGQueue) GetNextTask(ctx context.Context, agentID int64) (*storage.Task, error) {
	// Проверяем лимит одновременных задач, если лимитер настроен
	if q.limiter != nil {
		canAcquire, err := q.limiter.CanAcquire(ctx, agentID)
		if err != nil {
			return nil, err
		}
		if !canAcquire {
			return nil, ErrRateLimited
		}
		// CanAcquire уже увеличил счётчик, поэтому дополнительное увеличение не требуется
	}

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		// Если транзакция не началась, нужно откатить увеличение счётчика
		if q.limiter != nil {
			q.limiter.Release(ctx, agentID)
		}
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
			// Задач нет, нужно уменьшить счётчик, который мы увеличили ранее
			if q.limiter != nil {
				q.limiter.Release(ctx, agentID)
			}
			return nil, ErrNoTask
		}
		// Ошибка сканирования, также уменьшаем счётчик
		if q.limiter != nil {
			q.limiter.Release(ctx, agentID)
		}
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		// Ошибка коммита, уменьшаем счётчик
		if q.limiter != nil {
			q.limiter.Release(ctx, agentID)
		}
		return nil, err
	}

	// Успешно взяли задачу, счётчик уже увеличен
	return &t, nil
}

func (q *PGQueue) CompleteTask(ctx context.Context, agentID, taskID int64, exitCode int, resultJSON, stdout, stderr, logs string) error {
	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	finalStatus := "done"
	if exitCode != 0 {
		finalStatus = "failed"
	}

	res, err := tx.ExecContext(ctx, `
		UPDATE tasks
		SET status = $3, finished_at = NOW()
		WHERE id = $1 AND agent_id = $2
	`, taskID, agentID, finalStatus)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrTaskNotFound
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
			WHERE id = $1 AND agent_id = $2 AND retry_count < max_retries
		`, taskID, agentID)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	// После успешного коммита уменьшаем счётчик активных задач агента
	if q.limiter != nil {
		_, err = q.limiter.Release(ctx, agentID)
		if err != nil {
			// Логируем ошибку, но не прерываем выполнение, так как задача уже завершена
			// В будущем можно добавить повторные попытки или dead letter queue
		}
	}

	return nil
}