package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCredentialMigrationAndRotation(t *testing.T) {
	dir := t.TempDir()
	m := &Manager{configDir: dir, configPath: filepath.Join(dir, ConfigFileName)}
	legacy := `{"active_profile":"work","profiles":{"work":{"url":"https://example.org/private-link","api_key":"fixture-key"}}}`
	if err := os.WriteFile(m.configPath, []byte(legacy), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := m.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err = m.Save(cfg); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(m.configPath)
	if strings.Contains(string(data), "fixture-key") || strings.Contains(string(data), "private-link") {
		t.Fatal("secret in metadata")
	}
	first := cfg.CredentialStore
	info, err := os.Stat(filepath.Join(dir, first))
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("credential permissions")
	}
	loaded, err := m.Load()
	if err != nil || loaded.Profiles["work"].APIKey != "fixture-key" {
		t.Fatal("credential round trip", err)
	}
	if err = m.Save(loaded); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(filepath.Join(dir, first)); !os.IsNotExist(err) {
		t.Fatal("old generation retained")
	}
	if err = m.RemoveProfile("work"); err != nil {
		t.Fatal(err)
	}
	_, path := m.Paths()
	data, _ = os.ReadFile(path)
	var secrets map[string]credential
	if err = json.Unmarshal(data, &secrets); err != nil || len(secrets) != 0 {
		t.Fatal("removed credential retained")
	}
}

func TestCredentialStoreRejectsTraversal(t *testing.T) {
	dir := t.TempDir()
	m := &Manager{configDir: dir, configPath: filepath.Join(dir, ConfigFileName)}
	os.WriteFile(m.configPath, []byte(`{"credential_store":"../credentials-stolen.json"}`), 0600)
	if _, err := m.Load(); err == nil {
		t.Fatal("accepted traversal")
	}
}

func TestFailedConfigPublicationPreservesCredentials(t *testing.T) {
	dir := t.TempDir()
	m := &Manager{configDir: dir, configPath: filepath.Join(dir, ConfigFileName)}
	if err := m.AddProfileWithKey("work", "https://example.org/old", "old-key"); err != nil {
		t.Fatal(err)
	}
	cfg, _ := m.Load()
	old := cfg.CredentialStore
	// A directory at the metadata destination makes atomic publication fail.
	os.Rename(m.configPath, m.configPath+".backup")
	os.Mkdir(m.configPath, 0700)
	p := cfg.Profiles["work"]
	p.APIKey = "new-key"
	cfg.Profiles["work"] = p
	if err := m.Save(cfg); err == nil {
		t.Fatal("expected failed publication")
	}
	data, err := os.ReadFile(filepath.Join(dir, old))
	if err != nil || !strings.Contains(string(data), "old-key") {
		t.Fatal("old generation lost")
	}
}

func TestResetRecoversCorruptMetadataAndRemovesCredentialGenerations(t *testing.T) {
	dir := t.TempDir()
	m := &Manager{configDir: dir, configPath: filepath.Join(dir, ConfigFileName)}
	os.WriteFile(m.configPath, []byte("invalid"), 0600)
	secret := filepath.Join(dir, "credentials-fixture.json")
	os.WriteFile(secret, []byte(`{}`), 0600)
	if err := m.Reset(); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{m.configPath, secret} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Fatal("reset retained private configuration")
		}
	}
}

func TestMissingCredentialGenerationFailsInsteadOfDroppingAuthentication(t *testing.T) {
	dir := t.TempDir()
	m := &Manager{configDir: dir, configPath: filepath.Join(dir, ConfigFileName)}
	if err := m.AddProfileWithKey("work", "https://example.org/link", "fixture-key"); err != nil {
		t.Fatal(err)
	}
	_, path := m.Paths()
	os.Remove(path)
	if _, err := m.Load(); err == nil || !strings.Contains(err.Error(), "credential store is missing") {
		t.Fatal("missing credentials silently ignored", err)
	}
}
