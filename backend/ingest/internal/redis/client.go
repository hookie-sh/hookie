package redis

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	client *redis.Client
}

func NewClient(addr string) (*Client, error) {
	opts := &redis.Options{
		Addr: addr,
	}

	// Read password from environment
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		opts.Password = password
	}

	// Read database number from environment
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if db, err := strconv.Atoi(dbStr); err == nil {
			opts.DB = db
		}
	}

	// Read username from environment
	if username := os.Getenv("REDIS_USERNAME"); username != "" {
		opts.Username = username
	}

	// Configure connection pool for high concurrency
	opts.PoolSize = 100 // Default is 10 * numCPU, but we want more for high load
	opts.MinIdleConns = 10
	opts.MaxRetries = 3
	opts.PoolTimeout = 4 * time.Second

	rdb := redis.NewClient(opts)

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	log.Printf("Successfully connected to Redis at %s", addr)
	if opts.DB > 0 {
		log.Printf("Using Redis database %d", opts.DB)
	}
	log.Printf("Redis pool config: PoolSize=%d, MinIdleConns=%d, PoolTimeout=%v", opts.PoolSize, opts.MinIdleConns, opts.PoolTimeout)

	return &Client{client: rdb}, nil
}

func (c *Client) PublishWebhook(ctx context.Context, streamKey string, fields map[string]interface{}) error {
	// MaxLen configurable via environment variable
	// Increased default from 1000 to 10000 to prevent eviction under load
	// With consumer groups, messages are tracked separately, but higher MaxLen provides safety margin
	maxLen := 10000
	if maxLenStr := os.Getenv("REDIS_STREAM_MAX_LEN"); maxLenStr != "" {
		if ml, err := strconv.Atoi(maxLenStr); err == nil && ml > 0 {
			maxLen = ml
		}
	}

	err := c.client.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		MaxLen: int64(maxLen),
		Approx: true,
		Values: fields,
	}).Err()

	if err != nil {
		return fmt.Errorf("failed to publish webhook to redis stream: %w", err)
	}

	return nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

// GetRedisClient returns the underlying Redis client for advanced operations
// like sorted sets used in rate limiting.
func (c *Client) GetRedisClient() *redis.Client {
	return c.client
}

// HasConnectedClients checks if any relay instances have clients connected to a topic
func (c *Client) HasConnectedClients(ctx context.Context, topicID string) (bool, error) {
	topicKey := fmt.Sprintf("relay:topics:%s", topicID)
	count, err := c.client.SCard(ctx, topicKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check connected clients: %w", err)
	}
	return count > 0, nil
}
