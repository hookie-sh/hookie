package ratelimit

import (
	"context"
	"fmt"

	"github.com/hookie/ingest/internal/config"
	"github.com/hookie/ingest/internal/supabase"
	"github.com/redis/go-redis/v9"
)

type Resolver struct {
	supabaseClient *supabase.Client
	redisClient    *redis.Client
}

func NewResolver(supabaseClient *supabase.Client, redisClient *redis.Client) *Resolver {
	return &Resolver{
		supabaseClient: supabaseClient,
		redisClient:    redisClient,
	}
}

// ResolveAuthTier resolves the tier for an authenticated topic.
// Uses Redis cache with 5-minute TTL to avoid hitting Supabase on every request.
func (r *Resolver) ResolveAuthTier(ctx context.Context, topicID string) (Tier, error) {
	// Check cache first
	cacheKey := config.BuildTierCacheKey(topicID)
	cachedTier, err := r.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		// Cache hit
		tier := TierByName(cachedTier)
		return tier, nil
	}

	// Cache miss — look up in Supabase
	result, err := r.supabaseClient.LookupTopicTier(ctx, topicID)
	if err != nil {
		return Tier{}, fmt.Errorf("failed to resolve tier: %w", err)
	}

	// Convert result to Tier
	var tier Tier
	if result.IsEnterprise && result.CustomOverrides != nil {
		// Convert EnterpriseOverride to Tier
		override := result.CustomOverrides
		tier = Tier{Name: config.TierEnterprise}
		tier.BurstPerMinute = fallback(override.BurstPerMinute, Scale.BurstPerMinute)
		tier.DailyQuota = fallback(override.DailyQuota, Scale.DailyQuota)
		tier.MaxPayloadSize = fallback(override.MaxPayloadSize, Scale.MaxPayloadSize)
	} else {
		tier = TierByName(result.TierName)
	}

	// Cache the result
	r.redisClient.Set(ctx, cacheKey, tier.Name, config.TierCacheTTL)

	return tier, nil
}
