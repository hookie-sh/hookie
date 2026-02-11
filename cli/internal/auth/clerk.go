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

// DumpTokenPayload decodes and prints the JWT token payload for debugging
func DumpTokenPayload(token string) {
	// Extract token from "Bearer <token>" format if present
	token = strings.TrimPrefix(token, "Bearer ")
	token = strings.TrimSpace(token)

	// Split token into parts
	tokenParts := strings.Split(token, ".")
	if len(tokenParts) != 3 {
		fmt.Printf("Invalid token format: expected 3 parts, got %d\n", len(tokenParts))
		return
	}

	// Decode the payload (second part of JWT)
	payloadBytes, err := base64.RawURLEncoding.DecodeString(tokenParts[1])
	if err != nil {
		fmt.Printf("Failed to decode payload: %v\n", err)
		return
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		fmt.Printf("Failed to parse payload: %v\n", err)
		return
	}

	// Pretty print the payload
	prettyJSON, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fmt.Printf("Failed to format payload: %v\n", err)
		fmt.Printf("Raw payload: %s\n", string(payloadBytes))
		return
	}

	fmt.Println(string(prettyJSON))
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

// UserInfo contains user information extracted from token
type UserInfo struct {
	Name  string
	Email string
}

// GetUserInfoFromToken extracts user name and email from JWT token claims
// Returns the user's full name and email if available in the token
func GetUserInfoFromToken(ctx context.Context, token string) (*UserInfo, error) {
	publishableKey := GetPublishableKey()
	if publishableKey == "" {
		return nil, fmt.Errorf("publishable key not configured")
	}

	// Extract token from "Bearer <token>" format if present
	token = strings.TrimPrefix(token, "Bearer ")
	token = strings.TrimSpace(token)

	// Decode the token to access all claims (including custom ones)
	unsafeClaims, err := jwt.Decode(ctx, &jwt.DecodeParams{
		Token: token,
	})
	if err != nil {
		return nil, nil // Silent failure - token might not have name claims
	}

	// Extract Clerk instance from publishable key to build JWKS URL
	parts := strings.Split(publishableKey, "_")
	if len(parts) < 3 {
		return nil, nil
	}

	instanceEncoded := parts[2]
	instanceBytes, err := base64.StdEncoding.DecodeString(instanceEncoded)
	if err != nil {
		return nil, nil
	}
	instance := string(instanceBytes)
	instance = strings.TrimSuffix(instance, "$")

	// Build JWKS URL
	jwksURL := fmt.Sprintf("https://%s/.well-known/jwks.json", instance)

	// Fetch JWKS
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, jwksURL, nil)
	if err != nil {
		return nil, nil
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}

	var jwksResponse struct {
		Keys []map[string]interface{} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jwksResponse); err != nil {
		return nil, nil
	}

	// Find the JWK matching the key ID
	var jwkJSON []byte
	for _, key := range jwksResponse.Keys {
		if kid, ok := key["kid"].(string); ok && kid == unsafeClaims.KeyID {
			var err error
			jwkJSON, err = json.Marshal(key)
			if err != nil {
				return nil, nil
			}
			break
		}
	}

	if jwkJSON == nil {
		return nil, nil
	}

	// Parse the JWK
	var jwk clerk.JSONWebKey
	if err := json.Unmarshal(jwkJSON, &jwk); err != nil {
		return nil, nil
	}

	// Verify the token to ensure it's valid (but we'll decode payload for custom claims)
	_, err = jwt.Verify(ctx, &jwt.VerifyParams{
		Token: token,
		JWK:   &jwk,
	})
	if err != nil {
		return nil, nil
	}

	// Decode the payload directly to access custom claims
	// The verified claims might not expose custom fields, so we decode the payload
	tokenParts := strings.Split(token, ".")
	if len(tokenParts) != 3 {
		return nil, nil
	}

	// Decode the payload (second part of JWT)
	payloadBytes, err := base64.RawURLEncoding.DecodeString(tokenParts[1])
	if err != nil {
		return nil, nil
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, nil
	}

	// Extract firstName and lastName from payload (camelCase)
	// Also check snake_case for compatibility
	var firstName, lastName string
	if fn, ok := payload["firstName"].(string); ok && fn != "" {
		firstName = fn
	} else if fn, ok := payload["first_name"].(string); ok && fn != "" {
		firstName = fn
	}
	if ln, ok := payload["lastName"].(string); ok && ln != "" {
		lastName = ln
	} else if ln, ok := payload["last_name"].(string); ok && ln != "" {
		lastName = ln
	}

	// Extract email from payload
	var email string
	if e, ok := payload["email"].(string); ok && e != "" {
		email = e
	}

	// Build full name
	var fullName string
	if firstName != "" || lastName != "" {
		nameParts := []string{}
		if firstName != "" {
			nameParts = append(nameParts, firstName)
		}
		if lastName != "" {
			nameParts = append(nameParts, lastName)
		}
		fullName = strings.Join(nameParts, " ")
	}

	return &UserInfo{
		Name:  fullName,
		Email: email,
	}, nil
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

