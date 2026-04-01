# Craft CLI Full API Parity Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Bring craft-cli to 100% parity with the Craft Space API documented in `docs/HHRuPxZZTJ6-openapi.json`.

**Architecture:** Add missing endpoints as new API client methods following the existing `doRequest` pattern, new Cobra commands mirroring the existing command group structure (collections, connection, upload, comments), and enhance existing commands (search, blocks, list) with missing flags. Models go in `internal/models/document.go`, client methods in `internal/api/client.go`, commands in `cmd/`.

**Tech Stack:** Go, Cobra CLI framework, net/http, encoding/json, httptest for tests.

---

## Task 1: Add Collection Models

**Files:**
- Modify: `internal/models/document.go`

**Step 1: Add collection types to models**

Add after the existing `ErrorResponse` struct (line 212):

```go
// Collection represents a Craft collection (database)
type Collection struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ItemCount   int    `json:"itemCount"`
	DocumentID  string `json:"documentId"`
}

// CollectionList represents a list of collections
type CollectionList struct {
	Items []Collection `json:"items"`
}

// CollectionSchema represents a collection's schema
type CollectionSchema struct {
	Key                string                    `json:"key"`
	Name               string                    `json:"name"`
	ContentPropDetails *CollectionPropDetails    `json:"contentPropDetails,omitempty"`
	Properties         []CollectionProperty      `json:"properties"`
}

// CollectionPropDetails describes the content/title property
type CollectionPropDetails struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// CollectionProperty represents a property in a collection schema
type CollectionProperty struct {
	Key     string   `json:"key"`
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Options []string `json:"options,omitempty"`
}

// CollectionItem represents an item in a collection
type CollectionItem struct {
	ID         string            `json:"id"`
	Title      string            `json:"title"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Content    []Block           `json:"content,omitempty"`
}

// CollectionItemList represents a list of collection items
type CollectionItemList struct {
	Items []CollectionItem `json:"items"`
}

// ConnectionInfo represents the response from GET /connection
type ConnectionInfo struct {
	Space struct {
		ID           string `json:"id"`
		Timezone     string `json:"timezone"`
		Time         string `json:"time"`
		FriendlyDate string `json:"friendlyDate"`
	} `json:"space"`
	UTC struct {
		Time string `json:"time"`
	} `json:"utc"`
	URLTemplates struct {
		App string `json:"app"`
	} `json:"urlTemplates"`
}

// UploadResponse represents the response from POST /upload
type UploadResponse struct {
	BlockID  string `json:"blockId"`
	AssetURL string `json:"assetUrl"`
}

// CommentResponse represents the response from POST /comments
type CommentResponse struct {
	Items []struct {
		CommentID string `json:"commentId"`
	} `json:"items"`
}

// BlockSearchResult represents a single block search match
type BlockSearchResult struct {
	BlockID       string           `json:"blockId"`
	Markdown      string           `json:"markdown"`
	PageBlockPath []PageBlockEntry `json:"pageBlockPath,omitempty"`
	BeforeBlocks  []BlockContext   `json:"beforeBlocks,omitempty"`
	AfterBlocks   []BlockContext   `json:"afterBlocks,omitempty"`
}

// PageBlockEntry represents a path entry in block search results
type PageBlockEntry struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

// BlockContext represents a surrounding block in search results
type BlockContext struct {
	BlockID  string `json:"blockId"`
	Markdown string `json:"markdown"`
}

// BlockSearchResultList represents block search results
type BlockSearchResultList struct {
	Items []BlockSearchResult `json:"items"`
}
```

**Step 2: Run tests**

Run: `cd "/Users/nerveband/wavedepth Dropbox/Ashraf Ali/Mac (2)/Documents/GitHub/craft-cli" && go build ./...`
Expected: Compiles cleanly

**Step 3: Commit**

```bash
git add internal/models/document.go
git commit -m "feat: add models for collections, connection, upload, comments, block search"
```

---

## Task 2: Add Collection API Client Methods

**Files:**
- Modify: `internal/api/client.go`

**Step 1: Write tests for collection client methods**

Create `internal/api/collections_test.go`:

```go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetCollections(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/collections" {
			t.Errorf("expected /collections, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"id": "col1", "name": "Tasks", "itemCount": 5, "documentId": "doc1"},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.GetCollections("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 collection, got %d", len(result.Items))
	}
	if result.Items[0].Name != "Tasks" {
		t.Errorf("expected name Tasks, got %s", result.Items[0].Name)
	}
}

