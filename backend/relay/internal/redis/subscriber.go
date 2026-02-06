package redis

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Subscriber struct {
	client *redis.Client
	ctx    context.Context
}

func NewSubscriber(addr string) (*Subscriber, error) {
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

	return &Subscriber{
		client: rdb,
		ctx:    ctx,
	}, nil
}

// StreamEvent represents a webhook event from Redis stream
type StreamEvent struct {
	StreamKey string
	ID        string
	Fields    map[string]string
}

// SubscribeToApplication subscribes to all topics for an application
// topicIDs should be a list of topic IDs that belong to the application
func (s *Subscriber) SubscribeToApplication(topicIDs []string, eventsChan chan<- StreamEvent) error {
	if len(topicIDs) == 0 {
		return nil
	}

	streamKeys := make([]string, 0, len(topicIDs))
	for _, topicID := range topicIDs {
		streamKey := fmt.Sprintf("topics:%s", topicID)
		streamKeys = append(streamKeys, streamKey)
	}

	// Create consumer group for each stream
	consumerGroup := "relay-consumers"
	consumerName := fmt.Sprintf("consumer-%d", time.Now().UnixNano())

	for _, streamKey := range streamKeys {
		err := s.client.XGroupCreateMkStream(s.ctx, streamKey, consumerGroup, "0").Err()
		if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
			log.Printf("Warning: failed to create consumer group for %s: %v", streamKey, err)
		}
	}

	// Start reading from all streams
	go s.readFromStreams(streamKeys, consumerGroup, consumerName, eventsChan)

	return nil
}

// SubscribeToTopic subscribes to a specific topic
func (s *Subscriber) SubscribeToTopic(topicID string, eventsChan chan<- StreamEvent) error {
	streamKey := fmt.Sprintf("topics:%s", topicID)
	return s.subscribeToStream(streamKey, eventsChan)
}

// subscribeToPattern monitors multiple streams matching a pattern
func (s *Subscriber) subscribeToPattern(pattern string, eventsChan chan<- StreamEvent) error {
	// Use a consumer group for reliable delivery
	consumerGroup := "relay-consumers"
	consumerName := fmt.Sprintf("consumer-%d", time.Now().UnixNano())

	// Get all keys matching the pattern
	keys, err := s.client.Keys(s.ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys for pattern %s: %w", pattern, err)
	}

	if len(keys) == 0 {
		// No streams yet, but we should still monitor for new ones
		go s.monitorPattern(pattern, eventsChan, consumerGroup, consumerName)
		return nil
	}

	// Create consumer group for each stream
	for _, key := range keys {
		err := s.client.XGroupCreateMkStream(s.ctx, key, consumerGroup, "0").Err()
		if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
			log.Printf("Warning: failed to create consumer group for %s: %v", key, err)
		}
	}

	// Start reading from all streams
	go s.readFromStreams(keys, consumerGroup, consumerName, eventsChan)

	// Monitor for new streams matching the pattern
	go s.monitorPattern(pattern, eventsChan, consumerGroup, consumerName)

	return nil
}

// subscribeToStream subscribes to a single stream
func (s *Subscriber) subscribeToStream(streamKey string, eventsChan chan<- StreamEvent) error {
	consumerGroup := "relay-consumers"
	consumerName := fmt.Sprintf("consumer-%d", time.Now().UnixNano())

	// Create consumer group
	err := s.client.XGroupCreateMkStream(s.ctx, streamKey, consumerGroup, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}

	go s.readFromStream(streamKey, consumerGroup, consumerName, eventsChan)
	return nil
}

func (s *Subscriber) readFromStreams(streams []string, group, consumer string, eventsChan chan<- StreamEvent) {
	for {
		// XReadGroup requires: all stream keys first, then all IDs
		// Format: [stream1, stream2, ..., streamN, id1, id2, ..., idN]
		streamsList := make([]string, 0, len(streams)*2)
		// Add all stream keys first
		streamsList = append(streamsList, streams...)
		// Then add all IDs (">" for each stream means "new messages never delivered")
		for range streams {
			streamsList = append(streamsList, ">")
		}

		results, err := s.client.XReadGroup(s.ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: consumer,
			Streams:  streamsList,
			Count:    10,
			Block:   time.Second,
		}).Result()

		if err == redis.Nil {
			continue
		}
		if err != nil {
			log.Printf("Error reading from streams: %v", err)
			time.Sleep(time.Second)
			continue
		}

		for _, stream := range results {
			for _, msg := range stream.Messages {
				fields := make(map[string]string)
				for k, v := range msg.Values {
					fields[k] = fmt.Sprintf("%v", v)
				}
				eventsChan <- StreamEvent{
					StreamKey: stream.Stream,
					ID:        msg.ID,
					Fields:    fields,
				}
			}
		}
	}
}

func (s *Subscriber) readFromStream(streamKey, group, consumer string, eventsChan chan<- StreamEvent) {
	for {
		results, err := s.client.XReadGroup(s.ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: consumer,
			Streams:  []string{streamKey, ">"},
			Count:    10,
			Block:   time.Second,
		}).Result()

		if err == redis.Nil {
			continue
		}
		if err != nil {
			log.Printf("Error reading from stream %s: %v", streamKey, err)
			time.Sleep(time.Second)
			continue
		}

		for _, stream := range results {
			for _, msg := range stream.Messages {
				fields := make(map[string]string)
				for k, v := range msg.Values {
					fields[k] = fmt.Sprintf("%v", v)
				}
				eventsChan <- StreamEvent{
					StreamKey: streamKey,
					ID:        msg.ID,
					Fields:    fields,
				}
			}
		}
	}
}

