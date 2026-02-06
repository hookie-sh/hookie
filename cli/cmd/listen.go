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
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/hookie/cli/internal/config"
	"github.com/hookie/cli/internal/relay"
	"github.com/hookie/cli/proto"
)

// runListen is a shared function that handles listening to events
// Parameters:
//   - topicID: Topic ID to listen to (empty for app level)
//   - appID: Application ID to listen to (empty for topic level)
//   - orgID: Organization ID (used for access verification)
//   - endpointURL: Optional default URL to forward events to
//   - topicForwardMap: Optional map of topic_id -> forward URL for per-topic forwarding
func runListen(topicID, appID, orgID string, endpointURL *url.URL, topicForwardMap map[string]*url.URL) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Token == "" {
		// Fall back to anonymous mode
		return runAnonymousListen(endpointURL)
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
		<-sigChan
		fmt.Println("\nShutting down...")
		cancel()
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

	fmt.Printf("Listening for events (%s)...\n", subscriptionInfo)
	fmt.Println("Press Ctrl+C to stop\n")

	// Create HTTP client with timeout for forwarding
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	for {
		event, err := stream.Recv()
		if err != nil {
			if ctx.Err() != nil {
				return nil // Context cancelled
			}
			return fmt.Errorf("failed to receive event: %w", err)
		}

		// Check if this is a disconnect event
		if event.EventType == "disconnect" {
			fmt.Println(color.YellowString("\nDisconnected by server. Exiting..."))
			return nil
		}

		printEvent(event, debug)

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
		if forwardURL != nil {
			go forwardEvent(httpClient, event, forwardURL)
		}
	}
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

func forwardEvent(client *http.Client, event *proto.Event, baseURL *url.URL) {
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
		fmt.Printf("  %s failed to create request: %v\n", color.RedString("✗"), err)
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
		fmt.Printf("  %s failed to forward: %v\n", color.RedString("✗"), err)
		return
	}
	defer resp.Body.Close()

	// Log success
	fmt.Printf("  %s forwarded to %s (status: %d)\n",
		color.GreenString("→"),
		color.CyanString(forwardURL.String()),
		resp.StatusCode,
	)
}


