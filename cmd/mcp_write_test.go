package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCompleteMCPWriteSavesRevertInfo(t *testing.T) {
	origAllow := allowOutsideCWD
	allowOutsideCWD = true
	defer func() { allowOutsideCWD = origAllow }()

	dir := t.TempDir()
	revertPath := filepath.Join(dir, "revert.json")

	result := json.RawMessage(`{"content":{"revertInfo":{"blocks":[{"id":"b1","markdown":"old"}]}}}`)

	// Redirect stdout so outputRawJSON noise does not pollute test output.
	origStdout := os.Stdout
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open devnull: %v", err)
	}
	os.Stdout = devNull
	defer func() {
		os.Stdout = origStdout
		devNull.Close()
	}()

	if err := completeMCPWrite(nil, result, revertPath, false); err != nil {
		t.Fatalf("completeMCPWrite() error = %v", err)
	}

	data, err := os.ReadFile(revertPath)
	if err != nil {
		t.Fatalf("revert file not written: %v", err)
	}
	var saved struct {
		RevertInfo map[string]interface{} `json:"revertInfo"`
	}
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("revert file is not valid JSON: %v", err)
	}
	if saved.RevertInfo == nil {
		t.Fatalf("revert file missing revertInfo key: %s", data)
	}
}

func TestCompleteMCPWriteFailsWithoutRevertInfo(t *testing.T) {
	origAllow := allowOutsideCWD
	allowOutsideCWD = true
	defer func() { allowOutsideCWD = origAllow }()

	dir := t.TempDir()
	revertPath := filepath.Join(dir, "revert.json")
	result := json.RawMessage(`{"content":{"ok":true}}`)

	if err := completeMCPWrite(nil, result, revertPath, false); err == nil {
		t.Fatal("expected error when MCP result has no revertInfo")
	}
}
