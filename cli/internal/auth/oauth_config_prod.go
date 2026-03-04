//go:build !dev
// +build !dev

package auth

// OAuthAppConfig removed - no longer needed with sign-in token flow

// PublishableKey is the Clerk publishable key for PRODUCTION (set via env or ldflags at build time)
var PublishableKey = ""

// WebAppURL is the web application URL for PRODUCTION
// Set this to your production web app URL (e.g., https://your-domain.com)
var WebAppURL = "https://hookie.sh" // Set your production web app URL here