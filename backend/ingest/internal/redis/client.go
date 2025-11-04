package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	client *redis.Client
}

func NewClient(addr string) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
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

