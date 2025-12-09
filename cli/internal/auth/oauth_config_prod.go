//go:build !dev
// +build !dev

package auth

// OAuthAppConfig removed - no longer needed with sign-in token flow

// PublishableKey is the Clerk publishable key for PRODUCTION
// Set this to your production publishable key (pk_live_...)
var PublishableKey = "pk_test_c21vb3RoLWdsaWRlci05My5jbGVyay5hY2NvdW50cy5kZXYk" // Set your production publishable key here

// WebAppURL is the web application URL for PRODUCTION
// Set this to your production web app URL (e.g., https://your-domain.com)
var WebAppURL = "https://hookie.sh" // Set your production web app URL here