func TestGetCollectionSchema(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/collections/col1/schema" {
			t.Errorf("expected /collections/col1/schema, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"key": "tasks", "name": "Tasks",
			"properties": []map[string]interface{}{
				{"key": "status", "name": "Status", "type": "select"},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.GetCollectionSchema("col1", "schema")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "Tasks" {
		t.Errorf("expected name Tasks, got %s", result.Name)
	}
}

func TestGetCollectionItems(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/collections/col1/items" {
			t.Errorf("expected /collections/col1/items, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"id": "item1", "title": "Task 1", "properties": map[string]interface{}{"status": "Done"}},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.GetCollectionItems("col1", -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}
	if result.Items[0].Title != "Task 1" {
		t.Errorf("expected title Task 1, got %s", result.Items[0].Title)
	}
}

func TestAddCollectionItem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"id": "item-new", "title": "New Task", "properties": map[string]interface{}{"status": "Todo"}},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.AddCollectionItem("col1", "New Task", map[string]interface{}{"status": "Todo"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Items[0].ID != "item-new" {
		t.Errorf("expected id item-new, got %s", result.Items[0].ID)
	}
}

func TestUpdateCollectionItem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"id": "item1", "title": "Task 1", "properties": map[string]interface{}{"status": "Done"}},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.UpdateCollectionItem("col1", "item1", map[string]interface{}{"status": "Done"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteCollectionItem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{{"id": "item1"}},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.DeleteCollectionItem("col1", "item1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `cd "/Users/nerveband/wavedepth Dropbox/Ashraf Ali/Mac (2)/Documents/GitHub/craft-cli" && go test ./internal/api/ -run TestGetCollections -v`
Expected: FAIL - methods don't exist

**Step 3: Implement collection client methods**

Add to `internal/api/client.go` after the task operations section:

```go
// ========== Collection Operations ==========

// GetCollections retrieves all collections, optionally filtered by document IDs
func (c *Client) GetCollections(documentIDs string) (*models.CollectionList, error) {
	path := "/collections"
	if documentIDs != "" {
		path = fmt.Sprintf("/collections?documentIds=%s", url.QueryEscape(documentIDs))
	}

	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result models.CollectionList
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &result, nil
}

// GetCollectionSchema retrieves the schema for a collection
func (c *Client) GetCollectionSchema(collectionID, format string) (*models.CollectionSchema, error) {
	path := fmt.Sprintf("/collections/%s/schema", url.PathEscape(collectionID))
	if format != "" {
		path += "?format=" + url.QueryEscape(format)
	}

	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result models.CollectionSchema
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &result, nil
}

// GetCollectionItems retrieves items from a collection
func (c *Client) GetCollectionItems(collectionID string, maxDepth int) (*models.CollectionItemList, error) {
	path := fmt.Sprintf("/collections/%s/items", url.PathEscape(collectionID))
	if maxDepth != -1 {
		path += fmt.Sprintf("?maxDepth=%d", maxDepth)
	}

	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result models.CollectionItemList
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &result, nil
}

// AddCollectionItem adds an item to a collection
func (c *Client) AddCollectionItem(collectionID, title string, properties map[string]interface{}, allowNewOptions bool) (*models.CollectionItemList, error) {
	path := fmt.Sprintf("/collections/%s/items", url.PathEscape(collectionID))

	req := struct {
		Items []struct {
			Title      string                 `json:"title"`
			Properties map[string]interface{} `json:"properties,omitempty"`
		} `json:"items"`
		AllowNewSelectOptions bool `json:"allowNewSelectOptions"`
	}{
		Items: []struct {
			Title      string                 `json:"title"`
			Properties map[string]interface{} `json:"properties,omitempty"`
		}{{Title: title, Properties: properties}},
		AllowNewSelectOptions: allowNewOptions,
	}

	data, err := c.doRequest("POST", path, req)
	if err != nil {
		return nil, err
	}

	var result models.CollectionItemList
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &result, nil
}

// UpdateCollectionItem updates an item in a collection
func (c *Client) UpdateCollectionItem(collectionID, itemID string, properties map[string]interface{}, allowNewOptions bool) error {
	path := fmt.Sprintf("/collections/%s/items", url.PathEscape(collectionID))

	req := struct {
		ItemsToUpdate []struct {
			ID         string                 `json:"id"`
			Properties map[string]interface{} `json:"properties,omitempty"`
		} `json:"itemsToUpdate"`
		AllowNewSelectOptions bool `json:"allowNewSelectOptions"`
	}{
		ItemsToUpdate: []struct {
			ID         string                 `json:"id"`
			Properties map[string]interface{} `json:"properties,omitempty"`
		}{{ID: itemID, Properties: properties}},
		AllowNewSelectOptions: allowNewOptions,
	}

	_, err := c.doRequest("PUT", path, req)
	return err
}

// DeleteCollectionItem deletes an item from a collection
func (c *Client) DeleteCollectionItem(collectionID, itemID string) error {
	path := fmt.Sprintf("/collections/%s/items", url.PathEscape(collectionID))

	req := struct {
		IDsToDelete []string `json:"idsToDelete"`
	}{
		IDsToDelete: []string{itemID},
	}

	_, err := c.doRequest("DELETE", path, req)
	return err
}
```

**Step 4: Run tests**

Run: `cd "/Users/nerveband/wavedepth Dropbox/Ashraf Ali/Mac (2)/Documents/GitHub/craft-cli" && go test ./internal/api/ -run TestGetCollection -v && go test ./internal/api/ -run TestAddCollection -v && go test ./internal/api/ -run TestUpdateCollection -v && go test ./internal/api/ -run TestDeleteCollection -v`
Expected: All PASS

**Step 5: Commit**

```bash
git add internal/api/client.go internal/api/collections_test.go
git commit -m "feat: add collection API client methods with tests"
```

---

## Task 3: Add Connection, Upload, Comments, and Block Search API Client Methods

**Files:**
- Modify: `internal/api/client.go`

**Step 1: Write tests**

Create `internal/api/endpoints_test.go`:

```go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/connection" {
			t.Errorf("expected /connection, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"space": map[string]interface{}{
				"id": "space-123", "timezone": "America/New_York",
				"time": "2026-01-30T10:00:00", "friendlyDate": "Thursday, January 30, 2026",
			},
			"utc":          map[string]interface{}{"time": "2026-01-30T15:00:00Z"},
			"urlTemplates": map[string]interface{}{"app": "craftdocs://open?spaceId={spaceId}&documentId={documentId}"},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.GetConnection()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Space.ID != "space-123" {
		t.Errorf("expected space id space-123, got %s", result.Space.ID)
	}
	if result.Space.Timezone != "America/New_York" {
		t.Errorf("expected timezone America/New_York, got %s", result.Space.Timezone)
	}
}

func TestAddComment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/comments" {
			t.Errorf("expected /comments, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"commentId": "comment-123"},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.AddComment("block-1", "Test comment")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(result.Items))
	}
}

func TestSearchBlocks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/blocks/search" {
			t.Errorf("expected /blocks/search, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("blockId") != "doc1" {
			t.Errorf("expected blockId=doc1, got %s", r.URL.Query().Get("blockId"))
		}
		if r.URL.Query().Get("pattern") != "test.*pattern" {
			t.Errorf("expected pattern, got %s", r.URL.Query().Get("pattern"))
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{
					"blockId":  "block-5",
					"markdown": "test matching pattern here",
					"pageBlockPath": []map[string]interface{}{
						{"id": "0", "content": "Doc Title"},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.SearchBlocks("doc1", "test.*pattern", false, 5, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Items))
	}
	if result.Items[0].BlockID != "block-5" {
		t.Errorf("expected blockId block-5, got %s", result.Items[0].BlockID)
	}
}

func TestSearchDocumentsAdvanced(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("regexps") != "test.*" {
			t.Errorf("expected regexps param, got %s", r.URL.Query().Get("regexps"))
		}
		if r.URL.Query().Get("location") != "unsorted" {
			t.Errorf("expected location=unsorted, got %s", r.URL.Query().Get("location"))
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"documentId": "doc-1", "markdown": "test content"},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	opts := SearchOptions{
		Regexps:  "test.*",
		Location: "unsorted",
	}
	result, err := client.SearchDocumentsAdvanced("", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Items))
	}
}

func TestGetDocumentsAdvanced(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("fetchMetadata") != "true" {
			t.Errorf("expected fetchMetadata=true")
		}
		if r.URL.Query().Get("createdDateGte") != "2026-01-01" {
			t.Errorf("expected createdDateGte=2026-01-01")
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"id": "doc-1", "title": "Test"},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	opts := ListDocumentsOptions{
		FetchMetadata:  true,
		CreatedDateGte: "2026-01-01",
	}
	result, err := client.GetDocumentsAdvanced(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 document, got %d", len(result.Items))
	}
}

func TestGetBlockByDate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("date") != "today" {
			t.Errorf("expected date=today, got %s", r.URL.Query().Get("date"))
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "dn-123", "type": "page", "markdown": "Thursday, January 30",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	block, err := client.GetBlockByDate("today", -1, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if block.ID != "dn-123" {
		t.Errorf("expected id dn-123, got %s", block.ID)
	}
}

func TestAddBlockToDate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		pos := body["position"].(map[string]interface{})
		if pos["date"] != "today" {
			t.Errorf("expected position.date=today, got %v", pos["date"])
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"id": "block-new", "type": "text", "markdown": "New content"},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	block, err := client.AddBlockToDate("today", "New content", "end")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if block.ID != "block-new" {
		t.Errorf("expected id block-new, got %s", block.ID)
	}
}

func TestUploadFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/upload" {
			t.Errorf("expected /upload, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("pageId") != "doc1" {
			t.Errorf("expected pageId=doc1")
		}
		if r.URL.Query().Get("position") != "end" {
			t.Errorf("expected position=end")
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"blockId":  "block-img",
			"assetUrl": "https://res.luki.io/image.png",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.UploadFile([]byte("fake-image-data"), "doc1", "", "", "end")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.BlockID != "block-img" {
		t.Errorf("expected blockId block-img, got %s", result.BlockID)
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `cd "/Users/nerveband/wavedepth Dropbox/Ashraf Ali/Mac (2)/Documents/GitHub/craft-cli" && go test ./internal/api/ -run TestGetConnection -v`
Expected: FAIL

**Step 3: Implement all new API client methods**

Add to `internal/api/client.go`:

```go
// ========== Connection Operations ==========

// GetConnection retrieves space connection metadata
func (c *Client) GetConnection() (*models.ConnectionInfo, error) {
	data, err := c.doRequest("GET", "/connection", nil)
	if err != nil {
		return nil, err
	}

	var result models.ConnectionInfo
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &result, nil
}

// ========== Comment Operations ==========

// AddComment adds a comment to a block
func (c *Client) AddComment(blockID, content string) (*models.CommentResponse, error) {
	req := struct {
		Comments []struct {
			BlockID string `json:"blockId"`
			Content string `json:"content"`
		} `json:"comments"`
	}{
		Comments: []struct {
			BlockID string `json:"blockId"`
			Content string `json:"content"`
		}{{BlockID: blockID, Content: content}},
	}

	data, err := c.doRequest("POST", "/comments", req)
	if err != nil {
		return nil, err
	}

	var result models.CommentResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &result, nil
}

// ========== Block Search ==========

// SearchBlocks searches within a document using regex
func (c *Client) SearchBlocks(blockID, pattern string, caseSensitive bool, beforeCount, afterCount int) (*models.BlockSearchResultList, error) {
	params := []string{
		"blockId=" + url.QueryEscape(blockID),
		"pattern=" + url.QueryEscape(pattern),
	}
	if caseSensitive {
		params = append(params, "caseSensitive=true")
	}
	if beforeCount >= 0 {
		params = append(params, fmt.Sprintf("beforeBlockCount=%d", beforeCount))
	}
	if afterCount >= 0 {
		params = append(params, fmt.Sprintf("afterBlockCount=%d", afterCount))
	}

	path := "/blocks/search?" + strings.Join(params, "&")
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result models.BlockSearchResultList
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &result, nil
}

// ========== Advanced Search ==========

// SearchOptions holds optional filters for document search
type SearchOptions struct {
	Regexps             string
	Location            string
	FolderIDs           string
	DocumentIDs         string
	FetchMetadata       bool
	CreatedDateGte      string
	CreatedDateLte      string
	LastModifiedDateGte string
	LastModifiedDateLte string
	DailyNoteDateGte    string
	DailyNoteDateLte    string
}

// SearchDocumentsAdvanced searches documents with full filter support
func (c *Client) SearchDocumentsAdvanced(query string, opts SearchOptions) (*models.SearchResult, error) {
	var params []string
	if query != "" {
		params = append(params, "include="+url.QueryEscape(query))
	}
	if opts.Regexps != "" {
		params = append(params, "regexps="+url.QueryEscape(opts.Regexps))
	}
	if opts.Location != "" {
		params = append(params, "location="+url.QueryEscape(opts.Location))
	}
	if opts.FolderIDs != "" {
		params = append(params, "folderIds="+url.QueryEscape(opts.FolderIDs))
	}
	if opts.DocumentIDs != "" {
		params = append(params, "documentIds="+url.QueryEscape(opts.DocumentIDs))
	}
	if opts.FetchMetadata {
		params = append(params, "fetchMetadata=true")
	}
	if opts.CreatedDateGte != "" {
		params = append(params, "createdDateGte="+url.QueryEscape(opts.CreatedDateGte))
	}
	if opts.CreatedDateLte != "" {
		params = append(params, "createdDateLte="+url.QueryEscape(opts.CreatedDateLte))
	}
	if opts.LastModifiedDateGte != "" {
		params = append(params, "lastModifiedDateGte="+url.QueryEscape(opts.LastModifiedDateGte))
	}
	if opts.LastModifiedDateLte != "" {
		params = append(params, "lastModifiedDateLte="+url.QueryEscape(opts.LastModifiedDateLte))
	}
	if opts.DailyNoteDateGte != "" {
		params = append(params, "dailyNoteDateGte="+url.QueryEscape(opts.DailyNoteDateGte))
	}
	if opts.DailyNoteDateLte != "" {
		params = append(params, "dailyNoteDateLte="+url.QueryEscape(opts.DailyNoteDateLte))
	}

	path := "/documents/search"
	if len(params) > 0 {
		path += "?" + strings.Join(params, "&")
	}

	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result models.SearchResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &result, nil
}

// ========== Advanced Document Listing ==========

// ListDocumentsOptions holds optional filters for document listing
type ListDocumentsOptions struct {
	FolderID            string
	Location            string
	FetchMetadata       bool
	CreatedDateGte      string
	CreatedDateLte      string
	LastModifiedDateGte string
	LastModifiedDateLte string
	DailyNoteDateGte    string
	DailyNoteDateLte    string
}

// GetDocumentsAdvanced retrieves documents with full filter support
func (c *Client) GetDocumentsAdvanced(opts ListDocumentsOptions) (*models.DocumentList, error) {
	var params []string
	if opts.FolderID != "" {
		params = append(params, "folderId="+url.QueryEscape(opts.FolderID))
	}
	if opts.Location != "" {
		params = append(params, "location="+url.QueryEscape(opts.Location))
	}
	if opts.FetchMetadata {
		params = append(params, "fetchMetadata=true")
	}
	if opts.CreatedDateGte != "" {
		params = append(params, "createdDateGte="+url.QueryEscape(opts.CreatedDateGte))
	}
	if opts.CreatedDateLte != "" {
		params = append(params, "createdDateLte="+url.QueryEscape(opts.CreatedDateLte))
	}
	if opts.LastModifiedDateGte != "" {
		params = append(params, "lastModifiedDateGte="+url.QueryEscape(opts.LastModifiedDateGte))
	}
	if opts.LastModifiedDateLte != "" {
		params = append(params, "lastModifiedDateLte="+url.QueryEscape(opts.LastModifiedDateLte))
	}
	if opts.DailyNoteDateGte != "" {
		params = append(params, "dailyNoteDateGte="+url.QueryEscape(opts.DailyNoteDateGte))
	}
	if opts.DailyNoteDateLte != "" {
		params = append(params, "dailyNoteDateLte="+url.QueryEscape(opts.DailyNoteDateLte))
	}

	path := "/documents"
	if len(params) > 0 {
		path += "?" + strings.Join(params, "&")
	}

	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result models.DocumentList
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &result, nil
}

// ========== Daily Note Block Access ==========

// GetBlockByDate retrieves the root block of a daily note by date
func (c *Client) GetBlockByDate(date string, maxDepth int, fetchMetadata bool) (*models.Block, error) {
	params := []string{"date=" + url.QueryEscape(date)}
	if maxDepth != -1 {
		params = append(params, fmt.Sprintf("maxDepth=%d", maxDepth))
	}
	if fetchMetadata {
		params = append(params, "fetchMetadata=true")
	}

	path := "/blocks?" + strings.Join(params, "&")
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var block models.Block
	if err := json.Unmarshal(data, &block); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &block, nil
}

// AddBlockToDate adds a block to a daily note by date
func (c *Client) AddBlockToDate(date, markdown, position string) (*models.Block, error) {
	req := struct {
		Markdown string `json:"markdown"`
		Position struct {
			Date     string `json:"date"`
			Position string `json:"position"`
		} `json:"position"`
	}{
		Markdown: markdown,
	}
	req.Position.Date = date
	req.Position.Position = position

	data, err := c.doRequest("POST", "/blocks", req)
	if err != nil {
		return nil, err
	}

	var resp addBlockResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("no block returned from API")
	}

	return &models.Block{
		ID:       resp.Items[0].ID,
		Type:     resp.Items[0].Type,
		Markdown: resp.Items[0].Markdown,
	}, nil
}

// ========== Upload Operations ==========

// doRequestRaw performs an HTTP request with raw bytes body
func (c *Client) doRequestRaw(method, path string, body []byte, contentType string) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewBuffer(body)
	}

	reqURL := fmt.Sprintf("%s%s", c.baseURL, path)
	req, err := http.NewRequest(method, reqURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, c.handleErrorResponse(resp.StatusCode, respBody)
	}

	return respBody, nil
}

// UploadFile uploads a file to a document or daily note
// Provide either pageID or date (not both). siblingID is for before/after positioning.
func (c *Client) UploadFile(fileData []byte, pageID, date, siblingID, position string) (*models.UploadResponse, error) {
	params := []string{"position=" + url.QueryEscape(position)}
	if pageID != "" {
		params = append(params, "pageId="+url.QueryEscape(pageID))
	}
	if date != "" {
		params = append(params, "date="+url.QueryEscape(date))
	}
	if siblingID != "" {
		params = append(params, "siblingId="+url.QueryEscape(siblingID))
	}

	path := "/upload?" + strings.Join(params, "&")
	data, err := c.doRequestRaw("POST", path, fileData, "application/octet-stream")
	if err != nil {
		return nil, err
	}

	var result models.UploadResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &result, nil
}

// ========== Enhanced Block Get ==========

// GetBlockWithOptions retrieves a block with depth and metadata options
func (c *Client) GetBlockWithOptions(blockID string, maxDepth int, fetchMetadata bool) (*models.Block, error) {
	params := []string{"id=" + url.QueryEscape(blockID)}
	if maxDepth != -1 {
		params = append(params, fmt.Sprintf("maxDepth=%d", maxDepth))
	}
	if fetchMetadata {
		params = append(params, "fetchMetadata=true")
	}

	path := "/blocks?" + strings.Join(params, "&")
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var block models.Block
	if err := json.Unmarshal(data, &block); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &block, nil
}
```

**Step 4: Run all new tests**

Run: `cd "/Users/nerveband/wavedepth Dropbox/Ashraf Ali/Mac (2)/Documents/GitHub/craft-cli" && go test ./internal/api/ -v`
Expected: All PASS

**Step 5: Commit**

```bash
git add internal/api/client.go internal/api/endpoints_test.go
git commit -m "feat: add connection, upload, comments, block search, advanced search/list API methods"
```

---

## Task 4: Add Collections Command Group

**Files:**
- Create: `cmd/collections.go`

**Step 1: Implement collections commands**

Create `cmd/collections.go`:

```go
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ashrafali/craft-cli/internal/models"
	"github.com/spf13/cobra"
)

var collectionsCmd = &cobra.Command{
	Use:   "collections",
	Short: "Manage collections (databases)",
	Long: `Manage Craft collections - list, view schema, and manage items.

Examples:
  craft collections list
  craft collections schema COLLECTION_ID
  craft collections items COLLECTION_ID
  craft collections add COLLECTION_ID --title "New Item" --properties '{"status":"Todo"}'
  craft collections update COLLECTION_ID --item ITEM_ID --properties '{"status":"Done"}'
  craft collections delete COLLECTION_ID --item ITEM_ID`,
}

var collectionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all collections",
	Long:  "List all collections in the space, optionally filtered by document",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		result, err := client.GetCollections(collectionDocFilter)
		if err != nil {
			return err
		}

		format := getOutputFormat()
		return outputCollections(result.Items, format)
	},
}

