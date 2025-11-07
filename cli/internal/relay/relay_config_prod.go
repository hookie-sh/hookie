//go:build !dev
// +build !dev

package relay

// RelayURL is the default relay service URL for PRODUCTION
// This file is compiled by default (when not using -tags dev)
// Update this value to your production relay URL
var RelayURL = "hookie-relay.fly.dev:443" // Use port 443 for TLS termination

