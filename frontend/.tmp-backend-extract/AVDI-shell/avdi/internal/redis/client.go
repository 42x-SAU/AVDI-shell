package redis

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config представляет конфигурацию Redis.
type Config struct {
	URL      string
	Password string
	DB       int
}

// Client обёртка над redis.Client с дополнительными методами.
type Client struct {
	*redis.Client
	config Config
}

// NewClient создаёт новый клиент Redis на основе конфигурации.
func NewClient(config Config) (*Client, error) {
	opts, err := redis.ParseURL(config.URL)
	if err != nil {
		// Если не удалось распарсить как URL, используем обычные параметры
		opts = &redis.Options{
			Addr:     config.URL,
			Password: config.Password,
			DB:       config.DB,
		}
	}

	rdb := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Проверяем подключение
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("не удалось подключиться к Redis: %w", err)
	}

	return &Client{
		Client: rdb,
		config: config,
	}, nil
}

// NewClientFromEnv создаёт клиент Redis из переменных окружения.
func NewClientFromEnv() (*Client, error) {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		url = "redis://localhost:6379"
	}

	dbStr := os.Getenv("REDIS_DB")
	if dbStr == "" {
		dbStr = "0"
	}
	db, err := strconv.Atoi(dbStr)
	if err != nil {
		db = 0
	}

	config := Config{
		URL:      url,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	}
	return NewClient(config)
}

// Close закрывает соединение с Redis.
func (c *Client) Close() error {
	return c.Client.Close()
}

// IncrementWithLimit увеличивает значение ключа, если не превышен лимит.
// Возвращает новое значение и флаг, превышен ли лимит.
func (c *Client) IncrementWithLimit(ctx context.Context, key string, limit int64, ttl time.Duration) (int64, bool, error) {
	// Используем Lua-скрипт для атомарной проверки лимита
	script := redis.NewScript(`
		local current = redis.call('GET', KEYS[1])
		if current and tonumber(current) >= tonumber(ARGV[1]) then
			return {-1, 0}
		end
		local val = redis.call('INCR', KEYS[1])
		if val == 1 then
			redis.call('EXPIRE', KEYS[1], ARGV[2])
		end
		return {val, 1}
	`)

	keys := []string{key}
	args := []interface{}{limit, int(ttl.Seconds())}
	result, err := script.Run(ctx, c.Client, keys, args...).Result()
	if err != nil {
		return 0, false, err
	}

	arr := result.([]interface{})
	val := arr[0].(int64)
	success := arr[1].(int64)

	if success == 0 {
		// Лимит превышен
		return val, false, nil
	}
	return val, true, nil
}

// AddToSortedSet добавляет элемент в отсортированное множество с указанным score.
func (c *Client) AddToSortedSet(ctx context.Context, key string, score float64, member interface{}) error {
	return c.Client.ZAdd(ctx, key, redis.Z{Score: score, Member: member}).Err()
}

// GetFromSortedSetRange возвращает элементы из отсортированного множества в диапазоне score.
func (c *Client) GetFromSortedSetRange(ctx context.Context, key string, min, max string, offset, count int64) ([]string, error) {
	return c.Client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min:    min,
		Max:    max,
		Offset: offset,
		Count:  count,
	}).Result()
}

// RemoveFromSortedSet удаляет элементы из отсортированного множества.
func (c *Client) RemoveFromSortedSet(ctx context.Context, key string, members ...interface{}) error {
	return c.Client.ZRem(ctx, key, members...).Err()
}

// Publish публикует сообщение в канал.
func (c *Client) Publish(ctx context.Context, channel string, message interface{}) error {
	return c.Client.Publish(ctx, channel, message).Err()
}

// Subscribe подписывается на канал и возвращает канал для получения сообщений.
func (c *Client) Subscribe(ctx context.Context, channel string) <-chan *redis.Message {
	pubsub := c.Client.Subscribe(ctx, channel)
	return pubsub.Channel()
}