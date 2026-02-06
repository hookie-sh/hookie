package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
	"github.com/hookie/cli/internal/config"
	"github.com/hookie/cli/internal/relay"
	"github.com/spf13/cobra"
)

const repoConfigFileName = "hookie.yml"

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a hookie.yml configuration file",
	Long:  `Initialize a hookie.yml configuration file in the current directory. This will prompt you to select an application and optionally configure forwarding.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if authenticated
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if cfg.Token == "" {
			return fmt.Errorf("not authenticated. Run 'hookie login' first")
		}

		// Check if hookie.yml already exists in current directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}

		configPath := filepath.Join(cwd, repoConfigFileName)
		exists, existingPath := config.RepoConfigExists()
		if exists {
			// Check if it's in the current directory
			if existingPath == configPath {
				var overwrite bool
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewConfirm().
							Title(fmt.Sprintf("%s already exists in this directory", repoConfigFileName)).
							Description("Do you want to overwrite it?").
							Value(&overwrite),
					),
				)

				if err := form.Run(); err != nil {
					return fmt.Errorf("failed to prompt for overwrite: %w", err)
				}

				if !overwrite {
					fmt.Println("Cancelled.")
					return nil
				}
			} else {
				fmt.Printf("Note: Found %s at %s, but will create a new one in the current directory.\n", repoConfigFileName, existingPath)
			}
		}

		// Connect to relay
		client, err := relay.NewClient(cfg.Token)
		if err != nil {
			return fmt.Errorf("failed to connect to relay: %w", err)
		}
		defer client.Close()

		// Fetch applications
		applications, err := client.ListApplications(context.Background(), orgID)
		if err != nil {
			return fmt.Errorf("failed to list applications: %w", err)
		}

		if len(applications) == 0 {
			return fmt.Errorf("no applications found. Please create an application in the web application at https://app.hookie.sh first")
		}

		// Build options list
		var selectedAppID string
		var options []huh.Option[string]

		for _, app := range applications {
			displayName := app.Name
			if displayName == "" {
				displayName = app.Id
			}
			options = append(options, huh.NewOption(
				fmt.Sprintf("%s (%s)", displayName, app.Id),
				app.Id,
			))
		}

		// Create form to select application
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select an application").
					Description("Choose the application to use for this repository").
					Options(options...).
					Value(&selectedAppID),
			),
		)

		if err := form.Run(); err != nil {
			return fmt.Errorf("failed to select application: %w", err)
		}

		if selectedAppID == "" {
			return fmt.Errorf("no application selected")
		}

		// Prompt for optional forward URL
		var forwardURL string
		forwardForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Forward URL (optional)").
					Description("URL to forward events to (e.g., http://localhost:3001/webhooks). Leave empty to skip.").
					Value(&forwardURL),
			),
		)

		if err := forwardForm.Run(); err != nil {
			return fmt.Errorf("failed to prompt for forward URL: %w", err)
		}

		// Validate forward URL if provided
		if forwardURL != "" {
			parsedURL, err := url.Parse(forwardURL)
			if err != nil {
				return fmt.Errorf("invalid forward URL: %w", err)
			}
			if parsedURL.Scheme == "" || parsedURL.Host == "" {
				return fmt.Errorf("invalid forward URL: must include scheme and host (e.g., http://localhost:3001/webhooks)")
			}
		}

		// Create repo config
		repoConfig := &config.RepoConfig{
			AppID:   selectedAppID,
			Forward: forwardURL,
			Topics:  make(map[string]string), // Empty map for per-topic forwarding (users can add manually)
		}

		// Save config file
		if err := config.SaveRepoConfig(repoConfig, configPath); err != nil {
			return fmt.Errorf("failed to save config file: %w", err)
		}

		fmt.Printf("\n%s Created %s\n", color.GreenString("✓"), color.CyanString(configPath))
		fmt.Printf("\nConfiguration:\n")
		fmt.Printf("  App ID: %s\n", color.CyanString(selectedAppID))
		if forwardURL != "" {
			fmt.Printf("  Forward URL: %s\n", color.CyanString(forwardURL))
		}
		fmt.Printf("\nYou can now run %s without specifying flags.\n", color.CyanString("hookie listen"))
		fmt.Printf("To add per-topic forwarding, edit %s and add entries under 'topics'.\n", color.CyanString(repoConfigFileName))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
