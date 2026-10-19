package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/spf13/cobra"
)

const repoOwner = "nerveband"
const repoName = "craft-cli"
const upgradeChangelogLimit = 4

var githubAPIBaseURL = "https://api.github.com"

type githubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Body    string `json:"body"`
	HTMLURL string `json:"html_url"`
	Draft   bool   `json:"draft"`
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
