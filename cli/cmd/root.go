package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"runtime"
		"strings"

	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
	"github.com/hookie/cli/internal/config"
	"github.com/hookie/cli/internal/gui"
	"github.com/hookie/cli/internal/relay"
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
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Display debug information first if debug flag is set
		if debug {
			// Get command name by parsing os.Args directly
			// This is necessary because PersistentPreRun runs after flag parsing,
			// so args may not contain the subcommand name
			commandName := getCommandNameFromArgs()
			// Build command string without the full path - just use "hookie" as the command name
			commandParts := []string{"hookie"}
			if len(os.Args) > 1 {
				commandParts = append(commandParts, os.Args[1:]...)
			}
			fullCommand := strings.Join(commandParts, " ")
			printDebugInfo(commandName, orgID, fullCommand)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, color.RedString("Error: %v", err))
		os.Exit(1)
	}
}

var listenCmd = &cobra.Command{
	Use:   "listen [--forward <url>]",
	Short: "Listen for webhook events (anonymous or authenticated)",
	Long:  `Listen for webhook events. If unauthenticated, creates an anonymous ephemeral channel. If authenticated without flags, prompts for app or topic selection. Use --app-id to subscribe to all topics of an app, or --topic-id to subscribe to a specific topic. Optionally forward events to an endpoint URL using --forward flag.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		forwardURL, _ := cmd.Flags().GetString("forward")
		forwardExplicit := cmd.Flags().Changed("forward")
		showUI, _ := cmd.Flags().GetBool("ui")
		showGUI := !forwardExplicit || showUI
		topicID, _ := cmd.Flags().GetString("topic-id")
		appID, _ := cmd.Flags().GetString("app-id")

		// Load repository config (if exists)
		repoConfig, _, err := config.LoadRepoConfig()
		if err != nil {
			return fmt.Errorf("failed to load repository config: %w", err)
		}

		// Store original CLI flag values to check precedence
		cliAppID := appID
		cliTopicID := topicID

		// Priority: CLI flags > repo config > interactive selector
		// Use repo config values only if:
		// 1. The CLI flag for that field is empty, AND
		// 2. The conflicting CLI flag is also empty (to prevent mutual exclusion)
		if cliAppID == "" && cliTopicID == "" && repoConfig != nil && repoConfig.AppID != "" {
			appID = repoConfig.AppID
		}
		if cliTopicID == "" && cliAppID == "" && repoConfig != nil && repoConfig.TopicID != "" {
			topicID = repoConfig.TopicID
		}
		if forwardURL == "" && repoConfig != nil && repoConfig.Forward != "" {
			forwardURL = repoConfig.Forward
		}

		// Build topic forward map from repo config
		var topicForwardMap map[string]*url.URL
		if repoConfig != nil && repoConfig.Topics != nil && len(repoConfig.Topics) > 0 {
			topicForwardMap = make(map[string]*url.URL)
			for topicID, topicURL := range repoConfig.Topics {
				if topicURL != "" {
					parsedURL, err := url.Parse(topicURL)
					if err != nil {
						return fmt.Errorf("invalid forward URL for topic %s: %w", topicID, err)
					}
					if parsedURL.Scheme == "" || parsedURL.Host == "" {
						return fmt.Errorf("invalid forward URL for topic %s: must include scheme and host", topicID)
					}
					topicForwardMap[topicID] = parsedURL
				}
			}
		}

		// Validate flags - topic-id and app-id are mutually exclusive
		if topicID != "" && appID != "" {
			return fmt.Errorf("cannot specify both --topic-id and --app-id flags")
		}

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

		// Start GUI when: no --forward, or --forward + --ui
		var guiURL *url.URL
		if showGUI {
			port := gui.DefaultPort()
			var started bool
			var err error
			guiURL, started, err = gui.AcquireOrUseServer(port)
			if err != nil {
				return fmt.Errorf("GUI: %w", err)
			}
			if started {
				fmt.Println(color.CyanString("GUI available at %s/", guiURL.String()))
				openBrowserTo(guiURL.String())
			} else {
				fmt.Println(color.CyanString("Using existing GUI at %s/", guiURL.String()))
			}
		}

		// Check if authenticated
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if cfg.Token == "" {
			// Anonymous mode
			return runAnonymousListen(endpointURL, guiURL)
		}

		// Authenticated mode
		client, err := relay.NewClient(cfg.Token, debug)
		if err != nil {
			return fmt.Errorf("failed to connect to relay: %w", err)
		}
		defer client.Close()

		// Use persistent orgID flag
		effectiveOrgID := orgID

		// If flags are provided, subscribe directly
		if topicID != "" {
			return runListen(topicID, "", effectiveOrgID, endpointURL, topicForwardMap, guiURL)
		}
		if appID != "" {
			return runListen("", appID, effectiveOrgID, endpointURL, topicForwardMap, guiURL)
		}

		// No flags provided - show interactive selector with apps and topics
		// Fetch applications
		applications, err := client.ListApplications(context.Background(), effectiveOrgID)
		if err != nil {
			return fmt.Errorf("failed to list applications: %w", err)
		}

		// Fetch all topics
		topics, err := client.ListTopics(context.Background(), "")
		if err != nil {
			return fmt.Errorf("failed to list topics: %w", err)
		}

		// Handle empty case
		if len(applications) == 0 && len(topics) == 0 {
			return fmt.Errorf("no applications or topics found. Please create an application or topic in the web application at https://app.hookie.sh first")
		}

		// Build unified options list
		var selectedValue string
		var options []huh.Option[string]

		// Add applications
		for _, app := range applications {
			displayName := app.Name
			if displayName == "" {
				displayName = app.Id
			}
			options = append(options, huh.NewOption(
				fmt.Sprintf("App: %s (%s)", displayName, app.Id),
				fmt.Sprintf("app:%s", app.Id),
			))
		}

		// Add topics
		for _, topic := range topics {
			displayName := topic.Name
			if displayName == "" {
				displayName = topic.Id
			}
			// Include app ID if available for better context
			if topic.ApplicationId != "" {
				displayName = fmt.Sprintf("%s (%s)", displayName, topic.ApplicationId)
			}
			options = append(options, huh.NewOption(
				fmt.Sprintf("Topic: %s (%s)", displayName, topic.Id),
				fmt.Sprintf("topic:%s", topic.Id),
			))
		}

		// Create and run the form
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select an app or topic to listen to").
					Description("Choose an application to listen to all topics, or a specific topic").
					Options(options...).
					Value(&selectedValue),
			),
		)

		if err := form.Run(); err != nil {
			return fmt.Errorf("failed to select app or topic: %w", err)
		}

		if selectedValue == "" {
			return fmt.Errorf("no app or topic selected")
		}

		// Parse the selected value
		if len(selectedValue) > 4 && selectedValue[:4] == "app:" {
			// App selected
			appID := selectedValue[4:]
			return runListen("", appID, effectiveOrgID, endpointURL, topicForwardMap, guiURL)
		} else if len(selectedValue) > 6 && selectedValue[:6] == "topic:" {
			// Topic selected
			topicID := selectedValue[6:]
			return runListen(topicID, "", effectiveOrgID, endpointURL, topicForwardMap, guiURL)
		}

		return fmt.Errorf("invalid selection format")
	},
}

// getCommandNameFromArgs extracts the subcommand name from os.Args
// It skips the program name and any flags to find the actual command
func getCommandNameFromArgs() string {
	if len(os.Args) < 2 {
		return "root"
	}

	// Skip program name (os.Args[0])
	// Look for the first non-flag argument
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		// Skip flags (starting with -)
		if strings.HasPrefix(arg, "-") {
			// Skip flag value if it's not a flag itself
			if strings.Contains(arg, "=") {
				continue
			}
			// Check if next arg is a value (not a flag)
			if i+1 < len(os.Args) && !strings.HasPrefix(os.Args[i+1], "-") {
				i++ // Skip the flag value
			}
			continue
		}
		// Found a non-flag argument - this should be the command name
		return arg
	}

	return "root"
}

func openBrowserTo(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return
	}
	_ = cmd.Start()
}

func init() {
	rootCmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		return fmt.Errorf("%w\n\n%s", err, c.UsageString())
	})
	rootCmd.PersistentFlags().StringVar(&orgID, "org-id", "", "Organization ID (can be set globally or per command)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Show detailed information (headers, query params, body, etc.)")
	listenCmd.Flags().StringP("forward", "f", "", "Forward events to the specified endpoint URL")
	listenCmd.Flags().Bool("ui", false, "Show local UI when forwarding with --forward")
	listenCmd.Flags().StringP("topic-id", "t", "", "Subscribe to a specific topic")
	listenCmd.Flags().StringP("app-id", "a", "", "Subscribe to all topics of an application")
	rootCmd.AddCommand(listenCmd)
}

