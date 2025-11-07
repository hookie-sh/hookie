//go:build !dev
// +build !dev

package auth

// OAuthAppConfig contains the OAuth application configuration for PRODUCTION
// This file is compiled by default (when not using -tags dev)
// Update these values for your production Clerk instance
var OAuthAppConfig = OAuthConfig{
	ClientID:     "mUFZjFsoJS4L8S56", // Set your production OAuth client ID here
	AuthorizeURL: "https://smooth-glider-93.clerk.accounts.dev/oauth/authorize", // Set your production authorization URL here
	TokenURL:     "https://smooth-glider-93.clerk.accounts.dev/oauth/token", // Set your production token URL here
	UserInfoURL:  "https://smooth-glider-93.clerk.accounts.dev/oauth/userinfo", // Set your production userinfo URL here
	RedirectURI:  "", // Empty = use dynamic port discovery
}

// PublishableKey is the Clerk publishable key for PRODUCTION
// Set this to your production publishable key (pk_live_...)
var PublishableKey = "" // Set your production publishable key here

