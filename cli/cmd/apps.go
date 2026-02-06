package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/fatih/color"
	"github.com/hookie/cli/internal/config"
	"github.com/hookie/cli/internal/relay"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)


var appsCmd = &cobra.Command{
	Use:   "apps",
	Short: "List applications",
	Long:  `List all applications accessible to the authenticated user. Optionally filter by organization.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if cfg.Token == "" {
			return fmt.Errorf("not authenticated. Run 'hookie login' first")
		}

		client, err := relay.NewClient(cfg.Token)
		if err != nil {
			return fmt.Errorf("failed to connect to relay: %w", err)
		}
		defer client.Close()

		// Use flag value if set, otherwise use global orgID
		effectiveOrgID := orgIDFlag
		if effectiveOrgID == "" {
			effectiveOrgID = orgID
		}
		applications, err := client.ListApplications(context.Background(), effectiveOrgID)
		if err != nil {
			return fmt.Errorf("failed to list applications: %w", err)
		}

		if len(applications) == 0 {
			fmt.Println("No applications found.")
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header("ID", "NAME", "TOPICS", "DESCRIPTION")

		for _, app := range applications {
			desc := app.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}
			table.Append(
				color.CyanString(app.Id),
				app.Name,
				fmt.Sprintf("%d", app.TopicCount),
				desc,
			)
		}
		table.Render()

		return nil
	},
}

var appsListenCmd = &cobra.Command{
	Use:   "listen [app-id]",
	Short: "Listen to webhook events for an application",
	Long:  `Listen to webhook events for all topics in a specific application. Optionally forward events to an endpoint URL using --forward flag.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appID := args[0]
		forwardURL, _ := cmd.Flags().GetString("forward")

		// Parse and validate endpoint URL if provided
		var endpointURL *url.URL
		if forwardURL != "" {
			parsedURL, err := url.Parse(forwardURL)
			if err != nil {
				return fmt.Errorf("invalid endpoint URL: %w", err)
			}
			if parsedURL.Scheme == "" || parsedURL.Host == "" {
				return fmt.Errorf("invalid endpoint URL: must include scheme and host (e.g., http://localhost:3001/webhooks)")
			}
			endpointURL = parsedURL
		}

		return runListen("", appID, "", endpointURL)
	},
}

func init() {
	appsCmd.Flags().StringVar(&orgIDFlag, "org-id", "", "Filter by organization ID")
	appsListenCmd.Flags().StringP("forward", "f", "", "Forward events to the specified endpoint URL")
	appsCmd.AddCommand(appsListenCmd)
	rootCmd.AddCommand(appsCmd)
}

