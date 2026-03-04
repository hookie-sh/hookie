package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hookie-sh/hookie/backend/ingest/internal/config"
	"github.com/redis/go-redis/v9"
)

type RateLimitResult struct {
	Allowed        bool
	Remaining      int64
	Limit          int64
	RetryAfterSecs int
	Window         string // "minute" or "day"
}

type Limiter struct {
	client *redis.Client
}

func NewLimiter(client *redis.Client) *Limiter {
	return &Limiter{client: client}
}

// CheckRateLimit implements a sliding window rate limit using Redis sorted sets.
// Each request adds a unique member scored by timestamp. Expired entries are pruned
// in the same pipeline. On rejection, the added member is removed.
func (l *Limiter) CheckRateLimit(
	ctx context.Context,
	key string,
	limit int64,
	window time.Duration,
) (*RateLimitResult, error) {
	now := time.Now().UnixMilli()
	windowStart := now - window.Milliseconds()
	member := fmt.Sprintf("%d:%s", now, uuid.NewString()[:8]) // unique member

	pipe := l.client.Pipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: member})
	countCmd := pipe.ZCard(ctx, key)
	pipe.Expire(ctx, key, window) // auto-cleanup

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("rate limit check failed: %w", err)
	}

	count := countCmd.Val()
	allowed := count <= limit

	result := &RateLimitResult{
		Allowed:   allowed,
		Remaining: max(0, limit-count),
		Limit:     limit,
	}

	if !allowed {
		// Remove the entry we just added since it's denied
		l.client.ZRem(ctx, key, member)

		if window < time.Hour {
			result.RetryAfterSecs = config.RetryAfterMinute
			result.Window = config.WindowMinute
		} else {
			// Calculate seconds until midnight UTC
			nextDay := time.Now().UTC().Truncate(24 * time.Hour).Add(24 * time.Hour)
			result.RetryAfterSecs = int(time.Until(nextDay).Seconds())
			result.Window = config.WindowDay
		}
	}

	return result, nil
}

// EnforceRateLimits applies both per-minute burst and daily quota limits.
// Both must pass. Checks per-minute first (cheaper to reject early).
func (l *Limiter) EnforceRateLimits(ctx context.Context, key string, tier Tier) (*RateLimitResult, error) {
	// Check burst limit first
	minuteKey := key + ":min"
	minuteResult, err := l.CheckRateLimit(ctx, minuteKey, tier.BurstPerMinute, time.Minute)
	if err != nil {
		return nil, err
	}
	if !minuteResult.Allowed {
		return minuteResult, nil
	}

	// Check daily quota
	dayKey := key + ":day"
	dayResult, err := l.CheckRateLimit(ctx, dayKey, tier.DailyQuota, 24*time.Hour)
	if err != nil {
		return nil, err
	}
	return dayResult, nil
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
