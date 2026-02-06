package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/hookie/cli/internal/config"
	"github.com/hookie/cli/internal/relay"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var topicsCmd = &cobra.Command{
	Use:   "topics [app-id]",
	Short: "List topics for an application or all topics across all applications",
	Long:  `List all topics for a specific application, or all topics across all accessible applications if no app-id is provided.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var appID string
		if len(args) > 0 {
			appID = args[0]
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if cfg.Token == "" {
			return fmt.Errorf("not authenticated. Run 'hookie login' first")
		}

		client, err := relay.NewClient(cfg.Token, debug)
		if err != nil {
			return fmt.Errorf("failed to connect to relay: %w", err)
		}
		defer client.Close()

		topics, err := client.ListTopics(context.Background(), appID)
		if err != nil {
			return fmt.Errorf("failed to list topics: %w", err)
		}

		if len(topics) == 0 {
			if appID != "" {
				fmt.Printf("No topics found for application %s.\n", appID)
			} else {
				fmt.Println("No topics found.")
			}
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		
		// Include APP ID column when listing all topics (appID is empty)
		if appID == "" {
			table.Header("ID", "APP ID", "NAME", "DESCRIPTION")
		} else {
			table.Header("ID", "NAME", "DESCRIPTION")
		}

		for _, topic := range topics {
			desc := topic.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}
			
			if appID == "" {
				table.Append(
					color.CyanString(topic.Id),
					color.YellowString(topic.ApplicationId),
					topic.Name,
					desc,
				)
			} else {
				table.Append(
					color.CyanString(topic.Id),
					topic.Name,
					desc,
				)
			}
		}
		table.Render()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(topicsCmd)
}

