package realtime

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"nhooyr.io/websocket"
)

// Client represents a Supabase Realtime WebSocket client
type Client struct {
	conn        *websocket.Conn
	url         string
	apiKey      string
	channels    map[string]*Channel
	mu          sync.RWMutex
	refCounter  int64
	connected   int32
	ctx         context.Context
	cancel      context.CancelFunc
	heartbeatTicker *time.Ticker
}

// NewClient creates a new Realtime client
func NewClient() (*Client, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	apiKey := os.Getenv("SUPABASE_SECRET_KEY")

	if supabaseURL == "" || apiKey == "" {
		return nil, fmt.Errorf("SUPABASE_URL and SUPABASE_SECRET_KEY must be set")
	}

	wsURL, err := buildWebSocketURL(supabaseURL, apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to build WebSocket URL: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Client{
		url:        wsURL,
		apiKey:     apiKey,
		channels:   make(map[string]*Channel),
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

// buildWebSocketURL builds the WebSocket URL from SUPABASE_URL
func buildWebSocketURL(supabaseURL, apiKey string) (string, error) {
	parsedURL, err := url.Parse(supabaseURL)
	if err != nil {
		return "", fmt.Errorf("invalid SUPABASE_URL format: %s: %w", supabaseURL, err)
	}

	hostname := parsedURL.Hostname()
	var wsScheme string
	var wsHost string

	// Check if it's a production Supabase URL (contains .supabase.co)
	if strings.Contains(hostname, ".supabase.co") {
		wsScheme = "wss"
		wsHost = hostname
	} else {
		// Local Supabase instance
		wsScheme = "ws"
		// Use the port from the original URL, or default to 54321
		port := parsedURL.Port()
		if port == "" {
			port = "54321"
		}
		wsHost = fmt.Sprintf("%s:%s", hostname, port)
		log.Printf("Local Supabase detected, using WebSocket URL: %s://%s", wsScheme, wsHost)
	}

	wsURL := fmt.Sprintf("%s://%s/realtime/v1/websocket?apikey=%s&vsn=2.0.0", wsScheme, wsHost, apiKey)
	return wsURL, nil
}

// Connect establishes a WebSocket connection to the Realtime server
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if atomic.LoadInt32(&c.connected) == 1 {
		return fmt.Errorf("client already connected")
	}

	log.Printf("[RealtimeClient] Connecting to %s", c.url)

	conn, _, err := websocket.Dial(ctx, c.url, &websocket.DialOptions{
		CompressionMode: websocket.CompressionDisabled,
	})
	if err != nil {
		return fmt.Errorf("failed to dial WebSocket: %w", err)
	}

	c.conn = conn
	atomic.StoreInt32(&c.connected, 1)

	// Start message reader
	go c.readMessages()

	// Start heartbeat
	c.startHeartbeat()

	log.Printf("[RealtimeClient] Connected successfully")
	return nil
}

// Disconnect closes the WebSocket connection
func (c *Client) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if atomic.LoadInt32(&c.connected) == 0 {
		return
	}

	atomic.StoreInt32(&c.connected, 0)

	if c.heartbeatTicker != nil {
		c.heartbeatTicker.Stop()
	}

	if c.conn != nil {
		c.conn.Close(websocket.StatusNormalClosure, "client disconnecting")
		c.conn = nil
	}

	c.cancel()
}

// Channel returns a channel for the given topic
func (c *Client) Channel(topic string, config *ChannelConfig) *Channel {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ch, ok := c.channels[topic]; ok {
		return ch
	}

	ch := &Channel{
		client:  c,
		topic:   topic,
		config:  config,
		state:   SubscribeStateJoining,
		joinRef: fmt.Sprintf("%d", atomic.AddInt64(&c.refCounter, 1)),
	}

	c.channels[topic] = ch
	return ch
}

// nextRef generates the next reference number
func (c *Client) nextRef() string {
	return fmt.Sprintf("%d", atomic.AddInt64(&c.refCounter, 1))
}

// sendMessage sends a message to the server
func (c *Client) sendMessage(ctx context.Context, msg *Message) error {
	c.mu.RLock()
	conn := c.conn
	connected := atomic.LoadInt32(&c.connected) == 1
	c.mu.RUnlock()

	if !connected || conn == nil {
		return fmt.Errorf("client not connected")
	}

	data, err := msg.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return conn.Write(ctx, websocket.MessageText, data)
}

// readMessages reads messages from the WebSocket connection
func (c *Client) readMessages() {
	for {
		ctx, cancel := context.WithTimeout(c.ctx, 30*time.Second)
		_, data, err := c.conn.Read(ctx)
		cancel()

		if err != nil {
			if c.ctx.Err() != nil {
				// Context cancelled, normal shutdown
				return
			}
			log.Printf("[RealtimeClient] Error reading message: %v", err)
			return
		}

		msg, err := DeserializeMessage(data)
		if err != nil {
			log.Printf("[RealtimeClient] Failed to deserialize message: %v", err)
			continue
		}

		c.handleMessage(msg)
	}
}

// handleMessage routes incoming messages to the appropriate channel
func (c *Client) handleMessage(msg *Message) {
	// #region agent log
	log.Printf("[DEBUG] Received message: topic=%s, event=%s, payload=%+v", msg.Topic, msg.Event, msg.Payload)
	// #endregion

	c.mu.RLock()
	ch, ok := c.channels[msg.Topic]
	c.mu.RUnlock()

	if !ok {
		// Handle phoenix topic (heartbeat replies)
		if msg.Topic == "phoenix" {
			return
		}
		log.Printf("[RealtimeClient] Received message for unknown topic: %s", msg.Topic)
		return
	}

	ch.handleMessage(msg)
}

// startHeartbeat starts sending heartbeat messages every 25 seconds
func (c *Client) startHeartbeat() {
	c.heartbeatTicker = time.NewTicker(25 * time.Second)
	go func() {
		for {
			select {
			case <-c.heartbeatTicker.C:
				if atomic.LoadInt32(&c.connected) == 1 {
					ref := c.nextRef()
					heartbeat := &Message{
						Ref:     &ref,
						Topic:   "phoenix",
						Event:   "heartbeat",
						Payload: make(map[string]interface{}),
					}
					if err := c.sendMessage(c.ctx, heartbeat); err != nil {
						log.Printf("[RealtimeClient] Failed to send heartbeat: %v", err)
					}
				}
			case <-c.ctx.Done():
				return
			}
		}
	}()
}
