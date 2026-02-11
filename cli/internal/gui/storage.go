package gui

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const defaultCapacity = 1000

// StoredEvent is the JSON representation of an event for storage and API
type StoredEvent struct {
	ID        string            `json:"id"`
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Query     map[string]string `json:"query"`
	Headers   map[string]string `json:"headers"`
	Body      interface{}       `json:"body"`
	Timestamp string            `json:"timestamp"`
	AppID     string            `json:"appId,omitempty"`
	TopicID   string            `json:"topicId,omitempty"`
}

// IngestRequest is the JSON body for POST /api/ingest
type IngestRequest struct {
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	Query       string            `json:"query"`
	Headers     string            `json:"headers"`
	Body        string            `json:"body"`
	ContentType string            `json:"contentType,omitempty"`
	Timestamp   int64             `json:"timestamp"`
	AppID       string            `json:"appId,omitempty"`
	TopicID     string            `json:"topicId,omitempty"`
}

// Storage is an in-memory ring buffer for events
type Storage struct {
	events    []StoredEvent
	capacity  int
	nextID    uint64
	mu        sync.RWMutex
	subs      []chan<- StoredEvent
	subsMu    sync.RWMutex
}

// NewStorage creates a new storage with the given capacity
func NewStorage(capacity int) *Storage {
	if capacity <= 0 {
		capacity = defaultCapacity
	}
	return &Storage{
		events:   make([]StoredEvent, 0, capacity),
		capacity: capacity,
	}

}

// Add stores an event and returns it with ID
func (s *Storage) Add(req IngestRequest) StoredEvent {
	id := atomic.AddUint64(&s.nextID, 1)
	event := s.ingestToEvent(req, id)

	s.mu.Lock()
	s.events = append([]StoredEvent{event}, s.events...)
	if len(s.events) > s.capacity {
		s.events = s.events[:s.capacity]
	}
	s.mu.Unlock()

	s.subsMu.RLock()
	for _, ch := range s.subs {
		select {
		case ch <- event:
		default:
			// Channel full, skip
		}
	}
	s.subsMu.RUnlock()

	return event
}

func (s *Storage) ingestToEvent(req IngestRequest, id uint64) StoredEvent {
	var query map[string]string
	if req.Query != "" && req.Query != "{}" {
		_ = json.Unmarshal([]byte(req.Query), &query)
	}
	if query == nil {
		query = make(map[string]string)
	}

	var headers map[string]string
	if req.Headers != "" && req.Headers != "{}" {
		_ = json.Unmarshal([]byte(req.Headers), &headers)
	}
	if headers == nil {
		headers = make(map[string]string)
	}

	var body interface{}
	if req.Body != "" {
		// Body is base64 encoded
		decoded, err := decodeBase64(req.Body)
		if err == nil {
			_ = json.Unmarshal(decoded, &body)
		} else {
			body = req.Body
		}
	}
	if body == nil {
		body = nil
	}

	timestamp := formatTimestamp(req.Timestamp)

	return StoredEvent{
		ID:        formatID(id),
		Method:    req.Method,
		Path:      req.Path,
		Query:     query,
		Headers:   headers,
		Body:      body,
		Timestamp: timestamp,
		AppID:     req.AppID,
		TopicID:   req.TopicID,
	}
}

// Events returns events, optionally filtered by since (event id)
func (s *Storage) Events(since string) []StoredEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if since == "" {
		result := make([]StoredEvent, len(s.events))
		copy(result, s.events)
		return result
	}

	sinceID, err := strconv.ParseUint(since, 10, 64)
	if err != nil {
		// Invalid since, return all
		result := make([]StoredEvent, len(s.events))
		copy(result, s.events)
		return result
	}

	var result []StoredEvent
	for _, e := range s.events {
		eid, err := strconv.ParseUint(e.ID, 10, 64)
		if err != nil {
			continue
		}
		if eid > sinceID {
			result = append(result, e)
		} else {
			break
		}
	}
	return result
}

// Subscribe returns a channel that receives new events and a cancel function to unsubscribe
func (s *Storage) Subscribe() (<-chan StoredEvent, func()) {
	ch := make(chan StoredEvent, 100)
	s.subsMu.Lock()
	s.subs = append(s.subs, ch)
	s.subsMu.Unlock()
	cancel := func() {
		s.subsMu.Lock()
		defer s.subsMu.Unlock()
		for i, c := range s.subs {
			if c == ch {
				s.subs = append(s.subs[:i], s.subs[i+1:]...)
				close(ch)
				return
			}
		}
	}
	return ch, cancel
}

func formatID(id uint64) string {
	return strconv.FormatUint(id, 10)
}

func formatTimestamp(ts int64) string {
	if ts == 0 {
		return time.Now().UTC().Format(time.RFC3339)
	}
	// Timestamp can be nanoseconds (proto) or milliseconds
	if ts > 1e15 {
		return time.Unix(0, ts).UTC().Format(time.RFC3339)
	}
	return time.Unix(ts/1000, (ts%1000)*1e6).UTC().Format(time.RFC3339)
}

func decodeBase64(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

