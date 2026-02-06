package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const repoConfigFileName = "hookie.yml"

// RepoConfig represents the repository-specific configuration
type RepoConfig struct {
	AppID   string            `yaml:"app_id,omitempty"`
	TopicID string            `yaml:"topic_id,omitempty"`
	Forward string            `yaml:"forward,omitempty"`
	Topics  map[string]string `yaml:"topics,omitempty"`
}

// LoadRepoConfig searches for hookie.yml starting from the current working directory
// and walking up the directory tree until found or reaching filesystem root.
// Returns the config if found, or nil if not found.
func LoadRepoConfig() (*RepoConfig, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Walk up directory tree
	dir := cwd
	for {
		configPath := filepath.Join(dir, repoConfigFileName)
		
		// Check if file exists
		if _, err := os.Stat(configPath); err == nil {
			// File exists, read and parse it
			data, err := os.ReadFile(configPath)
			if err != nil {
				return nil, "", fmt.Errorf("failed to read %s: %w", configPath, err)
			}

			var config RepoConfig
			if err := yaml.Unmarshal(data, &config); err != nil {
				return nil, "", fmt.Errorf("failed to parse %s: %w", configPath, err)
			}

			// Validate config
			if err := validateRepoConfig(&config); err != nil {
				return nil, "", fmt.Errorf("invalid configuration in %s: %w", configPath, err)
			}

			return &config, configPath, nil
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	// Config file not found
	return nil, "", nil
}

// SaveRepoConfig writes the RepoConfig to a file at the specified path
func SaveRepoConfig(config *RepoConfig, filePath string) error {
	// Validate config before saving
	if err := validateRepoConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Marshal to YAML with indentation
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// validateRepoConfig validates the repository configuration
func validateRepoConfig(config *RepoConfig) error {
	// Check that app_id and topic_id are mutually exclusive
	if config.AppID != "" && config.TopicID != "" {
		return fmt.Errorf("cannot specify both app_id and topic_id")
	}

	// Validate forward URL if provided
	if config.Forward != "" {
		if err := validateURL(config.Forward); err != nil {
			return fmt.Errorf("invalid forward URL: %w", err)
		}
	}

	// Validate topic forward URLs if provided
	if config.Topics != nil {
		for topicID, topicURL := range config.Topics {
			if topicURL != "" {
				if err := validateURL(topicURL); err != nil {
					return fmt.Errorf("invalid forward URL for topic %s: %w", topicID, err)
				}
			}
		}
	}

	return nil
}

// validateURL validates that a URL has a scheme and host
func validateURL(urlStr string) error {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	if parsed.Scheme == "" {
		return fmt.Errorf("URL must include a scheme (e.g., http:// or https://)")
	}

	if parsed.Host == "" {
		return fmt.Errorf("URL must include a host")
	}

	return nil
}

// RepoConfigExists checks if hookie.yml exists in the current directory or any parent directory
func RepoConfigExists() (bool, string) {
	config, path, err := LoadRepoConfig()
	if err != nil {
		return false, ""
	}
	return config != nil, path
}