var collectionsSchemaCmd = &cobra.Command{
	Use:   "schema [collection-id]",
	Short: "Get collection schema",
	Long: `Retrieve the schema for a collection.

Formats:
  schema           - Editable structure with property details
  json-schema-items - JSON Schema for validation`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		result, err := client.GetCollectionSchema(args[0], collectionSchemaFormat)
		if err != nil {
			return err
		}

		return outputJSON(result)
	},
}

var collectionsItemsCmd = &cobra.Command{
	Use:   "items [collection-id]",
	Short: "List collection items",
	Long:  "Retrieve all items from a collection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		result, err := client.GetCollectionItems(args[0], collectionMaxDepth)
		if err != nil {
			return err
		}

		format := getOutputFormat()
		return outputCollectionItems(result.Items, format)
	},
}

var collectionsAddCmd = &cobra.Command{
	Use:   "add [collection-id]",
	Short: "Add an item to a collection",
	Long: `Add a new item to a collection with title and properties.

Examples:
  craft collections add col1 --title "New Task" --properties '{"status":"Todo","priority":"High"}'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if collectionItemTitle == "" {
			return fmt.Errorf("--title is required")
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		var props map[string]interface{}
		if collectionItemProps != "" {
			if err := json.Unmarshal([]byte(collectionItemProps), &props); err != nil {
				return fmt.Errorf("invalid --properties JSON: %w", err)
			}
		}

		if isDryRun() {
			fmt.Printf("[dry-run] Would add item \"%s\" to collection %s\n", collectionItemTitle, args[0])
			return nil
		}

		result, err := client.AddCollectionItem(args[0], collectionItemTitle, props, collectionAllowNewOpts)
		if err != nil {
			return err
		}

		if isQuiet() && len(result.Items) > 0 {
			fmt.Println(result.Items[0].ID)
			return nil
		}

		format := getOutputFormat()
		if format == "json" {
			return outputJSON(result)
		}
		if len(result.Items) > 0 {
			fmt.Printf("Item created: %s (ID: %s)\n", result.Items[0].Title, result.Items[0].ID)
		}
		return nil
	},
}

var collectionsUpdateCmd = &cobra.Command{
	Use:   "update [collection-id]",
	Short: "Update a collection item",
	Long: `Update properties of an existing collection item.

Examples:
  craft collections update col1 --item item1 --properties '{"status":"Done"}'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if collectionItemID == "" {
			return fmt.Errorf("--item is required")
		}
		if collectionItemProps == "" {
			return fmt.Errorf("--properties is required")
		}

		var props map[string]interface{}
		if err := json.Unmarshal([]byte(collectionItemProps), &props); err != nil {
			return fmt.Errorf("invalid --properties JSON: %w", err)
		}

		if isDryRun() {
			fmt.Printf("[dry-run] Would update item %s in collection %s\n", collectionItemID, args[0])
			return nil
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		if err := client.UpdateCollectionItem(args[0], collectionItemID, props, collectionAllowNewOpts); err != nil {
			return err
		}

		if !isQuiet() {
			fmt.Printf("Item %s updated\n", collectionItemID)
		}
		return nil
	},
}

