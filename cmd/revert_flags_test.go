package cmd

import "testing"

func TestCollectionsWriteCmdsRegisterRevertFlags(t *testing.T) {
	targets := []struct {
		name string
		has  func(flag string) bool
	}{
		{"collections create", func(f string) bool { return collectionsCreateCmd.Flags().Lookup(f) != nil }},
		{"collections rename", func(f string) bool { return collectionsRenameCmd.Flags().Lookup(f) != nil }},
		{"collections items add", func(f string) bool { return collectionsAddCmd.Flags().Lookup(f) != nil }},
		{"collections items update", func(f string) bool { return collectionsUpdateCmd.Flags().Lookup(f) != nil }},
		{"collections views create", func(f string) bool { return collectionsViewsCreateCmd.Flags().Lookup(f) != nil }},
		{"collections views update", func(f string) bool { return collectionsViewsUpdateCmd.Flags().Lookup(f) != nil }},
		{"collections views delete", func(f string) bool { return collectionsViewsDeleteCmd.Flags().Lookup(f) != nil }},
		{"collections active-view set", func(f string) bool { return collectionsActiveViewSetCmd.Flags().Lookup(f) != nil }},
	}
	for _, tc := range targets {
		if !tc.has("save-revert") {
			t.Errorf("%s missing --save-revert", tc.name)
		}
		if !tc.has("diff") {
			t.Errorf("%s missing --diff", tc.name)
		}
	}
}
