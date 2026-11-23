package cmd

import (
	"testing"

	"github.com/spf13/pflag"
)

func resetLinksFlags(t *testing.T) {
	t.Helper()
	linksCmd.Flags().VisitAll(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	})
}

func TestBuildLinksQuery(t *testing.T) {
	defer resetLinksFlags(t)

	for _, tc := range []struct {
		name  string
		flags map[string]string
		want  string
	}{
		{"defaults", nil, "level=document&limit=50"},
		{"backlinks", map[string]string{"to": "doc-1", "from-date": "yesterday", "include-self": "true", "limit": "5"},
			"fromDate=yesterday&includeSelfLinks=true&level=document&limit=5&toDocumentId=doc-1"},
		{"block level with cursor", map[string]string{"level": "block", "cursor": "doc-a:doc-b"},
			"cursor=doc-a%3Adoc-b&level=block&limit=50"},
	} {
		resetLinksFlags(t)
		for name, value := range tc.flags {
			if err := linksCmd.Flags().Set(name, value); err != nil {
				t.Fatal(err)
			}
		}
		q, err := buildLinksQuery(linksCmd)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if got := q.Encode(); got != tc.want {
			t.Errorf("%s: got %s, want %s", tc.name, got, tc.want)
		}
	}
}

func TestBuildLinksQueryRejectsInvalidInput(t *testing.T) {
	defer resetLinksFlags(t)

	for name, flags := range map[string]map[string]string{
		"from and from-date": {"from": "doc-1", "from-date": "today"},
		"to and to-date":     {"to": "doc-1", "to-date": "today"},
		"bad date":           {"to-date": "next week"},
		"bad level":          {"level": "paragraph"},
		"limit too high":     {"limit": "201"},
		"limit too low":      {"limit": "0"},
		"traversal id":       {"from": "../x"},
	} {
		resetLinksFlags(t)
		for flag, value := range flags {
			if err := linksCmd.Flags().Set(flag, value); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := buildLinksQuery(linksCmd); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}
