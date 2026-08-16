# Craft API Batching and Feature Improvements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add multi-ID batching to delete/move commands, daily-note date filtering to `list`, extended task repeat configuration, multi-scope search, and universal `--save-revert`/`--diff` flags across MCP-capable mutating commands.

**Architecture:** All Craft REST endpoints used here already accept arrays (`documentIds`, `folderIds`, `idsToDelete`, `blockIds`), so batching means adding plural client methods that send one request and converting the single-ID methods into thin wrappers. CLI commands switch from `cobra.ExactArgs(1)` to `cobra.MinimumNArgs(1)` and call the batch methods once. Revert/diff support extracts the existing MCP write plumbing from `cmd/blocks.go` into shared helpers (`cmd/mcp_write.go`) reused by collections and update.

**Tech Stack:** Go 1.25.6, cobra v1.10.2 / pflag v1.0.10, `net/http/httptest` for API tests. No new dependencies.

## Global Constraints

- Module path: `github.com/ashrafali/craft-cli`; Go version pinned at `go 1.25.6` in `go.mod`.
- Default `go test ./...` must never hit Craft; live tests stay gated behind `CRAFT_LIVE_TESTS=1` (AGENTS.md "Live tests").
- Single-ID invocations of every changed command MUST keep their current output byte-identical (existing agents parse it).
- Destructive commands keep `--dry-run` support; new multi-ID dry-runs must include a `count` field.
- MCP-only features on `--backend rest` return `CAPABILITY_UNAVAILABLE` via `newCLIError` (existing contract in `cmd/blocks.go:816`).
- `./craft audit agent-dx --format json` must remain 85/85 (the audit's max is `len(checks)`; do not add or remove checks).
- Exit codes stay 0/1/2/3 per AGENTS.md.
- No new third-party dependencies.
- Line numbers below reference the repo state at plan time; re-read the named region before editing if surrounding tasks have already landed.

**Design decision (recorded for reviewers):** The investigation asked for `--save-revert`/`--diff` on `tasks`, `collections`, and `update`. This repo's MCP `craft_write` grammar (built in `cmd/blocks.go:buildMCPBlocksCommand` and `cmd/collections.go`) covers only `blocks *` and `collections *` commands — there is no MCP task-write surface. Therefore: collections get full revert/diff plumbing; `update` routes title-only updates through MCP `blocks update --id <docID>` (the document root is a block, see `client.UpdateBlockMarkdown`); `tasks` register the flags for discoverability but return a structured `CAPABILITY_UNAVAILABLE` error pointing at `craft blocks update <block-id> --save-revert`, consistent with the AGENTS.md capability contract.

---

### Task 1: Batch document deletion in the API client

**Files:**
- Modify: `internal/api/client.go:461-471` (`DeleteDocument`)
- Create: `internal/api/batch_test.go`
- Test: `internal/api/batch_test.go`

**Interfaces:**
- Consumes: existing `(c *Client) doRequest(method, path string, body interface{}) ([]byte, error)` (`internal/api/client.go:75`).
- Produces: `func (c *Client) DeleteDocuments(ids []string) error` — one `DELETE /documents` with `{"documentIds": ids}`. `DeleteDocument(id string) error` remains and delegates. Tasks 5 and 19 rely on `DeleteDocuments`.

- [ ] **Step 1: Write the failing test**

Create `internal/api/batch_test.go`:

```go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestClient_DeleteDocuments(t *testing.T) {
	t.Run("sends all ids in one request", func(t *testing.T) {
		requests := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests++
			if r.Method != "DELETE" {
				t.Errorf("Expected DELETE method, got %s", r.Method)
			}
			if r.URL.Path != "/documents" {
				t.Errorf("Expected path /documents, got %s", r.URL.Path)
			}
			var body struct {
				DocumentIDs []string `json:"documentIds"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("Failed to decode request body: %v", err)
			}
			want := []string{"doc1", "doc2", "doc3"}
			if !reflect.DeepEqual(body.DocumentIDs, want) {
				t.Errorf("Expected documentIds %v, got %v", want, body.DocumentIDs)
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"items": []string{"doc1", "doc2", "doc3"}})
		}))
		defer server.Close()

		client := NewClient(server.URL)
		if err := client.DeleteDocuments([]string{"doc1", "doc2", "doc3"}); err != nil {
			t.Fatalf("DeleteDocuments() error = %v", err)
		}
		if requests != 1 {
			t.Fatalf("Expected exactly 1 request, got %d", requests)
		}
	})

	t.Run("empty slice is a no-op", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("no request expected for empty id slice")
		}))
		defer server.Close()

		client := NewClient(server.URL)
		if err := client.DeleteDocuments(nil); err != nil {
			t.Fatalf("DeleteDocuments(nil) error = %v", err)
		}
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/api -run TestClient_DeleteDocuments -v`
Expected: FAIL (build error) with `client.DeleteDocuments undefined (type *Client has no field or method DeleteDocuments)`

- [ ] **Step 3: Write minimal implementation**

Replace `DeleteDocument` at `internal/api/client.go:461-471` with:

```go
// DeleteDocuments soft-deletes documents by moving them to trash in one request.
func (c *Client) DeleteDocuments(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	req := struct {
		DocumentIDs []string `json:"documentIds"`
	}{
		DocumentIDs: ids,
	}

	_, err := c.doRequest("DELETE", "/documents", req)
	return err
}

// DeleteDocument soft-deletes a single document by moving it to trash.
func (c *Client) DeleteDocument(id string) error {
	return c.DeleteDocuments([]string{id})
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/api -run 'TestClient_DeleteDocuments|TestClient_DeleteDocument$' -v`
Expected: PASS (both the new test and the existing `TestClient_DeleteDocument` in `client_test.go:403`)

- [ ] **Step 5: Commit**

```bash
git add internal/api/client.go internal/api/batch_test.go
git commit -m "feat(api): add DeleteDocuments batch method"
```

---

### Task 2: Batch task deletion in the API client

**Files:**
- Modify: `internal/api/client.go:1101-1109` (`DeleteTask`)
- Test: `internal/api/batch_test.go` (append)

**Interfaces:**
- Consumes: `doRequest` and the existing `deleteTaskRequest` struct (`internal/api/client.go:1097-1099`, field `IDsToDelete []string \`json:"idsToDelete"\``).
- Produces: `func (c *Client) DeleteTasks(ids []string) error` — one `DELETE /tasks` with `{"idsToDelete": ids}`. `DeleteTask(taskID string) error` delegates. Task 7 relies on `DeleteTasks`.

- [ ] **Step 1: Write the failing test**

Append to `internal/api/batch_test.go`:

```go
func TestClient_DeleteTasks(t *testing.T) {
	t.Run("sends all ids in one request", func(t *testing.T) {
		requests := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests++
			if r.Method != "DELETE" {
				t.Errorf("Expected DELETE method, got %s", r.Method)
			}
			if r.URL.Path != "/tasks" {
				t.Errorf("Expected path /tasks, got %s", r.URL.Path)
			}
			var body struct {
				IDsToDelete []string `json:"idsToDelete"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("Failed to decode request body: %v", err)
			}
			want := []string{"task1", "task2"}
			if !reflect.DeepEqual(body.IDsToDelete, want) {
				t.Errorf("Expected idsToDelete %v, got %v", want, body.IDsToDelete)
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"items": []string{"task1", "task2"}})
		}))
		defer server.Close()

		client := NewClient(server.URL)
		if err := client.DeleteTasks([]string{"task1", "task2"}); err != nil {
			t.Fatalf("DeleteTasks() error = %v", err)
		}
		if requests != 1 {
			t.Fatalf("Expected exactly 1 request, got %d", requests)
		}
	})

	t.Run("empty slice is a no-op", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("no request expected for empty id slice")
		}))
		defer server.Close()

		client := NewClient(server.URL)
		if err := client.DeleteTasks(nil); err != nil {
			t.Fatalf("DeleteTasks(nil) error = %v", err)
		}
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/api -run TestClient_DeleteTasks -v`
Expected: FAIL (build error) with `client.DeleteTasks undefined (type *Client has no field or method DeleteTasks)`

- [ ] **Step 3: Write minimal implementation**

Replace `DeleteTask` at `internal/api/client.go:1101-1109` with:

```go
// DeleteTasks deletes tasks by ID in one request.
func (c *Client) DeleteTasks(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	req := deleteTaskRequest{
		IDsToDelete: ids,
	}

	_, err := c.doRequest("DELETE", "/tasks", req)
	return err
}

// DeleteTask deletes a single task.
func (c *Client) DeleteTask(taskID string) error {
	return c.DeleteTasks([]string{taskID})
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/api -run 'TestClient_DeleteTasks|TestClient_TaskPayloads' -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/api/client.go internal/api/batch_test.go
git commit -m "feat(api): add DeleteTasks batch method"
```

---

### Task 3: Batch folder deletion in the API client

**Files:**
- Modify: `internal/api/client.go:732-740` (`DeleteFolder`)
- Test: `internal/api/batch_test.go` (append)

**Interfaces:**
- Consumes: `doRequest` and the existing `deleteFolderRequest` struct (`internal/api/client.go:728-730`, field `FolderIDs []string \`json:"folderIds"\``).
- Produces: `func (c *Client) DeleteFolders(ids []string) error` — one `DELETE /folders` with `{"folderIds": ids}`. `DeleteFolder(folderID string) error` delegates. Task 8 relies on `DeleteFolders`.

- [ ] **Step 1: Write the failing test**

Append to `internal/api/batch_test.go`:

```go
func TestClient_DeleteFolders(t *testing.T) {
	t.Run("sends all ids in one request", func(t *testing.T) {
		requests := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests++
			if r.Method != "DELETE" {
				t.Errorf("Expected DELETE method, got %s", r.Method)
			}
			if r.URL.Path != "/folders" {
				t.Errorf("Expected path /folders, got %s", r.URL.Path)
			}
			var body struct {
				FolderIDs []string `json:"folderIds"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("Failed to decode request body: %v", err)
			}
			want := []string{"folder1", "folder2"}
			if !reflect.DeepEqual(body.FolderIDs, want) {
				t.Errorf("Expected folderIds %v, got %v", want, body.FolderIDs)
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"items": []string{"folder1", "folder2"}})
		}))
		defer server.Close()

		client := NewClient(server.URL)
		if err := client.DeleteFolders([]string{"folder1", "folder2"}); err != nil {
			t.Fatalf("DeleteFolders() error = %v", err)
		}
		if requests != 1 {
			t.Fatalf("Expected exactly 1 request, got %d", requests)
		}
	})

	t.Run("empty slice is a no-op", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("no request expected for empty id slice")
		}))
		defer server.Close()

		client := NewClient(server.URL)
		if err := client.DeleteFolders(nil); err != nil {
			t.Fatalf("DeleteFolders(nil) error = %v", err)
		}
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/api -run TestClient_DeleteFolders -v`
Expected: FAIL (build error) with `client.DeleteFolders undefined (type *Client has no field or method DeleteFolders)`

- [ ] **Step 3: Write minimal implementation**

Replace `DeleteFolder` at `internal/api/client.go:732-740` with:

```go
// DeleteFolders deletes folders by ID in one request.
func (c *Client) DeleteFolders(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	req := deleteFolderRequest{
		FolderIDs: ids,
	}

	_, err := c.doRequest("DELETE", "/folders", req)
	return err
}

