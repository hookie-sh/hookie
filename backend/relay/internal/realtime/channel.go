package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// Channel represents a Realtime channel
type Channel struct {
	client              *Client
	topic               string
	config              *ChannelConfig
	state               SubscribeState
	joinRef             string
	mu                  sync.RWMutex
	postgresChangeHandlers map[string][]func(PostgresChangeEvent)
	broadcastHandlers      map[string][]func(map[string]interface{}) // event name -> handlers
	presenceHandlers       []func(PresenceEvent)
	subscribeCallback      func(SubscribeState, error)
	subscribed          chan bool
	subscribeErr        chan error
}

// Subscribe subscribes to the channel
func (ch *Channel) Subscribe(ctx context.Context, callback func(SubscribeState, error)) error {
	ch.mu.Lock()
	ch.subscribeCallback = callback
	ch.subscribed = make(chan bool, 1)
	ch.subscribeErr = make(chan error, 1)
	ch.postgresChangeHandlers = make(map[string][]func(PostgresChangeEvent))
	ch.presenceHandlers = make([]func(PresenceEvent), 0)
	ch.broadcastHandlers = make(map[string][]func(map[string]interface{}))
	ch.mu.Unlock()

	// Build join payload - config is always required
	payload := make(map[string]interface{})
	config := make(map[string]interface{})
	
	if ch.config != nil {
		if ch.config.Broadcast != nil {
			config["broadcast"] = map[string]interface{}{
				"ack": ch.config.Broadcast.Ack,
				"self": ch.config.Broadcast.Self,
			}
		}

		if ch.config.Presence != nil {
			config["presence"] = map[string]interface{}{
				"enabled": ch.config.Presence.Enabled,
			}
			if ch.config.Presence.Key != "" {
				config["presence"].(map[string]interface{})["key"] = ch.config.Presence.Key
			}
		}

		if len(ch.config.PostgresChanges) > 0 {
			pgChanges := make([]map[string]interface{}, len(ch.config.PostgresChanges))
			for i, pgc := range ch.config.PostgresChanges {
				pgChanges[i] = map[string]interface{}{
					"event":  pgc.Event,
					"schema": pgc.Schema,
					"table":  pgc.Table,
				}
				if pgc.Filter != "" {
					pgChanges[i]["filter"] = pgc.Filter
				}
			}
			config["postgres_changes"] = pgChanges
		}

		if ch.config.Private {
			config["private"] = true
		}
	}

	// Always include config, even if empty
	payload["config"] = config

	ref := ch.client.nextRef()
	joinMsg := &Message{
		JoinRef: &ch.joinRef,
		Ref:     &ref,
		Topic:   ch.topic,
		Event:   "phx_join",
		Payload: payload,
	}

	// #region agent log
	if joinMsgBytes, err := joinMsg.Serialize(); err == nil {
		log.Printf("[DEBUG] Join message topic=%s, payload=%s, serialized=%s", ch.topic, fmt.Sprintf("%+v", payload), string(joinMsgBytes))
	}
	// #endregion

	if err := ch.client.sendMessage(ctx, joinMsg); err != nil {
		return fmt.Errorf("failed to send join message: %w", err)
	}

	// Wait for subscription confirmation
	select {
	case <-ch.subscribed:
		if callback != nil {
			callback(SubscribeStateSubscribed, nil)
		}
		return nil
	case err := <-ch.subscribeErr:
		if callback != nil {
			callback(SubscribeStateErrored, err)
		}
		return err
	case <-time.After(5 * time.Second):
		err := fmt.Errorf("timeout waiting for subscription confirmation")
		if callback != nil {
			callback(SubscribeStateErrored, err)
		}
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// OnPostgresChange registers a handler for Postgres change events
func (ch *Channel) OnPostgresChange(eventType string, handler func(PostgresChangeEvent)) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.postgresChangeHandlers == nil {
		ch.postgresChangeHandlers = make(map[string][]func(PostgresChangeEvent))
	}
	ch.postgresChangeHandlers[eventType] = append(ch.postgresChangeHandlers[eventType], handler)
}

// OnPresence registers a handler for presence events
func (ch *Channel) OnPresence(handler func(PresenceEvent)) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.presenceHandlers == nil {
		ch.presenceHandlers = make([]func(PresenceEvent), 0)
	}
	ch.presenceHandlers = append(ch.presenceHandlers, handler)
}

