//go:build dev
// +build dev

package auth

// OAuthAppConfig contains the OAuth application configuration for DEVELOPMENT
// This file is only compiled when building with: go build -tags dev
var OAuthAppConfig = OAuthConfig{
	ClientID:     "mUFZjFsoJS4L8S56", // Your dev OAuth client ID
	AuthorizeURL: "https://smooth-glider-93.clerk.accounts.dev/oauth/authorize",
	TokenURL:     "https://smooth-glider-93.clerk.accounts.dev/oauth/token",
	UserInfoURL:  "https://smooth-glider-93.clerk.accounts.dev/oauth/userinfo",
	RedirectURI:  "", // Empty = use dynamic port discovery
}

// PublishableKey is the Clerk publishable key for DEVELOPMENT
var PublishableKey = "pk_test_c21vb3RoLWdsaWRlci05My5jbGVyay5hY2NvdW50cy5kZXYk"

