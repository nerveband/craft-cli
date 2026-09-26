package cmd

import (
	"github.com/ashrafali/craft-cli/internal/config"
	"net/url"
	"strings"
)

func redactedURL(value string) string {
	if value == "" {
		return ""
	}
	u, err := url.Parse(value)
	if err != nil {
		return "[redacted URL]"
	}
	u.User = nil
	u.RawQuery = ""
	u.Fragment = ""
	parts := strings.Split(u.Path, "/")
	for i := 1; i < len(parts); i++ {
		if parts[i-1] == "links" || parts[i-1] == "link" {
			parts[i] = "REDACTED"
		}
	}
	u.Path = strings.Join(parts, "/")
	return u.String()
}
func safeProfiles(items []config.ProfileInfo) []config.ProfileInfo {
	for i := range items {
		items[i].URL = redactedURL(items[i].URL)
		items[i].MCPURL = redactedURL(items[i].MCPURL)
	}
	return items
}
