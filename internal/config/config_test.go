package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	if mgr.configDir == "" {
		t.Error("configDir should not be empty")
	}

	if mgr.configPath == "" {
		t.Error("configPath should not be empty")
	}
}

func TestManager_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()

	mgr := &Manager{
		configDir:  tmpDir,
		configPath: filepath.Join(tmpDir, ConfigFileName),
	}

	cfg := &Config{
		DefaultFormat: "json",
		ActiveProfile: "work",
		Profiles: map[string]Profile{
			"work": {URL: "https://connect.craft.do/links/work/api/v1"},
		},
	}

	err := mgr.Save(cfg)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loadedCfg, err := mgr.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loadedCfg.ActiveProfile != cfg.ActiveProfile {
		t.Errorf("ActiveProfile = %v, want %v", loadedCfg.ActiveProfile, cfg.ActiveProfile)
	}

	if loadedCfg.DefaultFormat != cfg.DefaultFormat {
		t.Errorf("DefaultFormat = %v, want %v", loadedCfg.DefaultFormat, cfg.DefaultFormat)
	}

	if loadedCfg.Profiles["work"].URL != cfg.Profiles["work"].URL {
		t.Errorf("Profile URL = %v, want %v", loadedCfg.Profiles["work"].URL, cfg.Profiles["work"].URL)
	}
}

func TestManager_LoadNonExistent(t *testing.T) {
	tmpDir := t.TempDir()

	mgr := &Manager{
		configDir:  tmpDir,
		configPath: filepath.Join(tmpDir, ConfigFileName),
	}

	cfg, err := mgr.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.DefaultFormat != "json" {
		t.Errorf("DefaultFormat = %v, want json", cfg.DefaultFormat)
	}

	if cfg.Profiles == nil {
		t.Error("Profiles should be initialized")
	}
}

func TestManager_AddProfile(t *testing.T) {
	tmpDir := t.TempDir()

	mgr := &Manager{
		configDir:  tmpDir,
		configPath: filepath.Join(tmpDir, ConfigFileName),
	}

	testURL := "https://connect.craft.do/links/test/api/v1"
	err := mgr.AddProfile("work", testURL)
	if err != nil {
		t.Fatalf("AddProfile() error = %v", err)
	}

	// First profile should be set as active
	cfg, err := mgr.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.ActiveProfile != "work" {
		t.Errorf("ActiveProfile = %v, want work", cfg.ActiveProfile)
	}

	if cfg.Profiles["work"].URL != testURL {
		t.Errorf("Profile URL = %v, want %v", cfg.Profiles["work"].URL, testURL)
	}
	if cfg.Profiles["work"].TypeOrDefault() != "rest" {
		t.Errorf("Profile type = %v, want rest", cfg.Profiles["work"].TypeOrDefault())
	}
}

func TestManager_AddTypedProfiles(t *testing.T) {
	tmpDir := t.TempDir()

	mgr := &Manager{
		configDir:  tmpDir,
		configPath: filepath.Join(tmpDir, ConfigFileName),
	}

	err := mgr.AddRESTProfile("work", "https://connect.craft.do/links/work/api/v1", "pdk_test", "api-key", "read-write", "all-documents")
	if err != nil {
		t.Fatalf("AddRESTProfile() error = %v", err)
	}
	err = mgr.AddMCPProfile("mcp", "https://mcp.craft.do/links/work/mcp", "public", "read-write", "connection-defined")
	if err != nil {
		t.Fatalf("AddMCPProfile() error = %v", err)
	}

	rest, err := mgr.GetProfile("work")
	if err != nil {
		t.Fatalf("GetProfile(work) error = %v", err)
	}
	if rest.TypeOrDefault() != "rest" {
		t.Errorf("REST profile type = %v, want rest", rest.TypeOrDefault())
	}
	if rest.APIKey != "pdk_test" {
		t.Errorf("REST API key not preserved")
	}
	if rest.Permission != "read-write" {
		t.Errorf("REST permission = %v, want read-write", rest.Permission)
	}

	mcp, err := mgr.GetProfile("mcp")
	if err != nil {
		t.Fatalf("GetProfile(mcp) error = %v", err)
	}
	if mcp.TypeOrDefault() != "mcp" {
		t.Errorf("MCP profile type = %v, want mcp", mcp.TypeOrDefault())
	}
	if mcp.MCPURL != "https://mcp.craft.do/links/work/mcp" {
		t.Errorf("MCP URL = %v", mcp.MCPURL)
	}

	profiles, err := mgr.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles() error = %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("Expected 2 profiles, got %d", len(profiles))
	}
	foundMCP := false
	for _, p := range profiles {
		if p.Name == "mcp" {
			foundMCP = true
			if p.Type != "mcp" {
				t.Errorf("ProfileInfo type = %v, want mcp", p.Type)
			}
			if p.MCPURL == "" {
				t.Errorf("ProfileInfo MCPURL should be set")
			}
		}
	}
	if !foundMCP {
		t.Fatal("Expected MCP profile in list")
	}
}

