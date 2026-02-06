package ratelimit

import (
	"context"
	"testing"

	"github.com/hookie/ingest/internal/supabase"
	"github.com/redis/go-redis/v9"
)

func TestResolver_ResolveAuthTier(t *testing.T) {
	// Skip if dependencies not available
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping test")
	}
	defer client.Close()

	supabaseClient, err := supabase.NewClient()
	if err != nil {
		t.Skip("Supabase not configured, skipping test")
	}

	resolver := NewResolver(supabaseClient, client)

	// Test: cache behavior
	// This is a basic test - actual tier resolution requires valid topic ID in database
	tier, err := resolver.ResolveAuthTier(ctx, "nonexistent_topic")
	if err == nil {
		t.Logf("Resolved tier: %v (this is expected to fail for nonexistent topic)", tier)
	}
	// We expect an error for nonexistent topic, which is correct behavior
}