var collectionsDeleteCmd = &cobra.Command{
	Use:   "delete [collection-id]",
	Short: "Delete a collection item",
	Long:  "Delete an item from a collection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if collectionItemID == "" {
			return fmt.Errorf("--item is required")
		}

		if isDryRun() {
			fmt.Printf("[dry-run] Would delete item %s from collection %s\n", collectionItemID, args[0])
			return nil
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		if err := client.DeleteCollectionItem(args[0], collectionItemID); err != nil {
			return err
		}

		if !isQuiet() {
			fmt.Printf("Item %s deleted\n", collectionItemID)
		}
		return nil
	},
}

// Flag variables
var (
	collectionDocFilter    string
	collectionSchemaFormat string
	collectionMaxDepth     int
	collectionItemTitle    string
	collectionItemProps    string
	collectionItemID       string
	collectionAllowNewOpts bool
)

func init() {
	rootCmd.AddCommand(collectionsCmd)

	collectionsCmd.AddCommand(collectionsListCmd)
	collectionsListCmd.Flags().StringVar(&collectionDocFilter, "document", "", "Filter by document ID")

	collectionsCmd.AddCommand(collectionsSchemaCmd)
	collectionsSchemaCmd.Flags().StringVar(&collectionSchemaFormat, "schema-format", "schema", "Schema format: schema or json-schema-items")

	collectionsCmd.AddCommand(collectionsItemsCmd)
	collectionsItemsCmd.Flags().IntVar(&collectionMaxDepth, "depth", -1, "Max depth for nested content (-1 for all)")

	collectionsCmd.AddCommand(collectionsAddCmd)
	collectionsAddCmd.Flags().StringVar(&collectionItemTitle, "title", "", "Item title")
	collectionsAddCmd.Flags().StringVar(&collectionItemProps, "properties", "", "Item properties as JSON")
	collectionsAddCmd.Flags().BoolVar(&collectionAllowNewOpts, "allow-new-options", false, "Allow creating new select options")
	collectionsAddCmd.MarkFlagRequired("title")

	collectionsCmd.AddCommand(collectionsUpdateCmd)
	collectionsUpdateCmd.Flags().StringVar(&collectionItemID, "item", "", "Item ID to update")
	collectionsUpdateCmd.Flags().StringVar(&collectionItemProps, "properties", "", "Properties to update as JSON")
	collectionsUpdateCmd.Flags().BoolVar(&collectionAllowNewOpts, "allow-new-options", false, "Allow creating new select options")
	collectionsUpdateCmd.MarkFlagRequired("item")
	collectionsUpdateCmd.MarkFlagRequired("properties")

	collectionsCmd.AddCommand(collectionsDeleteCmd)
	collectionsDeleteCmd.Flags().StringVar(&collectionItemID, "item", "", "Item ID to delete")
	collectionsDeleteCmd.MarkFlagRequired("item")
}