// OnBroadcast registers a handler for broadcast events
func (ch *Channel) OnBroadcast(eventName string, handler func(map[string]interface{})) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.broadcastHandlers == nil {
		ch.broadcastHandlers = make(map[string][]func(map[string]interface{}))
	}
	ch.broadcastHandlers[eventName] = append(ch.broadcastHandlers[eventName], handler)
}

// handleBroadcast handles broadcast events
func (ch *Channel) handleBroadcast(msg *Message) {
	ch.mu.RLock()
	handlers := ch.broadcastHandlers[msg.Event]
	// Also check for "*" handlers (all events)
	if handlersForAll, ok := ch.broadcastHandlers["*"]; ok {
		handlers = append(handlers, handlersForAll...)
	}
	ch.mu.RUnlock()

	if len(handlers) == 0 {
		return
	}

	for _, handler := range handlers {
		handler(msg.Payload)
	}
}

// SendBroadcast sends a broadcast message on this channel
func (ch *Channel) SendBroadcast(eventName string, payload map[string]interface{}) error {
	ch.mu.RLock()
	joinRef := ch.joinRef
	ch.mu.RUnlock()

	ref := ch.client.nextRef()
	broadcastPayload := map[string]interface{}{
		"type":    "broadcast",
		"event":   eventName,
		"payload": payload,
	}

	msg := &Message{
		JoinRef: &joinRef,
		Ref:     &ref,
		Topic:   ch.topic,
		Event:   "broadcast",
		Payload: broadcastPayload,
	}

	return ch.client.sendMessage(context.Background(), msg)
}

// Track sends a presence track message
func (ch *Channel) Track(payload map[string]interface{}) error {
	ch.mu.RLock()
	joinRef := ch.joinRef
	ch.mu.RUnlock()

	ref := ch.client.nextRef()
	trackPayload := map[string]interface{}{
		"type":    "presence",
		"event":   "track",
		"payload": payload,
	}

	msg := &Message{
		JoinRef: &joinRef,
		Ref:     &ref,
		Topic:   ch.topic,
		Event:   "presence",
		Payload: trackPayload,
	}

	return ch.client.sendMessage(context.Background(), msg)
}

// Untrack sends a presence untrack message
func (ch *Channel) Untrack() error {
	ch.mu.RLock()
	joinRef := ch.joinRef
	ch.mu.RUnlock()

	ref := ch.client.nextRef()
	untrackPayload := map[string]interface{}{
		"type":  "presence",
		"event": "untrack",
	}

	msg := &Message{
		JoinRef: &joinRef,
		Ref:     &ref,
		Topic:   ch.topic,
		Event:   "presence",
		Payload: untrackPayload,
	}

	return ch.client.sendMessage(context.Background(), msg)
}

// handleMessage processes incoming messages for this channel
func (ch *Channel) handleMessage(msg *Message) {
	switch msg.Event {
	case "phx_reply":
		ch.handleReply(msg)
	case "phx_close":
		ch.handleClose()
	case "phx_error":
		ch.handleError(msg)
	case "postgres_changes":
		ch.handlePostgresChange(msg)
	case "presence_state":
		ch.handlePresenceState(msg)
	case "presence_diff":
		ch.handlePresenceDiff(msg)
	default:
		// Handle broadcast events (any other event name)
		ch.handleBroadcast(msg)
	}
}

// handleReply handles phx_reply messages (subscription confirmations)
func (ch *Channel) handleReply(msg *Message) {
	// #region agent log
	log.Printf("[DEBUG] Received phx_reply for topic=%s, full_payload=%+v", ch.topic, msg.Payload)
	// #endregion

	status, ok := msg.Payload["status"].(string)
	if !ok {
		// #region agent log
		log.Printf("[DEBUG] Invalid reply: missing status, payload=%+v", msg.Payload)
		// #endregion
		ch.subscribeErr <- fmt.Errorf("invalid reply: missing status")
		return
	}

	if status == "ok" {
		ch.mu.Lock()
		ch.state = SubscribeStateSubscribed
		ch.mu.Unlock()
		ch.subscribed <- true
	} else {
		response, _ := msg.Payload["response"].(map[string]interface{})
		// #region agent log
		log.Printf("[DEBUG] Subscription failed: status=%s, response=%+v, topic=%s", status, response, ch.topic)
		// #endregion
		err := fmt.Errorf("subscription failed: %v", response)
		ch.subscribeErr <- err
	}
}

