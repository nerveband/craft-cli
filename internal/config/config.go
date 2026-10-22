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
	Type          string `json:"type,omitempty"` // rest or mcp; empty means rest for legacy profiles
	URL           string `json:"url,omitempty"`
	MCPURL        string `json:"mcp_url,omitempty"`
	APIKey        string `json:"api_key,omitempty"`
	AccessMode    string `json:"access_mode,omitempty"`    // public or api_key
	Permission    string `json:"permission,omitempty"`     // read-only, write-only, read-write
	DocumentScope string `json:"document_scope,omitempty"` // all-documents, selected-documents, daily-notes, connection-defined
}

// Config represents the application configuration
type Config struct {
	CredentialStore string             `json:"credential_store,omitempty"`
	DefaultFormat   string             `json:"default_format"`
	ActiveProfile   string             `json:"active_profile,omitempty"`
	Profiles        map[string]Profile `json:"profiles,omitempty"`
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

	configDir := os.Getenv("CRAFT_CONFIG_DIR")
	if configDir == "" {
		configDir = filepath.Join(homeDir, ConfigDirName)
	}
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

	credentials, err := m.readCredentials(cfg.CredentialStore)
	if err != nil {
		return nil, err
	}
	for name, profile := range cfg.Profiles {
		secret := credentials[name]
		if profile.URL == "" {
			profile.URL = secret.URL
		}
		if profile.MCPURL == "" {
			profile.MCPURL = secret.MCPURL
		}
		if profile.APIKey == "" {
			profile.APIKey = secret.APIKey
		}
		cfg.Profiles[name] = profile
	}
	return &cfg, nil
}

