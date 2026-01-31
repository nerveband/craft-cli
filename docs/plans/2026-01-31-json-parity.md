# Full JSON Parity + Styling Docs Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make CLI JSON output default to full API/MCP-shaped payloads, document styling/markdown capabilities with examples, and publish MCP/API/CLI payload snapshots plus a differences chart.

**Architecture:** Convert CLI outputs to emit full response structs (including `items` wrappers), add a compact format for legacy flattened output, and update docs/skill to highlight styling and parity. Capture payload snapshots for "Craft Everything In One" from MCP/API/CLI and store in `docs/`.

**Tech Stack:** Go, Cobra, encoding/json, Craft Connect API, Craft MCP server.

---

## Task 1: Add output mode for legacy/compact output

**Files:**
- Modify: `cmd/output.go`
- Modify: `cmd/output_blocks.go`
- Modify: `cmd/root.go`

**Step 1: Write failing test for output format default change**

Create `cmd/output_format_test.go`:

```go
package cmd

import "testing"

func TestDefaultFormatIsJson(t *testing.T) {
	if got := getOutputFormat(); got != "json" {
		t.Fatalf("expected default format json, got %s", got)
	}
}
```

**Step 2: Run test to verify it passes (baseline)**

Run: `go test ./cmd -run TestDefaultFormatIsJson`
Expected: PASS

**Step 3: Write failing test for compact format handling**

Update `cmd/output_format_test.go`:

```go
func TestCompactFormatIsSupported(t *testing.T) {
	if !isValidOutputFormat("compact") {
		t.Fatalf("expected compact to be a valid output format")
	}
}
```

**Step 4: Run test to verify it fails**

Run: `go test ./cmd -run TestCompactFormatIsSupported`
Expected: FAIL (format not listed)

**Step 5: Implement compact format support**

Update `ValidOutputFormats` to include `compact`, and ensure `--format compact` is accepted.

**Step 6: Run tests to verify they pass**

Run: `go test ./cmd -run TestCompactFormatIsSupported`
Expected: PASS

**Step 7: Commit**

```bash
git add cmd/output.go cmd/output_blocks.go cmd/root.go cmd/output_format_test.go
git commit -m "feat: add compact output format for legacy views"
```

---

## Task 2: Switch JSON outputs to API/MCP-shaped payloads

**Files:**
- Modify: `cmd/list.go`
- Modify: `cmd/search.go`
- Modify: `cmd/info.go`
- Modify: `cmd/create.go`
- Modify: `cmd/blocks.go`
- Modify: `cmd/folders.go`
- Modify: `cmd/tasks.go`
- Modify: `cmd/collections.go`
- Modify: `cmd/output.go`

**Step 1: Write failing tests for list/search JSON wrappers**

Create `cmd/json_payloads_test.go`:

```go
package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/ashrafali/craft-cli/internal/models"
)

func captureStdout(t *testing.T, fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	w.Close()
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
```

**Step 2: Run tests to verify they fail**

Run: `go test ./cmd -run TestOutputDocumentsJsonUsesWrapper`
Expected: FAIL (functions not defined)

**Step 3: Implement wrapper-aware output helpers**

Add `outputDocumentsPayload` and `outputSearchResultsPayload` to `cmd/output.go` and update commands to call them.

**Step 4: Run tests to verify they pass**

Run: `go test ./cmd -run TestOutputDocumentsJsonUsesWrapper`
Expected: PASS

**Step 5: Update commands to pass full payloads**

- `list`: pass `*models.DocumentList`
- `search`: pass `*models.SearchResult`
- `info`: pass `*models.DocumentList` (if applicable)
- `create`: wrap results in `DocumentList` or use a `CreateResult` wrapper
- `blocks`: return block response objects rather than flattening
- `folders`, `tasks`, `collections`: return list wrappers for JSON

**Step 6: Update compact output for legacy views**

Ensure `--format compact` preserves the previous flattened behavior for tables/markdown.

**Step 7: Run full tests**

Run: `go test ./...`
Expected: PASS

**Step 8: Commit**

```bash
git add cmd/*.go cmd/json_payloads_test.go

git commit -m "feat: default json outputs full api payloads"
```

---

## Task 3: Capture MCP/API/CLI JSON payloads

**Files:**
- Create: `docs/payloads/craft-everything-cli.json`
- Create: `docs/payloads/craft-everything-api.json`
- Create: `docs/payloads/craft-everything-mcp.json`
- Create: `docs/payloads/README.md`

**Step 1: Fetch document id for "Craft Everything In One"**

Run: `craft search "Craft Everything In One" --format json`
Expected: JSON with documentId

**Step 2: Fetch MCP JSON payload**

Run: `curl -sS https://mcp.craft.do/links/548eu6Zqdao/mcp -d '{...}' > docs/payloads/craft-everything-mcp.json`
Expected: JSON payload saved

**Step 3: Fetch API JSON payload**

Run: `curl -sS https://connect.craft.do/links/HHRuPxZZTJ6/api/v1/documents/<id> > docs/payloads/craft-everything-api.json`
Expected: JSON payload saved

**Step 4: Fetch CLI JSON payload**

Run: `craft get <id> --format json > docs/payloads/craft-everything-cli.json`
Expected: JSON payload saved

**Step 5: Commit**

```bash
git add docs/payloads

git commit -m "docs: add mcp api cli payload snapshots"
```

---

## Task 4: Documentation updates (README + skill)

**Files:**
- Modify: `README.md`
- Create: `docs/llm/README.md`
- Create: `docs/llm/styling-and-markdown.md`
- Create: `docs/llm/output-parity.md`
- Modify: `/Users/nerveband/.codex/skills/craftdocs/SKILL.md`

**Step 1: Add LLM-friendly sub-docs**

Write `docs/llm/README.md` that links to styling and parity docs.

**Step 2: Write styling doc with examples**

Include:
- Markdown shortcut list (from user)
- Styling via JSON block fields (color, font, alignment, listStyle, callout, quote, cardLayout)
- CLI examples for create/update and inspection

**Step 3: Write parity doc with differences chart**

Include a table comparing MCP/API/CLI output shapes and any differences.

**Step 4: Update README**

Add a section that links to `docs/llm/README.md` and include a short differences chart summary.

**Step 5: Update craftdocs skill**

Add explicit examples for styling and call out that default JSON is full API payloads.

**Step 6: Commit**

```bash
git add README.md docs/llm docs/payloads /Users/nerveband/.codex/skills/craftdocs/SKILL.md

git commit -m "docs: add styling guidance and output parity references"
```

---

## Task 5: Verification

Run: `go test ./...`
Expected: PASS

---

## Task 6: Final check

Run: `git status -sb`
Expected: clean
