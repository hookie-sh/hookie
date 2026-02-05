package realtime

import (
	"context"
	"fmt"
	"log"
	"time"
)

// BroadcastListener listens for broadcast messages to force disconnect clients
type BroadcastListener struct {
	client      *Client
	grpcService interface {
		DisconnectClientByMachineID(dbMachineID string)
	}
}

// NewBroadcastListener creates a new broadcast listener
func NewBroadcastListener(grpcService interface {
	DisconnectClientByMachineID(dbMachineID string)
}) (*BroadcastListener, error) {
	client, err := NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create realtime client: %w", err)
	}

	return &BroadcastListener{
		client:      client,
		grpcService: grpcService,
	}, nil
}

// Start starts listening for broadcast messages on machine_id channels
// It dynamically subscribes to channels as clients connect
func (bl *BroadcastListener) Start(ctx context.Context) error {
	// Connect to the Realtime server
	connectCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("[BroadcastListener] Connecting to Supabase Realtime server...")

	if err := bl.client.Connect(connectCtx); err != nil {
		return fmt.Errorf("failed to connect to realtime server: %w", err)
	}
	defer bl.client.Disconnect()

	log.Println("[BroadcastListener] Connected to Supabase Realtime server")
	log.Println("[BroadcastListener] Broadcast listener started (will subscribe to machine_id channels dynamically)")

	// Keep the connection alive until context is cancelled
	<-ctx.Done()
	log.Println("[BroadcastListener] Context cancelled, shutting down...")
	return nil
}

// SubscribeToMachineID subscribes to broadcast events for a specific machine_id channel
// The channel name is the machine_id itself
func (bl *BroadcastListener) SubscribeToMachineID(ctx context.Context, machineID string) error {
	channelName := machineID // Use machine_id as the channel name
	config := &ChannelConfig{
		Broadcast: &BroadcastConfig{
			Ack:  false,
			Self: false,
		},
	}
	channel := bl.client.Channel(channelName, config)

	// Subscribe to the channel
	subscribed := make(chan bool, 1)
	subErr := make(chan error, 1)

	err := channel.Subscribe(ctx, func(state SubscribeState, err error) {
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
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	// Wait for subscription confirmation
	select {
	case <-subscribed:
		log.Printf("[BroadcastListener] Subscribed to channel for machine ID: %s", machineID)
		// Listen for disconnect broadcasts
		channel.OnBroadcast("disconnect", func(payload map[string]interface{}) {
			log.Printf("[BroadcastListener] Received disconnect broadcast for machine ID: %s", machineID)
			bl.grpcService.DisconnectClientByMachineID(machineID)
		})
		return nil
	case err := <-subErr:
		return err
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout waiting for subscription confirmation")
	case <-ctx.Done():
		return ctx.Err()
	}
}
