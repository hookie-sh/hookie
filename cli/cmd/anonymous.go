package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/hookie/cli/internal/config"
	"github.com/hookie/cli/internal/relay"
	"github.com/hookie/cli/proto"
)

// runAnonymousListen handles anonymous ephemeral channel listening
func runAnonymousListen(endpointURL *url.URL, guiURL *url.URL) error {
	// Load repository config for forward URL (anonymous mode doesn't support app_id/topic_id)
	repoConfig, _, err := config.LoadRepoConfig()
	if err != nil {
		return fmt.Errorf("failed to load repository config: %w", err)
	}

	// Use repo config forward URL if CLI flag not provided
	if endpointURL == nil && repoConfig != nil && repoConfig.Forward != "" {
		parsedURL, err := url.Parse(repoConfig.Forward)
		if err != nil {
			return fmt.Errorf("invalid forward URL in repository config: %w", err)
		}
		if parsedURL.Scheme == "" || parsedURL.Host == "" {
			return fmt.Errorf("invalid forward URL in repository config: must include scheme and host")
		}
		endpointURL = parsedURL
	}
	// Connect to relay (no auth)
	client, err := relay.NewAnonymousClient(debug)
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

	// Create anonymous channel via gRPC
	resp, err := client.CreateAnonymousChannel(ctx)
	if err != nil {
		return fmt.Errorf("failed to create anonymous channel: %w", err)
	}

	// Store channel ID in client for subscription
	client.SetChannelID(resp.ChannelId)

	// Parse expiry time
	expiresAt := time.Unix(resp.ExpiresAt, 0)
	expiresIn := time.Until(expiresAt)

	// Print session banner
	bold := color.New(color.Bold)
	fmt.Println()
	fmt.Println(color.CyanString("╔═══════════════════════════════════════════════════════════╗"))
	fmt.Println(color.CyanString("║") + "  " + bold.Sprint("Anonymous Ephemeral Channel Created") + "                    " + color.CyanString("║"))
	fmt.Println(color.CyanString("╠═══════════════════════════════════════════════════════════╣"))
	fmt.Printf(color.CyanString("║")+"  Webhook URL: %-45s "+color.CyanString("║")+"\n", color.GreenString(resp.WebhookUrl))
	if endpointURL != nil {
		fmt.Printf(color.CyanString("║")+"  Forwarding to: %-42s "+color.CyanString("║")+"\n", color.YellowString(endpointURL.String()))
	}
	if guiURL != nil {
		fmt.Printf(color.CyanString("║")+"  GUI: %-50s "+color.CyanString("║")+"\n", color.YellowString(guiURL.String()))
	}
	fmt.Printf(color.CyanString("║")+"  Expires in: %-45s "+color.CyanString("║")+"\n", color.YellowString(formatDuration(expiresIn)))
	fmt.Printf(color.CyanString("║")+"  Rate limits: %d/min, %d/day, %d KB payload          "+color.CyanString("║")+"\n",
		resp.Limits.RequestsPerMinute,
		resp.Limits.RequestsPerDay,
		resp.Limits.MaxPayloadBytes/1024)
	fmt.Println(color.CyanString("╚═══════════════════════════════════════════════════════════╝"))
	fmt.Println()
	fmt.Println(color.YellowString("💡 Tip: Sign up at https://hookie.sh to create permanent channels"))
	fmt.Println()

	// Start expiry warning goroutine
	go func() {
		// Warn 15 minutes before expiry
		warnAt := expiresAt.Add(-15 * time.Minute)
		waitDuration := time.Until(warnAt)
		if waitDuration > 0 {
			time.Sleep(waitDuration)
			if ctx.Err() == nil {
				fmt.Println()
				fmt.Println(color.YellowString("⚠️  Warning: This anonymous channel will expire in 15 minutes"))
				fmt.Println()
			}
		}
	}()

	// Generate a temporary machine ID for anonymous subscriptions
	cfg, _ := config.Load()
	machineID := cfg.MachineID
	if machineID == "" {
		machineID = "anon_temp"
	}

	// Subscribe to the anonymous channel
	stream, err := client.Subscribe(ctx, "", resp.ChannelId, "", machineID)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	fmt.Printf("Listening for events on anonymous channel %s...\n", color.CyanString(resp.ChannelId))
	if endpointURL != nil {
		fmt.Printf("Events will be forwarded to: %s\n", color.CyanString(endpointURL.String()))
	}
	if guiURL != nil {
		fmt.Printf("Events will be visible in GUI at %s\n", color.CyanString(guiURL.String()))
	}
	fmt.Println("Press Ctrl+C to stop\n")

	// Reset event counter for new session
	atomic.StoreUint64(&eventCounter, 0)
	
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
			return nil // Context cancelled
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

		// Check if this is a disconnect event
		if event.EventType == "disconnect" {
			fmt.Println(color.YellowString("\nDisconnected by server. Exiting..."))
			return nil
		}

		// Get unique event ID for matching forwarding start/completion
		eventID := atomic.AddUint64(&eventCounter, 1)

		// Flush any pending forwarding results for earlier events before printing new event
		// This ensures forwarding completions appear right after their events, before the next event
		forwardingState.flushPendingResults(eventID - 1)

		// Protect event printing and forwarding log together to maintain order
		logMutex.Lock()
		printEvent(event, debug)
		if endpointURL != nil {
			// Log forwarding attempt immediately to maintain log order
			fmt.Printf("  %s forwarding to %s... [%d]\n",
				color.YellowString("→"),
				color.CyanString(endpointURL.String()),
				eventID,
			)
		}
		logMutex.Unlock()

		// Flush again after printing to catch any completions that arrived immediately
		forwardingState.flushPendingResults(eventID)

		if endpointURL != nil {
			go forwardEvent(httpClient, event, endpointURL, eventID, event.AppId, event.TopicId)
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

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if minutes == 0 {
		return fmt.Sprintf("%d hours", hours)
	}
	return fmt.Sprintf("%d hours %d minutes", hours, minutes)
}
