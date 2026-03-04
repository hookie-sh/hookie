package cmd

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/hookie-sh/hookie/cli/internal/config"
	"github.com/hookie-sh/hookie/cli/internal/relay"
	"github.com/hookie-sh/hookie/cli/proto"
)

// logMutex serializes all logging output to prevent interleaving
var logMutex sync.Mutex

// eventCounter provides sequential IDs for matching forwarding start/completion
var eventCounter uint64

// forwardingResult represents a forwarding completion message
type forwardingResult struct {
	eventID uint64
	message string
}

// forwardingResults queues forwarding completion messages for sequential printing
var forwardingResults = make(chan forwardingResult, 1000)

// forwardingLoggerState tracks state for sequential printing
type forwardingLoggerState struct {
	nextExpectedID uint64
	pendingResults map[uint64]string
	mutex          sync.Mutex
}

// flushPendingResults prints any pending forwarding results that are ready
// It prints all consecutive results starting from nextExpectedID until it hits a gap
// This version acquires logMutex internally - use flushPendingResultsUnlocked if you already hold it
func (s *forwardingLoggerState) flushPendingResults(maxEventID uint64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Print all consecutive results starting from nextExpectedID
	// Continue until we hit a gap
	for {
		if msg, exists := s.pendingResults[s.nextExpectedID]; exists {
			logMutex.Lock()
			fmt.Print(msg)
			logMutex.Unlock()
			delete(s.pendingResults, s.nextExpectedID)
			s.nextExpectedID++
		} else {
			// No more consecutive results available
			break
		}
	}
}

// flushPendingResultsUnlocked is like flushPendingResults but assumes logMutex is already held
// Caller must hold logMutex
func (s *forwardingLoggerState) flushPendingResultsUnlocked(maxEventID uint64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Print all consecutive results starting from nextExpectedID
	// Continue until we hit a gap
	// Note: logMutex is already held by caller
	for {
		if msg, exists := s.pendingResults[s.nextExpectedID]; exists {
			fmt.Print(msg)
			delete(s.pendingResults, s.nextExpectedID)
			s.nextExpectedID++
		} else {
			// No more consecutive results available
			break
		}
	}
}

// startForwardingLogger starts a goroutine that processes forwarding results
func startForwardingLogger(ctx context.Context) *forwardingLoggerState {
	state := &forwardingLoggerState{
		nextExpectedID: 1,
		pendingResults: make(map[uint64]string),
	}
	
	// Background goroutine collects results from channel and triggers flush to print them immediately
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case result := <-forwardingResults:
				state.mutex.Lock()
				state.pendingResults[result.eventID] = result.message
				state.mutex.Unlock()
				// Trigger flush to print this result immediately if it's the next expected one
				// flushPendingResults will print all consecutive results starting from nextExpectedID
				// Use non-blocking approach: try to acquire logMutex, but don't block forever
				// If we can't acquire it immediately, the main loop will flush it when it processes the next event
				select {
				case <-ctx.Done():
					return
				default:
					// Try to flush, but don't block the main event loop
					state.flushPendingResults(0)
				}
			}
		}
	}()
	
	return state
}

