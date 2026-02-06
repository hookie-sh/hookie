package cmd

import (
	"fmt"
	"net"
	"os"
	"runtime"
	runtimeDebug "runtime/debug"
	"strings"

	"github.com/fatih/color"
	"github.com/hookie/cli/internal/config"
	"github.com/hookie/cli/internal/relay"
	"gopkg.in/yaml.v3"
)

// printDebugInfo displays comprehensive debug information for troubleshooting
func printDebugInfo(commandName, orgID, fullCommand string) {
	fmt.Println()
	fmt.Println(color.CyanString("=== Debug Information ==="))
	fmt.Println("Copy this information to share with support:")
	fmt.Println()
	
	// Command executed
	fmt.Println("Command Executed:")
	fmt.Printf("  %s\n", fullCommand)
	fmt.Println()

	// System Information
	fmt.Println("System:")
	fmt.Printf("  OS: %s\n", runtime.GOOS)
	fmt.Printf("  Architecture: %s\n", runtime.GOARCH)
	fmt.Printf("  Go Version: %s\n", runtime.Version())
	
	// CLI Version
	cliVersion := getCLIVersion()
	fmt.Printf("  CLI Version: %s\n", cliVersion)
	fmt.Println()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Configuration:")
		fmt.Printf("  Error loading config: %v\n", err)
		fmt.Println()
		// Continue with limited info
		cfg = nil
	}

	// Determine relay URL source and effective URL
	relayURLEnv := os.Getenv("HOOKIE_RELAY_URL")
	relayURLFromConfig := ""
	if cfg != nil {
		relayURLFromConfig = cfg.RelayURL
	}
	relayURLDefault := relay.GetRelayURL()

	var relayURLSource string
	var effectiveRelayURL string

	if relayURLEnv != "" {
		effectiveRelayURL = relayURLEnv
		relayURLSource = "environment variable (HOOKIE_RELAY_URL)"
	} else if relayURLFromConfig != "" {
		effectiveRelayURL = relayURLFromConfig
		relayURLSource = "config file"
	} else {
		effectiveRelayURL = relayURLDefault
		relayURLSource = "default"
	}

	// Determine authentication status
	authStatus := "anonymous"
	if cfg != nil && cfg.Token != "" {
		authStatus = "authenticated"
	}

	// Determine TLS mode
	tlsMode := determineTLSMode(effectiveRelayURL)

	// Configuration
	fmt.Println("Configuration:")
	fmt.Printf("  Relay URL: %s (from %s)\n", effectiveRelayURL, relayURLSource)
	if cfg != nil {
		fmt.Printf("  Machine ID: %s\n", cfg.MachineID)
	} else {
		fmt.Println("  Machine ID: (config not loaded)")
	}
	fmt.Printf("  Authentication: %s\n", authStatus)
	fmt.Printf("  TLS Mode: %s\n", tlsMode)
	if orgID != "" {
		fmt.Printf("  Org ID: %s\n", orgID)
	}
	fmt.Println()

	// Repository Config
	repoConfig, repoConfigPath, err := config.LoadRepoConfig()
	if err != nil {
		fmt.Println("Repository Config:")
		fmt.Printf("  Error loading config: %v\n", err)
		fmt.Println()
	} else if repoConfig != nil {
		fmt.Println("Repository Config:")
		fmt.Printf("  Path: %s\n", repoConfigPath)
		// Marshal config to YAML for display
		configYAML, err := yaml.Marshal(repoConfig)
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(configYAML)), "\n")
			for _, line := range lines {
				fmt.Printf("  %s\n", line)
			}
		} else {
			fmt.Printf("  (unable to format config: %v)\n", err)
		}
		fmt.Println()
	} else {
		fmt.Println("Repository Config:")
		fmt.Println("  (not found)")
		fmt.Println()
	}

	// Environment Variables
	fmt.Println("Environment Variables:")
	if relayURLEnv != "" {
		fmt.Printf("  HOOKIE_RELAY_URL: %s\n", relayURLEnv)
	} else {
		fmt.Println("  HOOKIE_RELAY_URL: (not set)")
	}
	insecureTLS := os.Getenv("HOOKIE_INSECURE_TLS")
	if insecureTLS != "" {
		fmt.Printf("  HOOKIE_INSECURE_TLS: %s\n", insecureTLS)
	} else {
		fmt.Println("  HOOKIE_INSECURE_TLS: (not set)")
	}
	fmt.Println()

	fmt.Println(color.CyanString("==========================================="))
	fmt.Println()
}

// determineTLSMode determines the TLS mode based on relay URL and environment
func determineTLSMode(relayURL string) string {
	isLocal := isLocalhost(relayURL)
	insecureTLSEnv := os.Getenv("HOOKIE_INSECURE_TLS")

	if isLocal && insecureTLSEnv == "" {
		return "insecure (localhost, HOOKIE_INSECURE_TLS not set)"
	}
	return "secure"
}

// isLocalhost checks if the URL is pointing to localhost or 127.0.0.1
// Duplicated from relay/client.go since it's not exported
func isLocalhost(url string) bool {
	// Remove scheme if present
	host := strings.TrimPrefix(url, "grpc://")
	host = strings.TrimPrefix(host, "grpcs://")
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimPrefix(host, "https://")

	// Extract host:port and check host
	hostname, _, err := net.SplitHostPort(host)
	if err != nil {
		// No port, use entire string as host
		hostname = host
	}

	return hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1" || hostname == ""
}

// getCLIVersion retrieves the CLI version from build info or returns a default
func getCLIVersion() string {
	buildInfo, ok := runtimeDebug.ReadBuildInfo()
	if !ok {
		return "unknown (build info not available)"
	}

	// Try to get version from the main module
	if buildInfo.Main.Version != "" && buildInfo.Main.Version != "(devel)" {
		return buildInfo.Main.Version
	}

	// If version is "(devel)", try to get vcs revision info
	var versionParts []string
	for _, setting := range buildInfo.Settings {
		if setting.Key == "vcs.revision" && setting.Value != "" {
			// Use first 7 characters of commit hash (short commit)
			revision := setting.Value
			if len(revision) > 7 {
				revision = revision[:7]
			}
			versionParts = append(versionParts, revision)
		}
		if setting.Key == "vcs.modified" && setting.Value == "true" {
			versionParts = append(versionParts, "dirty")
		}
	}

	if len(versionParts) > 0 {
		return "dev (" + strings.Join(versionParts, ", ") + ")"
	}

	return "dev (no version info)"
}
