package realtime

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hookie-sh/hookie/backend/relay/internal/supabase"
)

// RelayPresenceListener listens for relay instance presence changes and handles disconnections
type RelayPresenceListener struct {
	client   *Client
	supabase *supabase.Client
}

// NewRelayPresenceListener creates a new listener for relay presence changes
func NewRelayPresenceListener(supabaseClient *supabase.Client) (*RelayPresenceListener, error) {
	client, err := NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create realtime client: %w", err)
	}

	return &RelayPresenceListener{
		client:   client,
		supabase: supabaseClient,
	}, nil
}

// Start starts listening for relay presence changes
func (rpl *RelayPresenceListener) Start(ctx context.Context) error {
	// Connect to the Realtime server
	connectCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("[RelayPresenceListener] Connecting to Supabase Realtime server...")

	if err := rpl.client.Connect(connectCtx); err != nil {
		return fmt.Errorf("failed to connect to realtime server: %w", err)
	}
	defer rpl.client.Disconnect()

	log.Println("[RelayPresenceListener] Connected to Supabase Realtime server")

	// Create a channel for relay instances
	channelName := "realtime:relay_instances"
	config := &ChannelConfig{
		Presence: &PresenceConfig{
			Enabled: true,
		},
	}
	channel := rpl.client.Channel(channelName, config)

	log.Printf("[RelayPresenceListener] Subscribing to %s channel...", channelName)

	// Subscribe to the channel
	subscribed := make(chan bool, 1)
	subErr := make(chan error, 1)

	err := channel.Subscribe(ctx, func(state SubscribeState, err error) {
		if err != nil {
			subErr <- fmt.Errorf("subscription error: %w", err)
			return
		}

		if state == SubscribeStateSubscribed {
			log.Printf("[RelayPresenceListener] Successfully subscribed to %s channel", channelName)
			subscribed <- true
		}
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	// Wait for subscription confirmation
	select {
	case <-subscribed:
		log.Println("[RelayPresenceListener] Relay presence listener started")
	case err := <-subErr:
		return err
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout waiting for subscription confirmation")
	case <-ctx.Done():
		return ctx.Err()
	}

	// Listen for presence events
	channel.OnPresence(func(event PresenceEvent) {
		rpl.handlePresenceEvent(ctx, event)
	})

	// Keep the connection alive until context is cancelled
	<-ctx.Done()
	log.Println("[RelayPresenceListener] Relay presence listener context cancelled, shutting down...")
	return nil
}

// handlePresenceEvent handles presence events (join, leave, sync)
func (rpl *RelayPresenceListener) handlePresenceEvent(ctx context.Context, event PresenceEvent) {
	log.Printf("[RelayPresenceListener] Received presence event: type=%s, key=%s", event.Type, event.Key)

	// Handle leave events - when a machine disconnects
	// The event.Key is now the machine ID directly (since we use machine ID as the presence key)
	if event.Type == "leave" {
		machineID := event.Key
		
		if machineID == "" {
			log.Printf("[RelayPresenceListener] Empty machine ID in leave event")
			return
		}

		log.Printf("[RelayPresenceListener] Machine %s disconnected", machineID)

		// Mark machine as disconnected
		if err := rpl.disconnectMachine(ctx, machineID); err != nil {
			log.Printf("[RelayPresenceListener] Failed to disconnect machine %s: %v", machineID, err)
		} else {
			log.Printf("[RelayPresenceListener] Successfully disconnected machine: %s", machineID)
		}
	}
}

// disconnectMachine marks a machine as disconnected using only the machine ID
func (rpl *RelayPresenceListener) disconnectMachine(ctx context.Context, machineID string) error {
	return rpl.supabase.DisconnectClientByMachineID(ctx, machineID)
}
