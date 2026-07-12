package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchRecentReleasesSkipsDrafts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/nerveband/craft-cli/releases" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("per_page"); got != "2" {
			t.Fatalf("per_page = %q, want 2", got)
		}
		fmt.Fprint(w, `[
			{"tag_name":"v1.11.3","name":"v1.11.3","body":"new","html_url":"https://example.com/1","draft":true},
			{"tag_name":"v1.11.2","name":"v1.11.2","body":"current","html_url":"https://example.com/2","draft":false},
			{"tag_name":"v1.11.1","name":"v1.11.1","body":"previous","html_url":"https://example.com/3","draft":false}
		]`)
	}))
	defer server.Close()

	oldBase := githubAPIBaseURL
	githubAPIBaseURL = server.URL
	t.Cleanup(func() {
		githubAPIBaseURL = oldBase
	})

	releases, err := fetchRecentReleases(context.Background(), repoOwner, repoName, 2)
	if err != nil {
		t.Fatalf("fetchRecentReleases() error = %v", err)
	}
	if len(releases) != 2 {
		t.Fatalf("len(releases) = %d, want 2", len(releases))
	}
	if releases[0].TagName != "v1.11.2" || releases[1].TagName != "v1.11.1" {
		t.Fatalf("unexpected releases: %#v", releases)
	}
}

func TestSummarizeReleaseBody(t *testing.T) {
	body := `## What's Changed

* Add image diagnostics
* Fix tasks default

` + "```" + `
ignored code
` + "```" + `
* Keep this
`
	lines := summarizeReleaseBody(body, 3)
	want := []string{"What's Changed", "* Add image diagnostics", "* Fix tasks default"}
	if len(lines) != len(want) {
		t.Fatalf("len(lines) = %d, want %d: %#v", len(lines), len(want), lines)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Fatalf("lines[%d] = %q, want %q", i, lines[i], want[i])
		}
	}
}
