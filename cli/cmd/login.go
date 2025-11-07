package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/hookie/cli/internal/auth"
	"github.com/hookie/cli/internal/config"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Clerk",
	Long:  `Authenticate with Clerk by opening a browser and completing the login flow.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Get OAuth configuration (compiled into binary, with optional env override)
		oauthConfig := auth.GetOAuthConfig()

		// Validate configuration
		if oauthConfig.ClientID == "" {
			return fmt.Errorf("OAuth client ID not configured. Please set CLERK_OAUTH_CLIENT_ID in oauth_config.go and rebuild")
		}
		if oauthConfig.AuthorizeURL == "" {
			return fmt.Errorf("OAuth authorize URL not configured. Please set CLERK_OAUTH_AUTHORIZE_URL in oauth_config.go and rebuild")
		}
		if oauthConfig.TokenURL == "" {
			return fmt.Errorf("OAuth token URL not configured. Please set CLERK_OAUTH_TOKEN_URL in oauth_config.go and rebuild")
		}

		// Get publishable key (compiled into binary, with optional env override)
		publishableKey := auth.GetPublishableKey()
		if publishableKey == "" {
			return fmt.Errorf("clerk publishable key not configured. please set publishablekey in oauth_config.go and rebuild")
		}

		// Check for discovery URL override (for dynamic endpoint discovery)
		if discoveryURL := os.Getenv("CLERK_OAUTH_DISCOVERY_URL"); discoveryURL != "" {
			authorizeURL, tokenURL, userInfoURL, err := auth.FetchOAuthEndpoints(discoveryURL)
			if err != nil {
				return fmt.Errorf("failed to fetch OAuth endpoints from discovery URL: %w", err)
			}
			oauthConfig.AuthorizeURL = authorizeURL
			oauthConfig.TokenURL = tokenURL
			oauthConfig.UserInfoURL = userInfoURL
		}

		fmt.Println("Starting authorization code flow with PKCE...")
		accessToken, idToken, userID, err := auth.StartLoginFlow(ctx, oauthConfig)
		if err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		// Exchange access token for a session token (JWT)
		// Prefers ID token if available, otherwise validates access token is a JWT
		sessionToken, err := auth.ExchangeAccessTokenForSessionToken(ctx, accessToken, idToken)
		if err != nil {
			return fmt.Errorf("failed to get session token: %w", err)
		}

		// Save session token to config
		cfg := &config.Config{
			Token:    sessionToken,
			UserID:   userID,
		}

		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("✓ Successfully authenticated as user %s\n", userID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

