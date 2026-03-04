//go:build dev
// +build dev

package auth

// OAuthAppConfig removed - no longer needed with sign-in token flow

// PublishableKey is the Clerk publishable key for DEVELOPMENT (set via env or ldflags at build time)
var PublishableKey = ""

// WebAppURL is the web application URL for DEVELOPMENT
var WebAppURL = "http://localhost:3000"

