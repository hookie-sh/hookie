package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// Pruner handles background cleanup of Redis streams
type Pruner struct {
	client *redis.Client
}

// NewPruner creates a new Pruner instance
func NewPruner(client *redis.Client) *Pruner {
	return &Pruner{
		client: client,
	}
}

// PruneStaleStreams scans all streams matching topics:* pattern and deletes
// streams that haven't received any events in the last staleDuration period.
func (p *Pruner) PruneStaleStreams(ctx context.Context, staleDuration time.Duration) (int, error) {
	pattern := "topics:*"
	keys, err := p.client.Keys(ctx, pattern).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to scan streams: %w", err)
	}

	deleted := 0
	cutoff := time.Now().Add(-staleDuration).UnixMilli()

	for _, key := range keys {
		// Get the last entry timestamp
		entries, err := p.client.XRevRangeN(ctx, key, "+", "-", 1).Result()
		if err != nil {
			log.Printf("Warning: failed to read stream %s: %v", key, err)
			continue
		}

		// If stream is empty or last entry is older than cutoff, delete it
		if len(entries) == 0 {
			if err := p.client.Del(ctx, key).Err(); err != nil {
				log.Printf("Warning: failed to delete empty stream %s: %v", key, err)
			} else {
				deleted++
				log.Printf("Deleted empty stream: %s", key)
			}
			continue
		}

		// Parse timestamp from entry ID (Redis stream IDs are timestamp-based)
		// Format: milliseconds-timestamp-sequenceNumber
		entryID := entries[0].ID
		timestamp, err := parseStreamIDTimestamp(entryID)
		if err != nil {
			log.Printf("Warning: failed to parse timestamp from entry ID %s: %v", entryID, err)
			continue
		}

		if timestamp < cutoff {
			if err := p.client.Del(ctx, key).Err(); err != nil {
				log.Printf("Warning: failed to delete stale stream %s: %v", key, err)
			} else {
				deleted++
				log.Printf("Deleted stale stream: %s (last entry: %d)", key, timestamp)
			}
		}
	}

	return deleted, nil
}

// CleanExpiredAnonChannels removes expired anonymous channels and their associated keys.
// This queries the anon:channels sorted set for entries with scores below now,
// then pipeline-deletes all related keys.
func (p *Pruner) CleanExpiredAnonChannels(ctx context.Context) (int, error) {
	now := time.Now().UnixMilli()

	// Get all expired channel IDs from sorted set
	expired, err := p.client.ZRangeByScore(ctx, "anon:channels", &redis.ZRangeBy{
		Min:   "0",
		Max:   fmt.Sprintf("%d", now),
		Count: 100, // Process in batches
	}).Result()

	if err != nil {
		return 0, fmt.Errorf("failed to query expired channels: %w", err)
	}

	if len(expired) == 0 {
		return 0, nil
	}

	deleted := 0
	for _, topicID := range expired {
		// Get IP from metadata for cleanup
		ip, _ := p.client.HGet(ctx, "anon:meta:"+topicID, "ip").Result()

		pipe := p.client.Pipeline()
		pipe.Del(ctx, "anon:topics:"+topicID)        // the event stream
		pipe.ZRem(ctx, "anon:channels", topicID)      // sorted set entry
		pipe.Del(ctx, "anon:meta:"+topicID)          // metadata hash
		pipe.Del(ctx, "anon:connected:"+topicID)     // connection tracking
		pipe.Del(ctx, "rl:anon:"+topicID+":min")     // rate limit burst
		pipe.Del(ctx, "rl:anon:"+topicID+":day")     // rate limit daily
		if ip != "" {
			pipe.SRem(ctx, "anon:ip:"+ip, topicID)   // IP tracking set
		}

		if _, err := pipe.Exec(ctx); err != nil {
			log.Printf("Warning: failed to clean up expired channel %s: %v", topicID, err)
		} else {
			deleted++
			log.Printf("Cleaned up expired anonymous channel: %s", topicID)
		}
	}

	return deleted, nil
}

// StartBackground starts background goroutines for periodic cleanup.
// It runs PruneStaleStreams every hour and CleanExpiredAnonChannels every 5 minutes.
func (p *Pruner) StartBackground(ctx context.Context) {
	// Stale stream cleanup: every hour
	staleTicker := time.NewTicker(1 * time.Hour)
	defer staleTicker.Stop()

	// Anonymous channel cleanup: every 5 minutes
	anonTicker := time.NewTicker(5 * time.Minute)
	defer anonTicker.Stop()

	// Run initial cleanup
	go func() {
		if count, err := p.PruneStaleStreams(ctx, 48*time.Hour); err != nil {
			log.Printf("Error pruning stale streams: %v", err)
		} else if count > 0 {
			log.Printf("Pruned %d stale streams on startup", count)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Pruner background goroutine stopped")
			return
		case <-staleTicker.C:
			if count, err := p.PruneStaleStreams(ctx, 48*time.Hour); err != nil {
				log.Printf("Error pruning stale streams: %v", err)
			} else if count > 0 {
				log.Printf("Pruned %d stale streams", count)
			}
		case <-anonTicker.C:
			if count, err := p.CleanExpiredAnonChannels(ctx); err != nil {
				log.Printf("Error cleaning expired anonymous channels: %v", err)
			} else if count > 0 {
				log.Printf("Cleaned up %d expired anonymous channels", count)
			}
		}
	}
}

// parseStreamIDTimestamp extracts the timestamp from a Redis stream ID.
// Stream IDs are in the format: milliseconds-timestamp-sequenceNumber
func parseStreamIDTimestamp(id string) (int64, error) {
	// Stream IDs are in format: "1234567890123-0" or "1234567890123-456"
	// We need to extract the milliseconds timestamp part
	var timestamp int64
	_, err := fmt.Sscanf(id, "%d-", &timestamp)
	if err != nil {
		return 0, fmt.Errorf("invalid stream ID format: %w", err)
	}
	return timestamp, nil
}
