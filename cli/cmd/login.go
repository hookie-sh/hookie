package cmd

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

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

		// Get publishable key (compiled into binary, with optional env override)
		publishableKey := auth.GetPublishableKey()
		if publishableKey == "" {
			return fmt.Errorf("clerk publishable key not configured. please set publishablekey in oauth_config.go and rebuild")
		}

		// Get web app URL
		webAppURL := auth.GetWebAppURL()
		if webAppURL == "" {
			return fmt.Errorf("web app URL not configured. please set WebAppURL in oauth_config.go and rebuild")
		}

		// Step 1: Find an available port for the callback server
		preferredPorts := []int{48443, 48444, 48445, 48446, 48447}
		var port int
		var listener net.Listener

		// Try preferred ports first, then fall back to any available port
		for _, p := range preferredPorts {
			l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", p))
			if err == nil {
				port = p
				listener = l
				break
			}
		}

		// If all preferred ports are busy, find any available port
		if listener == nil {
			l, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				return fmt.Errorf("failed to find available port: %w", err)
			}
			port = l.Addr().(*net.TCPAddr).Port
			listener = l
		}
		listener.Close() // Close immediately, we'll create a new server later

		// Step 2: Build redirect URL
		redirectURL := fmt.Sprintf("http://localhost:%d/callback", port)

		// Step 3: Start local server to receive token
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("Authorization required")
		fmt.Println(strings.Repeat("=", 60))
		fmt.Printf("\nStarting local server on port %d...\n", port)
		fmt.Printf("Opening browser to complete authentication...\n\n")

		// Start the callback server in a goroutine
		signInTokenChan := make(chan string, 1)
		errorChan := make(chan error, 1)

		go func() {
			token, err := auth.ReceiveSignInToken(ctx, port)
			if err != nil {
				errorChan <- err
			} else {
				signInTokenChan <- token
			}
		}()

		// Give server a brief moment to start listening
		time.Sleep(200 * time.Millisecond)

		// Step 4: Build authorization URL and open browser
		authURL, err := url.Parse(webAppURL)
		if err != nil {
			return fmt.Errorf("invalid web app URL: %w", err)
		}
		authURL.Path = "/cli"
		authURL.RawQuery = url.Values{
			"redirect_url": []string{redirectURL},
		}.Encode()

		authorizationURL := authURL.String()

		// Open browser automatically
		if err := auth.OpenBrowser(authorizationURL); err != nil {
			fmt.Printf("Warning: Failed to open browser automatically: %v\n", err)
			fmt.Printf("Please visit the URL manually: %s\n", authorizationURL)
		} else {
			fmt.Println("Opening browser...")
		}

		// Step 5: Wait for sign-in token callback
		var signInToken string
		select {
		case signInToken = <-signInTokenChan:
			// Successfully received token
		case err := <-errorChan:
			return fmt.Errorf("failed to receive sign-in token: %w", err)
		case <-time.After(5 * time.Minute):
			return fmt.Errorf("authorization timeout")
		}

		// Step 6: Complete sign-in using the sign-in token
		fmt.Println("Completing sign-in...")
		sessionToken, err := auth.CompleteSignInWithTicket(ctx, signInToken)
		if err != nil {
			return fmt.Errorf("failed to complete sign-in: %w", err)
		}

		// Step 7: Verify session token and extract user ID
		userID, err := auth.VerifyToken(ctx, sessionToken)
		if err != nil {
			return fmt.Errorf("failed to verify session token: %w", err)
		}

		// Step 8: Save session token to config
		cfg := &config.Config{
			Token:  sessionToken,
			UserID: userID,
		}

		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println("✓ Authentication successful!")
		fmt.Printf("✓ Successfully authenticated as user %s\n", userID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