// handleClose handles phx_close messages
func (ch *Channel) handleClose() {
	ch.mu.Lock()
	ch.state = SubscribeStateClosed
	ch.mu.Unlock()
}

// handleError handles phx_error messages
func (ch *Channel) handleError(msg *Message) {
	err := fmt.Errorf("channel error: %v", msg.Payload)
	ch.mu.Lock()
	ch.state = SubscribeStateErrored
	ch.mu.Unlock()
	if ch.subscribeErr != nil {
		select {
		case ch.subscribeErr <- err:
		default:
		}
	}
}

// handlePostgresChange handles postgres_changes events
func (ch *Channel) handlePostgresChange(msg *Message) {
	ch.mu.RLock()
	handlers := ch.postgresChangeHandlers
	ch.mu.RUnlock()

	// Extract event type from payload
	data, ok := msg.Payload["data"].(map[string]interface{})
	if !ok {
		log.Printf("[Channel] Invalid postgres_changes payload: missing data")
		return
	}

	eventType, ok := data["type"].(string)
	if !ok {
		log.Printf("[Channel] Invalid postgres_changes payload: missing type")
		return
	}

	// Get handlers for this event type and "*" (all events)
	handlersToCall := []func(PostgresChangeEvent){}
	if handlersForType, ok := handlers[eventType]; ok {
		handlersToCall = append(handlersToCall, handlersForType...)
	}
	if handlersForAll, ok := handlers["*"]; ok {
		handlersToCall = append(handlersToCall, handlersForAll...)
	}

	if len(handlersToCall) == 0 {
		return
	}

	// Serialize payload for the event
	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		log.Printf("[Channel] Failed to serialize postgres_changes payload: %v", err)
		return
	}

	event := PostgresChangeEvent{
		Payload: payloadBytes,
	}

	for _, handler := range handlersToCall {
		handler(event)
	}
}

// handlePresenceState handles presence_state events
func (ch *Channel) handlePresenceState(msg *Message) {
	// #region agent log
	logFile, _ := os.OpenFile("/Users/valentinprugnaud/Sites/CodyVal/hookie/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if logFile != nil {
		payloadJSON, _ := json.Marshal(msg.Payload)
		logData, _ := json.Marshal(map[string]interface{}{
			"location": "channel.go:309",
			"message": "Received presence_state event",
			"data": map[string]interface{}{
				"topic": ch.topic,
				"payload": string(payloadJSON),
			},
			"timestamp": time.Now().UnixMilli(),
			"sessionId": "debug-session",
			"runId": "run1",
			"hypothesisId": "B",
		})
		logFile.WriteString(string(logData) + "\n")
		logFile.Close()
	}
	// #endregion

	ch.mu.RLock()
	handlers := ch.presenceHandlers
	ch.mu.RUnlock()

	if len(handlers) == 0 {
		// #region agent log
		logFile, _ := os.OpenFile("/Users/valentinprugnaud/Sites/CodyVal/hookie/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if logFile != nil {
			logData, _ := json.Marshal(map[string]interface{}{
				"location": "channel.go:316",
				"message": "No presence handlers registered",
				"data": map[string]interface{}{"topic": ch.topic},
				"timestamp": time.Now().UnixMilli(),
				"sessionId": "debug-session",
				"runId": "run1",
				"hypothesisId": "B",
			})
			logFile.WriteString(string(logData) + "\n")
			logFile.Close()
		}
		// #endregion
		return
	}

	// presence_state payload is a map of keys to presence metadata
	presenceMap := msg.Payload

	// Create a sync event from the state for each key
	for key, metas := range presenceMap {
		event := PresenceEvent{
			Type:           "sync",
			Key:            key,
			CurrentPresence: make(map[string]interface{}),
		}

		if metasMap, ok := metas.(map[string]interface{}); ok {
			if metasList, ok := metasMap["metas"].([]interface{}); ok && len(metasList) > 0 {
				if firstMeta, ok := metasList[0].(map[string]interface{}); ok {
					event.CurrentPresence = firstMeta
				}
			}
		}

		// #region agent log
		logFile, _ := os.OpenFile("/Users/valentinprugnaud/Sites/CodyVal/hookie/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if logFile != nil {
			eventJSON, _ := json.Marshal(event)
			logData, _ := json.Marshal(map[string]interface{}{
				"location": "channel.go:340",
				"message": "Calling presence handler with sync event",
				"data": map[string]interface{}{
					"topic": ch.topic,
					"key": key,
					"event": string(eventJSON),
				},
				"timestamp": time.Now().UnixMilli(),
				"sessionId": "debug-session",
				"runId": "run1",
				"hypothesisId": "B",
			})
			logFile.WriteString(string(logData) + "\n")
			logFile.Close()
		}
		// #endregion

		for _, handler := range handlers {
			handler(event)
		}
	}
}

