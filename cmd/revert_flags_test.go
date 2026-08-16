package cmd

import (
	"errors"
	"testing"
)
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

func TestUpdateCmdRegistersRevertFlags(t *testing.T) {
	if updateCmd.Flags().Lookup("save-revert") == nil {
		t.Error("update missing --save-revert")
	}
	if updateCmd.Flags().Lookup("diff") == nil {
		t.Error("update missing --diff")
	}
}

func TestUpdateRevertFlagsRejectRESTBackend(t *testing.T) {
	origBackend, origSave, origDiff := backendName, updateSaveRevert, updateDiff
	defer func() { backendName, updateSaveRevert, updateDiff = origBackend, origSave, origDiff }()

	backendName = "rest"
	updateSaveRevert = "revert.json"
	updateDiff = false

	err := updateCmd.RunE(updateCmd, []string{"abc123"})
	if err == nil {
		t.Fatal("expected CAPABILITY_UNAVAILABLE error on --backend rest")
	}
	var cerr *cliCodeError
	if !errors.As(err, &cerr) || cerr.Code != "CAPABILITY_UNAVAILABLE" {
		t.Fatalf("expected cliCodeError CAPABILITY_UNAVAILABLE, got %v", err)
	}
}

func TestUpdateRevertFlagsRejectContentUpdates(t *testing.T) {
	origBackend, origSave, origDiff := backendName, updateSaveRevert, updateDiff
	origMarkdown, origTitle := updateMarkdown, updateTitle
	defer func() {
		backendName, updateSaveRevert, updateDiff = origBackend, origSave, origDiff
		updateMarkdown, updateTitle = origMarkdown, origTitle
	}()

	backendName = "auto"
	updateSaveRevert = "revert.json"
	updateDiff = false
	updateTitle = ""
	updateMarkdown = "# new body"

	err := updateCmd.RunE(updateCmd, []string{"abc123"})
	if err == nil {
		t.Fatal("expected CAPABILITY_UNAVAILABLE error for content update with --save-revert")
	}
	var cerr *cliCodeError
	if !errors.As(err, &cerr) || cerr.Code != "CAPABILITY_UNAVAILABLE" {
		t.Fatalf("expected cliCodeError CAPABILITY_UNAVAILABLE, got %v", err)
	}
}

func TestTasksCmdsRegisterRevertFlags(t *testing.T) {
	for _, tc := range []struct {
		name string
		has  func(flag string) bool
	}{
		{"tasks add", func(f string) bool { return tasksAddCmd.Flags().Lookup(f) != nil }},
		{"tasks update", func(f string) bool { return tasksUpdateCmd.Flags().Lookup(f) != nil }},
		{"tasks delete", func(f string) bool { return tasksDeleteCmd.Flags().Lookup(f) != nil }},
	} {
		if !tc.has("save-revert") {
			t.Errorf("%s missing --save-revert", tc.name)
		}
		if !tc.has("diff") {
			t.Errorf("%s missing --diff", tc.name)
		}
	}
}

func TestTasksRevertFlagsReturnCapabilityUnavailable(t *testing.T) {
	origSave, origDiff := taskSaveRevert, taskDiff
	defer func() { taskSaveRevert, taskDiff = origSave, origDiff }()

	taskSaveRevert = "revert.json"
	taskDiff = false

	err := tasksDeleteCmd.RunE(tasksDeleteCmd, []string{"task1"})
	if err == nil {
		t.Fatal("expected CAPABILITY_UNAVAILABLE error")
	}
	var cerr *cliCodeError
	if !errors.As(err, &cerr) || cerr.Code != "CAPABILITY_UNAVAILABLE" {
		t.Fatalf("expected cliCodeError CAPABILITY_UNAVAILABLE, got %v", err)
	}
}
