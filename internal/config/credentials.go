package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Credential URLs carry connection capabilities and belong with API keys.
type credential struct {
	URL    string `json:"url,omitempty"`
	MCPURL string `json:"mcp_url,omitempty"`
	APIKey string `json:"api_key,omitempty"`
}

func (m *Manager) readCredentials(name string) (map[string]credential, error) {
	result := map[string]credential{}
	if name == "" {
		return result, nil
	}
	if filepath.Base(name) != name || len(name) < len("credentials-.json") || name[:12] != "credentials-" || filepath.Ext(name) != ".json" {
		return nil, fmt.Errorf("invalid configuration credential store filename")
	}
	data, err := os.ReadFile(filepath.Join(m.configDir, name))
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration credential store is missing: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("read configuration credential store: %w", err)
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parse configuration credential store: %w", err)
	}
	if result == nil {
		return nil, fmt.Errorf("configuration credential store must be a JSON object")
	}
	return result, nil
}

func writePrivateJSON(path string, value interface{}) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".craft-config-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(append(data, '\n')); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}

// Paths reports locations only, never their contents.
func (m *Manager) Paths() (string, string) {
	cfg, err := m.Load()
	if err != nil || cfg.CredentialStore == "" {
		return m.configPath, ""
	}
	return m.configPath, filepath.Join(m.configDir, cfg.CredentialStore)
}

// DefaultFormatSource distinguishes an explicit file setting from Load defaults.
func (m *Manager) DefaultFormatSource() string {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return "default"
	}
	var metadata struct {
		DefaultFormat string `json:"default_format"`
	}
	if json.Unmarshal(data, &metadata) == nil && metadata.DefaultFormat != "" {
		return "config:default_format"
	}
	return "default"
}