// handlePresenceDiff handles presence_diff events
func (ch *Channel) handlePresenceDiff(msg *Message) {
	// #region agent log
	logFile, _ := os.OpenFile("/Users/valentinprugnaud/Sites/CodyVal/hookie/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if logFile != nil {
		payloadJSON, _ := json.Marshal(msg.Payload)
		logData, _ := json.Marshal(map[string]interface{}{
			"location": "channel.go:345",
			"message": "Received presence_diff event",
			"data": map[string]interface{}{
				"topic": ch.topic,
				"payload": string(payloadJSON),
			},
			"timestamp": time.Now().UnixMilli(),
			"sessionId": "debug-session",
			"runId": "run1",
			"hypothesisId": "C",
		})
		logFile.WriteString(string(logData) + "\n")
		logFile.Close()
	}
	// #endregion

	ch.mu.RLock()
	handlers := ch.presenceHandlers
	ch.mu.RUnlock()

	diff := msg.Payload

	// Handle joins
	if joins, ok := diff["joins"].(map[string]interface{}); ok {
		for key, metas := range joins {
			event := PresenceEvent{
				Type:           "join",
				Key:            key,
				CurrentPresence: make(map[string]interface{}),
			}

			if metasMap, ok := metas.(map[string]interface{}); ok {
				if metasList, ok := metasMap["metas"].([]interface{}); ok && len(metasList) > 0 {
					if firstMeta, ok := metasList[0].(map[string]interface{}); ok {
						event.CurrentPresence = firstMeta
					}
				}
			}

			// #region agent log
			logFile, _ := os.OpenFile("/Users/valentinprugnaud/Sites/CodyVal/hookie/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if logFile != nil {
				eventJSON, _ := json.Marshal(event)
				logData, _ := json.Marshal(map[string]interface{}{
					"location": "channel.go:370",
					"message": "Calling presence handler with join event",
					"data": map[string]interface{}{
						"topic": ch.topic,
						"key": key,
						"event": string(eventJSON),
					},
					"timestamp": time.Now().UnixMilli(),
					"sessionId": "debug-session",
					"runId": "run1",
					"hypothesisId": "C",
				})
				logFile.WriteString(string(logData) + "\n")
				logFile.Close()
			}
			// #endregion

			for _, handler := range handlers {
				handler(event)
			}
		}
	}

	// Handle leaves
	if leaves, ok := diff["leaves"].(map[string]interface{}); ok {
		for key, metas := range leaves {
			event := PresenceEvent{
				Type:           "leave",
				Key:            key,
				CurrentPresence: make(map[string]interface{}),
			}

			if metasMap, ok := metas.(map[string]interface{}); ok {
				if metasList, ok := metasMap["metas"].([]interface{}); ok && len(metasList) > 0 {
					if firstMeta, ok := metasList[0].(map[string]interface{}); ok {
						event.CurrentPresence = firstMeta
					}
				}
			}

			// #region agent log
			logFile, _ := os.OpenFile("/Users/valentinprugnaud/Sites/CodyVal/hookie/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if logFile != nil {
				eventJSON, _ := json.Marshal(event)
				logData, _ := json.Marshal(map[string]interface{}{
					"location": "channel.go:395",
					"message": "Calling presence handler with leave event",
					"data": map[string]interface{}{
						"topic": ch.topic,
						"key": key,
						"event": string(eventJSON),
					},
					"timestamp": time.Now().UnixMilli(),
					"sessionId": "debug-session",
					"runId": "run1",
					"hypothesisId": "C",
				})
				logFile.WriteString(string(logData) + "\n")
				logFile.Close()
			}
			// #endregion

			for _, handler := range handlers {
				handler(event)
			}
		}
	}
}