// Output functions

func outputCollections(collections []models.Collection, format string) error {
	switch format {
	case "json":
		return outputJSON(collections)
	case "table":
		return outputCollectionsTable(collections)
	case "markdown":
		return outputCollectionsMarkdown(collections)
	default:
		return outputJSON(collections)
	}
}

func outputCollectionsTable(collections []models.Collection) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if !hasNoHeaders() {
		fmt.Fprintln(w, "ID\tNAME\tITEMS\tDOCUMENT")
		fmt.Fprintln(w, "---\t----\t-----\t--------")
	}

	for _, c := range collections {
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", c.ID, c.Name, c.ItemCount, c.DocumentID)
	}

	return w.Flush()
}

func outputCollectionsMarkdown(collections []models.Collection) error {
	fmt.Println("# Collections")
	for _, c := range collections {
		fmt.Printf("## %s\n", c.Name)
		fmt.Printf("- **ID**: %s\n", c.ID)
		fmt.Printf("- **Items**: %d\n", c.ItemCount)
		fmt.Printf("- **Document**: %s\n", c.DocumentID)
		fmt.Println()
	}
	return nil
}

func outputCollectionItems(items []models.CollectionItem, format string) error {
	switch format {
	case "json":
		return outputJSON(items)
	case "table":
		return outputCollectionItemsTable(items)
	default:
		return outputJSON(items)
	}
}

