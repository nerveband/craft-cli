package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCommandEffectCoverage(t *testing.T) {
	var walk func(*cobra.Command)
	walk = func(c *cobra.Command) {
		if c.Runnable() {
			if _, ok := declaredEffect(c); !ok {
				t.Errorf("missing explicit effect: %s", c.CommandPath())
			}
		}
		for _, child := range c.Commands() {
			walk(child)
		}
	}
	walk(rootCmd)
}

func TestRESTStyleMapping(t *testing.T) {
	blocks := []map[string]interface{}{{"id": "page-a", "themeId": "soil", "bgColor": "#ffffff", "coverURL": "https://example.org/cover.png"}}
	applyRESTPageStyling(blocks)
	style := blocks[0]["styling"].(map[string]interface{})
	if style["themeId"] != "soil" || style["backgroundColor"] != "#ffffff" || blocks[0]["themeId"] != nil {
		t.Fatalf("incorrect styling %#v", blocks)
	}
}

func TestCanonicalVocabulary(t *testing.T) {
	banned := map[string]bool{"ls": true, "rm": true}
	var walk func(*cobra.Command)
	walk = func(c *cobra.Command) {
		if banned[c.Name()] {
			t.Errorf("banned canonical verb %s", c.CommandPath())
		}
		for _, child := range c.Commands() {
			walk(child)
		}
	}
	walk(rootCmd)
	for _, name := range []string{"format", "dry-run", "yes", "timeout", "profile"} {
		if rootCmd.PersistentFlags().Lookup(name) == nil {
			t.Errorf("missing canonical flag %s", name)
		}
	}
}

// Only commands that can lose data require --yes. Edits, moves and renames must
// stay usable by agents without a commitment flag.
func TestConfirmationGateCoversOnlyDataLoss(t *testing.T) {
	savedYes := yesFlag
	yesFlag = false
	defer func() { yesFlag = savedYes }()

	gated := []string{"delete", "clear", "blocks delete", "tasks delete", "folders delete", "collections delete", "collections views delete", "collections schema update", "whiteboards delete", "reminders delete", "config reset", "config remove", "profiles remove"}
	free := []string{"update", "move", "tasks update", "blocks update", "blocks move", "blocks revert", "folders move", "collections rename", "collections update", "collections views update", "collections active-view set", "reminders update", "whiteboards update"}

	for _, path := range gated {
		cmd, _, err := rootCmd.Find(strings.Fields(path))
		if err != nil {
			t.Fatalf("find %q: %v", path, err)
		}
		if err := commandPreflight(cmd, nil); err == nil || !strings.Contains(err.Error(), "--yes") {
			t.Errorf("%s: expected confirmation error, got %v", path, err)
		}
	}
	for _, path := range free {
		cmd, _, err := rootCmd.Find(strings.Fields(path))
		if err != nil {
			t.Fatalf("find %q: %v", path, err)
		}
		if err := commandPreflight(cmd, nil); err != nil {
			t.Errorf("%s: expected no confirmation gate, got %v", path, err)
		}
	}
}
