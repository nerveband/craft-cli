package cmd

import (
	"testing"

	"github.com/ashrafali/craft-cli/internal/models"
)

func TestBlockHasMarkdownUnsafeState(t *testing.T) {
	markdownNative := []models.Block{
		{Type: "text", TextStyle: "h2", Markdown: "## Heading"},
		{Type: "text", Decorations: []string{"callout"}, Markdown: "<callout>Note</callout>"},
		{Type: "line", LineStyle: "regular", Markdown: "*****"},
		{Type: "code", Language: "go", RawCode: "fmt.Println(1)", Markdown: "```go\nfmt.Println(1)\n```"},
	}
	for _, block := range markdownNative {
		if blockHasMarkdownUnsafeState(&block) {
			t.Fatalf("block %q should be markdown-native", block.Markdown)
		}
	}

	unsafe := []models.Block{
		{Type: "text", Color: "#9E1B1E", Markdown: "Red"},
		{Type: "text", Font: "serif", Markdown: "Serif"},
		{Type: "text", TextAlignment: "center", Markdown: "Centered"},
		{Type: "text", TaskInfo: &models.TaskInfo{State: "done"}, Markdown: "- [x] Done"},
		{Type: "image", URL: "https://example.com/image.png"},
	}
	for _, block := range unsafe {
		if !blockHasMarkdownUnsafeState(&block) {
			t.Fatalf("block %#v should be unsafe for markdown replace", block)
		}
	}
}

func TestCountReplaceStyleRisk(t *testing.T) {
	block := models.Block{
		Type:     "text",
		Markdown: "Parent",
		Content: []models.Block{
			{Type: "text", Markdown: "Plain"},
			{Type: "text", Color: "#9E1B1E", Markdown: "Styled"},
		},
	}
	var risk replaceStyleRisk
	countReplaceStyleRisk(&block, &risk)
	if risk.TotalBlocks != 3 {
		t.Fatalf("TotalBlocks = %d, want 3", risk.TotalBlocks)
	}
	if risk.StyledBlocks != 1 {
		t.Fatalf("StyledBlocks = %d, want 1", risk.StyledBlocks)
	}
}

func TestBuildSectionDeltaPlan(t *testing.T) {
	blocks := []models.Block{
		{ID: "intro", Type: "text", Markdown: "## Intro", Content: []models.Block{
			{ID: "intro-child", Type: "text", Markdown: "Nested"},
		}},
		{ID: "intro-body", Type: "text", Markdown: "Old intro"},
		{ID: "next", Type: "text", Markdown: "## Next"},
		{ID: "next-body", Type: "text", Markdown: "Keep me"},
	}

	plan, err := buildSectionDeltaPlan(blocks, "Intro", "New intro")
	if err != nil {
		t.Fatalf("buildSectionDeltaPlan() error = %v", err)
	}
	if plan.Replacement != "## Intro\n\nNew intro" {
		t.Fatalf("Replacement = %q", plan.Replacement)
	}
	if plan.InsertBeforeID != "next" {
		t.Fatalf("InsertBeforeID = %q, want next", plan.InsertBeforeID)
	}
	wantIDs := []string{"intro", "intro-child", "intro-body"}
	if len(plan.DeleteIDs) != len(wantIDs) {
		t.Fatalf("DeleteIDs = %#v, want %#v", plan.DeleteIDs, wantIDs)
	}
	for i, want := range wantIDs {
		if plan.DeleteIDs[i] != want {
			t.Fatalf("DeleteIDs[%d] = %q, want %q", i, plan.DeleteIDs[i], want)
		}
	}
}

func TestBuildSectionDeltaPlanReplacementWithHeading(t *testing.T) {
	blocks := []models.Block{
		{ID: "intro", Type: "text", Markdown: "## Intro"},
		{ID: "body", Type: "text", Markdown: "Old"},
	}

	plan, err := buildSectionDeltaPlan(blocks, "Intro", "## Renamed\n\nNew")
	if err != nil {
		t.Fatalf("buildSectionDeltaPlan() error = %v", err)
	}
	if plan.Replacement != "## Renamed\n\nNew" {
		t.Fatalf("Replacement = %q", plan.Replacement)
	}
	if plan.InsertBeforeID != "" {
		t.Fatalf("InsertBeforeID = %q, want empty for end insertion", plan.InsertBeforeID)
	}
}
