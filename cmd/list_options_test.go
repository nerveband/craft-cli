package cmd

import "testing"

func TestBuildListOptionsIncludesDailyNoteRange(t *testing.T) {
	origAfter, origBefore := listDailyNoteAfter, listDailyNoteBefore
	defer func() { listDailyNoteAfter, listDailyNoteBefore = origAfter, origBefore }()

	listDailyNoteAfter = "2026-08-01"
	listDailyNoteBefore = "2026-08-15"

	opts := buildListOptions()
	if opts.DailyNoteDateGte != "2026-08-01" {
		t.Errorf("DailyNoteDateGte = %q, want 2026-08-01", opts.DailyNoteDateGte)
	}
	if opts.DailyNoteDateLte != "2026-08-15" {
		t.Errorf("DailyNoteDateLte = %q, want 2026-08-15", opts.DailyNoteDateLte)
	}
	if !listNeedsAdvanced() {
		t.Error("listNeedsAdvanced() = false, want true when daily-note range set")
	}
}

func TestListNeedsAdvancedFalseByDefault(t *testing.T) {
	origAfter, origBefore := listDailyNoteAfter, listDailyNoteBefore
	defer func() { listDailyNoteAfter, listDailyNoteBefore = origAfter, origBefore }()
	listDailyNoteAfter, listDailyNoteBefore = "", ""

	if listCreatedAfter == "" && listCreatedBefore == "" && listModifiedAfter == "" &&
		listModifiedBefore == "" && !listMetadata && listNeedsAdvanced() {
		t.Error("listNeedsAdvanced() = true with no advanced filters set")
	}
}

func TestListCmdRegistersDailyNoteFlags(t *testing.T) {
	if listCmd.Flags().Lookup("daily-note-after") == nil {
		t.Error("missing --daily-note-after flag on list")
	}
	if listCmd.Flags().Lookup("daily-note-before") == nil {
		t.Error("missing --daily-note-before flag on list")
	}
}