func outputCollectionItemsTable(items []models.CollectionItem) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if !hasNoHeaders() {
		fmt.Fprintln(w, "ID\tTITLE\tPROPERTIES")
		fmt.Fprintln(w, "---\t-----\t----------")
	}

	for _, item := range items {
		propsJSON, _ := json.Marshal(item.Properties)
		title := item.Title
		if len(title) > 40 {
			title = title[:37] + "..."
		}
		props := string(propsJSON)
		if len(props) > 60 {
			props = props[:57] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", item.ID, title, props)
	}

	return w.Flush()
}
```

**Step 2: Build and verify**

Run: `cd "/Users/nerveband/wavedepth Dropbox/Ashraf Ali/Mac (2)/Documents/GitHub/craft-cli" && go build ./...`
Expected: Compiles cleanly

**Step 3: Commit**

```bash
git add cmd/collections.go
git commit -m "feat: add collections command group (list, schema, items, add, update, delete)"
```

---

## Task 5: Add Connection Command

**Files:**
- Create: `cmd/connection.go`

**Step 1: Implement connection command**

Create `cmd/connection.go`:

```go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var connectionCmd = &cobra.Command{
	Use:   "connection",
	Short: "Show space connection info",
	Long: `Display space metadata including timezone, current time, and deep link URL templates.

Examples:
  craft connection
  craft connection --format json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		info, err := client.GetConnection()
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == "json" {
			return outputJSON(info)
		}

		fmt.Println("Space Connection")
		fmt.Println("================")
		fmt.Printf("Space ID:      %s\n", info.Space.ID)
		fmt.Printf("Timezone:      %s\n", info.Space.Timezone)
		fmt.Printf("Local Time:    %s\n", info.Space.Time)
		fmt.Printf("Date:          %s\n", info.Space.FriendlyDate)
		fmt.Printf("UTC Time:      %s\n", info.UTC.Time)
		if info.URLTemplates.App != "" {
			fmt.Printf("Deep Link:     %s\n", info.URLTemplates.App)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(connectionCmd)
}
```

**Step 2: Build and verify**

Run: `cd "/Users/nerveband/wavedepth Dropbox/Ashraf Ali/Mac (2)/Documents/GitHub/craft-cli" && go build ./...`
Expected: Compiles cleanly

**Step 3: Commit**

```bash
git add cmd/connection.go
git commit -m "feat: add connection command for space metadata"
```

---

## Task 6: Add Upload Command

**Files:**
- Create: `cmd/upload.go`

**Step 1: Implement upload command**

Create `cmd/upload.go`:

```go
package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload [file-path]",
	Short: "Upload a file to a document",
	Long: `Upload an image, video, or file to a Craft document or daily note.

Examples:
  craft upload photo.png --page DOC_ID
  craft upload report.pdf --page DOC_ID --position start
  craft upload image.jpg --date today
  craft upload diagram.png --sibling BLOCK_ID --position after
  cat file.bin | craft upload --stdin --page DOC_ID`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var fileData []byte
		var err error

		if uploadStdin {
			fileData, err = io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read stdin: %w", err)
			}
		} else if len(args) == 0 {
			return fmt.Errorf("file path is required (or use --stdin)")
		} else {
			fileData, err = os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", args[0], err)
			}
		}

		if len(fileData) == 0 {
			return fmt.Errorf("file is empty")
		}

		if uploadPageID == "" && uploadDate == "" && uploadSiblingID == "" {
			return fmt.Errorf("one of --page, --date, or --sibling is required")
		}

		if isDryRun() {
			size := len(fileData)
			name := "stdin"
			if len(args) > 0 {
				name = filepath.Base(args[0])
			}
			fmt.Printf("[dry-run] Would upload %s (%d bytes)\n", name, size)
			return nil
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		result, err := client.UploadFile(fileData, uploadPageID, uploadDate, uploadSiblingID, uploadPosition)
		if err != nil {
			return err
		}

		if isQuiet() {
			fmt.Println(result.BlockID)
			return nil
		}

		format := getOutputFormat()
		if format == "json" {
			return outputJSON(result)
		}

		fmt.Printf("Uploaded: block %s\n", result.BlockID)
		fmt.Printf("Asset URL: %s\n", result.AssetURL)
		return nil
	},
}

var (
	uploadPageID    string
	uploadDate      string
	uploadSiblingID string
	uploadPosition  string
	uploadStdin     bool
)

func init() {
	rootCmd.AddCommand(uploadCmd)

	uploadCmd.Flags().StringVar(&uploadPageID, "page", "", "Target document/page ID")
	uploadCmd.Flags().StringVar(&uploadDate, "date", "", "Target daily note date (today, tomorrow, YYYY-MM-DD)")
	uploadCmd.Flags().StringVar(&uploadSiblingID, "sibling", "", "Sibling block ID for relative positioning")
	uploadCmd.Flags().StringVarP(&uploadPosition, "position", "p", "end", "Position: start, end, before, after")
	uploadCmd.Flags().BoolVar(&uploadStdin, "stdin", false, "Read file data from stdin")
}
```

**Step 2: Build and verify**

Run: `cd "/Users/nerveband/wavedepth Dropbox/Ashraf Ali/Mac (2)/Documents/GitHub/craft-cli" && go build ./...`
Expected: Compiles cleanly

**Step 3: Commit**

```bash
git add cmd/upload.go
git commit -m "feat: add upload command for files and images"
```

---

## Task 7: Add Comments Command

**Files:**
- Create: `cmd/comments.go`

**Step 1: Implement comments command**

Create `cmd/comments.go`:

```go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var commentsCmd = &cobra.Command{
	Use:   "comments",
	Short: "Manage comments (experimental)",
	Long:  "Manage comments on blocks. Note: this is an experimental API endpoint.",
}

var commentsAddCmd = &cobra.Command{
	Use:   "add [block-id]",
	Short: "Add a comment to a block",
	Long: `Add a comment to a specific block.

Examples:
  craft comments add BLOCK_ID --content "Review this section"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if commentContent == "" {
			return fmt.Errorf("--content is required")
		}

		if isDryRun() {
			fmt.Printf("[dry-run] Would add comment to block %s\n", args[0])
			return nil
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		result, err := client.AddComment(args[0], commentContent)
		if err != nil {
			return err
		}

		if isQuiet() && len(result.Items) > 0 {
			fmt.Println(result.Items[0].CommentID)
			return nil
		}

		format := getOutputFormat()
		if format == "json" {
			return outputJSON(result)
		}

		if len(result.Items) > 0 {
			fmt.Printf("Comment added: %s\n", result.Items[0].CommentID)
		}
		return nil
	},
}

var commentContent string

func init() {
	rootCmd.AddCommand(commentsCmd)

	commentsCmd.AddCommand(commentsAddCmd)
	commentsAddCmd.Flags().StringVar(&commentContent, "content", "", "Comment text")
	commentsAddCmd.MarkFlagRequired("content")
}
```

**Step 2: Build and verify**

Run: `cd "/Users/nerveband/wavedepth Dropbox/Ashraf Ali/Mac (2)/Documents/GitHub/craft-cli" && go build ./...`
Expected: Compiles cleanly

**Step 3: Commit**

```bash
git add cmd/comments.go
git commit -m "feat: add comments command (experimental API)"
```

---

## Task 8: Enhance Search Command with Block Search and Advanced Filters

**Files:**
- Modify: `cmd/search.go`

**Step 1: Rewrite search command with full capabilities**

Replace the entire contents of `cmd/search.go`:

```go
package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ashrafali/craft-cli/internal/api"
	"github.com/ashrafali/craft-cli/internal/models"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for documents or blocks",
	Long: `Search across documents or within a specific document.

Document search (default):
  craft search "meeting notes"
  craft search "project" --location unsorted
  craft search --regex "API.*endpoint"
  craft search "report" --created-after 2026-01-01 --modified-before 2026-01-30

Block search (within a document):
  craft search "pattern" --document DOC_ID
  craft search --regex "TODO.*fix" --document DOC_ID --context 3
  craft search "keyword" --document DOC_ID --case-sensitive`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := ""
		if len(args) > 0 {
			query = args[0]
		}

		if query == "" && searchRegex == "" {
			return fmt.Errorf("either a query argument or --regex is required")
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		format := getOutputFormat()

		// Block-level search within a document
		if searchDocumentID != "" {
			pattern := query
			if searchRegex != "" {
				pattern = searchRegex
			}
			if pattern == "" {
				return fmt.Errorf("a search pattern is required for document search")
			}

			result, err := client.SearchBlocks(searchDocumentID, pattern, searchCaseSensitive, searchContext, searchContext)
			if err != nil {
				return err
			}

			return outputBlockSearchResults(result.Items, format)
		}

		// Document-level search with advanced filters
		opts := api.SearchOptions{
			Regexps:             searchRegex,
			Location:            searchLocation,
			FolderIDs:           searchFolderID,
			FetchMetadata:       searchMetadata,
			CreatedDateGte:      searchCreatedAfter,
			CreatedDateLte:      searchCreatedBefore,
			LastModifiedDateGte: searchModifiedAfter,
			LastModifiedDateLte: searchModifiedBefore,
		}

		result, err := client.SearchDocumentsAdvanced(query, opts)
		if err != nil {
			return err
		}

		return outputSearchResults(result.Items, format)
	},
}

var (
	searchDocumentID    string
	searchRegex         string
	searchCaseSensitive bool
	searchContext        int
	searchLocation      string
	searchFolderID      string
	searchMetadata      bool
	searchCreatedAfter  string
	searchCreatedBefore string
	searchModifiedAfter string
	searchModifiedBefore string
)

