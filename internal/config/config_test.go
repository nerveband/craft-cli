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
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	
	mgr := &Manager{
		configDir:  tmpDir,
		configPath: filepath.Join(tmpDir, ConfigFileName),
	}

	// Test saving a new config
	cfg := &Config{
		APIURL:        "https://connect.craft.do/links/test/api/v1",
		DefaultFormat: "json",
	}

	err := mgr.Save(cfg)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Test loading the config
	loadedCfg, err := mgr.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loadedCfg.APIURL != cfg.APIURL {
		t.Errorf("APIURL = %v, want %v", loadedCfg.APIURL, cfg.APIURL)
	}

	if loadedCfg.DefaultFormat != cfg.DefaultFormat {
		t.Errorf("DefaultFormat = %v, want %v", loadedCfg.DefaultFormat, cfg.DefaultFormat)
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

	// Should return default config when file doesn't exist
	if cfg.DefaultFormat != "json" {
		t.Errorf("DefaultFormat = %v, want json", cfg.DefaultFormat)
	}
}

func TestManager_SetAPIURL(t *testing.T) {
	tmpDir := t.TempDir()
	
	mgr := &Manager{
		configDir:  tmpDir,
		configPath: filepath.Join(tmpDir, ConfigFileName),
	}

	testURL := "https://connect.craft.do/links/test/api/v1"
	err := mgr.SetAPIURL(testURL)
	if err != nil {
		t.Fatalf("SetAPIURL() error = %v", err)
	}

	url, err := mgr.GetAPIURL()
	if err != nil {
		t.Fatalf("GetAPIURL() error = %v", err)
	}

	if url != testURL {
		t.Errorf("GetAPIURL() = %v, want %v", url, testURL)
	}
}

func TestManager_Reset(t *testing.T) {
	tmpDir := t.TempDir()
	
	mgr := &Manager{
		configDir:  tmpDir,
		configPath: filepath.Join(tmpDir, ConfigFileName),
	}

	// Create a config file
	cfg := &Config{
		APIURL:        "https://connect.craft.do/links/test/api/v1",
		DefaultFormat: "json",
	}

	err := mgr.Save(cfg)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Reset the config
	err = mgr.Reset()
	if err != nil {
		t.Fatalf("Reset() error = %v", err)
	}

	// Verify the file is gone
	_, err = os.Stat(mgr.configPath)
	if !os.IsNotExist(err) {
		t.Error("config file should not exist after reset")
	}
}
