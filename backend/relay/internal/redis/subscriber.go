package redis

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Subscriber struct {
	client *redis.Client
	ctx    context.Context
}

func NewSubscriber(addr string) (*Subscriber, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
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
// Returns a channel that receives events matching the pattern webhook:events:{appId}:*
func (s *Subscriber) SubscribeToApplication(appID string, eventsChan chan<- StreamEvent) error {
	pattern := fmt.Sprintf("webhook:events:%s:*", appID)
	return s.subscribeToPattern(pattern, eventsChan)
}

// SubscribeToTopic subscribes to a specific application/topic combination
func (s *Subscriber) SubscribeToTopic(appID, topicID string, eventsChan chan<- StreamEvent) error {
	streamKey := fmt.Sprintf("webhook:events:%s:%s", appID, topicID)
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
		streamsList := make([]string, 0, len(streams)*2)
		for _, stream := range streams {
			streamsList = append(streamsList, stream, ">")
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

func (s *Subscriber) Close() error {
	return s.client.Close()
}