func init() {
	rootCmd.AddCommand(searchCmd)

	// Block search flags
	searchCmd.Flags().StringVar(&searchDocumentID, "document", "", "Search within a specific document (block-level search)")
	searchCmd.Flags().StringVar(&searchRegex, "regex", "", "RE2-compatible regex pattern")
	searchCmd.Flags().BoolVar(&searchCaseSensitive, "case-sensitive", false, "Case-sensitive search (block search only)")
	searchCmd.Flags().IntVar(&searchContext, "context", 5, "Number of surrounding blocks to include (block search only)")

	// Document search filters
	searchCmd.Flags().StringVar(&searchLocation, "location", "", "Filter by location: unsorted, trash, templates, daily_notes")
	searchCmd.Flags().StringVar(&searchFolderID, "folder", "", "Filter by folder ID")
	searchCmd.Flags().BoolVar(&searchMetadata, "metadata", false, "Include metadata in results")
	searchCmd.Flags().StringVar(&searchCreatedAfter, "created-after", "", "Filter: created on or after (YYYY-MM-DD or today/tomorrow/yesterday)")
	searchCmd.Flags().StringVar(&searchCreatedBefore, "created-before", "", "Filter: created on or before")
	searchCmd.Flags().StringVar(&searchModifiedAfter, "modified-after", "", "Filter: modified on or after")
	searchCmd.Flags().StringVar(&searchModifiedBefore, "modified-before", "", "Filter: modified on or before")
}

// Block search output

func outputBlockSearchResults(items []models.BlockSearchResult, format string) error {
	switch format {
	case "json":
		return outputJSON(items)
	case "table":
		return outputBlockSearchTable(items)
	default:
		return outputBlockSearchMarkdown(items)
	}
}

