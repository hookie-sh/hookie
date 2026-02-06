package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestCheckRateLimit(t *testing.T) {
	// Skip if no Redis connection available
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping test")
	}
	defer client.Close()

	limiter := NewLimiter(client)
	key := "test:ratelimit:" + time.Now().Format("20060102150405")

	// Test: first request should be allowed
	result, err := limiter.CheckRateLimit(ctx, key+":min", 5, time.Minute)
	if err != nil {
		t.Fatalf("CheckRateLimit failed: %v", err)
	}
	if !result.Allowed {
		t.Errorf("First request should be allowed, got Allowed=%v", result.Allowed)
	}
	if result.Remaining != 4 {
		t.Errorf("Expected Remaining=4, got %d", result.Remaining)
	}

	// Test: exhaust the limit
	for i := 0; i < 4; i++ {
		_, err := limiter.CheckRateLimit(ctx, key+":min", 5, time.Minute)
		if err != nil {
			t.Fatalf("CheckRateLimit failed: %v", err)
		}
	}

	// Test: 6th request should be rejected
	result, err = limiter.CheckRateLimit(ctx, key+":min", 5, time.Minute)
	if err != nil {
		t.Fatalf("CheckRateLimit failed: %v", err)
	}
	if result.Allowed {
		t.Errorf("6th request should be rejected, got Allowed=%v", result.Allowed)
	}
	if result.Window != "minute" {
		t.Errorf("Expected Window=minute, got %q", result.Window)
	}
}
