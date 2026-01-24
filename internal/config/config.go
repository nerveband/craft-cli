package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const (
	ConfigDirName  = ".craft-cli"
	ConfigFileName = "config.json"
)

// Profile represents a named API configuration
type Profile struct {
	URL string `json:"url"`
}

// Config represents the application configuration
type Config struct {
	DefaultFormat string             `json:"default_format"`
	ActiveProfile string             `json:"active_profile,omitempty"`
	Profiles      map[string]Profile `json:"profiles,omitempty"`
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
			return &Config{
				DefaultFormat: "json",
				Profiles:      make(map[string]Profile),
			}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if cfg.DefaultFormat == "" {
		cfg.DefaultFormat = "json"
	}
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
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

// AddProfile adds or updates a named profile
func (m *Manager) AddProfile(name, url string) error {
	cfg, err := m.Load()
	if err != nil {
		return err
	}

	cfg.Profiles[name] = Profile{URL: url}

	// If this is the first profile, make it active
	if len(cfg.Profiles) == 1 {
		cfg.ActiveProfile = name
	}

	return m.Save(cfg)
}

// RemoveProfile deletes a named profile
func (m *Manager) RemoveProfile(name string) error {
	cfg, err := m.Load()
	if err != nil {
		return err
	}

	if _, exists := cfg.Profiles[name]; !exists {
		return fmt.Errorf("profile '%s' not found", name)
	}

	delete(cfg.Profiles, name)

	// Clear active profile if it was the removed one
	if cfg.ActiveProfile == name {
		cfg.ActiveProfile = ""
	}

	return m.Save(cfg)
}

// UseProfile sets the active profile
func (m *Manager) UseProfile(name string) error {
	cfg, err := m.Load()
	if err != nil {
		return err
	}

	if _, exists := cfg.Profiles[name]; !exists {
		return fmt.Errorf("profile '%s' not found", name)
	}

	cfg.ActiveProfile = name
	return m.Save(cfg)
}

// ListProfiles returns all profiles with the active one marked
func (m *Manager) ListProfiles() ([]ProfileInfo, error) {
	cfg, err := m.Load()
	if err != nil {
		return nil, err
	}

	var profiles []ProfileInfo
	for name, profile := range cfg.Profiles {
		profiles = append(profiles, ProfileInfo{
			Name:   name,
			URL:    profile.URL,
			Active: name == cfg.ActiveProfile,
		})
	}

	// Sort by name for consistent output
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})

	return profiles, nil
}

// ProfileInfo contains profile details for display
type ProfileInfo struct {
	Name   string
	URL    string
	Active bool
}

// GetActiveURL returns the URL of the active profile
func (m *Manager) GetActiveURL() (string, error) {
	cfg, err := m.Load()
	if err != nil {
		return "", err
	}

	if cfg.ActiveProfile == "" {
		return "", fmt.Errorf("no active profile. Run 'craft config add <name> <url>' first")
	}

	profile, exists := cfg.Profiles[cfg.ActiveProfile]
	if !exists {
		return "", fmt.Errorf("active profile '%s' not found. Run 'craft config add <name> <url>' first", cfg.ActiveProfile)
	}

	return profile.URL, nil
}

// Reset clears the configuration
func (m *Manager) Reset() error {
	if err := os.RemoveAll(m.configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove config file: %w", err)
	}
	return nil
}
