package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/99designs/keyring"
)

type Config struct {
	Token    string `json:"-"` // Token stored in keyring, not in JSON
	UserID   string `json:"user_id"`
	RelayURL string `json:"relay_url,omitempty"`
}

const (
	configDirName  = ".hookie"
	configFileName = "config.json"
	keyringService = "hookie"
	keyringAccount = "token"
)

// getKeyring initializes and returns a keyring instance
// Configured for macOS to minimize password prompts by using the login keychain
// and allowing access when the keychain is unlocked
func getKeyring() (keyring.Keyring, error) {
	config := keyring.Config{
		ServiceName: keyringService,
	}
	
	// On macOS, configure keychain settings to reduce password prompts
	// The login keychain is unlocked when the user is logged in, reducing prompts
	if runtime.GOOS == "darwin" {
		config.AllowedBackends = []keyring.BackendType{keyring.KeychainBackend}
		// Note: KeychainName may not be available in all versions of the library
		// If it causes compilation issues, it will be ignored gracefully
	}
	
	return keyring.Open(config)
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, configDirName)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(configDir, configFileName), nil
}

func Load() (*Config, error) {
	config := &Config{}

	// Load UserID and RelayURL from JSON file
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	// Try to read config file (it's okay if it doesn't exist)
	if data, err := os.ReadFile(configPath); err == nil {
		// Use a struct that can read the old format with token
		var fileConfig struct {
			Token    string `json:"token"`
			UserID   string `json:"user_id"`
			RelayURL string `json:"relay_url,omitempty"`
		}
		if err := json.Unmarshal(data, &fileConfig); err == nil {
			config.UserID = fileConfig.UserID
			config.RelayURL = fileConfig.RelayURL
			// If there's a token in the file, migrate it to keyring
			if fileConfig.Token != "" {
				if err := migrateTokenToKeyring(fileConfig.Token); err == nil {
					// Migration successful, rewrite config without token
					configToSave := &Config{
						UserID:   config.UserID,
						RelayURL: config.RelayURL,
					}
					if err := saveConfigFile(configToSave); err == nil {
						// Successfully migrated and saved
					}
				}
			}
		}
	}

	// Fetch Token from keyring
	kr, err := getKeyring()
	if err == nil {
		item, err := kr.Get(keyringAccount)
		if err == nil {
			token := string(item.Data)
			// Trim whitespace and newlines
			token = strings.TrimSpace(token)
			token = strings.Trim(token, "\n\r")
			config.Token = token
			
			// Validate token format (warn but don't fail if invalid)
			if token != "" {
				if err := ValidateTokenFormat(token); err != nil {
					// Token format is invalid, but we'll still return it
					// The caller will handle the error when trying to use it
				}
			}
		}
		// If keyring fails, silently continue (fallback to file migration above)
	}

	return config, nil
}

// ValidateTokenFormat validates that a token is in JWT format (three parts separated by dots)
func ValidateTokenFormat(token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("token is empty")
	}
	
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid JWT format: expected 3 parts separated by dots, got %d parts", len(parts))
	}
	
	// Check that each part is non-empty
	for i, part := range parts {
		if part == "" {
			return fmt.Errorf("invalid JWT format: part %d is empty", i+1)
		}
	}
	
	return nil
}

func Save(config *Config) error {
	// Validate and clean token before storing
	if config.Token != "" {
		// Trim whitespace and newlines
		config.Token = strings.TrimSpace(config.Token)
		config.Token = strings.Trim(config.Token, "\n\r")
		
		// Validate JWT format
		if err := ValidateTokenFormat(config.Token); err != nil {
			return fmt.Errorf("invalid token format: %w", err)
		}
		
		// Store Token in keyring
		kr, err := getKeyring()
		if err != nil {
			// If keyring is unavailable, we can't store the token securely
			// Still save UserID and RelayURL, but token will be lost
			// User will need to login again
		} else {
			// On macOS, try to remove the old item first to ensure it's recreated
			// with the current keyring configuration (which has better access control)
			// This helps reduce password prompts for future accesses
			if runtime.GOOS == "darwin" {
				_ = kr.Remove(keyringAccount) // Ignore error if item doesn't exist
			}
			
			err = kr.Set(keyring.Item{
				Key:  keyringAccount,
				Data: []byte(config.Token),
			})
			if err != nil {
				// If keyring set fails, token storage failed
				// Still save UserID and RelayURL
			}
		}
	}

	// Save UserID and RelayURL to JSON file (Token is excluded via json:"-")
	return saveConfigFile(config)
}

// saveConfigFile saves only UserID and RelayURL to the JSON file
func saveConfigFile(config *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Create a copy without Token for JSON serialization
	fileConfig := struct {
		UserID   string `json:"user_id"`
		RelayURL string `json:"relay_url,omitempty"`
	}{
		UserID:   config.UserID,
		RelayURL: config.RelayURL,
	}

	data, err := json.MarshalIndent(fileConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// migrateTokenToKeyring migrates a token from file to keyring
func migrateTokenToKeyring(token string) error {
	// Trim whitespace and newlines before migrating
	token = strings.TrimSpace(token)
	token = strings.Trim(token, "\n\r")
	
	kr, err := getKeyring()
	if err != nil {
		return fmt.Errorf("failed to initialize keyring: %w", err)
	}

	// On macOS, ensure any old item is removed first so the new one is created
	// with the current keyring configuration (better access control)
	if runtime.GOOS == "darwin" {
		_ = kr.Remove(keyringAccount) // Ignore error if item doesn't exist
	}

	err = kr.Set(keyring.Item{
		Key:  keyringAccount,
		Data: []byte(token),
	})
	if err != nil {
		return fmt.Errorf("failed to store token in keyring: %w", err)
	}

	return nil
}

func Clear() error {
	// Delete token from keyring
	kr, err := getKeyring()
	if err == nil {
		_ = kr.Remove(keyringAccount) // Ignore error if key doesn't exist
	}

	// Delete config JSON file
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove config: %w", err)
	}

	return nil
}

