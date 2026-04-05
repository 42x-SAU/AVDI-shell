package queue

import (
	"context"
	"os"
	"strconv"
	"time"

	"diag-system/internal/redis"
	rd "github.com/redis/go-redis/v9"
)

// Limiter реализует ограничение одновременных задач на агента с использованием Redis.
type Limiter struct {
	redisClient *redis.Client
	// Максимальное количество одновременных задач на агента
	maxConcurrent int64
	// TTL для ключей Redis (по умолчанию 1 час)
	keyTTL time.Duration
}

// NewLimiter создаёт новый лимитер.
// Если maxConcurrent <= 0, используется значение из переменной окружения
// MAX_CONCURRENT_TASKS_PER_AGENT (по умолчанию 3).
func NewLimiter(redisClient *redis.Client, maxConcurrent int64) *Limiter {
	if maxConcurrent <= 0 {
		maxConcurrent = getMaxConcurrentFromEnv()
	}
	return &Limiter{
		redisClient:   redisClient,
		maxConcurrent: maxConcurrent,
		keyTTL:        time.Hour,
	}
}

// CanAcquire проверяет, может ли агент взять ещё одну задачу.
// Возвращает true, если количество активных задач меньше лимита.
// При успешной проверке увеличивает счётчик активных задач.
func (l *Limiter) CanAcquire(ctx context.Context, agentID int64) (bool, error) {
	key := l.agentKey(agentID)
	_, ok, err := l.redisClient.IncrementWithLimit(ctx, key, l.maxConcurrent, l.keyTTL)
	if err != nil {
		return false, err
	}
	return ok, nil
}

// Acquire увеличивает счётчик активных задач без проверки.
// Используется, когда проверка уже выполнена (например, через CanAcquire).
func (l *Limiter) Acquire(ctx context.Context, agentID int64) (int64, error) {
	key := l.agentKey(agentID)
	val, err := l.redisClient.Client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	// Устанавливаем TTL, если ключ только что создан
	if val == 1 {
		l.redisClient.Client.Expire(ctx, key, l.keyTTL)
	}
	return val, nil
}

// Release уменьшает счётчик активных задач.
// Если счётчик становится нулевым, ключ удаляется автоматически по истечении TTL.
func (l *Limiter) Release(ctx context.Context, agentID int64) (int64, error) {
	key := l.agentKey(agentID)
	val, err := l.redisClient.Client.Decr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	// Если значение стало отрицательным (не должно происходить), сбрасываем в 0
	if val < 0 {
		l.redisClient.Client.Del(ctx, key)
		return 0, nil
	}
	return val, nil
}

// GetCurrent возвращает текущее количество активных задач агента.
func (l *Limiter) GetCurrent(ctx context.Context, agentID int64) (int64, error) {
	key := l.agentKey(agentID)
	val, err := l.redisClient.Client.Get(ctx, key).Int64()
	if err != nil {
		// Если ключ не существует, значит активных задач нет
		if err == rd.Nil {
			return 0, nil
		}
		return 0, err
	}
	return val, nil
}

// agentKey формирует ключ Redis для хранения счётчика активных задач агента.
func (l *Limiter) agentKey(agentID int64) string {
	return "agent:" + strconv.FormatInt(agentID, 10) + ":active_tasks"
}

// getMaxConcurrentFromEnv читает значение MAX_CONCURRENT_TASKS_PER_AGENT из окружения.
func getMaxConcurrentFromEnv() int64 {
	val := os.Getenv("MAX_CONCURRENT_TASKS_PER_AGENT")
	if val == "" {
		return 3 // значение по умолчанию
	}
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil || n <= 0 {
		return 3
	}
	return n
}