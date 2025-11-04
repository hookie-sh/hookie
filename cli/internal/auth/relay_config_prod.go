//go:build !dev
// +build !dev

package auth

// RelayURL is the default relay service URL for PRODUCTION
// This file is compiled by default (when not using -tags dev)
// Update this value to your production relay URL
var RelayURL = "" // Set your production relay URL here (e.g., "relay.yourdomain.com:50051")

