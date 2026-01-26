package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ashrafali/craft-cli/internal/config"
	"github.com/creativeprojects/go-selfupdate"
	"github.com/spf13/cobra"
)

const repoOwner = "nerveband"
const repoName = "craft-cli"

// updateCheckCache stores the last version check
type updateCheckCache struct {
	LastCheck      time.Time `json:"last_check"`
	LatestVersion  string    `json:"latest_version"`
	UpdateRequired bool      `json:"update_required"`
}

// checkForUpdates checks if a new version is available (without installing)
func checkForUpdates() (hasUpdate bool, latestVersion string, err error) {
	// Check cache first (only check once per day)
	cached, err := loadUpdateCache()
	if err == nil && time.Since(cached.LastCheck) < 24*time.Hour {
		return cached.UpdateRequired, cached.LatestVersion, nil
	}

	source, err := selfupdate.NewGitHubSource(selfupdate.GitHubConfig{})
	if err != nil {
		return false, "", err
	}

	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Source:    source,
		Validator: &selfupdate.ChecksumValidator{UniqueFilename: "checksums.txt"},
	})
	if err != nil {
		return false, "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	latest, found, err := updater.DetectLatest(ctx, selfupdate.NewRepositorySlug(repoOwner, repoName))
	if err != nil || !found {
		return false, "", err
	}

	hasUpdate = latest.GreaterThan(version)
	latestVer := latest.Version()

	// Save to cache
	saveUpdateCache(updateCheckCache{
		LastCheck:      time.Now(),
		LatestVersion:  latestVer,
		UpdateRequired: hasUpdate,
	})

	return hasUpdate, latestVer, nil
}

// notifyUpdateAvailable shows update notification if not in quiet mode
func notifyUpdateAvailable() {
	if quietMode {
		return
	}

	hasUpdate, latestVersion, err := checkForUpdates()
	if err != nil || !hasUpdate {
		return
	}

	fmt.Fprintf(os.Stderr, "\n🎁 New version available: %s (current: %s)\n", latestVersion, version)
	fmt.Fprintf(os.Stderr, "Run 'craft upgrade' to update\n\n")
}

// loadUpdateCache loads the cached update check result
func loadUpdateCache() (*updateCheckCache, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	cachePath := filepath.Join(homeDir, config.ConfigDirName, "update_cache.json")
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}

	var cache updateCheckCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	return &cache, nil
}

// saveUpdateCache saves the update check result to cache
func saveUpdateCache(cache updateCheckCache) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, config.ConfigDirName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	cachePath := filepath.Join(configDir, "update_cache.json")
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cachePath, data, 0644)
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade craft-cli to the latest version",
	Long:  "Check for and install the latest version of craft-cli from GitHub releases",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUpgrade()
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}

func runUpgrade() error {
	fmt.Printf("Current version: %s\n", version)
	fmt.Printf("Checking for updates...\n")

	source, err := selfupdate.NewGitHubSource(selfupdate.GitHubConfig{})
	if err != nil {
		return fmt.Errorf("failed to create update source: %w", err)
	}

	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Source:    source,
		Validator: &selfupdate.ChecksumValidator{UniqueFilename: "checksums.txt"},
	})
	if err != nil {
		return fmt.Errorf("failed to create updater: %w", err)
	}

	latest, found, err := updater.DetectLatest(context.Background(), selfupdate.NewRepositorySlug(repoOwner, repoName))
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !found {
		fmt.Println("No releases found")
		return nil
	}

	if latest.LessOrEqual(version) {
		fmt.Printf("Already up to date (latest: %s)\n", latest.Version())
		return nil
	}

	fmt.Printf("New version available: %s\n", latest.Version())
	fmt.Printf("Downloading for %s/%s...\n", runtime.GOOS, runtime.GOARCH)

	exe, err := selfupdate.ExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	if err := updater.UpdateTo(context.Background(), latest, exe); err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	fmt.Printf("Successfully upgraded to %s\n", latest.Version())
	return nil
}
