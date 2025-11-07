package auth

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwks"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
)

type Verifier struct {
	jwksClient *jwks.Client
}

func NewVerifier(secretKey string) (*Verifier, error) {
	// Set the API key globally using resource-based approach
	clerk.SetKey(secretKey)

	// Configure TLS to use system certificates
	// This is important for containerized environments like Fly.io
	systemCerts, err := x509.SystemCertPool()
	if err != nil {
		// Fallback: create new pool (will be empty but won't crash)
		systemCerts = x509.NewCertPool()
	}

	// Configure default HTTP transport to use system certificates
	// The Clerk SDK uses this for JWKS requests
	if transport, ok := http.DefaultTransport.(*http.Transport); ok {
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{
				RootCAs: systemCerts,
			}
		} else {
			transport.TLSClientConfig.RootCAs = systemCerts
		}
	} else {
		http.DefaultTransport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: systemCerts,
			},
		}
	}

	// Create JWKS client for networkless JWT verification
	config := &clerk.ClientConfig{}
	config.Key = clerk.String(secretKey)
	jwksClient := jwks.NewClient(config)

	return &Verifier{
		jwksClient: jwksClient,
	}, nil
}

// TokenInfo contains user and organization information from a verified token
type TokenInfo struct {
	UserID string
	OrgID  string
}

// VerifyToken verifies a Clerk JWT session token using networkless verification
// as recommended by Clerk: https://clerk.com/docs/guides/sessions/session-tokens
// This function works with both session tokens and sign-in tokens that have been
// exchanged for session tokens via Clerk's sign-in API
func (v *Verifier) VerifyToken(ctx context.Context, token string) (*TokenInfo, error) {
	// Extract token from "Bearer <token>" format if present
	token = strings.TrimPrefix(token, "Bearer ")
	token = strings.TrimSpace(token)
	
	// Validate token format before attempting to decode
	if token == "" {
		return nil, fmt.Errorf("token is empty")
	}
	
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format: expected JWT with 3 parts separated by dots, got %d parts (token length: %d). This usually means the token is not a valid JWT or was corrupted during storage", len(parts), len(token))
	}
	
	// Check that each part is non-empty
	for i, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("invalid token format: JWT part %d is empty (expected 3 non-empty parts separated by dots)", i+1)
		}
	}

	// Step 1: Decode the session JWT to find the key ID
	unsafeClaims, err := jwt.Decode(ctx, &jwt.DecodeParams{
		Token: token,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT token: %w. Token format validation passed but JWT decoding failed. This may indicate the token is malformed or not a valid Clerk JWT", err)
	}

	// Step 2: Fetch the JSON Web Key (JWK) corresponding to the Key ID
	jwk, err := jwt.GetJSONWebKey(ctx, &jwt.GetJSONWebKeyParams{
		KeyID:      unsafeClaims.KeyID,
		JWKSClient: v.jwksClient,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get JSON web key for key ID %s: %w. This may indicate the token was issued by a different Clerk instance or the key has been rotated", unsafeClaims.KeyID, err)
	}

	// Step 3: Verify the session token using the JWK
	claims, err := jwt.Verify(ctx, &jwt.VerifyParams{
		Token: token,
		JWK:   jwk,
	})
	if err != nil {
		return nil, fmt.Errorf("token verification failed: %w. The token signature is invalid, expired, or not issued by the expected Clerk instance", err)
	}

	// Extract user ID from subject claim
	if claims.Subject == "" {
		return nil, fmt.Errorf("token does not contain user ID")
	}

	info := &TokenInfo{
		UserID: claims.Subject,
	}

	// Extract organization ID from active organization claim if available
	if claims.ActiveOrganizationID != "" {
		info.OrgID = claims.ActiveOrganizationID
	}

	return info, nil
}


