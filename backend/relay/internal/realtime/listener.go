package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type Listener struct {
	client      *Client
	grpcService interface {
		DisconnectClientByMachineID(dbMachineID string)
	}
}

func NewListener(grpcService interface {
	DisconnectClientByMachineID(dbMachineID string)
}) (*Listener, error) {
	client, err := NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create realtime client: %w", err)
	}

	return &Listener{
		client:      client,
		grpcService: grpcService,
	}, nil
}

func (l *Listener) Start(ctx context.Context) error {
	// Connect to the Realtime server
	connectCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("Connecting to Supabase Realtime server...")

	if err := l.client.Connect(connectCtx); err != nil {
		return fmt.Errorf("failed to connect to realtime server: %w", err)
	}
	defer l.client.Disconnect()

	log.Println("Connected to Supabase Realtime server")

	// Create a channel for Postgres changes
	channelName := "realtime:public:connected_clients"
	config := &ChannelConfig{
		PostgresChanges: []PostgresChangeConfig{
			{
				Event:  "UPDATE",
				Schema: "public",
				Table:  "connected_clients",
			},
		},
	}
	channel := l.client.Channel(channelName, config)

	log.Printf("Subscribing to %s channel...", channelName)

	// Subscribe to the channel
	subscribed := make(chan bool, 1)
	subErr := make(chan error, 1)

	err := channel.Subscribe(ctx, func(state SubscribeState, err error) {
		if err != nil {
			subErr <- fmt.Errorf("subscription error: %w", err)
			return
		}

		if state == SubscribeStateSubscribed {
			log.Println("Successfully subscribed to connected_clients changes")
			subscribed <- true
		}
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	// Wait for subscription confirmation
	select {
	case <-subscribed:
		log.Println("Realtime listener started for connected_clients table")
	case err := <-subErr:
		return err
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout waiting for subscription confirmation")
	case <-ctx.Done():
		return ctx.Err()
	}

	// Listen for UPDATE changes on the connected_clients table
	channel.OnPostgresChange("UPDATE", func(change PostgresChangeEvent) {
		l.handlePostgresChange(change)
	})

	// Keep the connection alive until context is cancelled
	<-ctx.Done()
	log.Println("Realtime listener context cancelled, shutting down...")
	return nil
}

func (l *Listener) handlePostgresChange(change PostgresChangeEvent) {
	// Parse the payload to extract the record
	var payload map[string]interface{}
	if err := json.Unmarshal(change.Payload, &payload); err != nil {
		log.Printf("[Listener] Failed to parse payload: %v", err)
		return
	}

	log.Printf("[Listener] Received Postgres change event, payload keys: %v", getMapKeys(payload))

	// The payload structure from Supabase Realtime has the data nested under "data"
	dataRaw, ok := payload["data"]
	if !ok {
		// Try direct access to "record" (fallback for different payload structures)
		if recordRaw, ok := payload["record"]; ok {
			log.Printf("[Listener] Found 'record' directly in payload")
			dataRaw = recordRaw
		} else {
			log.Printf("[Listener] No 'data' or 'record' field in UPDATE payload")
			return
		}
	}

	var newRecordRaw interface{}
	if data, ok := dataRaw.(map[string]interface{}); ok {
		// Get the new record (for UPDATE events)
		var found bool
		newRecordRaw, found = data["record"]
		if !found {
			log.Printf("[Listener] No 'record' field in UPDATE payload data")
			return
		}
	} else {
		// Data is already the record
		newRecordRaw = dataRaw
	}

	newRecordBytes, err := json.Marshal(newRecordRaw)
	if err != nil {
		log.Printf("[Listener] Error marshaling new record: %v", err)
		return
	}

	var record struct {
		ID             string  `json:"id"`
		DisconnectedAt *string `json:"disconnected_at"`
	}

	if err := json.Unmarshal(newRecordBytes, &record); err != nil {
		log.Printf("[Listener] Error unmarshaling new record: %v", err)
		return
	}

	log.Printf("[Listener] Parsed record - ID: %s, DisconnectedAt: %v", record.ID, record.DisconnectedAt)

	// Only process if disconnected_at is set
	if record.DisconnectedAt == nil || *record.DisconnectedAt == "" {
		log.Printf("[Listener] Skipping - disconnected_at is not set")
		return
	}

	log.Printf("[Listener] Received disconnect event for machine ID: %s", record.ID)

	// Disconnect the client
	l.grpcService.DisconnectClientByMachineID(record.ID)
}

// Helper function to get keys from a map for logging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
