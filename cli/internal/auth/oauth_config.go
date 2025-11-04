package auth

import "os"

// OAuthAppConfig and PublishableKey are declared in:
// - oauth_config_dev.go (compiled with -tags dev) for development
// - oauth_config_prod.go (compiled by default) for production
// These variables are populated by the build-tagged files

// GetOAuthConfig returns the OAuth configuration
// Configuration is loaded from oauth_config_dev.go (with -tags dev) or oauth_config_prod.go (default)
// Can be overridden by environment variables for development/testing
func GetOAuthConfig() OAuthConfig {
	// Start with the compiled-in configuration (from dev or prod build tag)
	config := OAuthAppConfig

	// Check for environment variable overrides (for development/testing)
	if clientID := getEnvOrDefault("CLERK_OAUTH_CLIENT_ID", ""); clientID != "" {
		config.ClientID = clientID
	}
	if authorizeURL := getEnvOrDefault("CLERK_OAUTH_AUTHORIZE_URL", ""); authorizeURL != "" {
		config.AuthorizeURL = authorizeURL
	}
	if tokenURL := getEnvOrDefault("CLERK_OAUTH_TOKEN_URL", ""); tokenURL != "" {
		config.TokenURL = tokenURL
	}
	if redirectURI := getEnvOrDefault("CLERK_OAUTH_REDIRECT_URI", ""); redirectURI != "" {
		config.RedirectURI = redirectURI
	}

	return config
}

// GetPublishableKey returns the Clerk publishable key for token verification
// Key is loaded from oauth_config_dev.go (with -tags dev) or oauth_config_prod.go (default)
func GetPublishableKey() string {
	return PublishableKey
}

// Helper function to get environment variable or return default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

