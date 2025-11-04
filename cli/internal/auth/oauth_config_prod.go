//go:build !dev
// +build !dev

package auth

// OAuthAppConfig contains the OAuth application configuration for PRODUCTION
// This file is compiled by default (when not using -tags dev)
// Update these values for your production Clerk instance
var OAuthAppConfig = OAuthConfig{
	ClientID:     "", // Set your production OAuth client ID here
	AuthorizeURL: "", // Set your production authorization URL here
	TokenURL:     "", // Set your production token URL here
	UserInfoURL:  "", // Set your production userinfo URL here
	RedirectURI:  "", // Empty = use dynamic port discovery
}

// PublishableKey is the Clerk publishable key for PRODUCTION
// Set this to your production publishable key (pk_live_...)
var PublishableKey = "" // Set your production publishable key here