// DeleteFolder deletes a single folder.
func (c *Client) DeleteFolder(folderID string) error {
	return c.DeleteFolders([]string{folderID})
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/api -run TestClient_DeleteFolders -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/api/client.go internal/api/batch_test.go
git commit -m "feat(api): add DeleteFolders batch method"
```

---

### Task 4: Batch document move in the API client

**Files:**
- Modify: `internal/api/client.go:750-763` (`MoveDocument`)
- Test: `internal/api/batch_test.go` (append)

**Interfaces:**
- Consumes: `doRequest` and the existing `moveDocumentRequest` struct (`internal/api/client.go:745-748`).
- Produces: `func (c *Client) MoveDocuments(ids []string, folderID, location string) error` — one `PUT /documents/move` with `{"documentIds": ids, "destination": {...}}`. `MoveDocument(docID, folderID, location string) error` delegates. Task 9 relies on `MoveDocuments`.

- [ ] **Step 1: Write the failing test**

Append to `internal/api/batch_test.go`:

```go
func TestClient_MoveDocuments(t *testing.T) {
	t.Run("moves all ids to a folder in one request", func(t *testing.T) {
		requests := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests++
			if r.Method != "PUT" {
				t.Errorf("Expected PUT method, got %s", r.Method)
			}
			if r.URL.Path != "/documents/move" {
				t.Errorf("Expected path /documents/move, got %s", r.URL.Path)
			}
			var body struct {
				DocumentIDs []string `json:"documentIds"`
				Destination struct {
					FolderID string `json:"folderId"`
				} `json:"destination"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("Failed to decode request body: %v", err)
			}
			want := []string{"doc1", "doc2"}
			if !reflect.DeepEqual(body.DocumentIDs, want) {
				t.Errorf("Expected documentIds %v, got %v", want, body.DocumentIDs)
			}
			if body.Destination.FolderID != "folder1" {
				t.Errorf("Expected destination.folderId folder1, got %s", body.Destination.FolderID)
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"items": []map[string]string{{"id": "doc1"}, {"id": "doc2"}}})
		}))
		defer server.Close()

		client := NewClient(server.URL)
		if err := client.MoveDocuments([]string{"doc1", "doc2"}, "folder1", ""); err != nil {
			t.Fatalf("MoveDocuments() error = %v", err)
		}
		if requests != 1 {
			t.Fatalf("Expected exactly 1 request, got %d", requests)
		}
	})

	t.Run("moves all ids to a virtual location", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body struct {
				DocumentIDs []string `json:"documentIds"`
				Destination struct {
					Destination string `json:"destination"`
				} `json:"destination"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("Failed to decode request body: %v", err)
			}
			if body.Destination.Destination != "unsorted" {
				t.Errorf("Expected destination.destination unsorted, got %s", body.Destination.Destination)
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"items": []map[string]string{{"id": "doc1"}}})
		}))
		defer server.Close()

		client := NewClient(server.URL)
		if err := client.MoveDocuments([]string{"doc1"}, "", "unsorted"); err != nil {
			t.Fatalf("MoveDocuments() error = %v", err)
		}
	})

	t.Run("empty slice is a no-op", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("no request expected for empty id slice")
		}))
		defer server.Close()

		client := NewClient(server.URL)
		if err := client.MoveDocuments(nil, "folder1", ""); err != nil {
			t.Fatalf("MoveDocuments(nil) error = %v", err)
		}
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/api -run TestClient_MoveDocuments -v`
Expected: FAIL (build error) with `client.MoveDocuments undefined (type *Client has no field or method MoveDocuments)`

- [ ] **Step 3: Write minimal implementation**

Replace `MoveDocument` at `internal/api/client.go:750-763` with:

```go
// MoveDocuments moves documents to a folder or location in one request.
func (c *Client) MoveDocuments(ids []string, folderID, location string) error {
	if len(ids) == 0 {
		return nil
	}
	req := moveDocumentRequest{
		DocumentIDs: ids,
	}
	if folderID != "" {
		req.Destination = map[string]string{"folderId": folderID}
	} else {
		req.Destination = map[string]string{"destination": location}
	}

	_, err := c.doRequest("PUT", "/documents/move", req)
	return err
}

