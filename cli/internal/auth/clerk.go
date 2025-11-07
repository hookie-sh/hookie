package auth

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"crypto/tls"
	"crypto/x509"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwks"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
)

// init initializes TLS configuration to use system certificates
// This ensures the Clerk SDK's HTTP client can verify TLS certificates
func init() {
	// Load system certificate pool
	systemCerts, err := x509.SystemCertPool()
	if err != nil {
		// Fallback: try to create a new pool and add common macOS cert locations
		systemCerts = x509.NewCertPool()
		
		// On macOS, try to load certificates from common locations
		if runtime.GOOS == "darwin" {
			// Try common macOS certificate locations
			certPaths := []string{
				"/etc/ssl/cert.pem",
				"/usr/local/etc/openssl/cert.pem",
			}
			for _, path := range certPaths {
				if certs, err := os.ReadFile(path); err == nil {
					systemCerts.AppendCertsFromPEM(certs)
				}
			}
		}
	}

	// Configure default HTTP transport with system certificates
	// This should be picked up by the Clerk SDK's HTTP client
	if transport, ok := http.DefaultTransport.(*http.Transport); ok {
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{
				RootCAs: systemCerts,
			}
		} else {
			transport.TLSClientConfig.RootCAs = systemCerts
		}
	} else {
		// If DefaultTransport is not a *http.Transport, replace it
		http.DefaultTransport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: systemCerts,
			},
		}
	}
}

// DeviceAuthorizationResponse represents the response from the device authorization endpoint
type DeviceAuthorizationResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"` // Polling interval in seconds
}

// TokenResponse represents the response from the token endpoint
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	IDToken      string `json:"id_token,omitempty"` // JWT ID token if available
}

// UserInfoResponse represents the response from the userinfo endpoint
type UserInfoResponse struct {
	Sub           string `json:"sub"`           // User ID (subject)
	Email         string `json:"email,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`
	Name          string `json:"name,omitempty"`
	GivenName     string `json:"given_name,omitempty"`
	FamilyName    string `json:"family_name,omitempty"`
	Picture       string `json:"picture,omitempty"`
}

// TokenErrorResponse represents an error response from the token endpoint
type TokenErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// generatePKCE generates a code verifier and code challenge for PKCE
func generatePKCE() (codeVerifier, codeChallenge string, err error) {
	// Generate a cryptographically random code verifier (43-128 characters)
	// Using 32 bytes = 256 bits, base64url encoded = 43 characters
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	codeVerifier = base64.RawURLEncoding.EncodeToString(randomBytes)

	// Generate code challenge using SHA256
	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge = base64.RawURLEncoding.EncodeToString(hash[:])

	return codeVerifier, codeChallenge, nil
}

// formatUserCode formats the user code for display (typically 8 characters, hyphenated)
func formatUserCode(code string) string {
	if len(code) >= 8 {
		return code[:4] + "-" + code[4:8]
	}
	return code
}

// OAuthConfig holds OAuth endpoint configuration
type OAuthConfig struct {
	AuthorizeURL string
	TokenURL     string
	UserInfoURL  string // UserInfo endpoint to get user details from access token
	ClientID     string
	RedirectURI  string // Must match what's configured in Clerk OAuth app
}

// DiscoveryResponse represents OAuth discovery document response
type DiscoveryResponse struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserInfoEndpoint      string `json:"userinfo_endpoint"`
}