// Save writes the configuration file
func (m *Manager) Save(cfg *Config) error {
	// Create config directory if it doesn't exist
	if err := os.MkdirAll(m.configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Publish a new credential generation, then atomically point metadata at it.
	// A failed config write leaves the old generation intact and readable.
	clean := *cfg
	clean.Profiles = make(map[string]Profile, len(cfg.Profiles))
	current := map[string]credential{}
	for name, profile := range cfg.Profiles {
		current[name] = credential{URL: profile.URL, MCPURL: profile.MCPURL, APIKey: profile.APIKey}
		profile.URL, profile.MCPURL, profile.APIKey = "", "", ""
		clean.Profiles[name] = profile
	}
	f, err := os.CreateTemp(m.configDir, "credentials-*.json")
	if err != nil {
		return err
	}
	credentialPath := f.Name()
	if err := f.Close(); err != nil {
		os.Remove(credentialPath)
		return err
	}
	clean.CredentialStore = filepath.Base(credentialPath)
	if err := writePrivateJSON(credentialPath, current); err != nil {
		os.Remove(credentialPath)
		return err
	}
	if err := writePrivateJSON(m.configPath, &clean); err != nil {
		os.Remove(credentialPath)
		return err
	}
	oldStore := cfg.CredentialStore
	cfg.CredentialStore = clean.CredentialStore
	if oldStore != "" && oldStore != clean.CredentialStore {
		_ = os.Remove(filepath.Join(m.configDir, oldStore))
	}
	return nil
}

// AddProfile adds or updates a named profile
func (m *Manager) AddProfile(name, url string) error {
	return m.AddProfileWithKey(name, url, "")
}

// AddProfileWithKey adds or updates a named profile with an optional API key
func (m *Manager) AddProfileWithKey(name, url, apiKey string) error {
	return m.AddRESTProfile(name, url, apiKey, "", "", "")
}

// AddRESTProfile adds or updates a REST API profile.
func (m *Manager) AddRESTProfile(name, url, apiKey, accessMode, permission, documentScope string) error {
	cfg, err := m.Load()
	if err != nil {
		return err
	}

	if accessMode == "" {
		if apiKey != "" {
			accessMode = "api-key"
		} else {
			accessMode = "public"
		}
	}

	cfg.Profiles[name] = Profile{
		Type:          "rest",
		URL:           url,
		APIKey:        apiKey,
		AccessMode:    accessMode,
		Permission:    permission,
		DocumentScope: documentScope,
	}

	// If this is the first profile, make it active
	if len(cfg.Profiles) == 1 {
		cfg.ActiveProfile = name
	}

	return m.Save(cfg)
}

// AddMCPProfile adds or updates a Craft MCP profile.
func (m *Manager) AddMCPProfile(name, mcpURL, accessMode, permission, documentScope string) error {
	cfg, err := m.Load()
	if err != nil {
		return err
	}

	if accessMode == "" {
		accessMode = "public"
	}
	if documentScope == "" {
		documentScope = "connection-defined"
	}

	cfg.Profiles[name] = Profile{
		Type:          "mcp",
		MCPURL:        mcpURL,
		AccessMode:    accessMode,
		Permission:    permission,
		DocumentScope: documentScope,
	}

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

	profiles := make([]ProfileInfo, 0)
	for name, profile := range cfg.Profiles {
		profiles = append(profiles, ProfileInfo{
			Name:          name,
			Type:          profile.TypeOrDefault(),
			URL:           profile.URL,
			MCPURL:        profile.MCPURL,
			Active:        name == cfg.ActiveProfile,
			HasAPIKey:     profile.APIKey != "",
			AccessMode:    profile.AccessMode,
			Permission:    profile.Permission,
			DocumentScope: profile.DocumentScope,
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
	Name          string `json:"name"`
	Type          string `json:"type"`
	URL           string `json:"url,omitempty"`
	MCPURL        string `json:"mcp_url,omitempty"`
	Active        bool   `json:"active"`
	HasAPIKey     bool   `json:"has_api_key"`
	AccessMode    string `json:"access_mode,omitempty"`
	Permission    string `json:"permission,omitempty"`
	DocumentScope string `json:"document_scope,omitempty"`
}

// TypeOrDefault returns the profile type, defaulting legacy profiles to REST.
func (p Profile) TypeOrDefault() string {
	if p.Type == "" {
		return "rest"
	}
	return p.Type
}

// GetProfile returns a named profile.
func (m *Manager) GetProfile(name string) (Profile, error) {
	cfg, err := m.Load()
	if err != nil {
		return Profile{}, err
	}
	profile, exists := cfg.Profiles[name]
	if !exists {
		return Profile{}, fmt.Errorf("profile '%s' not found", name)
	}
	return profile, nil
}

// GetActiveProfile returns the active profile.
func (m *Manager) GetActiveProfile() (Profile, error) {
	cfg, err := m.Load()
	if err != nil {
		return Profile{}, err
	}
	if cfg.ActiveProfile == "" {
		return Profile{}, fmt.Errorf("no active profile. Run 'craft config add <name> <url>' first")
	}
	profile, exists := cfg.Profiles[cfg.ActiveProfile]
	if !exists {
		return Profile{}, fmt.Errorf("active profile '%s' not found. Run 'craft config add <name> <url>' first", cfg.ActiveProfile)
	}
	return profile, nil
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
	if profile.TypeOrDefault() != "rest" {
		return "", fmt.Errorf("active profile '%s' is not a REST profile", cfg.ActiveProfile)
	}

	return profile.URL, nil
}

// GetActiveAPIKey returns the API key of the active profile (may be empty)
func (m *Manager) GetActiveAPIKey() (string, error) {
	cfg, err := m.Load()
	if err != nil {
		return "", err
	}

	if cfg.ActiveProfile == "" {
		return "", nil
	}

	profile, exists := cfg.Profiles[cfg.ActiveProfile]
	if !exists {
		return "", nil
	}
	if profile.TypeOrDefault() != "rest" {
		return "", nil
	}

	return profile.APIKey, nil
}

// Reset clears the configuration
func (m *Manager) Reset() error {
	paths, err := filepath.Glob(filepath.Join(m.configDir, "credentials-*.json"))
	if err != nil {
		return err
	}
	paths = append(paths, m.configPath)
	for _, path := range paths {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove configuration: %w", err)
		}
	}
	return nil
}
