package cmd

import (
	"testing"

	"github.com/ashrafali/craft-cli/internal/config"
	"github.com/ashrafali/craft-cli/internal/models"
)

func TestDefaultFormatIsJson(t *testing.T) {
	oldCfgManager := cfgManager
	oldOutputFormat := outputFormat
	t.Cleanup(func() {
		cfgManager = oldCfgManager
		outputFormat = oldOutputFormat
	})

	t.Setenv("HOME", t.TempDir())
	manager, err := config.NewManager()
	if err != nil {
		t.Fatalf("failed to create config manager: %v", err)
	}
	cfgManager = manager
	outputFormat = ""

	if got := getOutputFormat(); got != "json" {
		t.Fatalf("expected default format json, got %s", got)
	}
}

func TestCompactFormatIsSupported(t *testing.T) {
	if !IsValidFormat("compact") {
		t.Fatalf("expected compact to be a valid output format")
	}
}

func TestProjectItemsUsesJSONFieldNames(t *testing.T) {
	items := []models.SearchItem{{DocumentID: "doc1", Markdown: "match"}}

	projected, err := projectItems(items, "documentId")
	if err != nil {
		t.Fatalf("projectItems() error = %v", err)
	}
	if len(projected) != 1 {
		t.Fatalf("len(projected) = %d, want 1", len(projected))
	}
	if projected[0]["documentId"] != "doc1" {
		t.Fatalf("documentId = %v, want doc1", projected[0]["documentId"])
	}
}

func TestPayloadWithMetadataOnlyWhenTruncated(t *testing.T) {
	payload := payloadWithMetadata([]map[string]interface{}{{"id": "doc1"}}, 10, true, 10, 1, "craft list --limit 0")
	if payload["_metadata"] == nil {
		t.Fatal("expected _metadata for truncated payload")
	}

	payload = payloadWithMetadata([]map[string]interface{}{{"id": "doc1"}}, 1, false, 1, 1, "")
	if _, ok := payload["_metadata"]; ok {
		t.Fatal("did not expect _metadata for non-truncated payload")
	}
}
