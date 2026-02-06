package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hookie/ingest/internal/config"
	"github.com/hookie/ingest/internal/ratelimit"
	internalredis "github.com/hookie/ingest/internal/redis"
	"github.com/hookie/ingest/internal/supabase"
	"github.com/redis/go-redis/v9"
)

type contextKey string

const (
	tierKey   contextKey = "tier"
	resultKey contextKey = "rateLimitResult"
)

// ForTopics middleware resolves tier from Supabase and enforces rate limits.
func ForTopics(resolver *ratelimit.Resolver, limiter *ratelimit.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			topicID := r.PathValue("topicID")
			if topicID == "" {
				respondError(w, http.StatusBadRequest, config.ErrInvalidPathParams)
				return
			}

			ctx := r.Context()

			// Resolve tier
			tier, err := resolver.ResolveAuthTier(ctx, topicID)
			if err != nil {
				respondError(w, http.StatusNotFound, config.ErrTopicNotFound)
				return
			}

			// Apply payload size limit
			r.Body = http.MaxBytesReader(w, r.Body, tier.MaxPayloadSize)

			// Enforce rate limits
			key := ratelimit.RateLimitKey(config.RedisPrefixTopics, topicID)
			result, err := limiter.EnforceRateLimits(ctx, key, tier)
			if err != nil {
				log.Printf("Rate limit check failed: %v", err)
				respondError(w, http.StatusInternalServerError, config.ErrInternalServer)
				return
			}

			if !result.Allowed {
				respond429(w, result, tier)
				return
			}

			// Set rate limit headers
			setRateLimitHeaders(w, result)

			// Store in context
			ctx = context.WithValue(ctx, tierKey, tier)
			ctx = context.WithValue(ctx, resultKey, result)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ForAnon middleware enforces Anon tier limits for anonymous topics.
func ForAnon(redisClient *internalredis.Client, limiter *ratelimit.Limiter, supabaseClient *supabase.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			anonTopicID := r.PathValue("anonTopicID")
			if anonTopicID == "" {
				respondError(w, http.StatusBadRequest, config.ErrInvalidPathParams)
				return
			}

			ctx := r.Context()
			rdb := redisClient.GetRedisClient()

			// 1. Check anon:channels exists
			_, err := rdb.ZScore(ctx, config.RedisKeyAnonChannels, anonTopicID).Result()
			if err == redis.Nil {
				respondError(w, http.StatusNotFound, config.ErrTopicNotFound)
				return
			}
			if err != nil {
				log.Printf("Failed to check anon channel: %v", err)
				respondError(w, http.StatusInternalServerError, config.ErrInternalServer)
				return
			}

			// 2. Check disabled
			disabled, err := rdb.HGet(ctx, config.BuildAnonMetaKey(anonTopicID), "disabled").Result()
			if err != nil && err != redis.Nil {
				log.Printf("Failed to check disabled status: %v", err)
				respondError(w, http.StatusInternalServerError, config.ErrInternalServer)
				return
			}
			if disabled == "true" {
				respondError(w, http.StatusForbidden, config.ErrTopicDisabled)
				return
			}

			// 3. Apply payload size limit (Anon tier)
			r.Body = http.MaxBytesReader(w, r.Body, ratelimit.Anon.MaxPayloadSize)

			// 4. Rate limit by topic (Anon tier hardcoded)
			key := ratelimit.RateLimitKey(config.RedisPrefixAnon, anonTopicID)
			result, err := limiter.EnforceRateLimits(ctx, key, ratelimit.Anon)
			if err != nil {
				log.Printf("Rate limit check failed: %v", err)
				respondError(w, http.StatusInternalServerError, config.ErrInternalServer)
				return
			}

			if !result.Allowed {
				respond429(w, result, ratelimit.Anon)
				return
			}

			// Set rate limit headers
			setRateLimitHeaders(w, result)

			// Store in context
			ctx = context.WithValue(ctx, tierKey, ratelimit.Anon)
			ctx = context.WithValue(ctx, resultKey, result)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// TierFromContext extracts the tier from request context.
func TierFromContext(ctx context.Context) ratelimit.Tier {
	tier, ok := ctx.Value(tierKey).(ratelimit.Tier)
	if !ok {
		return ratelimit.Starter // fallback
	}
	return tier
}

// ResultFromContext extracts the rate limit result from request context.
func ResultFromContext(ctx context.Context) *ratelimit.RateLimitResult {
	result, ok := ctx.Value(resultKey).(*ratelimit.RateLimitResult)
	if !ok {
		return nil
	}
	return result
}

func respondError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func respond429(w http.ResponseWriter, result *ratelimit.RateLimitResult, tier ratelimit.Tier) {
	w.Header().Set("Retry-After", fmt.Sprintf("%d", result.RetryAfterSecs))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)

	message := config.ErrRateLimitExceededQuota
	if result.Window == config.WindowMinute {
		message = config.ErrRateLimitExceededBurst
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":       "rate_limit_exceeded",
		"message":     message,
		"limit":       result.Limit,
		"window":      result.Window,
		"retry_after": result.RetryAfterSecs,
	})
}

func setRateLimitHeaders(w http.ResponseWriter, result *ratelimit.RateLimitResult) {
	w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", result.Limit))
	w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))

	// Calculate reset time
	var resetTime int64
	if result.Window == config.WindowMinute {
		resetTime = (time.Now().Unix() / 60) * 60 + 60 // next minute
	} else {
		// Next midnight UTC
		nextDay := time.Now().UTC().Truncate(24 * time.Hour).Add(24 * time.Hour)
		resetTime = nextDay.Unix()
	}
	w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime))
}