// runListen is a shared function that handles listening to events
// Parameters:
//   - topicID: Topic ID to listen to (empty for app level)
//   - appID: Application ID to listen to (empty for topic level)
//   - orgID: Organization ID (used for access verification)
//   - endpointURL: Optional default URL to forward events to
//   - topicForwardMap: Optional map of topic_id -> forward URL for per-topic forwarding
//   - guiURL: Optional URL of local GUI server to ingest events into
func runListen(topicID, appID, orgID string, endpointURL *url.URL, topicForwardMap map[string]*url.URL, guiURL *url.URL) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Token == "" {
		// Fall back to anonymous mode
		return runAnonymousListen(endpointURL, guiURL)
	}

	client, err := relay.NewClient(cfg.Token, debug)
	if err != nil {
		return fmt.Errorf("failed to connect to relay: %w", err)
	}
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		fmt.Printf("\nReceived %v, shutting down...\n", sig)
		// Cancel context first to stop all goroutines
		cancel()
		// Close client connection to unblock Recv() calls
		client.Close()
		// Give a short timeout for graceful shutdown, then force exit
		time.Sleep(2 * time.Second)
		fmt.Println("Forcing exit...")
		os.Exit(0)
	}()

	stream, err := client.Subscribe(ctx, appID, topicID, orgID, cfg.MachineID)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	// Build subscription info string
	var subscriptionInfo string
	if topicID != "" {
		subscriptionInfo = fmt.Sprintf("topic: %s", color.CyanString(topicID))
	} else if appID != "" {
		subscriptionInfo = fmt.Sprintf("app: %s, all topics", color.CyanString(appID))
	}
	if endpointURL != nil {
		subscriptionInfo += fmt.Sprintf(", forwarding to: %s", color.CyanString(endpointURL.String()))
	}
	if guiURL != nil {
		subscriptionInfo += fmt.Sprintf(", GUI at %s", color.CyanString(guiURL.String()))
	}

	fmt.Printf("Listening for events (%s)...\n", subscriptionInfo)
	fmt.Println("Press Ctrl+C to stop")

	// Reset event counter for new session
	atomic.StoreUint64(&eventCounter, 0)
	
	// Track total events received for debugging
	var totalEventsReceived uint64
	
	// Start forwarding logger to ensure sequential output
	forwardingState := startForwardingLogger(ctx)

	// Create HTTP client with timeout for forwarding
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Process events with flow control: receive → process → send ready → receive next
	for {
		select {
		case <-ctx.Done():
			// Context cancelled - exit immediately
			return nil
		default:
		}

		// Receive event from stream
		event, err := stream.Recv()
		if err != nil {
			if ctx.Err() != nil {
				return nil // Context cancelled
			}
			return fmt.Errorf("failed to receive event: %w", err)
		}

		// Debug logging
		if debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] Received event from gRPC: appID=%s, topicID=%s, method=%s, path=%s\n", 
				event.AppId, event.TopicId, event.Method, event.Path)
		}

		// Check if this is a disconnect event
		if event.EventType == "disconnect" {
			fmt.Println(color.YellowString("\nDisconnected by server. Exiting..."))
			return nil
		}

		// Forward event to endpoint if provided
		// Priority: topic-specific URL > default URL
		var forwardURL *url.URL
		if topicForwardMap != nil && event.TopicId != "" {
			if topicURL, exists := topicForwardMap[event.TopicId]; exists {
				forwardURL = topicURL
			}
		}
		// Fall back to default forward URL if no topic-specific URL
		if forwardURL == nil {
			forwardURL = endpointURL
		}

		// Get unique event ID for matching forwarding start/completion
		eventID := atomic.AddUint64(&eventCounter, 1)

		// Track total events received
		totalEventsReceived++
		if debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] Total events received: %d, Processing event ID=%d: appID=%s, topicID=%s\n", 
				totalEventsReceived, eventID, event.AppId, event.TopicId)
		}

		// Protect event printing and forwarding log together to maintain order
		logMutex.Lock()
		// Flush any pending forwarding results for earlier events before printing new event
		// This ensures forwarding completions appear right after their events, before the next event
		forwardingState.flushPendingResultsUnlocked(eventID - 1)
		
		if debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] About to print event ID=%d\n", eventID)
		}
		printEvent(event, debug)
		if debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] Finished printing event ID=%d\n", eventID)
		}
		if forwardURL != nil {
			// Log forwarding attempt immediately to maintain log order
			fmt.Printf("  %s forwarding to %s... [%d]\n",
				color.YellowString("→"),
				color.CyanString(forwardURL.String()),
				eventID,
			)
		}
		// Flush again after printing to catch any completions that arrived immediately
		forwardingState.flushPendingResultsUnlocked(eventID)
		logMutex.Unlock()

		if forwardURL != nil {
			go forwardEvent(httpClient, event, forwardURL, eventID, event.AppId, event.TopicId)
		}

		if guiURL != nil {
			go ingestEventToGUI(httpClient, event, guiURL)
		}

		// Send Ready signal to relay to indicate we're ready for next event
		// This implements flow control: CLI controls the rate
		readyMsg := &proto.SubscribeMessage{
			Message: &proto.SubscribeMessage_Ready{
				Ready: &proto.Ready{},
			},
		}
		if err := stream.Send(readyMsg); err != nil {
			if ctx.Err() != nil {
				return nil // Context cancelled
			}
			return fmt.Errorf("failed to send Ready signal: %w", err)
		}
	}
}

// formatForwardingMessage formats a forwarding result message with event context
func formatForwardingMessage(appID, topicID, action, errorMsg string, eventID uint64, isError bool) string {
	var prefix string
	if isError {
		prefix = color.RedString("✗")
	} else {
		prefix = color.GreenString("→")
	}
	
	// Match the event header format: [app_id] with optional (topic: topic_id)
	context := fmt.Sprintf("[%s]", color.GreenString(appID))
	if topicID != "" {
		context += fmt.Sprintf(" (topic: %s)", color.CyanString(topicID))
	}
	
	if isError && errorMsg != "" {
		return fmt.Sprintf("  %s %s %s: %s [%d]\n", prefix, context, action, errorMsg, eventID)
	}
	return fmt.Sprintf("  %s %s %s [%d]\n", prefix, context, action, eventID)
}

