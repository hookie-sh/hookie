package realtime

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// PresenceTracker tracks relay instance presence with machine IDs
type PresenceTracker struct {
	client          *Client
	sharedChannel   *Channel // Single shared channel for all machines
	machineChannels map[string]*Channel // Map of machine ID to channel (all point to sharedChannel)
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewPresenceTracker creates a new presence tracker for a relay instance
func NewPresenceTracker() (*PresenceTracker, error) {
	client, err := NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create realtime client: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &PresenceTracker{
		client:          client,
		machineChannels: make(map[string]*Channel),
		ctx:             ctx,
		cancel:          cancel,
	}, nil
}

// Start connects to realtime and starts tracking presence
func (pt *PresenceTracker) Start() error {
	// Connect to the Realtime server
	connectCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("[PresenceTracker] Connecting to Supabase Realtime server...")

	if err := pt.client.Connect(connectCtx); err != nil {
		return fmt.Errorf("failed to connect to realtime server: %w", err)
	}

	log.Printf("[PresenceTracker] Connected to Supabase Realtime server")
	log.Printf("[PresenceTracker] Presence tracker started")

	return nil
}

// AddMachineID adds a machine ID and tracks it in presence
func (pt *PresenceTracker) AddMachineID(machineID string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	// Check if we already have a channel for this machine
	if _, exists := pt.machineChannels[machineID]; exists {
		log.Printf("[PresenceTracker] Machine %s already tracked", machineID)
		return
	}

	// All machines use the same topic name so the frontend can subscribe to one channel
	channelName := "realtime:relay_instances"
	
	// Get or create the shared channel
	var channel *Channel
	if pt.sharedChannel == nil {
		// Verify client is connected (using atomic load like the client does)
		if atomic.LoadInt32(&pt.client.connected) != 1 {
			log.Printf("[PresenceTracker] Warning: client not connected, cannot subscribe channel")
			return
		}
		
		// First machine - create channel with presence enabled
		config := &ChannelConfig{
			Presence: &PresenceConfig{
				Enabled: true,
				// Don't set Key here - we'll include machine_id in payload
			},
		}
		channel = pt.client.Channel(channelName, config)
		pt.sharedChannel = channel
		
		// Subscribe to the channel
		subscribed := make(chan bool, 1)
		subErr := make(chan error, 1)

		err := channel.Subscribe(pt.ctx, func(state SubscribeState, err error) {
			if err != nil {
				select {
				case subErr <- fmt.Errorf("subscription error: %w", err):
				default:
				}
				return
			}

			if state == SubscribeStateSubscribed {
				select {
				case subscribed <- true:
				default:
				}
			}
		})

		if err != nil {
			log.Printf("[PresenceTracker] Warning: failed to subscribe channel: %v", err)
			return
		}

		// Wait for subscription confirmation
		select {
		case <-subscribed:
			log.Printf("[PresenceTracker] Channel subscribed successfully")
		case err := <-subErr:
			log.Printf("[PresenceTracker] Warning: subscription error: %v", err)
			return
		case <-time.After(10 * time.Second):
			log.Printf("[PresenceTracker] Warning: timeout waiting for subscription confirmation")
			return
		case <-pt.ctx.Done():
			log.Printf("[PresenceTracker] Context cancelled while subscribing")
			return
		}
	} else {
		// Use the existing shared channel
		channel = pt.sharedChannel
		
		// Verify channel is subscribed
		channel.mu.RLock()
		isSubscribed := channel.state == SubscribeStateSubscribed
		channel.mu.RUnlock()
		
		if !isSubscribed {
			log.Printf("[PresenceTracker] Warning: channel not subscribed, cannot track machine %s", machineID)
			return
		}
	}

	// Track presence for all machines in a single payload
	// Since Supabase Realtime overwrites presence on each Track() call from the same client,
	// we need to include all tracked machines in one payload
	machineIDs := make([]string, 0, len(pt.machineChannels)+1)
	for id := range pt.machineChannels {
		machineIDs = append(machineIDs, id)
	}
	machineIDs = append(machineIDs, machineID)
	
	payload := map[string]interface{}{
		"updated_at": time.Now().Format(time.RFC3339),
		"machine_ids": machineIDs, // Array of all tracked machine IDs
	}

	if err := channel.Track(payload); err != nil {
		log.Printf("[PresenceTracker] Warning: failed to track presence for machine %s: %v", machineID, err)
		return
	}

	pt.machineChannels[machineID] = channel
	log.Printf("[PresenceTracker] Tracking presence for machine: %s (total: %d)", machineID, len(machineIDs))
}

// RemoveMachineID removes a machine ID from presence tracking
func (pt *PresenceTracker) RemoveMachineID(machineID string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	channel, exists := pt.machineChannels[machineID]
	if !exists {
		log.Printf("[PresenceTracker] Machine %s not found in tracked machines", machineID)
		return
	}

	// Remove from map first
	delete(pt.machineChannels, machineID)
	
	// Update presence to reflect remaining machines
	if pt.sharedChannel != nil && channel.state == SubscribeStateSubscribed {
		machineIDs := make([]string, 0, len(pt.machineChannels))
		for id := range pt.machineChannels {
			machineIDs = append(machineIDs, id)
		}
		
		if len(machineIDs) > 0 {
			// Track remaining machines
			payload := map[string]interface{}{
				"updated_at": time.Now().Format(time.RFC3339),
				"machine_ids": machineIDs,
			}
			if err := channel.Track(payload); err != nil {
				log.Printf("[PresenceTracker] Warning: failed to update presence after removing machine %s: %v", machineID, err)
			}
		} else {
			// No more machines, untrack completely
			if err := channel.Untrack(); err != nil {
				log.Printf("[PresenceTracker] Warning: failed to untrack presence for machine %s: %v", machineID, err)
			}
		}
	}

	log.Printf("[PresenceTracker] Stopped tracking presence for machine: %s", machineID)
}


// Stop stops tracking presence and disconnects
func (pt *PresenceTracker) Stop() {
	log.Printf("[PresenceTracker] Stopping presence tracker...")
	
	pt.mu.Lock()
	// Untrack all machines
	for machineID, channel := range pt.machineChannels {
		if err := channel.Untrack(); err != nil {
			log.Printf("[PresenceTracker] Warning: failed to untrack presence for machine %s: %v", machineID, err)
		}
	}
	pt.machineChannels = make(map[string]*Channel)
	pt.mu.Unlock()

	pt.cancel()
	if pt.client != nil {
		pt.client.Disconnect()
	}
}
