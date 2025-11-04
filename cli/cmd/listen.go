package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/hookie/cli/internal/auth"
	"github.com/hookie/cli/internal/config"
	"github.com/hookie/cli/internal/relay"
	"github.com/hookie/cli/proto"
	"github.com/spf13/cobra"
)

var (
	appIDFlag   string
	topicIDFlag string
)

var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Listen to webhook events",
	Long:  `Listen to webhook events for an application and/or specific topic. Can be run multiple times with different parameters.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if appIDFlag == "" {
			return fmt.Errorf("--app-id is required")
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if cfg.Token == "" {
			return fmt.Errorf("not authenticated. Run 'hookie login' first")
		}

		relayURL := cfg.RelayURL
		if relayURL == "" {
			relayURL = auth.GetRelayURL()
		}

		client, err := relay.NewClient(relayURL, cfg.Token)
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

		// Use flag value if set, otherwise use global orgID
		effectiveOrgID := orgIDFlag
		if effectiveOrgID == "" {
			effectiveOrgID = orgID
		}
		stream, err := client.Subscribe(ctx, appIDFlag, topicIDFlag, effectiveOrgID)
		if err != nil {
			return fmt.Errorf("failed to subscribe: %w", err)
		}

		subscriptionInfo := fmt.Sprintf("app: %s", color.CyanString(appIDFlag))
		if topicIDFlag != "" {
			subscriptionInfo += fmt.Sprintf(", topic: %s", color.CyanString(topicIDFlag))
		} else {
			subscriptionInfo += ", all topics"
		}
		if effectiveOrgID != "" {
			subscriptionInfo += fmt.Sprintf(", org: %s", color.CyanString(effectiveOrgID))
		}

		fmt.Printf("Listening for events (%s)...\n", subscriptionInfo)
		fmt.Println("Press Ctrl+C to stop\n")

		for {
			event, err := stream.Recv()
			if err != nil {
				if ctx.Err() != nil {
					return nil // Context cancelled
				}
				return fmt.Errorf("failed to receive event: %w", err)
			}

			printEvent(event)
		}
	},
}

func printEvent(event *proto.Event) {
	timestamp := time.Unix(0, event.Timestamp).Format(time.RFC3339)
	
	fmt.Printf("%s [%s] %s %s\n",
		color.YellowString(timestamp),
		color.GreenString(event.AppId),
		color.MagentaString(event.Method),
		event.Path,
	)

	if event.TopicId != "" {
		fmt.Printf("  Topic: %s\n", color.CyanString(event.TopicId))
	}

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

func init() {
	listenCmd.Flags().StringVar(&appIDFlag, "app-id", "", "Application ID (required)")
	listenCmd.Flags().StringVar(&topicIDFlag, "topic-id", "", "Topic ID (optional, listens to all topics if not specified)")
	listenCmd.Flags().StringVar(&orgIDFlag, "org-id", "", "Organization ID (optional, for org-owned applications)")
	rootCmd.AddCommand(listenCmd)
}