// MoveDocument moves a single document to a folder or location.
func (c *Client) MoveDocument(docID, folderID, location string) error {
	return c.MoveDocuments([]string{docID}, folderID, location)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/api -run 'TestClient_MoveDocuments|TestClient_MoveDocument$' -v`
Expected: PASS (including the existing `TestClient_MoveDocument` at `client_test.go:433`)

- [ ] **Step 5: Commit**

```bash
git add internal/api/client.go internal/api/batch_test.go
git commit -m "feat(api): add MoveDocuments batch method"
```

---

### Task 5: Multi-ID `craft delete`

**Files:**
- Modify: `cmd/delete.go:14-117` (Use string, Args, RunE, `runDeleteRaw`)
- Create: `cmd/batch_args_test.go`
- Test: `cmd/batch_args_test.go`

**Interfaces:**
- Consumes: `DeleteDocuments(ids []string) error` (Task 1); existing helpers `validateResourceID(id, label string) error` (`cmd/validate.go:40`), `isDryRun() bool`, `dryRunOutput(action string, target map[string]interface{}) error` (`cmd/root.go:375`), `outputDeleted(docID string)` (`cmd/output.go:315`), `outputJSON(v interface{}) error`.
- Produces: `craft delete <document-id...>` accepting 1..N positional IDs; multi-ID JSON output `{"deleted": N, "ids": [...]}`. Single-ID behavior unchanged.

- [ ] **Step 1: Write the failing test**

Create `cmd/batch_args_test.go`:

```go
package cmd

import "testing"

func TestDeleteCmdAcceptsMultipleIDs(t *testing.T) {
	if err := deleteCmd.Args(deleteCmd, []string{"doc1", "doc2", "doc3"}); err != nil {
		t.Fatalf("deleteCmd.Args rejected multiple ids: %v", err)
	}
	if err := deleteCmd.Args(deleteCmd, []string{}); err == nil {
		t.Fatal("deleteCmd.Args accepted zero ids")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd -run TestDeleteCmdAcceptsMultipleIDs -v`
Expected: FAIL with `deleteCmd.Args rejected multiple ids: accepts 1 arg(s), received 3`

- [ ] **Step 3: Write minimal implementation**

In `cmd/delete.go`, change the command definition (lines 14-74). New `Use`, `Long` examples, `Args`, and `RunE`:

```go
var deleteCmd = &cobra.Command{
	Use:   "delete <document-id...>",
	Short: "Move one or more documents to trash",
	Long: `Soft-delete documents by moving them to Craft trash.

This uses DELETE /documents (you can restore via documents/move).
Multiple IDs are sent in a single API request.

Use --dry-run to preview what would be deleted without making changes.

Examples:
  craft delete abc123
  craft delete abc123 def456 ghi789   # Batch delete in one request
  craft delete --json '{"documentIds":["abc123"]}' --dry-run
  craft delete abc123 --dry-run       # Preview without deleting
  craft delete abc123 -q              # Silent delete`,
	Args: func(cmd *cobra.Command, args []string) error {
		if deleteJSON != "" || deleteStdin {
			return cobra.NoArgs(cmd, args)
		}
		return cobra.MinimumNArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if deleteJSON != "" || deleteStdin {
			payload, err := readRawPayload(deleteJSON, deleteStdin)
			if err != nil {
				return err
			}
			return runDeleteRaw(payload)
		}

		for _, id := range args {
			if err := validateResourceID(id, "document-id"); err != nil {
				return err
			}
		}

		if isDryRun() {
			client, err := getAPIClient()
			if err != nil {
				return err
			}
			if len(args) == 1 {
				doc, err := client.GetDocument(args[0])
				if err != nil {
					return fmt.Errorf("document not found: %s", args[0])
				}
				return dryRunOutput("delete", map[string]interface{}{
					"id": doc.ID, "title": doc.Title, "reversible": true,
				})
			}
			targets := make([]map[string]interface{}, 0, len(args))
			for _, id := range args {
				doc, err := client.GetDocument(id)
				if err != nil {
					return fmt.Errorf("document not found: %s", id)
				}
				targets = append(targets, map[string]interface{}{"id": doc.ID, "title": doc.Title})
			}
			return dryRunOutput("delete", map[string]interface{}{
				"documents": targets, "count": len(targets), "reversible": true,
			})
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		if err := client.DeleteDocuments(args); err != nil {
			return err
		}

		if len(args) == 1 {
			outputDeleted(args[0])
			return nil
		}
		return outputJSON(map[string]interface{}{"deleted": len(args), "ids": args})
	},
}
```

In `runDeleteRaw` (`cmd/delete.go:82-117`), replace the per-ID loop (lines 104-115) with one batched call:

```go
	docIDs := make([]string, 0, len(ids))
	for _, id := range ids {
		docID, ok := id.(string)
		if !ok || docID == "" {
			return fmt.Errorf("documentIds must contain strings")
		}
		docIDs = append(docIDs, docID)
	}
	if err := client.DeleteDocuments(docIDs); err != nil {
		return err
	}
	return outputJSON(map[string]interface{}{"deleted": len(docIDs)})
```

(The final `return outputJSON(...)` replaces the old line 116.)

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd -run TestDeleteCmdAcceptsMultipleIDs -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/delete.go cmd/batch_args_test.go
git commit -m "feat(cli): accept multiple ids in craft delete"
```

---

### Task 6: Multi-ID `craft blocks delete`

**Files:**
- Modify: `cmd/blocks.go:366-396` (`blocksDeleteCmd`)
- Test: `cmd/batch_args_test.go` (append)

**Interfaces:**
- Consumes: existing `func (c *Client) DeleteBlocks(blockIDs []string) error` (`internal/api/client.go:499` — already batched, no client change needed).
- Produces: `craft blocks delete <block-id...>` accepting 1..N IDs in one `DELETE /blocks` request.

- [ ] **Step 1: Write the failing test**

Append to `cmd/batch_args_test.go`:

```go
func TestBlocksDeleteCmdAcceptsMultipleIDs(t *testing.T) {
	if err := blocksDeleteCmd.Args(blocksDeleteCmd, []string{"b1", "b2"}); err != nil {
		t.Fatalf("blocksDeleteCmd.Args rejected multiple ids: %v", err)
	}
	if err := blocksDeleteCmd.Args(blocksDeleteCmd, []string{}); err == nil {
		t.Fatal("blocksDeleteCmd.Args accepted zero ids")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd -run TestBlocksDeleteCmdAcceptsMultipleIDs -v`
Expected: FAIL with `blocksDeleteCmd.Args rejected multiple ids: accepts 1 arg(s), received 2`

- [ ] **Step 3: Write minimal implementation**

Replace `blocksDeleteCmd` at `cmd/blocks.go:366-396` with:

```go
var blocksDeleteCmd = &cobra.Command{
	Use:   "delete <block-id...>",
	Short: "Delete one or more blocks",
	Long:  "Delete specific blocks from a document. Multiple IDs are sent in a single API request.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, id := range args {
			if err := validateResourceID(id, "block-id"); err != nil {
				return err
			}
		}
		if isDryRun() {
			if len(args) == 1 {
				return dryRunOutput("delete block", map[string]interface{}{
					"id": args[0], "destructive": true,
				})
			}
			return dryRunOutput("delete blocks", map[string]interface{}{
				"ids": args, "count": len(args), "destructive": true,
			})
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		if err := client.DeleteBlocks(args); err != nil {
			return err
		}

		if !isQuiet() {
			if len(args) == 1 {
				fmt.Printf("Block %s deleted\n", args[0])
			} else {
				fmt.Printf("%d blocks deleted\n", len(args))
			}
		}
		return nil
	},
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd -run TestBlocksDeleteCmdAcceptsMultipleIDs -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/blocks.go cmd/batch_args_test.go
git commit -m "feat(cli): accept multiple ids in craft blocks delete"
```

---

### Task 7: Multi-ID `craft tasks delete`

**Files:**
- Modify: `cmd/tasks.go:193-237` (`tasksDeleteCmd`)
- Test: `cmd/batch_args_test.go` (append)

**Interfaces:**
- Consumes: `DeleteTasks(ids []string) error` (Task 2).
- Produces: `craft tasks delete <task-id...>` accepting 1..N IDs in one `DELETE /tasks` request. The `--json`/`--stdin` raw path is unchanged (already batched via `DeleteTasksRaw`).

- [ ] **Step 1: Write the failing test**

Append to `cmd/batch_args_test.go`:

```go
func TestTasksDeleteCmdAcceptsMultipleIDs(t *testing.T) {
	if err := tasksDeleteCmd.Args(tasksDeleteCmd, []string{"t1", "t2"}); err != nil {
		t.Fatalf("tasksDeleteCmd.Args rejected multiple ids: %v", err)
	}
	if err := tasksDeleteCmd.Args(tasksDeleteCmd, []string{}); err == nil {
		t.Fatal("tasksDeleteCmd.Args accepted zero ids")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd -run TestTasksDeleteCmdAcceptsMultipleIDs -v`
Expected: FAIL with `tasksDeleteCmd.Args rejected multiple ids: accepts 1 arg(s), received 2`

- [ ] **Step 3: Write minimal implementation**

Replace `tasksDeleteCmd` at `cmd/tasks.go:193-237` with:

```go
var tasksDeleteCmd = &cobra.Command{
	Use:   "delete <task-id...>",
	Short: "Delete one or more tasks",
	Long: `Delete tasks by ID. Multiple IDs are sent in a single API request.

Examples:
  craft tasks delete ID
  craft tasks delete ID1 ID2 ID3
  craft tasks delete --json '{"idsToDelete":["ID"]}' --dry-run`,
	Args: func(cmd *cobra.Command, args []string) error {
		if taskJSON != "" || taskStdin {
			return cobra.NoArgs(cmd, args)
		}
		return cobra.MinimumNArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if taskJSON != "" || taskStdin {
			payload, err := readTaskPayload(taskJSON, taskStdin)
			if err != nil {
				return err
			}
			return runTasksDeleteRaw(payload)
		}

		if isDryRun() {
			if len(args) == 1 {
				return dryRunOutput("delete task", map[string]interface{}{
					"id": args[0], "destructive": true,
				})
			}
			return dryRunOutput("delete tasks", map[string]interface{}{
				"ids": args, "count": len(args), "destructive": true,
			})
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		if err := client.DeleteTasks(args); err != nil {
			return err
		}

		if !isQuiet() {
			if len(args) == 1 {
				fmt.Printf("Task %s deleted\n", args[0])
			} else {
				fmt.Printf("%d tasks deleted\n", len(args))
			}
		}
		return nil
	},
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd -run TestTasksDeleteCmdAcceptsMultipleIDs -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/tasks.go cmd/batch_args_test.go
git commit -m "feat(cli): accept multiple ids in craft tasks delete"
```

---

### Task 8: Multi-ID `craft folders delete`

**Files:**
- Modify: `cmd/folders.go:205-271` (`foldersDeleteCmd`)
- Test: `cmd/batch_args_test.go` (append)

**Interfaces:**
- Consumes: `DeleteFolders(ids []string) error` (Task 3).
- Produces: `craft folders delete <folder-id...>` accepting 1..N IDs in one `DELETE /folders` request; the raw `--json` path also switches to one batched call.

- [ ] **Step 1: Write the failing test**

Append to `cmd/batch_args_test.go`:

```go
func TestFoldersDeleteCmdAcceptsMultipleIDs(t *testing.T) {
	if err := foldersDeleteCmd.Args(foldersDeleteCmd, []string{"f1", "f2"}); err != nil {
		t.Fatalf("foldersDeleteCmd.Args rejected multiple ids: %v", err)
	}
	if err := foldersDeleteCmd.Args(foldersDeleteCmd, []string{}); err == nil {
		t.Fatal("foldersDeleteCmd.Args accepted zero ids")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd -run TestFoldersDeleteCmdAcceptsMultipleIDs -v`
Expected: FAIL with `foldersDeleteCmd.Args rejected multiple ids: accepts 1 arg(s), received 2`

- [ ] **Step 3: Write minimal implementation**

Replace `foldersDeleteCmd` at `cmd/folders.go:205-271` with:

```go
var foldersDeleteCmd = &cobra.Command{
	Use:   "delete <folder-id...>",
	Short: "Delete one or more folders",
	Long:  "Delete folders. Contents will be moved to the parent folder. Multiple IDs are sent in a single API request.",
	Args: func(cmd *cobra.Command, args []string) error {
		if folderDeleteJSON != "" || folderDeleteStdin {
			return cobra.NoArgs(cmd, args)
		}
		return cobra.MinimumNArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if folderDeleteJSON != "" || folderDeleteStdin {
			payload, err := readRawPayload(folderDeleteJSON, folderDeleteStdin)
			if err != nil {
				return err
			}
			ids, ok := payloadArray(payload, "folderIds")
			if !ok {
				if id := payloadString(payload, "id", "folderId"); id != "" {
					ids = []interface{}{id}
					ok = true
				}
			}
			if !ok || len(ids) == 0 {
				return fmt.Errorf("raw folder delete payload must include non-empty \"folderIds\" array or \"id\"")
			}
			if isDryRun() {
				return dryRunOutput("delete folders", map[string]interface{}{"payload": payload, "count": len(ids), "destructive": true})
			}
			client, err := getAPIClient()
			if err != nil {
				return err
			}
			folderIDs := make([]string, 0, len(ids))
			for _, id := range ids {
				folderID, ok := id.(string)
				if !ok || folderID == "" {
					return fmt.Errorf("folderIds must contain strings")
				}
				folderIDs = append(folderIDs, folderID)
			}
			if err := client.DeleteFolders(folderIDs); err != nil {
				return err
			}
			return outputJSON(map[string]interface{}{"deleted": len(folderIDs)})
		}

		if isDryRun() {
			if len(args) == 1 {
				return dryRunOutput("delete folder", map[string]interface{}{
					"id": args[0], "destructive": true,
				})
			}
			return dryRunOutput("delete folders", map[string]interface{}{
				"ids": args, "count": len(args), "destructive": true,
			})
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		if err := client.DeleteFolders(args); err != nil {
			return err
		}

		if !isQuiet() {
			if len(args) == 1 {
				fmt.Printf("Folder %s deleted\n", args[0])
			} else {
				fmt.Printf("%d folders deleted\n", len(args))
			}
		}
		return nil
	},
}
```

Note: `Folder %s deleted` matches the exact success string printed today at `cmd/folders.go:267`, so single-folder output stays byte-identical.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd -run TestFoldersDeleteCmdAcceptsMultipleIDs -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/folders.go cmd/batch_args_test.go
git commit -m "feat(cli): accept multiple ids in craft folders delete"
```

---

### Task 9: Multi-ID `craft move`

**Files:**
- Modify: `cmd/move.go:16-137` (`moveCmd`, `runMoveRaw`)
- Test: `cmd/batch_args_test.go` (append)

**Interfaces:**
- Consumes: `MoveDocuments(ids []string, folderID, location string) error` (Task 4).
- Produces: `craft move <document-id...> --to-folder|--to-location` accepting 1..N IDs in one `PUT /documents/move` request; raw path batched too.

- [ ] **Step 1: Write the failing test**

Append to `cmd/batch_args_test.go`:

```go
func TestMoveCmdAcceptsMultipleIDs(t *testing.T) {
	if err := moveCmd.Args(moveCmd, []string{"doc1", "doc2"}); err != nil {
		t.Fatalf("moveCmd.Args rejected multiple ids: %v", err)
	}
	if err := moveCmd.Args(moveCmd, []string{}); err == nil {
		t.Fatal("moveCmd.Args accepted zero ids")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd -run TestMoveCmdAcceptsMultipleIDs -v`
Expected: FAIL with `moveCmd.Args rejected multiple ids: accepts 1 arg(s), received 2`

- [ ] **Step 3: Write minimal implementation**

In `cmd/move.go`, update `Use`, examples, `Args`, and `RunE` (lines 16-86):

```go
var moveCmd = &cobra.Command{
	Use:   "move <document-id...>",
	Short: "Move one or more documents to a folder or location",
	Long: `Move documents to a different folder or special location.
Multiple IDs are sent in a single API request.

Locations:
  unsorted   - Move to unsorted documents
  trash      - Move to trash

Examples:
  craft move abc123 --to-folder def456        # Move to folder
  craft move abc123 def456 --to-folder xyz789 # Batch move
  craft move abc123 --to-location unsorted    # Move to unsorted
  craft move abc123 --to-location trash       # Move to trash
  craft move --json '{"documentIds":["abc123"],"destination":"unsorted"}' --dry-run`,
	Args: func(cmd *cobra.Command, args []string) error {
		if moveJSON != "" || moveStdin {
			return cobra.NoArgs(cmd, args)
		}
		return cobra.MinimumNArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if moveJSON != "" || moveStdin {
			payload, err := readRawPayload(moveJSON, moveStdin)
			if err != nil {
				return err
			}
			return runMoveRaw(payload)
		}

		if moveTargetFolder == "" && moveTargetLocation == "" {
			return fmt.Errorf("either --to-folder or --to-location is required")
		}
		for _, id := range args {
			if err := validateResourceID(id, "document-id"); err != nil {
				return err
			}
		}
		if moveTargetFolder != "" {
			if err := validateResourceID(moveTargetFolder, "folder-id"); err != nil {
				return err
			}
		}

		if isDryRun() {
			target := map[string]interface{}{}
			if len(args) == 1 {
				target["id"] = args[0]
			} else {
				target["ids"] = args
				target["count"] = len(args)
			}
			if moveTargetFolder != "" {
				target["destination_folder"] = moveTargetFolder
			} else {
				target["destination_location"] = moveTargetLocation
			}
			return dryRunOutput("move", target)
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		if err := client.MoveDocuments(args, moveTargetFolder, moveTargetLocation); err != nil {
			return err
		}

		if !isQuiet() {
			destination := moveTargetLocation
			if moveTargetFolder != "" {
				destination = "folder " + moveTargetFolder
			}
			if len(args) == 1 {
				if moveTargetFolder != "" {
					fmt.Printf("Document %s moved to folder %s\n", args[0], moveTargetFolder)
				} else {
					fmt.Printf("Document %s moved to %s\n", args[0], moveTargetLocation)
				}
			} else {
				fmt.Printf("%d documents moved to %s\n", len(args), destination)
			}
		}
		return nil
	},
}
```

In `runMoveRaw` (`cmd/move.go:96-137`), replace the per-ID loop (lines 119-135) with:

```go
	docIDs := make([]string, 0, len(ids))
	for _, id := range ids {
		docID, ok := id.(string)
		if !ok || docID == "" {
			return fmt.Errorf("documentIds must contain strings")
		}
		docIDs = append(docIDs, docID)
	}
	if err := client.MoveDocuments(docIDs, folderID, location); err != nil {
		return err
	}
	return outputJSON(map[string]interface{}{"moved": len(docIDs)})
```

(The final `return outputJSON(...)` replaces the old line 136.)

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd -run TestMoveCmdAcceptsMultipleIDs -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/move.go cmd/batch_args_test.go
git commit -m "feat(cli): accept multiple ids in craft move"
```

---

### Task 10: Daily-note date range flags on `craft list`

**Files:**
- Modify: `cmd/list.go:11-122`
- Create: `cmd/list_options_test.go`
- Test: `cmd/list_options_test.go`

**Interfaces:**
- Consumes: existing `api.ListDocumentsOptions` — fields `DailyNoteDateGte`, `DailyNoteDateLte` already exist and are already sent as `dailyNoteDateGte`/`dailyNoteDateLte` query params (`internal/api/client.go:1468-1510`). No API client change is needed.
- Produces: `--daily-note-after` / `--daily-note-before` flags; helper `func buildListOptions() api.ListDocumentsOptions` and `func listNeedsAdvanced() bool` used by `listCmd.RunE`. Task 19's schema test relies on the flag names.

- [ ] **Step 1: Write the failing test**

Create `cmd/list_options_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd -run 'TestBuildListOptions|TestListNeedsAdvanced|TestListCmdRegistersDailyNoteFlags' -v`
Expected: FAIL (build error) with `undefined: listDailyNoteAfter` (and `buildListOptions`, `listNeedsAdvanced`)

- [ ] **Step 3: Write minimal implementation**

In `cmd/list.go`, add to the var block (after line 17, `listModifiedBefore`):

```go
	listDailyNoteAfter  string
	listDailyNoteBefore string
```

Add the helpers after the `listCmd` definition (after line 106):

```go
// listNeedsAdvanced reports whether any filter requires GetDocumentsAdvanced.
func listNeedsAdvanced() bool {
	return listCreatedAfter != "" || listCreatedBefore != "" ||
		listModifiedAfter != "" || listModifiedBefore != "" ||
		listDailyNoteAfter != "" || listDailyNoteBefore != "" ||
		listMetadata
}

// buildListOptions assembles the advanced listing options from CLI flags.
func buildListOptions() api.ListDocumentsOptions {
	return api.ListDocumentsOptions{
		FolderID:            listFolderID,
		Location:            listLocation,
		FetchMetadata:       listMetadata,
		CreatedDateGte:      listCreatedAfter,
		CreatedDateLte:      listCreatedBefore,
		LastModifiedDateGte: listModifiedAfter,
		LastModifiedDateLte: listModifiedBefore,
		DailyNoteDateGte:    listDailyNoteAfter,
		DailyNoteDateLte:    listDailyNoteBefore,
	}
}
```

In `listCmd.RunE`, replace lines 58-75 (the `useAdvanced` computation and options literal) with:

```go
		useAdvanced := listNeedsAdvanced()

		var result *models.DocumentList
		if useAdvanced {
			result, err = client.GetDocumentsAdvanced(buildListOptions())
		} else {
			result, err = client.GetDocumentsFiltered(listFolderID, listLocation)
		}
```

In `init()` (after line 115), register the flags:

```go
	listCmd.Flags().StringVar(&listDailyNoteAfter, "daily-note-after", "", "Filter daily notes dated on or after this date (YYYY-MM-DD)")
	listCmd.Flags().StringVar(&listDailyNoteBefore, "daily-note-before", "", "Filter daily notes dated on or before this date (YYYY-MM-DD)")
```

In the `Long` help text (after line 38), add:

```
  --daily-note-after DATE  Daily notes dated on or after DATE (YYYY-MM-DD)
  --daily-note-before DATE Daily notes dated on or before DATE (YYYY-MM-DD)
```

and an example:

```
  craft list --location daily_notes --daily-note-after 2026-08-01 --daily-note-before 2026-08-15
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd -run 'TestBuildListOptions|TestListNeedsAdvanced|TestListCmdRegistersDailyNoteFlags' -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/list.go cmd/list_options_test.go
git commit -m "feat(cli): add daily-note date range filters to craft list"
```

---

### Task 11: Extend `RepeatConfig` model

**Files:**
- Modify: `internal/models/document.go:73-79` (`RepeatConfig`)
- Create: `internal/models/document_test.go`
- Test: `internal/models/document_test.go`

**Interfaces:**
- Consumes: nothing new.
- Produces: extended `models.RepeatConfig`:

```go
type RepeatConfig struct {
	Type         string `json:"type,omitempty"`
	Frequency    string `json:"frequency,omitempty"`
	Interval     int    `json:"interval,omitempty"`
	Weekdays     []int  `json:"weekdays,omitempty"`
	EndDate      string `json:"endDate,omitempty"`
	SkipWeekends bool   `json:"skipWeekends,omitempty"`
	DynamicDays  bool   `json:"dynamicDays,omitempty"`
	Reminder     string `json:"reminder,omitempty"`
}
```

Tasks 12 and 13 rely on these exact field names and JSON keys.

- [ ] **Step 1: Write the failing test**

Create `internal/models/document_test.go`:

```go
package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRepeatConfigJSONRoundTrip(t *testing.T) {
	cfg := RepeatConfig{
		Type:         "weekly",
		Frequency:    "weekly",
		Interval:     2,
		Weekdays:     []int{1, 3, 5},
		EndDate:      "2026-12-31",
		SkipWeekends: true,
		DynamicDays:  true,
		Reminder:     "09:00",
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	for _, key := range []string{`"type":"weekly"`, `"frequency":"weekly"`, `"interval":2`,
		`"weekdays":[1,3,5]`, `"endDate":"2026-12-31"`, `"skipWeekends":true`,
		`"dynamicDays":true`, `"reminder":"09:00"`} {
		if !strings.Contains(string(data), key) {
			t.Errorf("marshaled RepeatConfig missing %s in %s", key, data)
		}
	}

	var decoded RepeatConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if !repeatConfigEqual(decoded, cfg) {
		t.Errorf("round trip mismatch: got %+v want %+v", decoded, cfg)
	}
}

func repeatConfigEqual(a, b RepeatConfig) bool {
	if a.Type != b.Type || a.Frequency != b.Frequency || a.Interval != b.Interval ||
		a.EndDate != b.EndDate || a.SkipWeekends != b.SkipWeekends ||
		a.DynamicDays != b.DynamicDays || a.Reminder != b.Reminder ||
		len(a.Weekdays) != len(b.Weekdays) {
		return false
	}
	for i := range a.Weekdays {
		if a.Weekdays[i] != b.Weekdays[i] {
			return false
		}
	}
	return true
}


func TestRepeatConfigOmitsEmptyFields(t *testing.T) {
	data, err := json.Marshal(RepeatConfig{Type: "daily"})
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	if string(data) != `{"type":"daily"}` {
		t.Errorf("expected omitempty on all optional fields, got %s", data)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/models -run TestRepeatConfig -v`
Expected: FAIL (build error) with `unknown field Frequency in struct literal of type RepeatConfig` (also `SkipWeekends`, `DynamicDays`, `Reminder`)

- [ ] **Step 3: Write minimal implementation**

Replace `RepeatConfig` at `internal/models/document.go:73-79` with:

```go
// RepeatConfig represents task repeat configuration
type RepeatConfig struct {
	Type         string `json:"type,omitempty"`         // daily, weekly, monthly, yearly
	Frequency    string `json:"frequency,omitempty"`    // alternate frequency key used by some Craft payloads
	Interval     int    `json:"interval,omitempty"`     // every N days/weeks/etc
	Weekdays     []int  `json:"weekdays,omitempty"`     // 0=Sunday, 6=Saturday
	EndDate      string `json:"endDate,omitempty"`      // YYYY-MM-DD
	SkipWeekends bool   `json:"skipWeekends,omitempty"` // skip Saturday/Sunday occurrences
	DynamicDays  bool   `json:"dynamicDays,omitempty"`  // reschedule relative to completion
	Reminder     string `json:"reminder,omitempty"`     // HH:MM reminder time
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/models -run TestRepeatConfig -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/models/document.go internal/models/document_test.go
git commit -m "feat(models): extend RepeatConfig with frequency, skipWeekends, dynamicDays, reminder"
```

---

### Task 12: Repeat support in `AddTask` / `UpdateTask`

**Files:**
- Modify: `internal/api/client.go:949` (`AddTask` signature/body), `internal/api/client.go:1064` (`UpdateTask` signature/body)
- Modify (callsites): `cmd/tasks.go:119`, `cmd/tasks.go:182`, `internal/api/client_test.go:994`, `internal/api/client_test.go:1046`, `internal/api/live_test.go:104`, `internal/api/live_test.go:113`
- Create: `internal/api/task_repeat_test.go`
- Test: `internal/api/task_repeat_test.go`

**Interfaces:**
- Consumes: `models.RepeatConfig` (Task 11).
- Produces (clean cutover — every caller updated in this task):

```go
func (c *Client) AddTask(markdown, location, docID, scheduleDate, deadlineDate string, repeat *models.RepeatConfig) (*models.Task, error)
func (c *Client) UpdateTask(taskID, state, scheduleDate, deadlineDate string, repeat *models.RepeatConfig) error
```

When `repeat != nil`, `taskInfo["repeat"] = repeat` is included in the payload. Task 13 relies on these signatures.

- [ ] **Step 1: Write the failing test**

Create `internal/api/task_repeat_test.go`:

```go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ashrafali/craft-cli/internal/models"
)

func TestClient_AddTaskWithRepeat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Tasks []struct {
				Markdown string `json:"markdown"`
				TaskInfo struct {
					Repeat *models.RepeatConfig `json:"repeat"`
				} `json:"taskInfo"`
			} `json:"tasks"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}
		if len(body.Tasks) != 1 {
			t.Fatalf("Expected 1 task, got %d", len(body.Tasks))
		}
		repeat := body.Tasks[0].TaskInfo.Repeat
		if repeat == nil {
			t.Fatal("Expected taskInfo.repeat in payload")
		}
		if repeat.Type != "weekly" || repeat.Interval != 2 || !repeat.SkipWeekends || repeat.Reminder != "09:00" {
			t.Errorf("Unexpected repeat payload: %+v", repeat)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{{"id": "task1", "markdown": "Water plants", "state": "todo"}},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	repeat := &models.RepeatConfig{Type: "weekly", Interval: 2, SkipWeekends: true, Reminder: "09:00"}
	task, err := client.AddTask("Water plants", "inbox", "", "", "", repeat)
	if err != nil {
		t.Fatalf("AddTask() error = %v", err)
	}
	if task.ID != "task1" {
		t.Errorf("task.ID = %q, want task1", task.ID)
	}
}

func TestClient_UpdateTaskWithRepeat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			TasksToUpdate []struct {
				ID       string `json:"id"`
				TaskInfo struct {
					Repeat *models.RepeatConfig `json:"repeat"`
				} `json:"taskInfo"`
			} `json:"tasksToUpdate"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}
		if len(body.TasksToUpdate) != 1 || body.TasksToUpdate[0].ID != "task1" {
			t.Fatalf("Unexpected tasksToUpdate: %+v", body.TasksToUpdate)
		}
		repeat := body.TasksToUpdate[0].TaskInfo.Repeat
		if repeat == nil || repeat.Frequency != "daily" || !repeat.DynamicDays {
			t.Errorf("Unexpected repeat payload: %+v", repeat)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"items": []string{"task1"}})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	repeat := &models.RepeatConfig{Frequency: "daily", DynamicDays: true}
	if err := client.UpdateTask("task1", "", "", "", repeat); err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/api -run 'TestClient_AddTaskWithRepeat|TestClient_UpdateTaskWithRepeat' -v`
Expected: FAIL (build error) with `too many arguments in call to client.AddTask` / `too many arguments in call to client.UpdateTask`

- [ ] **Step 3: Write minimal implementation**

In `internal/api/client.go:949`, change the `AddTask` signature and taskInfo assembly:

```go
// AddTask creates a new task
func (c *Client) AddTask(markdown, location, docID, scheduleDate, deadlineDate string, repeat *models.RepeatConfig) (*models.Task, error) {
	taskInfo := map[string]interface{}{}
	if scheduleDate != "" {
		taskInfo["scheduleDate"] = scheduleDate
	}
	if deadlineDate != "" {
		taskInfo["deadlineDate"] = deadlineDate
	}
	if repeat != nil {
		taskInfo["repeat"] = repeat
	}
	if len(taskInfo) == 0 {
		taskInfo = nil
	}
```

(the rest of the function body from line 961 onward is unchanged).

In `internal/api/client.go:1064`, change `UpdateTask`:

```go
// UpdateTask updates a task's state, dates, or repeat rule
func (c *Client) UpdateTask(taskID, state, scheduleDate, deadlineDate string, repeat *models.RepeatConfig) error {
	taskInfo := map[string]interface{}{}
	if state != "" {
		taskInfo["state"] = state
	}
	if scheduleDate != "" {
		taskInfo["scheduleDate"] = scheduleDate
	}
	if deadlineDate != "" {
		taskInfo["deadlineDate"] = deadlineDate
	}
	if repeat != nil {
		taskInfo["repeat"] = repeat
	}
```

(the rest unchanged).

Update every caller to pass `nil` for now (Task 13 wires real values):
- `cmd/tasks.go:119` → `client.AddTask(description, taskLocation, taskDocumentID, taskScheduleDate, taskDeadlineDate, nil)`
- `cmd/tasks.go:182` → `client.UpdateTask(taskID, taskState, taskScheduleDate, taskDeadlineDate, nil)`
- `internal/api/client_test.go:994` → `client.AddTask("Review PR", "document", "doc1", "2026-02-01", "2026-02-15", nil)`
- `internal/api/client_test.go:1046` → `client.UpdateTask("task1", "done", "2026-02-01", "", nil)`
- `internal/api/live_test.go:104` → `client.AddTask(prefix+" task", "inbox", "", "", "", nil)`
- `internal/api/live_test.go:113` → `client.UpdateTask(task.ID, "done", "", "", nil)`

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/api ./cmd -v -run 'TestClient_AddTask|TestClient_UpdateTask|TestClient_TaskPayloads'`
Expected: PASS (new tests plus all pre-existing task payload tests)

- [ ] **Step 5: Commit**

```bash
git add internal/api/client.go internal/api/client_test.go internal/api/live_test.go internal/api/task_repeat_test.go cmd/tasks.go
git commit -m "feat(api): accept repeat config in AddTask and UpdateTask"
```

---

### Task 13: Repeat flags on `craft tasks add` / `craft tasks update`

**Files:**
- Modify: `cmd/tasks.go:69-77` (vars), `cmd/tasks.go:100-135` (add RunE), `cmd/tasks.go:159-190` (update RunE), `cmd/tasks.go:239-264` (init)
- Create: `cmd/tasks_repeat_test.go`
- Test: `cmd/tasks_repeat_test.go`

**Interfaces:**
- Consumes: `models.RepeatConfig` (Task 11); `AddTask(..., repeat *models.RepeatConfig)` / `UpdateTask(..., repeat *models.RepeatConfig)` (Task 12).
- Produces: flags `--repeat`, `--repeat-frequency`, `--repeat-interval`, `--repeat-weekdays`, `--repeat-end`, `--repeat-reminder`, `--repeat-skip-weekends`, `--repeat-dynamic-days` on both `tasks add` and `tasks update`; helper:

```go
type repeatFlagValues struct {
	Type         string
	Frequency    string
	Interval     int
	Weekdays     []int
	EndDate      string
	Reminder     string
	SkipWeekends bool
	DynamicDays  bool
}
func buildRepeatConfig(v repeatFlagValues) (*models.RepeatConfig, error)
```

- [ ] **Step 1: Write the failing test**

Create `cmd/tasks_repeat_test.go`:

```go
package cmd

import "testing"

func TestBuildRepeatConfig(t *testing.T) {
	t.Run("returns nil when nothing set", func(t *testing.T) {
		cfg, err := buildRepeatConfig(repeatFlagValues{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg != nil {
			t.Fatalf("expected nil config, got %+v", cfg)
		}
	})

	t.Run("builds full config", func(t *testing.T) {
		cfg, err := buildRepeatConfig(repeatFlagValues{
			Type:         "weekly",
			Interval:     2,
			Weekdays:     []int{1, 3},
			EndDate:      "2026-12-31",
			Reminder:     "09:00",
			SkipWeekends: true,
			DynamicDays:  true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Type != "weekly" || cfg.Interval != 2 || len(cfg.Weekdays) != 2 ||
			cfg.EndDate != "2026-12-31" || cfg.Reminder != "09:00" ||
			!cfg.SkipWeekends || !cfg.DynamicDays {
			t.Errorf("unexpected config: %+v", cfg)
		}
	})

	t.Run("rejects invalid type", func(t *testing.T) {
		if _, err := buildRepeatConfig(repeatFlagValues{Type: "hourly"}); err == nil {
			t.Fatal("expected error for invalid repeat type")
		}
	})

	t.Run("rejects modifiers without a rule", func(t *testing.T) {
		if _, err := buildRepeatConfig(repeatFlagValues{Interval: 2}); err == nil {
			t.Fatal("expected error when --repeat-interval set without --repeat/--repeat-frequency")
		}
	})

	t.Run("rejects out-of-range weekday", func(t *testing.T) {
		if _, err := buildRepeatConfig(repeatFlagValues{Type: "weekly", Weekdays: []int{7}}); err == nil {
			t.Fatal("expected error for weekday 7")
		}
	})

	t.Run("rejects bad end date", func(t *testing.T) {
		if _, err := buildRepeatConfig(repeatFlagValues{Type: "daily", EndDate: "31-12-2026"}); err == nil {
			t.Fatal("expected error for non-ISO end date")
		}
	})
}

func TestTasksCmdsRegisterRepeatFlags(t *testing.T) {
	for _, name := range []string{"repeat", "repeat-frequency", "repeat-interval",
		"repeat-weekdays", "repeat-end", "repeat-reminder",
		"repeat-skip-weekends", "repeat-dynamic-days"} {
		if tasksAddCmd.Flags().Lookup(name) == nil {
			t.Errorf("tasks add missing --%s", name)
		}
		if tasksUpdateCmd.Flags().Lookup(name) == nil {
			t.Errorf("tasks update missing --%s", name)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd -run 'TestBuildRepeatConfig|TestTasksCmdsRegisterRepeatFlags' -v`
Expected: FAIL (build error) with `undefined: buildRepeatConfig` and `undefined: repeatFlagValues`

- [ ] **Step 3: Write minimal implementation**

In `cmd/tasks.go`, extend the var block (lines 69-77):

```go
var (
	taskMarkdown           string
	taskLocation           string
	taskScheduleDate       string
	taskDeadlineDate       string
	taskState              string
	taskJSON               string
	taskStdin              bool
	taskRepeatType         string
	taskRepeatFrequency    string
	taskRepeatInterval     int
	taskRepeatWeekdays     []int
	taskRepeatEnd          string
	taskRepeatReminder     string
	taskRepeatSkipWeekends bool
	taskRepeatDynamicDays  bool
)
```

Add the helper (place it just above `readTaskPayload`, `cmd/tasks.go:266`). It needs `"time"` in the imports:

```go
// repeatFlagValues carries repeat flag inputs so validation is testable.
type repeatFlagValues struct {
	Type         string
	Frequency    string
	Interval     int
	Weekdays     []int
	EndDate      string
	Reminder     string
	SkipWeekends bool
	DynamicDays  bool
}

// buildRepeatConfig validates repeat flags and assembles a RepeatConfig.
// Returns (nil, nil) when no repeat flag was provided.
func buildRepeatConfig(v repeatFlagValues) (*models.RepeatConfig, error) {
	hasRule := v.Type != "" || v.Frequency != ""
	hasModifier := v.Interval != 0 || len(v.Weekdays) > 0 || v.EndDate != "" ||
		v.Reminder != "" || v.SkipWeekends || v.DynamicDays
	if !hasRule && !hasModifier {
		return nil, nil
	}
	if !hasRule {
		return nil, fmt.Errorf("repeat modifiers require --repeat or --repeat-frequency")
	}
	validRules := map[string]bool{"daily": true, "weekly": true, "monthly": true, "yearly": true}
	if v.Type != "" && !validRules[v.Type] {
		return nil, fmt.Errorf("invalid --repeat %q (expected daily, weekly, monthly, or yearly)", v.Type)
	}
	if v.Frequency != "" && !validRules[v.Frequency] {
		return nil, fmt.Errorf("invalid --repeat-frequency %q (expected daily, weekly, monthly, or yearly)", v.Frequency)
	}
	if v.Interval < 0 {
		return nil, fmt.Errorf("--repeat-interval must be >= 1")
	}
	for _, day := range v.Weekdays {
		if day < 0 || day > 6 {
			return nil, fmt.Errorf("--repeat-weekdays values must be 0 (Sunday) through 6 (Saturday), got %d", day)
		}
	}
	if v.EndDate != "" {
		if _, err := time.Parse("2006-01-02", v.EndDate); err != nil {
			return nil, fmt.Errorf("invalid --repeat-end %q (expected YYYY-MM-DD)", v.EndDate)
		}
	}
	return &models.RepeatConfig{
		Type:         v.Type,
		Frequency:    v.Frequency,
		Interval:     v.Interval,
		Weekdays:     v.Weekdays,
		EndDate:      v.EndDate,
		Reminder:     v.Reminder,
		SkipWeekends: v.SkipWeekends,
		DynamicDays:  v.DynamicDays,
	}, nil
}

func repeatFlagsFromVars() repeatFlagValues {
	return repeatFlagValues{
		Type:         taskRepeatType,
		Frequency:    taskRepeatFrequency,
		Interval:     taskRepeatInterval,
		Weekdays:     taskRepeatWeekdays,
		EndDate:      taskRepeatEnd,
		Reminder:     taskRepeatReminder,
		SkipWeekends: taskRepeatSkipWeekends,
		DynamicDays:  taskRepeatDynamicDays,
	}
}
```

Wire into `tasksAddCmd.RunE` — replace `cmd/tasks.go:119`:

```go
		repeat, err := buildRepeatConfig(repeatFlagsFromVars())
		if err != nil {
			return err
		}

		task, err := client.AddTask(description, taskLocation, taskDocumentID, taskScheduleDate, taskDeadlineDate, repeat)
```

Wire into `tasksUpdateCmd.RunE` — replace the emptiness check at `cmd/tasks.go:168-170` and the call at line 182:

```go
		repeat, err := buildRepeatConfig(repeatFlagsFromVars())
		if err != nil {
			return err
		}
		if taskState == "" && taskScheduleDate == "" && taskDeadlineDate == "" && repeat == nil {
			return fmt.Errorf("at least one of --state, --schedule, --deadline, or a --repeat flag is required")
		}
```

```go
		if err := client.UpdateTask(taskID, taskState, taskScheduleDate, taskDeadlineDate, repeat); err != nil {
```

Register flags in `init()` — add after `cmd/tasks.go:252` for add, and after line 259 for update (identical blocks, one per command; the shared vars are fine because add and update never run in the same process invocation):

```go
	for _, c := range []*cobra.Command{tasksAddCmd, tasksUpdateCmd} {
		c.Flags().StringVar(&taskRepeatType, "repeat", "", "Repeat rule: daily, weekly, monthly, yearly")
		c.Flags().StringVar(&taskRepeatFrequency, "repeat-frequency", "", "Repeat frequency key (daily, weekly, monthly, yearly)")
		c.Flags().IntVar(&taskRepeatInterval, "repeat-interval", 0, "Repeat every N periods")
		c.Flags().IntSliceVar(&taskRepeatWeekdays, "repeat-weekdays", nil, "Repeat weekdays, 0=Sunday..6=Saturday (comma-separated)")
		c.Flags().StringVar(&taskRepeatEnd, "repeat-end", "", "Repeat end date (YYYY-MM-DD)")
		c.Flags().StringVar(&taskRepeatReminder, "repeat-reminder", "", "Reminder time (HH:MM)")
		c.Flags().BoolVar(&taskRepeatSkipWeekends, "repeat-skip-weekends", false, "Skip weekend occurrences")
		c.Flags().BoolVar(&taskRepeatDynamicDays, "repeat-dynamic-days", false, "Reschedule relative to completion date")
	}
```

Add examples to `tasksAddCmd.Long` (after line 92):

```
  craft tasks add "Water plants" --repeat weekly --repeat-interval 2 --repeat-weekdays 1,3,5
  craft tasks add "Standup" --repeat daily --repeat-skip-weekends --repeat-reminder 09:00
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd -run 'TestBuildRepeatConfig|TestTasksCmdsRegisterRepeatFlags' -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/tasks.go cmd/tasks_repeat_test.go
git commit -m "feat(cli): add repeat rule flags to craft tasks add/update"
```

---

### Task 14: Multi-scope `craft search` (`--folder` and `--document` accept multiple IDs)

**Files:**
- Modify: `cmd/search.go:14-176`
- Create: `cmd/search_scope_test.go`
- Test: `cmd/search_scope_test.go`

**Interfaces:**
- Consumes: existing `api.SearchOptions.FolderIDs` / `DocumentIDs` (comma-separated strings, already sent as `folderIDs`/`documentIDs` query params — `internal/api/client.go:1400-1427`). No API client change needed.
- Produces: `--folder` and `--document` become `StringSlice` flags (repeatable and comma-separated). Semantics: exactly one `--document` → block search (unchanged behavior); two or more `--document` values → document search scoped to those documents. Helper:

```go
type searchScope struct {
	BlockDocID  string // set when exactly one --document: block search mode
	FolderIDs   string // comma-joined for SearchOptions.FolderIDs
	DocumentIDs string // comma-joined for SearchOptions.DocumentIDs (multi-document scope)
}
func resolveSearchScope(folders, documents []string) searchScope
```

- [ ] **Step 1: Write the failing test**

Create `cmd/search_scope_test.go`:

```go
package cmd

import "testing"

func TestResolveSearchScope(t *testing.T) {
	t.Run("single document means block search", func(t *testing.T) {
		scope := resolveSearchScope(nil, []string{"doc1"})
		if scope.BlockDocID != "doc1" {
			t.Errorf("BlockDocID = %q, want doc1", scope.BlockDocID)
		}
		if scope.DocumentIDs != "" {
			t.Errorf("DocumentIDs = %q, want empty", scope.DocumentIDs)
		}
	})

	t.Run("multiple documents mean scoped document search", func(t *testing.T) {
		scope := resolveSearchScope(nil, []string{"doc1", "doc2"})
		if scope.BlockDocID != "" {
			t.Errorf("BlockDocID = %q, want empty", scope.BlockDocID)
		}
		if scope.DocumentIDs != "doc1,doc2" {
			t.Errorf("DocumentIDs = %q, want doc1,doc2", scope.DocumentIDs)
		}
	})

	t.Run("folders join comma-separated", func(t *testing.T) {
		scope := resolveSearchScope([]string{"f1", "f2", "f3"}, nil)
		if scope.FolderIDs != "f1,f2,f3" {
			t.Errorf("FolderIDs = %q, want f1,f2,f3", scope.FolderIDs)
		}
	})

	t.Run("empty inputs produce empty scope", func(t *testing.T) {
		scope := resolveSearchScope(nil, nil)
		if scope.BlockDocID != "" || scope.FolderIDs != "" || scope.DocumentIDs != "" {
			t.Errorf("expected empty scope, got %+v", scope)
		}
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd -run TestResolveSearchScope -v`
Expected: FAIL (build error) with `undefined: resolveSearchScope`

- [ ] **Step 3: Write minimal implementation**

In `cmd/search.go`, change the vars (lines 15 and 20):

```go
	searchDocuments      []string
	searchFolders        []string
```

(delete `searchDocument string` and `searchFolder string`).

Add the helper after the var block:

```go
// searchScope resolves multi-value --folder/--document flags into search parameters.
type searchScope struct {
	BlockDocID  string
	FolderIDs   string
	DocumentIDs string
}

// resolveSearchScope maps flag values to a search scope. Exactly one --document
// keeps the existing block-search behavior; two or more scope a document search.
func resolveSearchScope(folders, documents []string) searchScope {
	scope := searchScope{FolderIDs: strings.Join(folders, ",")}
	switch len(documents) {
	case 0:
	case 1:
		scope.BlockDocID = documents[0]
	default:
		scope.DocumentIDs = strings.Join(documents, ",")
	}
	return scope
}
```

Update `searchCmd.RunE` (lines 48-63):

```go
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		format := getOutputFormat()
		scope := resolveSearchScope(searchFolders, searchDocuments)

		// Block search mode: exactly one --document
		if scope.BlockDocID != "" {
			return runBlockSearch(client, args, format, scope.BlockDocID)
		}

		// Document search mode (default; optionally scoped to folders/documents)
		return runDocumentSearch(client, args, format, scope)
	},
```

Update `runBlockSearch` signature (line 86) and its `SearchBlocks` call (line 99):

```go
func runBlockSearch(client *api.Client, args []string, format, docID string) error {
```

```go
	result, err := client.SearchBlocks(docID, pattern, searchCaseSensitive, searchContext, searchContext)
```

Update `runDocumentSearch` signature (line 119) and the options literal (lines 130-139):

```go
func runDocumentSearch(client *api.Client, args []string, format string, scope searchScope) error {
```

```go
	opts := api.SearchOptions{
		Regexps:             searchRegex,
		Location:            searchLocation,
		FolderIDs:           scope.FolderIDs,
		DocumentIDs:         scope.DocumentIDs,
		FetchMetadata:       searchMetadata,
		CreatedDateGte:      searchCreatedAfter,
		CreatedDateLte:      searchCreatedBefore,
		LastModifiedDateGte: searchModifiedAfter,
		LastModifiedDateLte: searchModifiedBefore,
	}
```

Update flag registration in `init()` (lines 69 and 74):

```go
	searchCmd.Flags().StringSliceVar(&searchDocuments, "document", nil, "Document ID(s). One ID: block-level search. Multiple IDs (repeat or comma-separate): scope document search")
	searchCmd.Flags().StringSliceVar(&searchFolders, "folder", nil, "Folder ID(s) to scope the search (repeat or comma-separate)")
```

Update `searchCmd.Long` (after line 41) with:

```
Multi-scope search:
  craft search "budget" --folder ID1 --folder ID2
  craft search "budget" --folder ID1,ID2
  craft search "roadmap" --document DOC1,DOC2   # scoped document search
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd -run TestResolveSearchScope -v && go build ./...`
Expected: PASS and clean build (the build proves every use of the removed `searchDocument`/`searchFolder` vars was migrated)

- [ ] **Step 5: Commit**

```bash
git add cmd/search.go cmd/search_scope_test.go
git commit -m "feat(cli): support multiple --folder/--document scopes in craft search"
```

---

### Task 15: Extract shared MCP write helpers (`cmd/mcp_write.go`)

**Files:**
- Create: `cmd/mcp_write.go`
- Modify: `cmd/blocks.go:824-867` (`runMCPBlocksMutation` delegates to the new helper)
- Create: `cmd/mcp_write_test.go`
- Test: `cmd/mcp_write_test.go`

**Interfaces:**
- Consumes: existing `saveRevertInfo(raw json.RawMessage, path string) error` (`cmd/blocks.go:976`), `outputMCPMutationWithReview(client *craftmcp.Client, result json.RawMessage) error` (`cmd/blocks.go:869`), `outputRawJSON`, `dryRunOutput`, `getMCPClient`, `enhanceMCPWriteError`, `yesFlag`, `isDryRun`.
- Produces (Tasks 16 and 17 rely on both):

```go
func completeMCPWrite(client *craftmcp.Client, result json.RawMessage, saveRevertPath string, diff bool) error
func runMCPWriteCommand(operation, command string, capabilities []string, saveRevertPath string, diff bool) error
```

`runMCPWriteCommand` performs: dry-run preview (including `save_revert`/`diff` keys), `--yes` gate, `craft_write` call, then `completeMCPWrite`.

- [ ] **Step 1: Write the failing test**

Create `cmd/mcp_write_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd -run TestCompleteMCPWrite -v`
Expected: FAIL (build error) with `undefined: completeMCPWrite`

- [ ] **Step 3: Write minimal implementation**

Create `cmd/mcp_write.go`:

```go
package cmd

import (
	"encoding/json"
	"fmt"

	craftmcp "github.com/ashrafali/craft-cli/internal/mcp"
)

// completeMCPWrite finishes a craft_write mutation: persists revertInfo when
// requested, then emits the result (with edit-review metadata when diff is set).
func completeMCPWrite(client *craftmcp.Client, result json.RawMessage, saveRevertPath string, diff bool) error {
	if saveRevertPath != "" {
		if err := saveRevertInfo(result, saveRevertPath); err != nil {
			return err
		}
	}
	if diff {
		return outputMCPMutationWithReview(client, result)
	}
	return outputRawJSON(result)
}

// runMCPWriteCommand executes a craft_write command with the standard
// dry-run preview, --yes confirmation gate, and revert/diff handling.
func runMCPWriteCommand(operation, command string, capabilities []string, saveRevertPath string, diff bool) error {
	target := map[string]interface{}{
		"backend":      "mcp",
		"tool":         "craft_write",
		"command":      command,
		"capabilities": capabilities,
	}
	if saveRevertPath != "" {
		target["save_revert"] = saveRevertPath
	}
	if diff {
		target["diff"] = map[string]interface{}{
			"available": false,
			"hint":      "MCP edit-review metadata is captured from real craft_write results; dry-run shows the planned command.",
		}
	}
	if isDryRun() {
		return dryRunOutput(operation, target)
	}
	if !yesFlag {
		return fmt.Errorf("%s uses craft_write; rerun with --yes after reviewing --dry-run", operation)
	}
	client, err := getMCPClient()
	if err != nil {
		return err
	}
	result, err := client.CallTool("craft_write", map[string]interface{}{"command": command})
	if err != nil {
		return enhanceMCPWriteError(command, err)
	}
	return completeMCPWrite(client, result, saveRevertPath, diff)
}
```

Replace the body of `runMCPBlocksMutation` (`cmd/blocks.go:824-867`) with a delegation:

```go
func runMCPBlocksMutation(cmd *cobra.Command, action string, blocks []map[string]interface{}, position map[string]interface{}) error {
	command, err := buildMCPBlocksCommand(action, blocks, position)
	if err != nil {
		return err
	}
	return runMCPWriteCommand(
		"blocks "+action,
		command,
		[]string{"mcp", "write", "blocks.revert", "blocks.style"},
		blockSaveRevert,
		blockDiff,
	)
}
```

Note: the `--yes` gate message for blocks changes from `blocks %s with --backend mcp uses craft_write; ...` to `blocks %s uses craft_write; rerun with --yes after reviewing --dry-run` — a deliberate unification with the collections wording.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd -run 'TestCompleteMCPWrite' -v && go build ./...`
Expected: PASS and clean build

- [ ] **Step 5: Commit**

```bash
git add cmd/mcp_write.go cmd/mcp_write_test.go cmd/blocks.go
git commit -m "refactor(cli): extract shared MCP craft_write helpers"
```

---

### Task 16: `--save-revert` / `--diff` on collections MCP writes

**Files:**
- Modify: `cmd/collections.go:550-614` (init flag registration), `cmd/collections.go:640-661` (`runMCPCollectionCommand`)
- Create: `cmd/revert_flags_test.go`
- Test: `cmd/revert_flags_test.go`

**Interfaces:**
- Consumes: `runMCPWriteCommand` (Task 15).
- Produces: package vars `collectionSaveRevert string`, `collectionDiff bool`; flags `--save-revert` and `--diff` on every collections command that calls `runMCPCollectionCommand(..., "craft_write")`: `collections create` (MCP path), `collections rename`, `collections items add` (MCP path), `collections items update` (MCP path), `collections views create`, `collections views update`, `collections views delete`, `collections active-view set`.

- [ ] **Step 1: Write the failing test**

Create `cmd/revert_flags_test.go`:

```go
package cmd

import "testing"

func TestCollectionsWriteCmdsRegisterRevertFlags(t *testing.T) {
	targets := []struct {
		name string
		has  func(flag string) bool
	}{
		{"collections create", func(f string) bool { return collectionsCreateCmd.Flags().Lookup(f) != nil }},
		{"collections rename", func(f string) bool { return collectionsRenameCmd.Flags().Lookup(f) != nil }},
		{"collections items add", func(f string) bool { return collectionsAddCmd.Flags().Lookup(f) != nil }},
		{"collections items update", func(f string) bool { return collectionsUpdateCmd.Flags().Lookup(f) != nil }},
		{"collections views create", func(f string) bool { return collectionsViewsCreateCmd.Flags().Lookup(f) != nil }},
		{"collections views update", func(f string) bool { return collectionsViewsUpdateCmd.Flags().Lookup(f) != nil }},
		{"collections views delete", func(f string) bool { return collectionsViewsDeleteCmd.Flags().Lookup(f) != nil }},
		{"collections active-view set", func(f string) bool { return collectionsActiveViewSetCmd.Flags().Lookup(f) != nil }},
	}
	for _, tc := range targets {
		if !tc.has("save-revert") {
			t.Errorf("%s missing --save-revert", tc.name)
		}
		if !tc.has("diff") {
			t.Errorf("%s missing --diff", tc.name)
		}
	}
}
```

Identifier note: the items add/update commands are `collectionsAddCmd` (`cmd/collections.go:208`) and `collectionsUpdateCmd` (`cmd/collections.go:308`); the collection create/rename/views/active-view commands are `collectionsCreateCmd` (`:95`), `collectionsRenameCmd` (`:143`), `collectionsViewsCreateCmd` (`:472`), `collectionsViewsUpdateCmd` (`:494`), `collectionsViewsDeleteCmd` (`:519`), and `collectionsActiveViewSetCmd` (`:537`).

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd -run TestCollectionsWriteCmdsRegisterRevertFlags -v`
Expected: FAIL with `collections rename missing --save-revert` (and the other four commands)

- [ ] **Step 3: Write minimal implementation**

In `cmd/collections.go`, add package vars near the other collection vars:

```go
var (
	collectionSaveRevert string
	collectionDiff       bool
)
```

In `init()` (`cmd/collections.go:550-614`), register on each craft_write-capable command (create, rename, items add, items update, views create, views update, views delete, active-view set):

```go
	for _, c := range []*cobra.Command{
		collectionsCreateCmd, collectionsRenameCmd,
		collectionsAddCmd, collectionsUpdateCmd,
		collectionsViewsCreateCmd, collectionsViewsUpdateCmd,
		collectionsViewsDeleteCmd, collectionsActiveViewSetCmd,
	} {
		c.Flags().StringVar(&collectionSaveRevert, "save-revert", "", "Save MCP revertInfo JSON to a file")
		c.Flags().BoolVar(&collectionDiff, "diff", false, "Include MCP edit-review metadata in output when possible")
	}
```

Replace `runMCPCollectionCommand` (`cmd/collections.go:640-661`) with:

```go
func runMCPCollectionCommand(operation, command, tool string) error {
	if tool == "craft_write" {
		return runMCPWriteCommand(operation, command, []string{"mcp", "collections.views"}, collectionSaveRevert, collectionDiff)
	}
	if isDryRun() {
		return dryRunOutput(operation, map[string]interface{}{
			"backend":      "mcp",
			"tool":         tool,
			"command":      command,
			"capabilities": []string{"mcp", "collections.views"},
		})
	}
	client, err := getMCPClient()
	if err != nil {
		return err
	}
	result, err := client.CallTool(tool, map[string]interface{}{"command": command})
	if err != nil {
		return err
	}
	return outputRawJSON(result)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd -run TestCollectionsWriteCmdsRegisterRevertFlags -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/collections.go cmd/revert_flags_test.go
git commit -m "feat(cli): add --save-revert/--diff to collections MCP writes"
```

---

### Task 17: `--save-revert` / `--diff` on `craft update` (title via MCP root-block route)

**Files:**
- Modify: `cmd/update.go:14-23` (vars), `cmd/update.go:61-77` (RunE opening), `cmd/update.go:372-382` (init)
- Test: `cmd/revert_flags_test.go` (append)

**Interfaces:**
- Consumes: `runMCPWriteCommand` (Task 15); `newCLIError(code, message string) error` (`cmd/root.go:57`); `quoteMCPArg` (`cmd/mcp.go:379`); `backendName` package var.
- Produces: `--save-revert`/`--diff` flags on `craft update`. Supported combination: `--title` only (routed as MCP `blocks update --id <docID> --markdown <title>`, since the document root is a block — see `client.UpdateBlockMarkdown`, `internal/api/client.go:511`). Content/section updates with these flags return `CAPABILITY_UNAVAILABLE` pointing at `craft blocks update`.

- [ ] **Step 1: Write the failing test**

Append to `cmd/revert_flags_test.go` (add `"errors"` to the imports):

```go
func TestUpdateCmdRegistersRevertFlags(t *testing.T) {
	if updateCmd.Flags().Lookup("save-revert") == nil {
		t.Error("update missing --save-revert")
	}
	if updateCmd.Flags().Lookup("diff") == nil {
		t.Error("update missing --diff")
	}
}

func TestUpdateRevertFlagsRejectRESTBackend(t *testing.T) {
	origBackend, origSave, origDiff := backendName, updateSaveRevert, updateDiff
	defer func() { backendName, updateSaveRevert, updateDiff = origBackend, origSave, origDiff }()

	backendName = "rest"
	updateSaveRevert = "revert.json"
	updateDiff = false

	err := updateCmd.RunE(updateCmd, []string{"abc123"})
	if err == nil {
		t.Fatal("expected CAPABILITY_UNAVAILABLE error on --backend rest")
	}
	var cerr *cliCodeError
	if !errors.As(err, &cerr) || cerr.Code != "CAPABILITY_UNAVAILABLE" {
		t.Fatalf("expected cliCodeError CAPABILITY_UNAVAILABLE, got %v", err)
	}
}

func TestUpdateRevertFlagsRejectContentUpdates(t *testing.T) {
	origBackend, origSave, origDiff := backendName, updateSaveRevert, updateDiff
	origMarkdown, origTitle := updateMarkdown, updateTitle
	defer func() {
		backendName, updateSaveRevert, updateDiff = origBackend, origSave, origDiff
		updateMarkdown, updateTitle = origMarkdown, origTitle
	}()

	backendName = "auto"
	updateSaveRevert = "revert.json"
	updateDiff = false
	updateTitle = ""
	updateMarkdown = "# new body"

	err := updateCmd.RunE(updateCmd, []string{"abc123"})
	if err == nil {
		t.Fatal("expected CAPABILITY_UNAVAILABLE error for content update with --save-revert")
	}
	var cerr *cliCodeError
	if !errors.As(err, &cerr) || cerr.Code != "CAPABILITY_UNAVAILABLE" {
		t.Fatalf("expected cliCodeError CAPABILITY_UNAVAILABLE, got %v", err)
	}
}
```

`backendName` defaults to `"auto"` (registered at `cmd/root.go:136`); `"auto"` with revert flags routes to MCP, which is why the content-update guard must fire before any MCP call.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd -run 'TestUpdateCmdRegistersRevertFlags|TestUpdateRevertFlags' -v`
Expected: FAIL (build error) with `undefined: updateSaveRevert` and `undefined: updateDiff`

- [ ] **Step 3: Write minimal implementation**

Add to the var block (`cmd/update.go:14-23`):

```go
	updateSaveRevert string
	updateDiff       bool
```

Restructure the opening of `updateCmd.RunE` (`cmd/update.go:61-77`) so validation and the MCP guard run before any REST client is created:

```go
	RunE: func(cmd *cobra.Command, args []string) error {
		docID := args[0]
		if err := validateResourceID(docID, "document-id"); err != nil {
			return err
		}

		if updateSaveRevert != "" || updateDiff {
			if backendName == "rest" {
				return newCLIError("CAPABILITY_UNAVAILABLE", "--save-revert/--diff require MCP capabilities: blocks.revert")
			}
			if updateTitle == "" || updateMarkdown != "" || updateFile != "" || updateStdin || updateSection != "" || updateJSON != "" {
				return newCLIError("CAPABILITY_UNAVAILABLE", "--save-revert/--diff on update support --title-only updates; for content edits use craft blocks update BLOCK_ID --save-revert")
			}
			command := "blocks update --id " + quoteMCPArg(docID) + " --markdown " + quoteMCPArg(updateTitle)
			return runMCPWriteCommand("update title", command, []string{"mcp", "write", "blocks.revert"}, updateSaveRevert, updateDiff)
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		if updateJSON != "" {
```

(the remainder of RunE from the `updateJSON` handling at old line 71 onward is unchanged; only the ordering of `getAPIClient` and docID validation moved).

Register flags in `init()` (`cmd/update.go:372-382`):

```go
	updateCmd.Flags().StringVar(&updateSaveRevert, "save-revert", "", "Save MCP revertInfo JSON to a file (title-only updates)")
	updateCmd.Flags().BoolVar(&updateDiff, "diff", false, "Include MCP edit-review metadata in output when possible (title-only updates)")
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd -run 'TestUpdateCmdRegistersRevertFlags|TestUpdateRevertFlags' -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/update.go cmd/revert_flags_test.go
git commit -m "feat(cli): add --save-revert/--diff to craft update title path"
```

---

### Task 18: `--save-revert` / `--diff` gate on `craft tasks` mutations

**Files:**
- Modify: `cmd/tasks.go` (vars, add/update/delete RunE openings, init)
- Test: `cmd/revert_flags_test.go` (append)

**Interfaces:**
- Consumes: `newCLIError` (`cmd/root.go:57`).
- Produces: `--save-revert`/`--diff` flags registered on `tasks add`, `tasks update`, `tasks delete` for discoverability. Craft MCP exposes no task-write commands in this repo's `craft_write` grammar, so using either flag returns a structured `CAPABILITY_UNAVAILABLE` error naming the working alternative.

- [ ] **Step 1: Write the failing test**

Append to `cmd/revert_flags_test.go`:

```go
func TestTasksCmdsRegisterRevertFlags(t *testing.T) {
	for _, tc := range []struct {
		name string
		has  func(flag string) bool
	}{
		{"tasks add", func(f string) bool { return tasksAddCmd.Flags().Lookup(f) != nil }},
		{"tasks update", func(f string) bool { return tasksUpdateCmd.Flags().Lookup(f) != nil }},
		{"tasks delete", func(f string) bool { return tasksDeleteCmd.Flags().Lookup(f) != nil }},
	} {
		if !tc.has("save-revert") {
			t.Errorf("%s missing --save-revert", tc.name)
		}
		if !tc.has("diff") {
			t.Errorf("%s missing --diff", tc.name)
		}
	}
}

func TestTasksRevertFlagsReturnCapabilityUnavailable(t *testing.T) {
	origSave, origDiff := taskSaveRevert, taskDiff
	defer func() { taskSaveRevert, taskDiff = origSave, origDiff }()

	taskSaveRevert = "revert.json"
	taskDiff = false

	err := tasksDeleteCmd.RunE(tasksDeleteCmd, []string{"task1"})
	if err == nil {
		t.Fatal("expected CAPABILITY_UNAVAILABLE error")
	}
	var cerr *cliCodeError
	if !errors.As(err, &cerr) || cerr.Code != "CAPABILITY_UNAVAILABLE" {
		t.Fatalf("expected cliCodeError CAPABILITY_UNAVAILABLE, got %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd -run 'TestTasksCmdsRegisterRevertFlags|TestTasksRevertFlagsReturnCapabilityUnavailable' -v`
Expected: FAIL (build error) with `undefined: taskSaveRevert` and `undefined: taskDiff`

- [ ] **Step 3: Write minimal implementation**

Add to the tasks var block in `cmd/tasks.go`:

```go
	taskSaveRevert string
	taskDiff       bool
```

Add a guard helper near `readTaskPayload`:

```go
// rejectTaskRevertFlags returns a structured error when --save-revert/--diff
// are used on tasks: Craft MCP exposes no task-write commands, so there is no
// revert metadata to capture for task mutations.
func rejectTaskRevertFlags() error {
	if taskSaveRevert == "" && !taskDiff {
		return nil
	}
	return newCLIError("CAPABILITY_UNAVAILABLE",
		"Craft MCP does not expose task writes; revert metadata is unavailable for tasks. For the task's underlying block, use craft blocks update <block-id> --save-revert")
}
```

Insert as the first statement of `tasksAddCmd.RunE`, `tasksUpdateCmd.RunE`, and `tasksDeleteCmd.RunE`:

```go
		if err := rejectTaskRevertFlags(); err != nil {
			return err
		}
```

Register flags in `init()` (extend the existing loop from Task 13 or add a parallel loop including `tasksDeleteCmd`):

```go
	for _, c := range []*cobra.Command{tasksAddCmd, tasksUpdateCmd, tasksDeleteCmd} {
		c.Flags().StringVar(&taskSaveRevert, "save-revert", "", "Not supported for tasks; returns CAPABILITY_UNAVAILABLE (use craft blocks update --save-revert)")
		c.Flags().BoolVar(&taskDiff, "diff", false, "Not supported for tasks; returns CAPABILITY_UNAVAILABLE")
	}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd -run 'TestTasksCmdsRegisterRevertFlags|TestTasksRevertFlagsReturnCapabilityUnavailable' -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/tasks.go cmd/revert_flags_test.go
git commit -m "feat(cli): register --save-revert/--diff capability gate on tasks mutations"
```

---

### Task 19: Schema coverage, docs regeneration, and full verification

**Files:**
- Create: `cmd/schema_features_test.go`
- Modify: `docs/command-reference.json` (regenerated), `README.md`, `docs/llm/README.md`, `AGENTS.md`
- Test: `cmd/schema_features_test.go` plus the full suite

**Interfaces:**
- Consumes: everything above; existing helpers `findCommandSchema` and `hasFlagSchema` from `cmd/schema_test.go:280-305` (same package).
- Produces: schema regression coverage for all new flags; regenerated machine manifest; updated docs.

- [ ] **Step 1: Write the failing-then-passing schema test**

Create `cmd/schema_features_test.go` (this test passes immediately if Tasks 5-18 landed correctly; it exists to lock the contract):

```go
package cmd

import (
	"strings"
	"testing"
)

func TestBuildSchemaIncludesBatchingAndFeatureFlags(t *testing.T) {
	rootSchema := buildSchema(rootCmd)

	del := findCommandSchema(rootSchema.Subcommands, "delete")
	if del == nil {
		t.Fatal("expected delete command in schema")
	}
	if !strings.Contains(del.Usage, "<document-id...>") {
		t.Errorf("delete usage %q does not advertise multiple ids", del.Usage)
	}

	mv := findCommandSchema(rootSchema.Subcommands, "move")
	if mv == nil {
		t.Fatal("expected move command in schema")
	}
	if !strings.Contains(mv.Usage, "<document-id...>") {
		t.Errorf("move usage %q does not advertise multiple ids", mv.Usage)
	}

	list := findCommandSchema(rootSchema.Subcommands, "list")
	if list == nil {
		t.Fatal("expected list command in schema")
	}
	for _, flag := range []string{"--daily-note-after", "--daily-note-before"} {
		if !hasFlagSchema(list.Flags, flag) {
			t.Errorf("list missing %s flag", flag)
		}
	}

	tasks := findCommandSchema(rootSchema.Subcommands, "tasks")
	if tasks == nil {
		t.Fatal("expected tasks command in schema")
	}
	tasksAdd := findCommandSchema(tasks.Subcommands, "add")
	if tasksAdd == nil {
		t.Fatal("expected tasks add in schema")
	}
	for _, flag := range []string{"--repeat", "--repeat-interval", "--repeat-weekdays",
		"--repeat-end", "--repeat-skip-weekends", "--repeat-dynamic-days",
		"--repeat-reminder", "--repeat-frequency", "--save-revert", "--diff"} {
		if !hasFlagSchema(tasksAdd.Flags, flag) {
			t.Errorf("tasks add missing %s flag", flag)
		}
	}

	upd := findCommandSchema(rootSchema.Subcommands, "update")
	if upd == nil {
		t.Fatal("expected update command in schema")
	}
	for _, flag := range []string{"--save-revert", "--diff"} {
		if !hasFlagSchema(upd.Flags, flag) {
			t.Errorf("update missing %s flag", flag)
		}
	}
}
```

- [ ] **Step 2: Run the schema test**

Run: `go test ./cmd -run TestBuildSchemaIncludesBatchingAndFeatureFlags -v`
Expected: PASS. If any assertion fails, a prior task was not completed — fix that task, not this test.

- [ ] **Step 3: Regenerate the machine manifest and update docs**

```bash
go build -o craft .
./craft schema > docs/command-reference.json
git diff --stat docs/command-reference.json   # confirm new flags/usages appear
```

Documentation edits (exact content, locate the matching existing sections):

1. `README.md` — in the `craft delete`, `craft move`, `craft tasks delete`, `craft folders delete`, and `craft blocks delete` example sections, add one batch example each, e.g.:

```
craft delete DOC_ID1 DOC_ID2 DOC_ID3        # Batch delete in one API request
craft move DOC_ID1 DOC_ID2 --to-folder FID  # Batch move
```

   In the `craft list` section add:

```
craft list --location daily_notes --daily-note-after 2026-08-01 --daily-note-before 2026-08-15
```

   In the `craft tasks` section add:

```
craft tasks add "Standup" --repeat daily --repeat-skip-weekends --repeat-reminder 09:00
```

   In the `craft search` section add:

```
craft search "budget" --folder ID1,ID2          # Multi-folder scope
craft search "roadmap" --document DOC1,DOC2     # Multi-document scope
```

2. `AGENTS.md` — in "Agent guardrails", add one line:

```
- Batch deletes and moves accept multiple positional IDs and send one API request: `craft delete ID1 ID2`, `craft move ID1 ID2 --to-folder FID`, `craft tasks delete ID1 ID2`, `craft folders delete ID1 ID2`, `craft blocks delete ID1 ID2`
```

   In "Common pitfalls", add:

```
- `--save-revert`/`--diff` on `craft tasks ...` always return `CAPABILITY_UNAVAILABLE`: Craft MCP has no task-write surface. Use `craft blocks update <block-id> --save-revert` on the task's block instead. On `craft update` they support title-only updates.
```

3. `docs/llm/README.md` — note that `docs/command-reference.json` was regenerated and list the new flags in whichever discovery section enumerates flags (follow the file's existing structure).

- [ ] **Step 4: Full verification**

```bash
go test ./...
go vet ./...
go build -o craft .
./craft audit agent-dx --format json
./craft schema | head -c 400   # sanity: valid JSON manifest
```

Expected: all tests PASS; vet clean; audit reports `"score": 85, "max": 85`. If the score regressed, inspect `"failures"` in the audit JSON — the checks are computed from `buildSchema(rootCmd)` and repo files (`cmd/audit.go:83`), so a failure means a doc or schema contract broke in Tasks 5-18.

- [ ] **Step 5: Commit**

```bash
git add cmd/schema_features_test.go docs/command-reference.json README.md AGENTS.md docs/llm/README.md
git commit -m "docs: regenerate command reference and document batching, repeat, scope, and revert flags"
```

---

## Self-Review Notes (completed at plan time)

- **Spec coverage:** (1) multi-ID batching → Tasks 1-9; (2) daily-note range → Task 10 (client fields pre-existed at `internal/api/client.go:1476-1477`); (3) repeat config → Tasks 11-13; (4) search multi-scope → Task 14 (client fields pre-existed at `internal/api/client.go:1400-1401`); (5) universal revert/diff → Tasks 15-18; (6) schema + verification → Task 19.
- **Deviation from investigation, deliberate:** the investigation named `DeleteBlocks([]string)` as new API work — it already exists (`internal/api/client.go:499`), so Task 6 is CLI-only. It also assumed tasks/update are MCP-capable; they are not (see Global Constraints design decision).
- **Type consistency:** `DeleteDocuments/DeleteTasks/DeleteFolders/MoveDocuments` names match between Tasks 1-4 (producers) and 5-9 (consumers). `repeatFlagValues`/`buildRepeatConfig` match between Tasks 13's test and implementation. `completeMCPWrite`/`runMCPWriteCommand` signatures match between Task 15 (producer) and Tasks 16-17 (consumers). `RepeatConfig` field names match Tasks 11, 12, 13.
- **Placeholder scan:** every code block is complete and compiles as written; no TBD/TODO markers remain.
