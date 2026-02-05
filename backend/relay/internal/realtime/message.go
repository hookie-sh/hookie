package realtime

import (
	"encoding/json"
	"fmt"
)

// Message represents a Phoenix Channels protocol message (version 2.0.0)
// Format: [join_ref, ref, topic, event, payload]
type Message struct {
	JoinRef *string                `json:"join_ref,omitempty"`
	Ref     *string                `json:"ref,omitempty"`
	Topic   string                 `json:"topic"`
	Event   string                 `json:"event"`
	Payload map[string]interface{} `json:"payload"`
}

// Serialize converts a Message to the Phoenix protocol 2.0.0 JSON array format
func (m *Message) Serialize() ([]byte, error) {
	var joinRef interface{}
	if m.JoinRef != nil {
		joinRef = *m.JoinRef
	} else {
		joinRef = nil
	}

	var ref interface{}
	if m.Ref != nil {
		ref = *m.Ref
	} else {
		ref = nil
	}

	// Protocol 2.0.0 uses JSON array format: [join_ref, ref, topic, event, payload]
	msgArray := []interface{}{
		joinRef,
		ref,
		m.Topic,
		m.Event,
		m.Payload,
	}

	return json.Marshal(msgArray)
}

// DeserializeMessage parses a Phoenix protocol 2.0.0 JSON array message
func DeserializeMessage(data []byte) (*Message, error) {
	var msgArray []interface{}
	if err := json.Unmarshal(data, &msgArray); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message array: %w", err)
	}

	if len(msgArray) != 5 {
		return nil, fmt.Errorf("invalid message format: expected 5 elements, got %d", len(msgArray))
	}

	msg := &Message{}

	// Parse join_ref
	if msgArray[0] != nil {
		if joinRefStr, ok := msgArray[0].(string); ok {
			msg.JoinRef = &joinRefStr
		}
	}

	// Parse ref
	if msgArray[1] != nil {
		if refStr, ok := msgArray[1].(string); ok {
			msg.Ref = &refStr
		}
	}

	// Parse topic
	if topicStr, ok := msgArray[2].(string); ok {
		msg.Topic = topicStr
	} else {
		return nil, fmt.Errorf("invalid topic type: expected string")
	}

	// Parse event
	if eventStr, ok := msgArray[3].(string); ok {
		msg.Event = eventStr
	} else {
		return nil, fmt.Errorf("invalid event type: expected string")
	}

	// Parse payload
	if payloadMap, ok := msgArray[4].(map[string]interface{}); ok {
		msg.Payload = payloadMap
	} else {
		msg.Payload = make(map[string]interface{})
	}

	return msg, nil
}

// PostgresChangeEvent represents a Postgres change event
type PostgresChangeEvent struct {
	Payload []byte
}

// PresenceEvent represents a presence event
type PresenceEvent struct {
	Type           string                 `json:"type"`
	Key            string                 `json:"key"`
	CurrentPresence map[string]interface{} `json:"current_presence"`
	Joins          map[string]interface{} `json:"joins,omitempty"`
	Leaves         map[string]interface{} `json:"leaves,omitempty"`
}

// SubscribeState represents the subscription state
type SubscribeState int

const (
	SubscribeStateJoining SubscribeState = iota
	SubscribeStateSubscribed
	SubscribeStateClosed
	SubscribeStateErrored
)

// ChannelConfig represents channel configuration
type ChannelConfig struct {
	Broadcast      *BroadcastConfig      `json:"broadcast,omitempty"`
	Presence      *PresenceConfig        `json:"presence,omitempty"`
	PostgresChanges []PostgresChangeConfig `json:"postgres_changes,omitempty"`
	Private       bool                  `json:"private,omitempty"`
}

// BroadcastConfig represents broadcast configuration
type BroadcastConfig struct {
	Ack   bool `json:"ack,omitempty"`
	Self  bool `json:"self,omitempty"`
}

// PresenceConfig represents presence configuration
type PresenceConfig struct {
	Enabled bool   `json:"enabled,omitempty"`
	Key     string `json:"key,omitempty"`
}

// PostgresChangeConfig represents Postgres change subscription configuration
type PostgresChangeConfig struct {
	Event  string `json:"event"`
	Schema string `json:"schema"`
	Table  string `json:"table"`
	Filter string `json:"filter,omitempty"`
}