// GetUserInfo fetches user information from Clerk's client API using a session token
// Returns the user's full name and email if available
func GetUserInfo(ctx context.Context, sessionToken string) (*UserInfo, error) {
	publishableKey := GetPublishableKey()
	if publishableKey == "" {
		return nil, fmt.Errorf("publishable key not configured")
	}

	// Extract Clerk instance from publishable key
	parts := strings.Split(publishableKey, "_")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid publishable key format")
	}

	instanceEncoded := parts[2]
	instanceBytes, err := base64.StdEncoding.DecodeString(instanceEncoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode instance from publishable key: %w", err)
	}
	instance := string(instanceBytes)
	instance = strings.TrimSuffix(instance, "$")

	// Build the API URL
	apiURL := fmt.Sprintf("https://%s/v1/client", instance)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers - Clerk client API requires Authorization header with session token
	// Also need Origin header (can be empty) and other standard headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", sessionToken))
	req.Header.Set("Origin", "")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Make the request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Don't return error, just return nil - we'll handle gracefully
		return nil, nil
	}

	// Parse response - try multiple possible structures
	var clientResponse map[string]interface{}
	if err := json.Unmarshal(body, &clientResponse); err != nil {
		return nil, nil
	}

	// Try to extract user info from various possible response structures
	var firstName, lastName, email string

	// Try sessions[0].user structure (most common)
	if sessions, ok := clientResponse["sessions"].([]interface{}); ok && len(sessions) > 0 {
		if session, ok := sessions[0].(map[string]interface{}); ok {
			if user, ok := session["user"].(map[string]interface{}); ok {
				// Try camelCase first, then snake_case
				if fn, ok := user["firstName"].(string); ok && fn != "" {
					firstName = fn
				} else if fn, ok := user["first_name"].(string); ok && fn != "" {
					firstName = fn
				}
				if ln, ok := user["lastName"].(string); ok && ln != "" {
					lastName = ln
				} else if ln, ok := user["last_name"].(string); ok && ln != "" {
					lastName = ln
				}
				// Extract email
				if e, ok := user["emailAddresses"].([]interface{}); ok && len(e) > 0 {
					if emailObj, ok := e[0].(map[string]interface{}); ok {
						if e, ok := emailObj["emailAddress"].(string); ok && e != "" {
							email = e
						}
					}
				} else if e, ok := user["email"].(string); ok && e != "" {
					email = e
				}
			}
		}
	}

	// Try direct user structure
	if firstName == "" && lastName == "" {
		if user, ok := clientResponse["user"].(map[string]interface{}); ok {
			if fn, ok := user["firstName"].(string); ok && fn != "" {
				firstName = fn
			} else if fn, ok := user["first_name"].(string); ok && fn != "" {
				firstName = fn
			}
			if ln, ok := user["lastName"].(string); ok && ln != "" {
				lastName = ln
			} else if ln, ok := user["last_name"].(string); ok && ln != "" {
				lastName = ln
			}
			if e, ok := user["email"].(string); ok && e != "" {
				email = e
			}
		}
	}

	// Try client.sessions[0].user structure (from sign-in response)
	if firstName == "" && lastName == "" {
		if client, ok := clientResponse["client"].(map[string]interface{}); ok {
			if sessions, ok := client["sessions"].([]interface{}); ok && len(sessions) > 0 {
				if session, ok := sessions[0].(map[string]interface{}); ok {
					if user, ok := session["user"].(map[string]interface{}); ok {
						if fn, ok := user["firstName"].(string); ok && fn != "" {
							firstName = fn
						} else if fn, ok := user["first_name"].(string); ok && fn != "" {
							firstName = fn
						}
						if ln, ok := user["lastName"].(string); ok && ln != "" {
							lastName = ln
						} else if ln, ok := user["last_name"].(string); ok && ln != "" {
							lastName = ln
						}
						if e, ok := user["email"].(string); ok && e != "" {
							email = e
						}
					}
				}
			}
		}
	}

	// Build full name from first and last name
	var fullName string
	if firstName != "" || lastName != "" {
		nameParts := []string{}
		if firstName != "" {
			nameParts = append(nameParts, firstName)
		}
		if lastName != "" {
			nameParts = append(nameParts, lastName)
		}
		fullName = strings.Join(nameParts, " ")
	}

	if fullName == "" && email == "" {
		return nil, nil
	}

	return &UserInfo{
		Name:  fullName,
		Email: email,
	}, nil
}
