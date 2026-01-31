package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/ashrafali/craft-cli/internal/models"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String()
}

func TestOutputDocumentsJsonUsesWrapper(t *testing.T) {
	payload := &models.DocumentList{Items: []models.Document{{ID: "doc1"}}, Total: 1}
	out := captureStdout(t, func() {
		_ = outputDocumentsPayload(payload, "json")
	})
	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if _, ok := decoded["items"]; !ok {
		t.Fatalf("expected items wrapper in output")
	}
}

func TestOutputSearchJsonUsesWrapper(t *testing.T) {
	payload := &models.SearchResult{Items: []models.SearchItem{{DocumentID: "doc1"}}, Total: 1}
	out := captureStdout(t, func() {
		_ = outputSearchResultsPayload(payload, "json")
	})
	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if _, ok := decoded["items"]; !ok {
		t.Fatalf("expected items wrapper in output")
	}
}
