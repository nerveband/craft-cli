package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExplicitEmbeddedInputExpansion(t *testing.T) {
	old := expandFiles
	defer func() { expandFiles = old }()
	dir := t.TempDir()
	text := filepath.Join(dir, "notes.txt")
	nested := filepath.Join(dir, "data.json")
	binary := filepath.Join(dir, "data.bin")
	os.WriteFile(text, []byte("Alice's notes\nSecond line"), 0600)
	os.WriteFile(nested, []byte(`{"n":9007199254740993,"literal":"@text://not-followed"}`), 0600)
	os.WriteFile(binary, []byte{0, 1, 255}, 0600)
	payload, _ := json.Marshal(map[string]interface{}{"text": "@text://" + text, "object": "@file://" + nested, "base64": "@data://" + binary, "escape": "@@text://keep"})
	var unchanged map[string]interface{}
	expandFiles = false
	if err := decodeInputJSON(payload, &unchanged); err != nil {
		t.Fatal(err)
	}
	if unchanged["text"] != "@text://"+text {
		t.Fatal("expanded without opt-in")
	}
	expandFiles = true
	var result map[string]json.RawMessage
	if err := decodeInputJSON(payload, &result); err != nil {
		t.Fatal(err)
	}
	if string(result["base64"]) != `"AAH/"` || !strings.Contains(string(result["object"]), "9007199254740993") || string(result["escape"]) != `"@text://keep"` {
		t.Fatal(result)
	}
	if err := decodeInputJSON([]byte(`{"v":"@file://`+dir+`"}`), &result); err == nil {
		t.Fatal("directory accepted")
	}
	if err := decodeInputJSON([]byte(`{} {}`), &result); err == nil {
		t.Fatal("multiple JSON values accepted")
	}
	budget := 1
	if _, err := expandInputValue("@text://"+text, &budget); err == nil {
		t.Fatal("budget ignored")
	}
}

func TestExpansionBudgetBoundsSourceReads(t *testing.T) {
	old := expandFiles
	expandFiles = true
	defer func() { expandFiles = old }()
	if _, err := readStructuredBytes(strings.NewReader(strings.Repeat("x", maxEmbeddedBytes+1))); err == nil {
		t.Fatal("oversized structured source accepted")
	}
}
