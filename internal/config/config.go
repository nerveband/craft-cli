package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	ConfigDirName  = ".craft-cli"
	ConfigFileName = "config.json"
)

// Config represents the application configuration
type Config struct {
	APIURL        string `json:"api_url"`
	DefaultFormat string `json:"default_format"`
}

// Manager handles configuration operations
type Manager struct {
	configDir  string
	configPath string
}

// NewManager creates a new configuration manager
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ConfigDirName)
	configPath := filepath.Join(configDir, ConfigFileName)

	return &Manager{
		configDir:  configDir,
		configPath: configPath,
	}, nil
}

// Load reads the configuration file
func (m *Manager) Load() (*Config, error) {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{DefaultFormat: "json"}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set default format if not specified
	if cfg.DefaultFormat == "" {
		cfg.DefaultFormat = "json"
	}

	return &cfg, nil
}

// Save writes the configuration file
func (m *Manager) Save(cfg *Config) error {
	// Create config directory if it doesn't exist
	if err := os.MkdirAll(m.configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// SetAPIURL updates the API URL in the configuration
func (m *Manager) SetAPIURL(url string) error {
	cfg, err := m.Load()
	if err != nil {
		return err
	}

	cfg.APIURL = url
	return m.Save(cfg)
}

// GetAPIURL retrieves the API URL from the configuration
func (m *Manager) GetAPIURL() (string, error) {
	cfg, err := m.Load()
	if err != nil {
		return "", err
	}

	return cfg.APIURL, nil
}

// Reset clears the configuration
func (m *Manager) Reset() error {
	if err := os.RemoveAll(m.configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove config file: %w", err)
	}
	return nil
}