func outputBlockSearchTable(items []models.BlockSearchResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if !hasNoHeaders() {
		fmt.Fprintln(w, "BLOCK_ID\tMATCH\tPATH")
		fmt.Fprintln(w, "--------\t-----\t----")
	}

	for _, item := range items {
		match := item.Markdown
		if len(match) > 50 {
			match = match[:47] + "..."
		}
		path := ""
		for i, p := range item.PageBlockPath {
			if i > 0 {
				path += " > "
			}
			path += p.Content
		}
		if len(path) > 40 {
			path = path[:37] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", item.BlockID, match, path)
	}

	return w.Flush()
}

func outputBlockSearchMarkdown(items []models.BlockSearchResult) error {
	for i, item := range items {
		if i > 0 {
			fmt.Println("---")
		}

		// Show path
		if len(item.PageBlockPath) > 0 {
			path := ""
			for j, p := range item.PageBlockPath {
				if j > 0 {
					path += " > "
				}
				path += p.Content
			}
			fmt.Printf("**Path**: %s\n", path)
		}

		// Show before context
		for _, b := range item.BeforeBlocks {
			fmt.Printf("  %s\n", b.Markdown)
		}

		// Show match
		fmt.Printf("**>>> %s** (block: %s)\n", item.Markdown, item.BlockID)

		// Show after context
		for _, b := range item.AfterBlocks {
			fmt.Printf("  %s\n", b.Markdown)
		}
		fmt.Println()
	}
	return nil
}
```

**Step 2: Build and verify**

Run: `cd "/Users/nerveband/wavedepth Dropbox/Ashraf Ali/Mac (2)/Documents/GitHub/craft-cli" && go build ./...`
Expected: Compiles cleanly

**Step 3: Commit**

```bash
git add cmd/search.go
git commit -m "feat: enhance search with block search, regex, date filters, and location filters"
```

---

## Task 9: Enhance Blocks Command with Daily Note Access and Depth/Metadata

**Files:**
- Modify: `cmd/blocks.go`

**Step 1: Add --date, --depth, and --metadata flags to blocks get and blocks add**

In `cmd/blocks.go`, add new flag variables alongside the existing ones:

```go
var (
	blockMarkdown   string
	blockPosition   string
	blockSiblingID  string
	blockTargetPage string
	blockDate       string
	blockDepth      int
	blockMetadata   bool
)
```

Modify `blocksGetCmd` RunE to support `--date`, `--depth`, and `--metadata`:

Replace the `blocksGetCmd` definition with:

```go
var blocksGetCmd = &cobra.Command{
	Use:   "get [block-id]",
	Short: "Get a specific block or daily note",
	Long: `Retrieve a specific block by ID or a daily note by date.

Examples:
  craft blocks get BLOCK_ID
  craft blocks get BLOCK_ID --depth 0
  craft blocks get --date today
  craft blocks get --date 2026-01-30 --metadata`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		var block *models.Block

		if blockDate != "" {
			block, err = client.GetBlockByDate(blockDate, blockDepth, blockMetadata)
		} else {
			if len(args) == 0 {
				return fmt.Errorf("block-id is required (or use --date)")
			}
			block, err = client.GetBlockWithOptions(args[0], blockDepth, blockMetadata)
		}

		if err != nil {
			return err
		}

		format := getOutputFormat()
		switch format {
		case FormatStructured, "json":
			return outputJSON(block)
		case FormatCraft:
			var sb strings.Builder
			renderBlockCraft(&sb, block, 0)
			fmt.Print(sb.String())
			return nil
		case FormatRich:
			var sb strings.Builder
			renderBlockRich(&sb, block, 0)
			fmt.Print(sb.String())
			return nil
		default:
			fmt.Println(block.Markdown)
			return nil
		}
	},
}
```

Modify `blocksAddCmd` to support `--date`:

Replace the `blocksAddCmd` definition with:

```go
var blocksAddCmd = &cobra.Command{
	Use:   "add [page-id]",
	Short: "Add a block to a document or daily note",
	Long: `Add a new block to a document at a specified position.

Positions:
  start  - Add at the beginning of the page
  end    - Add at the end of the page (default)
  before - Add before a sibling block (requires --sibling)
  after  - Add after a sibling block (requires --sibling)

Examples:
  craft blocks add PAGE_ID --markdown "Hello"
  craft blocks add --date today --markdown "Daily note entry"
  craft blocks add --sibling BLOCK_ID --position after --markdown "After this"`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if blockMarkdown == "" {
			return fmt.Errorf("--markdown is required")
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		var block *models.Block

		if blockSiblingID != "" {
			if blockPosition != "before" && blockPosition != "after" {
				return fmt.Errorf("--position must be 'before' or 'after' when using --sibling")
			}
			block, err = client.AddBlockRelative(blockSiblingID, blockMarkdown, blockPosition)
		} else if blockDate != "" {
			if blockPosition == "" {
				blockPosition = "end"
			}
			block, err = client.AddBlockToDate(blockDate, blockMarkdown, blockPosition)
		} else {
			if len(args) == 0 {
				return fmt.Errorf("page-id is required when not using --sibling or --date")
			}
			pageID := args[0]
			if blockPosition == "" {
				blockPosition = "end"
			}
			block, err = client.AddBlock(pageID, blockMarkdown, blockPosition)
		}

		if err != nil {
			return err
		}

		if isQuiet() {
			fmt.Println(block.ID)
			return nil
		}

		format := getOutputFormat()
		if format == "json" {
			return outputJSON(block)
		}
		fmt.Printf("Block created: %s\n", block.ID)
		return nil
	},
}
```

Update the `init()` function to register the new flags:

```go
func init() {
	rootCmd.AddCommand(blocksCmd)

	blocksCmd.AddCommand(blocksGetCmd)
	blocksGetCmd.Flags().StringVar(&blockDate, "date", "", "Daily note date (today, tomorrow, yesterday, YYYY-MM-DD)")
	blocksGetCmd.Flags().IntVar(&blockDepth, "depth", -1, "Max depth (-1 for all, 0 for block only)")
	blocksGetCmd.Flags().BoolVar(&blockMetadata, "metadata", false, "Include metadata (created/modified info)")

	blocksCmd.AddCommand(blocksAddCmd)
	blocksAddCmd.Flags().StringVarP(&blockMarkdown, "markdown", "m", "", "Markdown content for the block")
	blocksAddCmd.Flags().StringVarP(&blockPosition, "position", "p", "end", "Position: start, end, before, after")
	blocksAddCmd.Flags().StringVar(&blockSiblingID, "sibling", "", "Sibling block ID for relative positioning")
	blocksAddCmd.Flags().StringVar(&blockDate, "date", "", "Daily note date (today, tomorrow, yesterday, YYYY-MM-DD)")
	blocksAddCmd.MarkFlagRequired("markdown")

	blocksCmd.AddCommand(blocksUpdateCmd)
	blocksUpdateCmd.Flags().StringVarP(&blockMarkdown, "markdown", "m", "", "New markdown content")
	blocksUpdateCmd.MarkFlagRequired("markdown")

	blocksCmd.AddCommand(blocksDeleteCmd)

	blocksCmd.AddCommand(blocksMoveCmd)
	blocksMoveCmd.Flags().StringVar(&blockTargetPage, "to", "", "Target page ID")
	blocksMoveCmd.Flags().StringVarP(&blockPosition, "position", "p", "end", "Position: start, end")
	blocksMoveCmd.MarkFlagRequired("to")
}
```

**Step 2: Build and verify**

Run: `cd "/Users/nerveband/wavedepth Dropbox/Ashraf Ali/Mac (2)/Documents/GitHub/craft-cli" && go build ./...`
Expected: Compiles cleanly

**Step 3: Commit**

```bash
git add cmd/blocks.go
git commit -m "feat: enhance blocks with daily note access, depth control, and metadata"
```

---

## Task 10: Enhance Document List with Date Filters and Metadata

**Files:**
- Modify: `cmd/list.go` (or wherever `list` command is defined)

**Step 1: Find and read the list command**

The list command is likely in `cmd/list.go` or `cmd/root.go`. Read it first.

**Step 2: Add date filter and metadata flags**

Add these flag variables and registration to the list command:

```go
var (
	listCreatedAfter  string
	listCreatedBefore string
	listModifiedAfter string
	listModifiedBefore string
	listMetadata      bool
)
```

Modify the RunE to use `GetDocumentsAdvanced`:

```go
RunE: func(cmd *cobra.Command, args []string) error {
	client, err := getAPIClient()
	if err != nil {
		return err
	}

	// Use advanced listing if any filter is set
	hasAdvanced := listCreatedAfter != "" || listCreatedBefore != "" ||
		listModifiedAfter != "" || listModifiedBefore != "" || listMetadata

	var result *models.DocumentList

	if hasAdvanced {
		opts := api.ListDocumentsOptions{
			FolderID:            listFolderID,
			Location:            listLocation,
			FetchMetadata:       listMetadata,
			CreatedDateGte:      listCreatedAfter,
			CreatedDateLte:      listCreatedBefore,
			LastModifiedDateGte: listModifiedAfter,
			LastModifiedDateLte: listModifiedBefore,
		}
		result, err = client.GetDocumentsAdvanced(opts)
	} else {
		result, err = client.GetDocumentsFiltered(listFolderID, listLocation)
	}

	if err != nil {
		return err
	}

	format := getOutputFormat()
	return outputDocuments(result.Items, format)
},
```

Register flags in init:

```go
listCmd.Flags().StringVar(&listCreatedAfter, "created-after", "", "Filter: created on or after (YYYY-MM-DD)")
listCmd.Flags().StringVar(&listCreatedBefore, "created-before", "", "Filter: created on or before")
listCmd.Flags().StringVar(&listModifiedAfter, "modified-after", "", "Filter: modified on or after")
listCmd.Flags().StringVar(&listModifiedBefore, "modified-before", "", "Filter: modified on or before")
listCmd.Flags().BoolVar(&listMetadata, "metadata", false, "Include metadata in results")
```

**Step 3: Build and verify**

Run: `cd "/Users/nerveband/wavedepth Dropbox/Ashraf Ali/Mac (2)/Documents/GitHub/craft-cli" && go build ./...`
Expected: Compiles cleanly

**Step 4: Commit**

```bash
git add cmd/list.go
git commit -m "feat: enhance list with date filters and metadata support"
```

---

## Task 11: Run Full Test Suite and Final Build

**Step 1: Run all tests**

Run: `cd "/Users/nerveband/wavedepth Dropbox/Ashraf Ali/Mac (2)/Documents/GitHub/craft-cli" && go test ./... -v`
Expected: All PASS

**Step 2: Build binary**

Run: `cd "/Users/nerveband/wavedepth Dropbox/Ashraf Ali/Mac (2)/Documents/GitHub/craft-cli" && go build -o craft .`
Expected: Binary builds cleanly

**Step 3: Smoke test new commands**

Run these to verify help text renders:

```bash
./craft collections --help
./craft collections list --help
./craft connection --help
./craft upload --help
./craft comments --help
./craft search --help
./craft blocks get --help
./craft blocks add --help
./craft list --help
```

Expected: Each shows updated help with new flags/subcommands

**Step 4: Commit and tag**

```bash
git add -A
git commit -m "feat: full API parity - collections, connection, upload, comments, block search, advanced filters

Adds:
- collections command group (list, schema, items, add, update, delete)
- connection command for space metadata
- upload command for files/images
- comments add command (experimental)
- Block-level search with regex and context
- Advanced search filters (date ranges, location, folder, regex)
- Daily note access via --date on blocks get/add
- Depth and metadata flags on blocks get
- Date filters and metadata on document list"
```

---

## Summary

| Task | What | New Files | Modified Files |
|------|------|-----------|----------------|
| 1 | Collection + new models | - | `internal/models/document.go` |
| 2 | Collection API client | `internal/api/collections_test.go` | `internal/api/client.go` |
| 3 | Connection/upload/comments/search API | `internal/api/endpoints_test.go` | `internal/api/client.go` |
| 4 | Collections commands | `cmd/collections.go` | - |
| 5 | Connection command | `cmd/connection.go` | - |
| 6 | Upload command | `cmd/upload.go` | - |
| 7 | Comments command | `cmd/comments.go` | - |
| 8 | Enhanced search | - | `cmd/search.go` |
| 9 | Enhanced blocks | - | `cmd/blocks.go` |
| 10 | Enhanced list | - | `cmd/list.go` |
| 11 | Full test + build | - | - |

**Total: 9 new API methods, 4 new command files, 3 enhanced commands, ~1200 lines of new code.**