// StartLoginFlow implements OAuth 2.0 Authorization Code Flow with PKCE
// Uses Clerk OAuth app endpoints
// Returns access token, ID token, and user ID
func StartLoginFlow(ctx context.Context, config OAuthConfig) (string, string, string, error) {
	// Generate PKCE parameters
	codeVerifier, codeChallenge, err := generatePKCE()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate PKCE: %w", err)
	}

	// Generate state for CSRF protection
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", "", "", fmt.Errorf("failed to generate state: %w", err)
	}
	state := base64.RawURLEncoding.EncodeToString(stateBytes)

	// Step 1: Find an available port for the callback server
	// Try common ports that are less likely to be in use
	preferredPorts := []int{48443, 48444, 48445, 48446, 48447}
	var port int
	var listener net.Listener

	// Try preferred ports first, then fall back to any available port
	for _, p := range preferredPorts {
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", p))
		if err == nil {
			port = p
			listener = l
			break
		}
	}

	// If all preferred ports are busy, find any available port
	if listener == nil {
		l, err := net.Listen("tcp", ":0")
		if err != nil {
			return "", "", "", fmt.Errorf("failed to find available port: %w", err)
		}
		port = l.Addr().(*net.TCPAddr).Port
		listener = l
	}
	listener.Close() // Close immediately, we'll create a new server later

	// Step 2: Build redirect URI with the available port
	redirectURI := config.RedirectURI
	if redirectURI == "" {
		// Use the found port
		redirectURI = fmt.Sprintf("http://localhost:%d/callback", port)
	} else {
		// If redirect URI is specified, try to extract port or use the found port
		u, err := url.Parse(redirectURI)
		if err == nil && u.Port() == "" {
			// No port specified, use the dynamically found port
			u.Host = fmt.Sprintf("localhost:%d", port)
			redirectURI = u.String()
		}
	}

	// Step 3: Build authorization URL with PKCE
	authURL, err := url.Parse(config.AuthorizeURL)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid authorize URL: %w", err)
	}

	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", config.ClientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("scope", "openid profile email")
	params.Set("code_challenge", codeChallenge)
	params.Set("code_challenge_method", "S256")
	params.Set("state", state)

	authURL.RawQuery = params.Encode()
	authorizationURL := authURL.String()

	// Step 4: Display instructions to user
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Authorization required")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("\nTo complete authentication, please visit:\n\n")
	fmt.Printf("  %s\n\n", authorizationURL)
	fmt.Printf("After authorizing, you will be redirected to:\n")
	fmt.Printf("  %s\n\n", redirectURI)
	fmt.Printf("Press Enter to open the browser, or visit the URL above manually...\n")
	fmt.Println(strings.Repeat("=", 60))

	// Step 5: Start local server to capture callback
	// Start server in background before waiting for user input
	codeChan := make(chan string, 1)
	errorChan := make(chan error, 1)
	
	// Start the callback server in a goroutine
	// captureAuthorizationCode starts its own server internally, so we call it async
	go func() {
		code, err := captureAuthorizationCode(ctx, redirectURI, port)
		if err != nil {
			errorChan <- err
		} else {
			codeChan <- code
		}
	}()

	// Give server a brief moment to start listening
	time.Sleep(200 * time.Millisecond)
	fmt.Printf("Listening on port %d for callback...\n", port)

	// Wait for user to press Enter, then open browser
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
	
	// Open browser automatically
	if err := openBrowser(authorizationURL); err != nil {
		fmt.Printf("Warning: Failed to open browser automatically: %v\n", err)
		fmt.Printf("Please visit the URL manually: %s\n", authorizationURL)
	} else {
		fmt.Println("Opening browser...")
	}

	// Wait for callback
	var code string
	select {
	case code = <-codeChan:
		// Successfully received code
	case <-errorChan:
		// Fallback: manual code entry
		fmt.Println("\nIf automatic capture failed, please enter the authorization code manually:")
		fmt.Print("Code: ")
		var manualCode string
		fmt.Scanln(&manualCode)
		code = manualCode
	case <-time.After(5 * time.Minute):
		return "", "", "", fmt.Errorf("authorization timeout")
	}

	// Step 4: Exchange authorization code for tokens
	tokenResp, err := exchangeAuthorizationCode(ctx, config.TokenURL, code, codeVerifier, config.ClientID, redirectURI)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	// Step 5: Get user ID from userinfo endpoint or ID token
	var userID string
	if tokenResp.IDToken != "" {
		// Try to extract user ID from ID token (JWT)
		publishableKey := GetPublishableKey()
		if publishableKey != "" {
			userID, err = verifyTokenJWT(ctx, tokenResp.IDToken, publishableKey)
			if err == nil {
				// Successfully extracted from ID token
			}
		}
	}

	// If ID token didn't work or wasn't available, use userinfo endpoint
	if userID == "" && config.UserInfoURL != "" {
		userID, err = getUserIDFromUserInfo(ctx, config.UserInfoURL, tokenResp.AccessToken)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to get user ID from userinfo: %w", err)
		}
	}

	if userID == "" {
		return "", "", "", fmt.Errorf("failed to extract user ID from token response")
	}

	fmt.Println("✓ Authentication successful!")
	return tokenResp.AccessToken, tokenResp.IDToken, userID, nil
}

