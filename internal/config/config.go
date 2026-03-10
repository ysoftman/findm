package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	appName    = "findm"
	configFile = "config.json"
)

// Config holds the application configuration.
type Config struct {
	YouTubeAPIKey string `json:"youtube_api_key"`
}

// ConfigDir returns the path to the config directory (~/.config/findm/).
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".config", appName), nil
}

// Load reads the config from disk, falling back to environment variable.
// Returns a valid Config even if no API key is found (API key is optional).
func Load() (*Config, error) {
	cfg := &Config{}

	// Try environment variable first
	if key := os.Getenv("YOUTUBE_API_KEY"); key != "" {
		cfg.YouTubeAPIKey = key
		return cfg, nil
	}

	// Try config file
	dir, err := ConfigDir()
	if err != nil {
		return cfg, nil
	}

	data, err := os.ReadFile(filepath.Join(dir, configFile))
	if err != nil {
		return cfg, nil
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}

// Save writes the config to disk.
func Save(cfg *Config) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(filepath.Join(dir, configFile), data, 0o600)
}