func printEvent(event *proto.Event, debug bool) {
	timestamp := time.Unix(0, event.Timestamp).Format(time.RFC3339)
	
	fmt.Printf("%s [%s] %s %s",
		color.YellowString(timestamp),
		color.GreenString(event.AppId),
		color.MagentaString(event.Method),
		event.Path,
	)

	if event.TopicId != "" {
		fmt.Printf(" (topic: %s)", color.CyanString(event.TopicId))
	}

	fmt.Println()

	// Only show detailed info in debug mode
	if debug {
		// Parse and print headers
		if event.Headers != "" {
			var headers map[string]interface{}
			if err := json.Unmarshal([]byte(event.Headers), &headers); err == nil {
				fmt.Println("  Headers:")
				for k, v := range headers {
					fmt.Printf("    %s: %v\n", k, v)
				}
			}
		}

		// Parse and print query params
		if event.Query != "" && event.Query != "{}" {
			var query map[string]interface{}
			if err := json.Unmarshal([]byte(event.Query), &query); err == nil && len(query) > 0 {
				fmt.Println("  Query:")
				for k, v := range query {
					fmt.Printf("    %s: %v\n", k, v)
				}
			}
		}

		// Decode and print body
		if event.Body != "" {
			bodyBytes, err := base64.StdEncoding.DecodeString(event.Body)
			if err == nil {
				var bodyJSON interface{}
				if err := json.Unmarshal(bodyBytes, &bodyJSON); err == nil {
					bodyPretty, _ := json.MarshalIndent(bodyJSON, "    ", "  ")
					fmt.Println("  Body:")
					fmt.Println(string(bodyPretty))
				} else {
					fmt.Printf("  Body: %s\n", string(bodyBytes))
				}
			}
		}

		fmt.Println()
	}
}

func ingestEventToGUI(client *http.Client, event *proto.Event, guiURL *url.URL) {
	ingestURL := guiURL.String() + "/api/ingest"
	req := map[string]interface{}{
		"method":    event.Method,
		"path":      event.Path,
		"query":     event.Query,
		"headers":   event.Headers,
		"body":      event.Body,
		"contentType": event.ContentType,
		"timestamp": event.Timestamp,
		"appId":     event.AppId,
		"topicId":   event.TopicId,
	}
	body, _ := json.Marshal(req)
	resp, err := client.Post(ingestURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

func forwardEvent(client *http.Client, event *proto.Event, baseURL *url.URL, eventID uint64, appID, topicID string) {
	// Build the full URL with query parameters
	forwardURL := *baseURL
	
	// Parse and add query parameters from the event
	if event.Query != "" && event.Query != "{}" {
		var queryParams map[string]interface{}
		if err := json.Unmarshal([]byte(event.Query), &queryParams); err == nil {
			query := forwardURL.Query()
			for k, v := range queryParams {
				// Convert value to string
				var val string
				switch v := v.(type) {
				case string:
					val = v
				case []interface{}:
					// Handle arrays by taking first element or joining
					if len(v) > 0 {
						val = fmt.Sprintf("%v", v[0])
					}
				default:
					val = fmt.Sprintf("%v", v)
				}
				query.Add(k, val)
			}
			forwardURL.RawQuery = query.Encode()
		}
	}

	// Decode body
	var bodyReader io.Reader
	if event.Body != "" {
		bodyBytes, err := base64.StdEncoding.DecodeString(event.Body)
		if err == nil {
			bodyReader = bytes.NewReader(bodyBytes)
		}
	}

	// Create request with original method
	req, err := http.NewRequest(event.Method, forwardURL.String(), bodyReader)
	if err != nil {
		// Queue error message for sequential printing with event context
		msg := formatForwardingMessage(appID, topicID, "failed to create request", err.Error(), eventID, true)
		select {
		case forwardingResults <- forwardingResult{eventID: eventID, message: msg}:
		default:
			// Channel full, fall back to direct logging
			logMutex.Lock()
			fmt.Print(msg)
			logMutex.Unlock()
		}
		return
	}

	// Parse and set headers
	hasContentType := false
	if event.Headers != "" {
		var headers map[string]interface{}
		if err := json.Unmarshal([]byte(event.Headers), &headers); err == nil {
			for k, v := range headers {
				// Skip Host header as it will be set automatically
				if k == "Host" {
					continue
				}
				// Track if Content-Type is in headers
				if k == "Content-Type" {
					hasContentType = true
				}
				// Convert value to string
				var val string
				switch v := v.(type) {
				case string:
					val = v
				case []interface{}:
					if len(v) > 0 {
						val = fmt.Sprintf("%v", v[0])
					}
				default:
					val = fmt.Sprintf("%v", v)
				}
				req.Header.Set(k, val)
			}
		}
	}

	// Set Content-Type if not already set from headers
	if !hasContentType && event.ContentType != "" {
		req.Header.Set("Content-Type", event.ContentType)
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		// Queue error message for sequential printing with event context
		msg := formatForwardingMessage(appID, topicID, fmt.Sprintf("failed to forward to %s", forwardURL.String()), err.Error(), eventID, true)
		select {
		case forwardingResults <- forwardingResult{eventID: eventID, message: msg}:
		default:
			// Channel full, fall back to direct logging
			logMutex.Lock()
			fmt.Print(msg)
			logMutex.Unlock()
		}
		return
	}
	defer resp.Body.Close()

	// Queue completion message for sequential printing with event context
	msg := formatForwardingMessage(appID, topicID, fmt.Sprintf("forwarded to %s (status: %d)", forwardURL.String(), resp.StatusCode), "", eventID, false)
	select {
	case forwardingResults <- forwardingResult{eventID: eventID, message: msg}:
	default:
		// Channel full, fall back to direct logging
		logMutex.Lock()
		fmt.Print(msg)
		logMutex.Unlock()
	}
}


