package cmd

import (
	"os"
	"testing"
)

func TestNormalizeAddBlocksDefaultsMissingTypeToText(t *testing.T) {
	blocks := []map[string]interface{}{
		{"markdown": "Catering menu requirement"},
	}
	normalizeAddBlocks(blocks)
	if blocks[0]["type"] != "text" {
		t.Fatalf("type = %v, want text", blocks[0]["type"])
	}
}

func TestNormalizeAddBlocksInfersNonTextTypes(t *testing.T) {
	blocks := []map[string]interface{}{
		{"lineStyle": "strong"},
		{"language": "go", "rawCode": "fmt.Println()"},
		{"fileName": "menu.pdf"},
		{"title": "Reference", "url": "https://example.com"},
	}
	normalizeAddBlocks(blocks)
	want := []string{"line", "code", "file", "richUrl"}
	for i := range want {
		if blocks[i]["type"] != want[i] {
			t.Fatalf("blocks[%d].type = %v, want %s", i, blocks[i]["type"], want[i])
		}
	}
}

func TestReadBlocksInputFromJSONFile(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "blocks-*.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString(`[{"markdown":"from file"}]`); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	blocks, err := readBlocksInput("", file.Name(), false)
	if err != nil {
		t.Fatalf("readBlocksInput() error = %v", err)
	}
	if got := blocks[0]["markdown"]; got != "from file" {
		t.Fatalf("markdown = %v, want from file", got)
	}
}

func TestReadBlocksInputRejectsMultipleSources(t *testing.T) {
	_, err := readBlocksInput(`{"markdown":"inline"}`, "blocks.json", false)
	if err == nil {
		t.Fatal("expected error")
	}
}