// captureAuthorizationCode starts a local HTTP server to capture the OAuth callback
func captureAuthorizationCode(ctx context.Context, redirectURI string, port int) (string, error) {
	// Parse redirect URI to get callback path
	u, err := url.Parse(redirectURI)
	if err != nil {
		return "", fmt.Errorf("invalid redirect URI: %w", err)
	}

	server := &http.Server{
		Addr: ":" + strconv.Itoa(port),
	}

	codeChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	mux := http.NewServeMux()
	callbackPath := u.Path
	if callbackPath == "" {
		callbackPath = "/callback"
	}
	mux.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		// Check for errors first
		if errParam := r.URL.Query().Get("error"); errParam != "" {
			errDesc := r.URL.Query().Get("error_description")
			http.Error(w, fmt.Sprintf("OAuth error: %s - %s", errParam, errDesc), http.StatusBadRequest)
			errorChan <- fmt.Errorf("oauth error: %s - %s", errParam, errDesc)
			return
		}

		// Get authorization code
		code := r.URL.Query().Get("code")
		if code == "" {
			// Log the full URL for debugging
			fmt.Printf("\n[DEBUG] Callback received but no code parameter. Full URL: %s\n", r.URL.String())
			fmt.Printf("[DEBUG] Query parameters: %v\n", r.URL.Query())
			
			http.Error(w, "Missing authorization code", http.StatusBadRequest)
			errorChan <- fmt.Errorf("missing authorization code in callback URL: %s", r.URL.String())
			return
		}

		// Verify state parameter (CSRF protection)
		// Note: state validation should be added here if we stored it

		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
			<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<title>Authentication Successful</title>
				<style>
					* {
						margin: 0;
						padding: 0;
						box-sizing: border-box;
					}
					
					:root {
						--primary: #ff6b3d;
						--primary-dark: #e85a2e;
						--background: #fafafa;
						--foreground: #2d2d2d;
						--card: #ffffff;
						--border: #e5e5e5;
						--muted: #f5f5f5;
						--success: #10b981;
					}
					
					@media (prefers-color-scheme: dark) {
						:root {
							--primary: #ff7a4d;
							--primary-dark: #ff6b3d;
							--background: #2d2d2d;
							--foreground: #fafafa;
							--card: #363636;
							--border: #4a4a4a;
							--muted: #2d2d2d;
							--success: #34d399;
						}
					}
					
					body {
						font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
						background: var(--background);
						color: var(--foreground);
						display: flex;
						align-items: center;
						justify-content: center;
						min-height: 100vh;
						padding: 20px;
						line-height: 1.6;
					}
					
					.container {
						background: var(--card);
						border-radius: 16px;
						padding: 48px 32px;
						box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
						max-width: 480px;
						width: 100%;
						text-align: center;
						border: 1px solid var(--border);
					}
					
					.checkmark {
						width: 80px;
						height: 80px;
						border-radius: 50%;
						background: linear-gradient(135deg, var(--success), #059669);
						display: flex;
						align-items: center;
						justify-content: center;
						margin: 0 auto 24px;
						animation: scaleIn 0.5s ease-out;
					}
					
					.checkmark svg {
						width: 48px;
						height: 48px;
						color: white;
					}
					
					@keyframes scaleIn {
						from {
							transform: scale(0);
							opacity: 0;
						}
						to {
							transform: scale(1);
							opacity: 1;
						}
					}
					
					h1 {
						font-size: 28px;
						font-weight: 600;
						color: var(--foreground);
						margin-bottom: 12px;
						letter-spacing: -0.5px;
					}
					
					p {
						font-size: 16px;
						color: var(--foreground);
						opacity: 0.7;
						margin-bottom: 32px;
					}
					
					.button {
						background: var(--primary);
						color: white;
						border: none;
						padding: 12px 32px;
						font-size: 16px;
						font-weight: 500;
						border-radius: 8px;
						cursor: pointer;
						transition: all 0.2s ease;
						display: inline-block;
						text-decoration: none;
						font-family: inherit;
					}
					
					.button:hover {
						background: var(--primary-dark);
						transform: translateY(-1px);
						box-shadow: 0 4px 12px rgba(255, 107, 61, 0.3);
					}
					
					.button:active {
						transform: translateY(0);
					}
					
					.button:focus {
						outline: 2px solid var(--primary);
						outline-offset: 2px;
					}
				</style>
			</head>
			<body>
				<div class="container">
					<div class="checkmark">
						<svg fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7"></path>
						</svg>
					</div>
					<h1>Authentication Successful!</h1>
					<p>You can close this window and return to the terminal.</p>
					<button class="button" onclick="window.close()">Close Window</button>
				</div>
			</body>
			</html>
		`))

		codeChan <- code
	})

	server.Handler = mux

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errorChan <- err
		}
	}()

	// Wait for callback or timeout
	select {
	case code := <-codeChan:
		server.Shutdown(ctx)
		return code, nil
	case err := <-errorChan:
		server.Shutdown(ctx)
		return "", err
	case <-ctx.Done():
		server.Shutdown(ctx)
		return "", ctx.Err()
	case <-time.After(5 * time.Minute):
		server.Shutdown(ctx)
		return "", fmt.Errorf("authorization timeout")
	}
}

// getUserIDFromUserInfo fetches user information from the userinfo endpoint
func getUserIDFromUserInfo(ctx context.Context, userInfoURL, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("userinfo request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var userInfo UserInfoResponse
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return "", fmt.Errorf("failed to decode userinfo response: %w", err)
	}

	if userInfo.Sub == "" {
		return "", fmt.Errorf("userinfo response missing user ID (sub)")
	}

	return userInfo.Sub, nil
}

// exchangeAuthorizationCode exchanges an authorization code for an access token using PKCE
func exchangeAuthorizationCode(ctx context.Context, tokenURL, authCode, codeVerifier, clientID, redirectURI string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", authCode)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", clientID)
	data.Set("code_verifier", codeVerifier)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var tokenErr TokenErrorResponse
		if err := json.Unmarshal(body, &tokenErr); err == nil {
			return nil, fmt.Errorf("%s: %s", tokenErr.Error, tokenErr.ErrorDescription)
		}
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}

// FetchOAuthEndpoints fetches OAuth endpoints from a discovery document
func FetchOAuthEndpoints(discoveryURL string) (authorizeURL, tokenURL, userInfoURL string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", discoveryURL, nil)
	if err != nil {
		return "", "", "", err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("discovery document request failed with status %d", resp.StatusCode)
	}

	var discovery DiscoveryResponse
	if err := json.NewDecoder(resp.Body).Decode(&discovery); err != nil {
		return "", "", "", fmt.Errorf("failed to decode discovery document: %w", err)
	}

	if discovery.AuthorizationEndpoint == "" || discovery.TokenEndpoint == "" {
		return "", "", "", fmt.Errorf("discovery document missing required endpoints")
	}

	return discovery.AuthorizationEndpoint, discovery.TokenEndpoint, discovery.UserInfoEndpoint, nil
}

// getClerkInstance gets the Clerk instance domain from environment variables
// Clerk instance format: <instance-name>.clerk.accounts.dev
func getClerkInstance() string {
	// Try CLERK_INSTANCE first (most explicit)
	if instance := os.Getenv("CLERK_INSTANCE"); instance != "" {
		// Remove https:// if present
		instance = strings.TrimPrefix(instance, "https://")
		instance = strings.TrimPrefix(instance, "http://")
		return instance
	}

	// Try CLERK_FRONTEND_API (alternative naming)
	if instance := os.Getenv("CLERK_FRONTEND_API"); instance != "" {
		instance = strings.TrimPrefix(instance, "https://")
		instance = strings.TrimPrefix(instance, "http://")
		return instance
	}

	// Try CLERK_DOMAIN (another common name)
	if instance := os.Getenv("CLERK_DOMAIN"); instance != "" {
		instance = strings.TrimPrefix(instance, "https://")
		instance = strings.TrimPrefix(instance, "http://")
		return instance
	}

	return ""
}

// requestDeviceAuthorization requests device authorization from Clerk
func requestDeviceAuthorization(ctx context.Context, authURL, clientID, codeChallenge string) (deviceCode, userCode, verificationURI string, err error) {
	// Note: Clerk may not have a native device flow endpoint
	// This is a placeholder implementation following OAuth 2.0 Device Flow spec
	// You may need to adjust the endpoint URL based on Clerk's actual API

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("scope", "openid")
	data.Set("code_challenge", codeChallenge)
	data.Set("code_challenge_method", "S256")

	req, err := http.NewRequestWithContext(ctx, "POST", authURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", "", "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		// If Clerk doesn't support device flow, fall back to a simulated approach
		// Generate our own device code and user code
		return simulateDeviceAuthorization(clientID, codeChallenge)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Fallback to simulated device flow if endpoint doesn't exist
		return simulateDeviceAuthorization(clientID, codeChallenge)
	}

	var deviceAuthResp DeviceAuthorizationResponse
	if err := json.NewDecoder(resp.Body).Decode(&deviceAuthResp); err != nil {
		return "", "", "", fmt.Errorf("failed to decode device authorization response: %w", err)
	}

	return deviceAuthResp.DeviceCode, deviceAuthResp.UserCode, deviceAuthResp.VerificationURI, nil
}

// simulateDeviceAuthorization creates a simulated device flow when Clerk doesn't support it natively
// This creates a user code and verification URL that users can visit
// NOTE: This is a fallback - if Clerk doesn't support Device Flow, you'll need a backend service
func simulateDeviceAuthorization(clientID, codeChallenge string) (deviceCode, userCode, verificationURI string, err error) {
	// Generate a random device code (store this securely in a real implementation)
	deviceBytes := make([]byte, 32)
	if _, err := rand.Read(deviceBytes); err != nil {
		return "", "", "", err
	}
	deviceCode = base64.RawURLEncoding.EncodeToString(deviceBytes)

	// Generate a user-friendly code (8 characters, alphanumeric)
	userBytes := make([]byte, 4)
	if _, err := rand.Read(userBytes); err != nil {
		return "", "", "", err
	}
	userCode = base64.RawURLEncoding.EncodeToString(userBytes)[:8]

	// Get Clerk instance to construct proper verification URI
	clerkInstance := getClerkInstance()
	if clerkInstance == "" {
		return "", "", "", fmt.Errorf("CLERK_INSTANCE required for device flow")
	}

	// Construct verification URI pointing to Clerk's device verification page
	// Format: https://<instance>/device?user_code=XXXX-XXXX
	verificationURI = fmt.Sprintf("https://%s/device?user_code=%s", clerkInstance, formatUserCode(userCode))

	return deviceCode, userCode, verificationURI, nil
}

// pollForToken polls the token endpoint until authorization is complete
func pollForToken(ctx context.Context, tokenURL, deviceCode, codeVerifier, clientID string) (string, error) {
	// Default polling interval (will be overridden by server response if provided)
	interval := 5 * time.Second
	maxAttempts := 120 // 10 minutes max wait time

	for attempt := 0; attempt < maxAttempts; attempt++ {
		token, err := exchangeDeviceCode(ctx, tokenURL, deviceCode, codeVerifier, clientID)
		if err == nil {
			return token, nil
		}

		// Check if it's an "authorization_pending" error (keep polling)
		if strings.Contains(err.Error(), "authorization_pending") {
			time.Sleep(interval)
			continue
		}

		// Check if it's an "slow_down" error (increase interval)
		if strings.Contains(err.Error(), "slow_down") {
			interval += 5 * time.Second
			time.Sleep(interval)
			continue
		}

		// Other errors are fatal
		return "", err
	}

	return "", fmt.Errorf("authorization timeout: user did not complete authorization in time")
}

// exchangeDeviceCode exchanges the device code for an access token
func exchangeDeviceCode(ctx context.Context, tokenURL, deviceCode, codeVerifier, clientID string) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	data.Set("device_code", deviceCode)
	data.Set("client_id", clientID)
	data.Set("code_verifier", codeVerifier)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		var tokenErr TokenErrorResponse
		if err := json.Unmarshal(body, &tokenErr); err == nil {
			return "", fmt.Errorf("%s: %s", tokenErr.Error, tokenErr.ErrorDescription)
		}
		return "", fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	return tokenResp.AccessToken, nil
}

// ValidateJWTFormat validates that a token is in JWT format (three parts separated by dots)
func ValidateJWTFormat(token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("token is empty")
	}
	
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid JWT format: expected 3 parts separated by dots, got %d parts", len(parts))
	}
	
	// Check that each part is non-empty
	for i, part := range parts {
		if part == "" {
			return fmt.Errorf("invalid JWT format: part %d is empty", i+1)
		}
	}
	
	return nil
}

// ExchangeAccessTokenForSessionToken exchanges an OAuth access token for a Clerk session token (JWT)
// Prefers ID token if available (it's a JWT), otherwise validates that access token is a JWT
func ExchangeAccessTokenForSessionToken(ctx context.Context, accessToken, idToken string) (string, error) {
	// First, try to use ID token if available (it's already a JWT)
	if idToken != "" {
		if err := ValidateJWTFormat(idToken); err == nil {
			// ID token is a valid JWT, use it
			return strings.TrimSpace(idToken), nil
		}
	}
	
	// Fallback: validate that access token is a JWT
	accessToken = strings.TrimSpace(accessToken)
	if err := ValidateJWTFormat(accessToken); err != nil {
		return "", fmt.Errorf("access token is not a valid JWT (expected 3 parts separated by dots): %w", err)
	}
	
	// Access token appears to be a JWT, use it
	return accessToken, nil
}

// VerifyToken verifies a Clerk token and returns the user ID
// Uses JWT verification with JWKS (no secret key required)
func VerifyToken(ctx context.Context, token string) (string, error) {
	publishableKey := GetPublishableKey()
	if publishableKey == "" {
		return "", fmt.Errorf("publishable key not configured")
	}
	return verifyTokenJWT(ctx, token, publishableKey)
}

// verifyTokenJWT verifies a Clerk JWT session token using networkless verification
// This method only requires the publishable key, making it safe for CLI use
func verifyTokenJWT(ctx context.Context, token, publishableKey string) (string, error) {
	// Extract token from "Bearer <token>" format if present
	token = strings.TrimPrefix(token, "Bearer ")
	token = strings.TrimSpace(token)

	// Step 1: Decode the session JWT to find the key ID
	unsafeClaims, err := jwt.Decode(ctx, &jwt.DecodeParams{
		Token: token,
	})
	if err != nil {
		return "", fmt.Errorf("failed to decode token: %w", err)
	}

	// Step 2: Create JWKS client with publishable key configured
	// The SDK needs the key to determine the correct endpoint
	config := &clerk.ClientConfig{}
	if publishableKey != "" {
		// Set the publishable key so SDK can determine the correct Clerk instance
		// Note: SDK might use this to construct the JWKS URL
		config.Key = clerk.String(publishableKey)
	}
	
	jwksClient := jwks.NewClient(config)

	// Step 3: Fetch the JSON Web Key (JWK) corresponding to the Key ID
	jwk, err := jwt.GetJSONWebKey(ctx, &jwt.GetJSONWebKeyParams{
		KeyID:      unsafeClaims.KeyID,
		JWKSClient: jwksClient,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get JSON web key: %w", err)
	}

	// Step 4: Verify the session token using the JWK
	claims, err := jwt.Verify(ctx, &jwt.VerifyParams{
		Token: token,
		JWK:   jwk,
	})
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	// Extract user ID from subject claim
	if claims.Subject == "" {
		return "", fmt.Errorf("token does not contain user ID")
	}

	return claims.Subject, nil
}

// openBrowser opens the specified URL in the user's default browser
// Supports macOS, Linux, and Windows
func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Run()
}

