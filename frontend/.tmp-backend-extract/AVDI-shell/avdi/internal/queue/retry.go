package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"diag-system/internal/redis"
	"diag-system/internal/storage"
)

// RetryManager управляет повторными попытками выполнения задач с экспоненциальной задержкой.
type RetryManager struct {
	redisClient *redis.Client
	queue       *PGQueue
	// Базовая задержка перед повторной попыткой (в секундах)
	baseDelay int64
	// Максимальная задержка (в секундах)
	maxDelay int64
	// Ключ Redis для отсортированного множества отложенных задач
	delayedSetKey string
	// Канал для остановки фонового воркера
	stopChan chan struct{}
}

// DelayedTask представляет задачу, отложенную для повторного выполнения.
type DelayedTask struct {
	TaskID  int64 `json:"task_id"`
	AgentID int64 `json:"agent_id"`
	// Номер текущей попытки (уже применённой)
	RetryCount int `json:"retry_count"`
	// Максимальное количество попыток
	MaxRetries int `json:"max_retries"`
	// Время, когда задача должна быть возвращена в очередь (Unix timestamp)
	ExecuteAt int64 `json:"execute_at"`
}

// NewRetryManager создаёт новый менеджер повторных попыток.
func NewRetryManager(redisClient *redis.Client, queue *PGQueue) *RetryManager {
	baseDelay := getBaseDelayFromEnv()
	maxDelay := getMaxDelayFromEnv()
	return &RetryManager{
		redisClient:   redisClient,
		queue:         queue,
		baseDelay:     baseDelay,
		maxDelay:      maxDelay,
		delayedSetKey: "delayed_tasks",
		stopChan:      make(chan struct{}),
	}
}

// ScheduleRetry планирует повторное выполнение задачи через определённую задержку.
// Возвращает true, если задача была запланирована, false если превышен лимит попыток.
func (rm *RetryManager) ScheduleRetry(ctx context.Context, task *storage.Task) (bool, error) {
	if task.RetryCount >= task.MaxRetries {
		return false, nil
	}

	// Вычисляем задержку с экспоненциальным откатом
	delay := rm.baseDelay * (1 << uint(task.RetryCount)) // baseDelay * 2^retryCount
	if delay > rm.maxDelay {
		delay = rm.maxDelay
	}

	executeAt := time.Now().Add(time.Duration(delay) * time.Second).Unix()

	delayedTask := DelayedTask{
		TaskID:     task.ID,
		AgentID:    task.AgentID,
		RetryCount: task.RetryCount,
		MaxRetries: task.MaxRetries,
		ExecuteAt:  executeAt,
	}

	// Сериализуем задачу в JSON
	member, err := json.Marshal(delayedTask)
	if err != nil {
		return false, fmt.Errorf("ошибка сериализации отложенной задачи: %w", err)
	}

	// Добавляем в отсортированное множество с score = executeAt
	err = rm.redisClient.AddToSortedSet(ctx, rm.delayedSetKey, float64(executeAt), member)
	if err != nil {
		return false, fmt.Errorf("ошибка добавления задачи в отсортированное множество: %w", err)
	}

	return true, nil
}

// ProcessDueTasks извлекает задачи, время выполнения которых наступило, и возвращает их в очередь.
func (rm *RetryManager) ProcessDueTasks(ctx context.Context) (int, error) {
	now := time.Now().Unix()
	// Получаем задачи с score <= now
	members, err := rm.redisClient.GetFromSortedSetRange(ctx, rm.delayedSetKey, "-inf", strconv.FormatInt(now, 10), 0, 100)
	if err != nil {
		return 0, fmt.Errorf("ошибка получения задач из отсортированного множества: %w", err)
	}

	processed := 0
	for _, member := range members {
		var delayedTask DelayedTask
		if err := json.Unmarshal([]byte(member), &delayedTask); err != nil {
			// Если не удалось распарсить, удаляем из множества
			rm.redisClient.RemoveFromSortedSet(ctx, rm.delayedSetKey, member)
			continue
		}

		// Обновляем задачу в PostgreSQL: устанавливаем статус pending
		err := rm.queue.RetryTask(ctx, delayedTask.TaskID, delayedTask.AgentID)
		if err != nil {
			// Логируем ошибку, но продолжаем обработку остальных задач
			fmt.Printf("Ошибка при повторной постановке задачи %d: %v\n", delayedTask.TaskID, err)
			continue
		}

		// Удаляем задачу из отсортированного множества
		err = rm.redisClient.RemoveFromSortedSet(ctx, rm.delayedSetKey, member)
		if err != nil {
			fmt.Printf("Ошибка удаления задачи %d из отсортированного множества: %v\n", delayedTask.TaskID, err)
		}

		processed++
	}

	return processed, nil
}

// StartWorker запускает фоновый воркер, который периодически проверяет и обрабатывает отложенные задачи.
func (rm *RetryManager) StartWorker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			processed, err := rm.ProcessDueTasks(ctx)
			if err != nil {
				fmt.Printf("Ошибка обработки отложенных задач: %v\n", err)
			} else if processed > 0 {
				fmt.Printf("Обработано отложенных задач: %d\n", processed)
			}
		case <-rm.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

// StopWorker останавливает фоновый воркер.
func (rm *RetryManager) StopWorker() {
	close(rm.stopChan)
}


// Вспомогательные функции для чтения конфигурации из окружения

func getBaseDelayFromEnv() int64 {
	val := os.Getenv("BASE_RETRY_DELAY_SECONDS")
	if val == "" {
		return 30 // значение по умолчанию
	}
	delay, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 30
	}
	return delay
}

func getMaxDelayFromEnv() int64 {
	val := os.Getenv("MAX_RETRY_DELAY_SECONDS")
	if val == "" {
		return 3600 // значение по умолчанию
	}
	delay, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 3600
	}
	return delay
}