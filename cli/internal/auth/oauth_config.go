package auth

import "os"

// OAuthAppConfig and PublishableKey are declared in:
// - oauth_config_dev.go (compiled with -tags dev) for development
// - oauth_config_prod.go (compiled by default) for production
// These variables are populated by the build-tagged files

// WebAppURL is declared in:
// - oauth_config_dev.go (compiled with -tags dev) for development
// - oauth_config_prod.go (compiled by default) for production

// OAuthConfig and GetOAuthConfig removed - no longer needed with sign-in token flow

// GetPublishableKey returns the Clerk publishable key for token verification.
// Env CLERK_PUBLISHABLE_KEY takes precedence; otherwise the build-injected PublishableKey is used (e.g. via ldflags in CI).
func GetPublishableKey() string {
	if k := os.Getenv("CLERK_PUBLISHABLE_KEY"); k != "" {
		return k
	}
	return PublishableKey
}

// GetWebAppURL returns the web application URL for CLI authentication
// URL is loaded from oauth_config_dev.go (with -tags dev) or oauth_config_prod.go (default)
// Can be overridden by HOOKIE_WEB_APP_URL environment variable
func GetWebAppURL() string {
	url := WebAppURL
	if envURL := getEnvOrDefault("HOOKIE_WEB_APP_URL", ""); envURL != "" {
		url = envURL
	}
	return url
}

// Helper function to get environment variable or return default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

