package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	relayURL  string
	orgID     string
	orgIDFlag string
	debug     bool
)

var rootCmd = &cobra.Command{
	Use:   "hookie",
	Short: "Hookie CLI - Webhook event streaming tool",
	Long:  `Hookie CLI allows you to authenticate, list applications/topics, and stream webhook events in real-time.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&relayURL, "relay-url", "", "Relay service URL (default: localhost:50051 or HOOKIE_RELAY_URL env var)")
	rootCmd.PersistentFlags().StringVar(&orgID, "org-id", "", "Organization ID (can be set globally or per command)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Show detailed information (headers, query params, body, etc.)")
}

