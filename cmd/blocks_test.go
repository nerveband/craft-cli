package cmd

import (
	"os"
	"reflect"
	"strings"
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

func TestApplyRESTPageStylingMapsContractFields(t *testing.T) {
	blocks := []map[string]interface{}{{
		"id": "page", "themeId": "default", "bgColor": "#ffffff", "pageWidth": "wide", "fontFamily": "serif",
		"coverURL": "https://example.com/c.jpg", "coverAttribution": "Photo by A", "coverCrop": `{"x":0,"y":0.1,"width":1,"height":0.5}`,
	}}
	if err := applyRESTPageStyling(blocks); err != nil {
		t.Fatal(err)
	}
	want := map[string]interface{}{
		"themeId": "default", "backgroundColor": "#ffffff", "pageWidth": "wide", "fontFamily": "serif",
		"coverImage": map[string]interface{}{
			"url": "https://example.com/c.jpg", "unsplashAttribution": "Photo by A",
			"cropMask": map[string]interface{}{"x": 0.0, "y": 0.1, "width": 1.0, "height": 0.5},
		},
	}
	if !reflect.DeepEqual(blocks[0]["styling"], want) || len(blocks[0]) != 2 {
		t.Fatalf("unexpected block: %#v", blocks[0])
	}
}

func TestApplyRESTPageStylingRejectsInvalidInput(t *testing.T) {
	for name, block := range map[string]map[string]interface{}{
		"unknown font family":        {"fontFamily": "comic"},
		"crop without cover url":     {"coverCrop": `{"x":0}`},
		"attribution without cover":  {"coverAttribution": "Photo by A"},
		"crop that is not an object": {"coverURL": "https://example.com/c.jpg", "coverCrop": "[1]"},
	} {
		if err := applyRESTPageStyling([]map[string]interface{}{block}); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

func TestMCPBlocksRejectRESTOnlyPageStyling(t *testing.T) {
	for _, key := range []string{"pageWidth", "fontFamily"} {
		_, err := buildMCPBlocksCommand("update", []map[string]interface{}{{"id": "page", key: "x"}}, nil)
		if err == nil || !strings.Contains(err.Error(), "REST-only") {
			t.Errorf("%s: expected REST-only error, got %v", key, err)
		}
	}
}
