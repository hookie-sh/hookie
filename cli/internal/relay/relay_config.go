package relay

// RelayURL is declared in:
// - relay_config_dev.go (compiled with -tags dev) for development
// - relay_config_prod.go (compiled by default) for production
// These variables are populated by the build-tagged files

// GetRelayURL returns the default relay URL
// URL is loaded from relay_config_dev.go (with -tags dev) or relay_config_prod.go (default)
func GetRelayURL() string {
	return RelayURL
}

