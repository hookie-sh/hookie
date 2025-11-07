package redis

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

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

	rdb := redis.NewClient(opts)

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	log.Printf("Successfully connected to Redis at %s", addr)
	if opts.DB > 0 {
		log.Printf("Using Redis database %d", opts.DB)
	}

	return &Client{client: rdb}, nil
}

func (c *Client) PublishWebhook(ctx context.Context, streamKey string, fields map[string]interface{}) error {
	err := c.client.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
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