func (s *Subscriber) monitorPattern(pattern string, eventsChan chan<- StreamEvent, group, consumer string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	knownStreams := make(map[string]bool)

	for range ticker.C {
		keys, err := s.client.Keys(s.ctx, pattern).Result()
		if err != nil {
			continue
		}

		newStreams := []string{}
		for _, key := range keys {
			if !knownStreams[key] {
				knownStreams[key] = true
				newStreams = append(newStreams, key)

				err := s.client.XGroupCreateMkStream(s.ctx, key, group, "0").Err()
				if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
					log.Printf("Warning: failed to create consumer group for %s: %v", key, err)
				}
			}
		}

		if len(newStreams) > 0 {
			go s.readFromStreams(newStreams, group, consumer, eventsChan)
		}
	}
}

func (s *Subscriber) Client() *redis.Client {
	return s.client
}

func (s *Subscriber) Close() error {
	return s.client.Close()
}

// StreamKey returns the Redis stream key for a topic, with optional anonymous prefix
func StreamKey(topicID string, anonymous bool) string {
	if anonymous {
		return fmt.Sprintf("anon:topics:%s", topicID)
	}
	return fmt.Sprintf("topics:%s", topicID)
}

// ValidateAnonChannel validates that an anonymous channel exists and is not expired
func (s *Subscriber) ValidateAnonChannel(ctx context.Context, topicID string) error {
	// Check if topicID starts with "anon_"
	if !strings.HasPrefix(topicID, "anon_") {
		return fmt.Errorf("invalid anonymous topic ID format")
	}

	// Check if channel exists and is not expired
	score, err := s.client.ZScore(ctx, "anon:channels", topicID).Result()
	if err == redis.Nil {
		return fmt.Errorf("anonymous channel not found")
	}
	if err != nil {
		return fmt.Errorf("failed to check anonymous channel: %w", err)
	}

	// Check if expired (score is expiry timestamp in milliseconds)
	now := time.Now().UnixMilli()
	if score <= float64(now) {
		return fmt.Errorf("anonymous channel expired")
	}

	// Check if disabled
	disabled, err := s.client.HGet(ctx, fmt.Sprintf("anon:meta:%s", topicID), "disabled").Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to check disabled status: %w", err)
	}
	if disabled == "true" {
		return fmt.Errorf("anonymous channel disabled")
	}

	return nil
}

// CreateAnonChannel creates an anonymous channel in Redis with all required keys
func (s *Subscriber) CreateAnonChannel(ctx context.Context, topicID, ip string, expiresAt time.Time) error {
	expiresAtMs := expiresAt.UnixMilli()
	createdAt := time.Now().UTC().Format(time.RFC3339)
	expiresAtStr := expiresAt.UTC().Format(time.RFC3339)
	ttl := expiresAt.Sub(time.Now())

	pipe := s.client.Pipeline()
	// Add to sorted set for expiry tracking
	pipe.ZAdd(ctx, "anon:channels", redis.Z{
		Score:  float64(expiresAtMs),
		Member: topicID,
	})
	// Store metadata
	pipe.HSet(ctx, fmt.Sprintf("anon:meta:%s", topicID), map[string]interface{}{
		"ip":         ip,
		"created_at": createdAt,
		"expires_at": expiresAtStr,
		"disabled":   "false",
	})
	pipe.Expire(ctx, fmt.Sprintf("anon:meta:%s", topicID), ttl)
	// Track per-IP
	pipe.SAdd(ctx, fmt.Sprintf("anon:ip:%s", ip), topicID)
	pipe.Expire(ctx, fmt.Sprintf("anon:ip:%s", ip), ttl)

	_, err := pipe.Exec(ctx)
	return err
}

// CheckAnonIPCount returns the number of active anonymous channels for an IP
func (s *Subscriber) CheckAnonIPCount(ctx context.Context, ip string) (int64, error) {
	return s.client.SCard(ctx, fmt.Sprintf("anon:ip:%s", ip)).Result()
}

// TrackAnonConnection marks that a CLI is connected to an anonymous channel
func (s *Subscriber) TrackAnonConnection(ctx context.Context, topicID string) error {
	return s.client.Set(ctx, fmt.Sprintf("anon:connected:%s", topicID), "1", 2*time.Hour).Err()
}

// RemoveAnonConnection removes the connection tracking for an anonymous channel
func (s *Subscriber) RemoveAnonConnection(ctx context.Context, topicID string) error {
	return s.client.Del(ctx, fmt.Sprintf("anon:connected:%s", topicID)).Err()
}

// SubscribeToAnonymousTopic subscribes to an anonymous topic stream
func (s *Subscriber) SubscribeToAnonymousTopic(topicID string, eventsChan chan<- StreamEvent) error {
	streamKey := StreamKey(topicID, true)
	return s.subscribeToStream(streamKey, eventsChan)
}


