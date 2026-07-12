package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/ashrafali/craft-cli/internal/config"
	"github.com/creativeprojects/go-selfupdate"
	"github.com/spf13/cobra"
)

const repoOwner = "nerveband"
const repoName = "craft-cli"
const upgradeChangelogLimit = 4

var githubAPIBaseURL = "https://api.github.com"

// updateCheckCache stores the last version check
type updateCheckCache struct {
	LastCheck      time.Time `json:"last_check"`
	LatestVersion  string    `json:"latest_version"`
	UpdateRequired bool      `json:"update_required"`
	LastNotified   time.Time `json:"last_notified"`
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Body    string `json:"body"`
	HTMLURL string `json:"html_url"`
	Draft   bool   `json:"draft"`
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

// notifyUpdateAvailable shows update notification at most once per day
func notifyUpdateAvailable() {
	if quietMode {
		return
	}

	// Check if we already notified recently (once per day max)
	cached, cacheErr := loadUpdateCache()
	if cacheErr == nil && !cached.LastNotified.IsZero() && time.Since(cached.LastNotified) < 24*time.Hour {
		return
	}

	hasUpdate, latestVersion, err := checkForUpdates()
	if err != nil || !hasUpdate {
		return
	}

	fmt.Fprintf(os.Stderr, "\nNew version available: %s (current: %s)\n", latestVersion, version)
	fmt.Fprintf(os.Stderr, "Run 'craft upgrade' to update\n\n")

	// Mark that we showed the notification
	if cached == nil {
		cached = &updateCheckCache{}
	}
	cached.LastNotified = time.Now()
	saveUpdateCache(*cached)
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
	if err := printRecentChangelog(context.Background(), version, latest.Version(), upgradeChangelogLimit); err != nil {
		fmt.Printf("\nCould not fetch changelog: %v\n", err)
	}
	return nil
}

func printRecentChangelog(ctx context.Context, previousVersion, currentVersion string, limit int) error {
	releases, err := fetchRecentReleases(ctx, repoOwner, repoName, limit)
	if err != nil {
		return err
	}
	if len(releases) == 0 {
		return nil
	}

	fmt.Printf("\nRecent changes:\n")
	for _, release := range releases {
		tag := strings.TrimPrefix(release.TagName, "v")
		marker := " "
		switch tag {
		case currentVersion:
			marker = "*"
		case previousVersion:
			marker = "-"
		}

		title := release.Name
		if title == "" {
			title = release.TagName
		}
		fmt.Printf("\n%s %s\n", marker, title)
		if tag == currentVersion {
			fmt.Printf("  installed now\n")
		} else if tag == previousVersion {
			fmt.Printf("  previous version\n")
		}
		for _, line := range summarizeReleaseBody(release.Body, 6) {
			fmt.Printf("  %s\n", line)
		}
		if release.HTMLURL != "" {
			fmt.Printf("  %s\n", release.HTMLURL)
		}
	}
	fmt.Printf("\nLegend: * installed version, - previous version\n")
	return nil
}

func fetchRecentReleases(ctx context.Context, owner, name string, limit int) ([]githubRelease, error) {
	if limit <= 0 {
		return nil, nil
	}
	url := fmt.Sprintf("%s/repos/%s/%s/releases?per_page=%d", strings.TrimRight(githubAPIBaseURL, "/"), owner, name, limit)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "craft-cli/"+version)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("GitHub releases request failed: %s %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var releases []githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	filtered := make([]githubRelease, 0, len(releases))
	for _, release := range releases {
		if release.Draft {
			continue
		}
		filtered = append(filtered, release)
		if len(filtered) == limit {
			break
		}
	}
	return filtered, nil
}

func summarizeReleaseBody(body string, maxLines int) []string {
	body = strings.ReplaceAll(body, "\r\n", "\n")
	lines := strings.Split(body, "\n")
	summary := make([]string, 0, maxLines)
	inFence := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "```") {
			inFence = !inFence
			continue
		}
		if line == "" || inFence {
			continue
		}
		if strings.HasPrefix(line, "<!--") {
			continue
		}
		line = strings.TrimLeft(line, "#")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		summary = append(summary, line)
		if len(summary) == maxLines {
			break
		}
	}
	if len(summary) == 0 && strings.TrimSpace(body) != "" {
		summary = append(summary, truncateForChangelog(strings.TrimSpace(body), 120))
	}
	return summary
}

func truncateForChangelog(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	if limit <= 3 {
		return value[:limit]
	}
	return value[:limit-3] + "..."
}
