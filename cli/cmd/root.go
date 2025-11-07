package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	orgID     string
	orgIDFlag string
	debug     bool
)

var rootCmd = &cobra.Command{
	Use:   "hookie",
	Short: "Hookie CLI - Webhook event streaming tool",
	Long:  `Hookie CLI allows you to authenticate, list applications/topics, and stream webhook events in real-time.`,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, color.RedString("Error: %v", err))
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		return fmt.Errorf("%w\n\n%s", err, c.UsageString())
	})
	rootCmd.PersistentFlags().StringVar(&orgID, "org-id", "", "Organization ID (can be set globally or per command)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Show detailed information (headers, query params, body, etc.)")
}