func TestManager_AddMultipleProfiles(t *testing.T) {
	tmpDir := t.TempDir()

	mgr := &Manager{
		configDir:  tmpDir,
		configPath: filepath.Join(tmpDir, ConfigFileName),
	}

	mgr.AddProfile("work", "https://work.example.com")
	mgr.AddProfile("personal", "https://personal.example.com")

	cfg, err := mgr.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// First profile should still be active
	if cfg.ActiveProfile != "work" {
		t.Errorf("ActiveProfile = %v, want work", cfg.ActiveProfile)
	}

	if len(cfg.Profiles) != 2 {
		t.Errorf("Expected 2 profiles, got %d", len(cfg.Profiles))
	}
}

func TestManager_UseProfile(t *testing.T) {
	tmpDir := t.TempDir()

	mgr := &Manager{
		configDir:  tmpDir,
		configPath: filepath.Join(tmpDir, ConfigFileName),
	}

	mgr.AddProfile("work", "https://work.example.com")
	mgr.AddProfile("personal", "https://personal.example.com")

	err := mgr.UseProfile("personal")
	if err != nil {
		t.Fatalf("UseProfile() error = %v", err)
	}

	cfg, err := mgr.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.ActiveProfile != "personal" {
		t.Errorf("ActiveProfile = %v, want personal", cfg.ActiveProfile)
	}
}

func TestManager_UseProfileNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	mgr := &Manager{
		configDir:  tmpDir,
		configPath: filepath.Join(tmpDir, ConfigFileName),
	}

	err := mgr.UseProfile("nonexistent")
	if err == nil {
		t.Error("UseProfile() should error for nonexistent profile")
	}
}

func TestManager_RemoveProfile(t *testing.T) {
	tmpDir := t.TempDir()

	mgr := &Manager{
		configDir:  tmpDir,
		configPath: filepath.Join(tmpDir, ConfigFileName),
	}

	mgr.AddProfile("work", "https://work.example.com")
	mgr.AddProfile("personal", "https://personal.example.com")

	err := mgr.RemoveProfile("work")
	if err != nil {
		t.Fatalf("RemoveProfile() error = %v", err)
	}

	cfg, err := mgr.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if _, exists := cfg.Profiles["work"]; exists {
		t.Error("Profile 'work' should be removed")
	}

	// Active profile should be cleared since we removed it
	if cfg.ActiveProfile != "" {
		t.Errorf("ActiveProfile should be empty, got %v", cfg.ActiveProfile)
	}
}

func TestManager_RemoveProfileNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	mgr := &Manager{
		configDir:  tmpDir,
		configPath: filepath.Join(tmpDir, ConfigFileName),
	}

	err := mgr.RemoveProfile("nonexistent")
	if err == nil {
		t.Error("RemoveProfile() should error for nonexistent profile")
	}
}

func TestManager_ListProfiles(t *testing.T) {
	tmpDir := t.TempDir()

	mgr := &Manager{
		configDir:  tmpDir,
		configPath: filepath.Join(tmpDir, ConfigFileName),
	}

	mgr.AddProfile("work", "https://work.example.com")
	mgr.AddProfile("personal", "https://personal.example.com")

	profiles, err := mgr.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles() error = %v", err)
	}

	if len(profiles) != 2 {
		t.Errorf("Expected 2 profiles, got %d", len(profiles))
	}

	// Should be sorted alphabetically
	if profiles[0].Name != "personal" {
		t.Errorf("First profile should be 'personal', got %v", profiles[0].Name)
	}

	// Work should be active (first added)
	for _, p := range profiles {
		if p.Name == "work" && !p.Active {
			t.Error("'work' profile should be active")
		}
	}
}

func TestManager_GetActiveURL(t *testing.T) {
	tmpDir := t.TempDir()

	mgr := &Manager{
		configDir:  tmpDir,
		configPath: filepath.Join(tmpDir, ConfigFileName),
	}

	testURL := "https://connect.craft.do/links/test/api/v1"
	mgr.AddProfile("work", testURL)

	url, err := mgr.GetActiveURL()
	if err != nil {
		t.Fatalf("GetActiveURL() error = %v", err)
	}

	if url != testURL {
		t.Errorf("GetActiveURL() = %v, want %v", url, testURL)
	}
}

func TestManager_GetActiveURLNoProfile(t *testing.T) {
	tmpDir := t.TempDir()

	mgr := &Manager{
		configDir:  tmpDir,
		configPath: filepath.Join(tmpDir, ConfigFileName),
	}

	_, err := mgr.GetActiveURL()
	if err == nil {
		t.Error("GetActiveURL() should error when no profile is set")
	}
}

func TestManager_Reset(t *testing.T) {
	tmpDir := t.TempDir()

	mgr := &Manager{
		configDir:  tmpDir,
		configPath: filepath.Join(tmpDir, ConfigFileName),
	}

	mgr.AddProfile("work", "https://work.example.com")

	err := mgr.Reset()
	if err != nil {
		t.Fatalf("Reset() error = %v", err)
	}

	_, err = os.Stat(mgr.configPath)
	if !os.IsNotExist(err) {
		t.Error("config file should not exist after reset")
	}
}
