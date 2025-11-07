package auth

import (
	"context"
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
	"strings"
	"time"

	"crypto/tls"
	"crypto/x509"

	"github.com/clerk/clerk-sdk-go/v2"
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

// VerifyToken verifies a Clerk token and returns the user ID
// Uses JWT verification with JWKS (no secret key required)
func VerifyToken(ctx context.Context, token string) (string, error) {
	publishableKey := GetPublishableKey()
	if publishableKey == "" {
		return "", fmt.Errorf("publishable key not configured")
	}
	return verifyTokenJWT(ctx, token, publishableKey)
}

// verifyTokenJWT verifies a Clerk JWT session token using JWKS
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

	// Step 2: Extract Clerk instance from publishable key to build JWKS URL
	parts := strings.Split(publishableKey, "_")
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid publishable key format")
	}
	
	instanceEncoded := parts[2]
	instanceBytes, err := base64.StdEncoding.DecodeString(instanceEncoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode instance from publishable key: %w", err)
	}
	instance := string(instanceBytes)
	instance = strings.TrimSuffix(instance, "$")
	
	// Step 3: Build JWKS URL (public endpoint, no authentication needed)
	jwksURL := fmt.Sprintf("https://%s/.well-known/jwks.json", instance)
	
	// Step 4: Fetch JWKS from public endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, jwksURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create JWKS request: %w", err)
	}
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("JWKS request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	var jwksResponse struct {
		Keys []map[string]interface{} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jwksResponse); err != nil {
		return "", fmt.Errorf("failed to decode JWKS response: %w", err)
	}
	
	// Step 5: Find the JWK matching the key ID
	var jwkJSON []byte
	for _, key := range jwksResponse.Keys {
		if kid, ok := key["kid"].(string); ok && kid == unsafeClaims.KeyID {
			var err error
			jwkJSON, err = json.Marshal(key)
			if err != nil {
				return "", fmt.Errorf("failed to marshal JWK: %w", err)
			}
			break
		}
	}
	
	if jwkJSON == nil {
		return "", fmt.Errorf("JWK not found for key ID: %s", unsafeClaims.KeyID)
	}
	
	// Step 6: Parse the JWK into Clerk's JSONWebKey type
	var jwk clerk.JSONWebKey
	if err := json.Unmarshal(jwkJSON, &jwk); err != nil {
		return "", fmt.Errorf("failed to parse JWK: %w", err)
	}
	
	// Step 7: Verify the session token using the JWK
	claims, err := jwt.Verify(ctx, &jwt.VerifyParams{
		Token: token,
		JWK:   &jwk,
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

// OpenBrowser opens the specified URL in the user's default browser
// Supports macOS, Linux, and Windows
func OpenBrowser(url string) error {
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

// ReceiveSignInToken starts a local HTTP server to receive the sign-in token from the web app
// The server listens on the specified port and waits for a GET request to /callback with a token query parameter
// Returns the sign-in token when received, or an error if timeout or failure occurs
func ReceiveSignInToken(ctx context.Context, port int) (string, error) {
	server := &http.Server{
		Addr: fmt.Sprintf("127.0.0.1:%d", port), // Bind to localhost only for security
	}

	tokenChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// Only accept GET requests
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			errorChan <- fmt.Errorf("invalid method: %s", r.Method)
			return
		}

		// Validate request is from localhost (security)
		if r.RemoteAddr != "" {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err == nil && host != "127.0.0.1" && host != "::1" && host != "localhost" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				errorChan <- fmt.Errorf("request not from localhost: %s", host)
				return
			}
		}

		// Get token from query parameter
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "Missing token parameter", http.StatusBadRequest)
			errorChan <- fmt.Errorf("missing token in callback URL")
			return
		}

		// Send success HTML page
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

		tokenChan <- token
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
	case token := <-tokenChan:
		server.Shutdown(ctx)
		return token, nil
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

// CompleteSignInWithTicket completes a sign-in using a sign-in token (ticket)
// Makes a direct HTTP request to Clerk's client API
// Returns the session token (JWT) that can be used for authenticated requests
func CompleteSignInWithTicket(ctx context.Context, signInToken string) (string, error) {
	// Get publishable key to determine Clerk instance
	publishableKey := GetPublishableKey()
	if publishableKey == "" {
		return "", fmt.Errorf("publishable key not configured")
	}

	// Extract Clerk instance from publishable key
	parts := strings.Split(publishableKey, "_")
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid publishable key format")
	}
	
	instanceEncoded := parts[2]
	instanceBytes, err := base64.StdEncoding.DecodeString(instanceEncoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode instance from publishable key: %w", err)
	}
	instance := string(instanceBytes)
	instance = strings.TrimSuffix(instance, "$")
	
	// Build the API URL
	apiURL := fmt.Sprintf("https://%s/v1/client/sign_ins", instance)
	
	// Prepare form-encoded data
	data := url.Values{}
	data.Set("strategy", "ticket")
	data.Set("ticket", signInToken)
	
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Origin", "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Make the request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	// Parse response
	var signInResponse signInResponse
	if err := json.Unmarshal(body, &signInResponse); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Log the parsed response for debugging
	responseJSON, _ := json.MarshalIndent(signInResponse, "", "  ")
	fmt.Printf("Parsed sign-in response: %s\n", string(responseJSON))
	fmt.Printf("Raw response body: %s\n", string(body))

	// Check the status - for ticket strategy, it should be "complete"
	if signInResponse.Response.Status != "complete" {
		return "", fmt.Errorf("sign-in not complete: status=%s (expected 'complete')", signInResponse.Response.Status)
	}

	// Extract session token from client.sessions[0].last_active_token.jwt
	if len(signInResponse.Client.Sessions) == 0 {
		return "", fmt.Errorf("no sessions found in response")
	}

	sessionToken := signInResponse.Client.Sessions[0].LastActiveToken.JWT
	if sessionToken == "" {
		return "", fmt.Errorf("session token (JWT) not found in response")
	}

	return sessionToken, nil
}

// signInResponse represents the response from Clerk's client sign-in API
type signInResponse struct {
	Response struct {
		Status          string `json:"status"`
		CreatedSessionID string `json:"created_session_id,omitempty"`
	} `json:"response"`
	Client struct {
		Sessions []struct {
			ID              string `json:"id"`
			LastActiveToken struct {
				JWT string `json:"jwt"`
			} `json:"last_active_token"`
		} `json:"sessions"`
	} `json:"client"`
}
