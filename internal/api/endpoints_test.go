package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ashrafali/craft-cli/internal/models"
)

func TestClient_GetConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/connection" {
			t.Errorf("Expected path /connection, got %s", r.URL.Path)
		}

		response := models.ConnectionInfo{}
		response.Space.ID = "space-123"
		response.Space.Timezone = "America/New_York"
		response.Space.Time = "2025-01-15T10:30:00"
		response.Space.FriendlyDate = "January 15, 2025"
		response.UTC.Time = "2025-01-15T15:30:00Z"
		response.URLTemplates.App = "craftdocs://open?spaceId={spaceId}&blockId={blockId}"

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.GetConnection()

	if err != nil {
		t.Fatalf("GetConnection() error = %v", err)
	}

	if result.Space.ID != "space-123" {
		t.Errorf("Expected space ID 'space-123', got %s", result.Space.ID)
	}
	if result.Space.Timezone != "America/New_York" {
		t.Errorf("Expected timezone 'America/New_York', got %s", result.Space.Timezone)
	}
	if result.UTC.Time != "2025-01-15T15:30:00Z" {
		t.Errorf("Expected UTC time '2025-01-15T15:30:00Z', got %s", result.UTC.Time)
	}
	if result.URLTemplates.App != "craftdocs://open?spaceId={spaceId}&blockId={blockId}" {
		t.Errorf("Expected URL template, got %s", result.URLTemplates.App)
	}
}

func TestClient_AddComment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/comments" {
			t.Errorf("Expected path /comments, got %s", r.URL.Path)
		}

		var body struct {
			Comments []struct {
				BlockID string `json:"blockId"`
				Content string `json:"content"`
			} `json:"comments"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if len(body.Comments) != 1 {
			t.Errorf("Expected 1 comment, got %d", len(body.Comments))
		}
		if body.Comments[0].BlockID != "block-abc" {
			t.Errorf("Expected blockId 'block-abc', got %s", body.Comments[0].BlockID)
		}
		if body.Comments[0].Content != "Nice work!" {
			t.Errorf("Expected content 'Nice work!', got %s", body.Comments[0].Content)
		}

		response := models.CommentResponse{
			Items: []struct {
				CommentID string `json:"commentId"`
			}{{CommentID: "comment-1"}},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.AddComment("block-abc", "Nice work!")

	if err != nil {
		t.Fatalf("AddComment() error = %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(result.Items))
	}
	if result.Items[0].CommentID != "comment-1" {
		t.Errorf("Expected commentId 'comment-1', got %s", result.Items[0].CommentID)
	}
}

func TestClient_SearchBlocks(t *testing.T) {
	t.Run("case insensitive search", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET method, got %s", r.Method)
			}
			if r.URL.Path != "/blocks/search" {
				t.Errorf("Expected path /blocks/search, got %s", r.URL.Path)
			}
			if r.URL.Query().Get("blockId") != "doc-1" {
				t.Errorf("Expected blockId 'doc-1', got %s", r.URL.Query().Get("blockId"))
			}
			if r.URL.Query().Get("pattern") != "hello" {
				t.Errorf("Expected pattern 'hello', got %s", r.URL.Query().Get("pattern"))
			}
			// caseSensitive should NOT be present when false
			if r.URL.Query().Get("caseSensitive") != "" {
				t.Errorf("Expected no caseSensitive param, got %s", r.URL.Query().Get("caseSensitive"))
			}
			if r.URL.Query().Get("beforeBlockCount") != "5" {
				t.Errorf("Expected beforeBlockCount '5', got %s", r.URL.Query().Get("beforeBlockCount"))
			}
			if r.URL.Query().Get("afterBlockCount") != "5" {
				t.Errorf("Expected afterBlockCount '5', got %s", r.URL.Query().Get("afterBlockCount"))
			}

			response := models.BlockSearchResultList{
				Items: []models.BlockSearchResult{
					{
						BlockID:  "block-match",
						Markdown: "Hello world",
						BeforeBlocks: []models.BlockContext{
							{BlockID: "before-1", Markdown: "before text"},
						},
						AfterBlocks: []models.BlockContext{
							{BlockID: "after-1", Markdown: "after text"},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		result, err := client.SearchBlocks("doc-1", "hello", false, 5, 5)

		if err != nil {
			t.Fatalf("SearchBlocks() error = %v", err)
		}

		if len(result.Items) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(result.Items))
		}
		if result.Items[0].BlockID != "block-match" {
			t.Errorf("Expected blockId 'block-match', got %s", result.Items[0].BlockID)
		}
		if result.Items[0].Markdown != "Hello world" {
			t.Errorf("Expected markdown 'Hello world', got %s", result.Items[0].Markdown)
		}
		if len(result.Items[0].BeforeBlocks) != 1 {
			t.Errorf("Expected 1 before block, got %d", len(result.Items[0].BeforeBlocks))
		}
		if len(result.Items[0].AfterBlocks) != 1 {
			t.Errorf("Expected 1 after block, got %d", len(result.Items[0].AfterBlocks))
		}
	})

	t.Run("case sensitive search", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("caseSensitive") != "true" {
				t.Errorf("Expected caseSensitive 'true', got %s", r.URL.Query().Get("caseSensitive"))
			}
			if r.URL.Query().Get("beforeBlockCount") != "3" {
				t.Errorf("Expected beforeBlockCount '3', got %s", r.URL.Query().Get("beforeBlockCount"))
			}
			if r.URL.Query().Get("afterBlockCount") != "10" {
				t.Errorf("Expected afterBlockCount '10', got %s", r.URL.Query().Get("afterBlockCount"))
			}

			response := models.BlockSearchResultList{Items: []models.BlockSearchResult{}}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		_, err := client.SearchBlocks("doc-1", "Hello", true, 3, 10)
		if err != nil {
			t.Fatalf("SearchBlocks() error = %v", err)
		}
	})
}

func TestClient_SearchDocumentsAdvanced(t *testing.T) {
	t.Run("with all options", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET method, got %s", r.Method)
			}
			if r.URL.Path != "/documents/search" {
				t.Errorf("Expected path /documents/search, got %s", r.URL.Path)
			}

			q := r.URL.Query()
			if q.Get("include") != "meeting notes" {
				t.Errorf("Expected include 'meeting notes', got %s", q.Get("include"))
			}
			if q.Get("location") != "personal" {
				t.Errorf("Expected location 'personal', got %s", q.Get("location"))
			}
			if q.Get("folderIDs") != "folder-1" {
				t.Errorf("Expected folderIDs 'folder-1', got %s", q.Get("folderIDs"))
			}
			if q.Get("fetchMetadata") != "true" {
				t.Errorf("Expected fetchMetadata 'true', got %s", q.Get("fetchMetadata"))
			}
			if q.Get("createdDateGte") != "2025-01-01" {
				t.Errorf("Expected createdDateGte '2025-01-01', got %s", q.Get("createdDateGte"))
			}
			if q.Get("lastModifiedDateLte") != "2025-12-31" {
				t.Errorf("Expected lastModifiedDateLte '2025-12-31', got %s", q.Get("lastModifiedDateLte"))
			}

			response := models.SearchResult{
				Items: []models.SearchItem{
					{DocumentID: "doc-found", Markdown: "Meeting Notes - Q1"},
				},
				Total: 1,
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		opts := SearchOptions{
			Location:            "personal",
			FolderIDs:           "folder-1",
			FetchMetadata:       true,
			CreatedDateGte:      "2025-01-01",
			LastModifiedDateLte: "2025-12-31",
		}
		result, err := client.SearchDocumentsAdvanced("meeting notes", opts)

		if err != nil {
			t.Fatalf("SearchDocumentsAdvanced() error = %v", err)
		}

		if len(result.Items) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(result.Items))
		}
		if result.Items[0].DocumentID != "doc-found" {
			t.Errorf("Expected documentId 'doc-found', got %s", result.Items[0].DocumentID)
		}
	})

	t.Run("minimal options", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			if q.Get("include") != "test" {
				t.Errorf("Expected include 'test', got %s", q.Get("include"))
			}
			// Verify optional params are absent
			if q.Get("location") != "" {
				t.Errorf("Expected no location param, got %s", q.Get("location"))
			}
			if q.Get("fetchMetadata") != "" {
				t.Errorf("Expected no fetchMetadata param, got %s", q.Get("fetchMetadata"))
			}

			response := models.SearchResult{Items: []models.SearchItem{}, Total: 0}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		result, err := client.SearchDocumentsAdvanced("test", SearchOptions{})
		if err != nil {
			t.Fatalf("SearchDocumentsAdvanced() error = %v", err)
		}
		if len(result.Items) != 0 {
			t.Errorf("Expected 0 results, got %d", len(result.Items))
		}
	})
}

func TestClient_GetDocumentsAdvanced(t *testing.T) {
	t.Run("with options", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET method, got %s", r.Method)
			}
			if r.URL.Path != "/documents" {
				t.Errorf("Expected path /documents, got %s", r.URL.Path)
			}

			q := r.URL.Query()
			if q.Get("folderId") != "folder-abc" {
				t.Errorf("Expected folderId 'folder-abc', got %s", q.Get("folderId"))
			}
			if q.Get("location") != "personal" {
				t.Errorf("Expected location 'personal', got %s", q.Get("location"))
			}
			if q.Get("fetchMetadata") != "true" {
				t.Errorf("Expected fetchMetadata 'true', got %s", q.Get("fetchMetadata"))
			}
			if q.Get("dailyNoteDateGte") != "2025-01-01" {
				t.Errorf("Expected dailyNoteDateGte '2025-01-01', got %s", q.Get("dailyNoteDateGte"))
			}
			if q.Get("dailyNoteDateLte") != "2025-01-31" {
				t.Errorf("Expected dailyNoteDateLte '2025-01-31', got %s", q.Get("dailyNoteDateLte"))
			}

			response := models.DocumentList{
				Items: []models.Document{
					{ID: "doc-1", Title: "Daily Note Jan 15"},
				},
				Total: 1,
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		opts := ListDocumentsOptions{
			FolderID:         "folder-abc",
			Location:         "personal",
			FetchMetadata:    true,
			DailyNoteDateGte: "2025-01-01",
			DailyNoteDateLte: "2025-01-31",
		}
		result, err := client.GetDocumentsAdvanced(opts)

		if err != nil {
			t.Fatalf("GetDocumentsAdvanced() error = %v", err)
		}

		if len(result.Items) != 1 {
			t.Fatalf("Expected 1 document, got %d", len(result.Items))
		}
		if result.Items[0].ID != "doc-1" {
			t.Errorf("Expected document ID 'doc-1', got %s", result.Items[0].ID)
		}
	})

	t.Run("empty options", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/documents" {
				t.Errorf("Expected path /documents, got %s", r.URL.Path)
			}
			if r.URL.RawQuery != "" {
				t.Errorf("Expected no query params, got %s", r.URL.RawQuery)
			}

			response := models.DocumentList{Items: []models.Document{}, Total: 0}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		result, err := client.GetDocumentsAdvanced(ListDocumentsOptions{})
		if err != nil {
			t.Fatalf("GetDocumentsAdvanced() error = %v", err)
		}
		if len(result.Items) != 0 {
			t.Errorf("Expected 0 documents, got %d", len(result.Items))
		}
	})
}

func TestClient_GetBlockByDate(t *testing.T) {
	t.Run("with all params", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET method, got %s", r.Method)
			}
			if r.URL.Path != "/blocks" {
				t.Errorf("Expected path /blocks, got %s", r.URL.Path)
			}

			q := r.URL.Query()
			if q.Get("date") != "2025-01-15" {
				t.Errorf("Expected date '2025-01-15', got %s", q.Get("date"))
			}
			if q.Get("maxDepth") != "3" {
				t.Errorf("Expected maxDepth '3', got %s", q.Get("maxDepth"))
			}
			if q.Get("fetchMetadata") != "true" {
				t.Errorf("Expected fetchMetadata 'true', got %s", q.Get("fetchMetadata"))
			}

			response := models.Block{
				ID:       "daily-block",
				Type:     "page",
				Markdown: "January 15, 2025",
				Content: []models.Block{
					{ID: "child-1", Type: "text", Markdown: "Meeting at 10am"},
				},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		result, err := client.GetBlockByDate("2025-01-15", 3, true)

		if err != nil {
			t.Fatalf("GetBlockByDate() error = %v", err)
		}

		if result.ID != "daily-block" {
			t.Errorf("Expected block ID 'daily-block', got %s", result.ID)
		}
		if result.Markdown != "January 15, 2025" {
			t.Errorf("Expected markdown 'January 15, 2025', got %s", result.Markdown)
		}
		if len(result.Content) != 1 {
			t.Fatalf("Expected 1 child block, got %d", len(result.Content))
		}
		if result.Content[0].Markdown != "Meeting at 10am" {
			t.Errorf("Expected child markdown 'Meeting at 10am', got %s", result.Content[0].Markdown)
		}
	})

	t.Run("without maxDepth and fetchMetadata", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			if q.Get("date") != "2025-06-01" {
				t.Errorf("Expected date '2025-06-01', got %s", q.Get("date"))
			}
			if q.Get("maxDepth") != "" {
				t.Errorf("Expected no maxDepth param, got %s", q.Get("maxDepth"))
			}
			if q.Get("fetchMetadata") != "" {
				t.Errorf("Expected no fetchMetadata param, got %s", q.Get("fetchMetadata"))
			}

			response := models.Block{ID: "daily-2", Type: "page", Markdown: "June 1"}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		result, err := client.GetBlockByDate("2025-06-01", -1, false)
		if err != nil {
			t.Fatalf("GetBlockByDate() error = %v", err)
		}
		if result.ID != "daily-2" {
			t.Errorf("Expected block ID 'daily-2', got %s", result.ID)
		}
	})
}

func TestClient_AddBlockToDate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/blocks" {
			t.Errorf("Expected path /blocks, got %s", r.URL.Path)
		}

		var body struct {
			Markdown string `json:"markdown"`
			Position struct {
				Date     string `json:"date"`
				Position string `json:"position"`
			} `json:"position"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if body.Markdown != "New daily note entry" {
			t.Errorf("Expected markdown 'New daily note entry', got %s", body.Markdown)
		}
		if body.Position.Date != "2025-01-15" {
			t.Errorf("Expected date '2025-01-15', got %s", body.Position.Date)
		}
		if body.Position.Position != "end" {
			t.Errorf("Expected position 'end', got %s", body.Position.Position)
		}

		response := struct {
			Items []struct {
				ID       string `json:"id"`
				Type     string `json:"type"`
				Markdown string `json:"markdown"`
			} `json:"items"`
		}{
			Items: []struct {
				ID       string `json:"id"`
				Type     string `json:"type"`
				Markdown string `json:"markdown"`
			}{{ID: "new-block-1", Type: "text", Markdown: "New daily note entry"}},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.AddBlockToDate("2025-01-15", "New daily note entry", "end")

	if err != nil {
		t.Fatalf("AddBlockToDate() error = %v", err)
	}

	if result.ID != "new-block-1" {
		t.Errorf("Expected block ID 'new-block-1', got %s", result.ID)
	}
	if result.Type != "text" {
		t.Errorf("Expected block type 'text', got %s", result.Type)
	}
	if result.Markdown != "New daily note entry" {
		t.Errorf("Expected markdown 'New daily note entry', got %s", result.Markdown)
	}
}

func TestClient_UploadFile(t *testing.T) {
	t.Run("upload to page", func(t *testing.T) {
		fileContent := []byte("fake-image-binary-data")

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}
			if r.URL.Path != "/upload" {
				t.Errorf("Expected path /upload, got %s", r.URL.Path)
			}
			if r.Header.Get("Content-Type") != "application/octet-stream" {
				t.Errorf("Expected Content-Type 'application/octet-stream', got %s", r.Header.Get("Content-Type"))
			}

			q := r.URL.Query()
			if q.Get("pageId") != "page-123" {
				t.Errorf("Expected pageId 'page-123', got %s", q.Get("pageId"))
			}
			if q.Get("position") != "end" {
				t.Errorf("Expected position 'end', got %s", q.Get("position"))
			}

			// Verify the body is the raw file data
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("Failed to read request body: %v", err)
			}
			if string(body) != "fake-image-binary-data" {
				t.Errorf("Expected body 'fake-image-binary-data', got %s", string(body))
			}

			response := models.UploadResponse{
				BlockID:  "upload-block-1",
				AssetURL: "https://cdn.craft.do/assets/upload-block-1.png",
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		result, err := client.UploadFile(fileContent, "page-123", "", "", "end")

		if err != nil {
			t.Fatalf("UploadFile() error = %v", err)
		}

		if result.BlockID != "upload-block-1" {
			t.Errorf("Expected blockId 'upload-block-1', got %s", result.BlockID)
		}
		if result.AssetURL != "https://cdn.craft.do/assets/upload-block-1.png" {
			t.Errorf("Expected assetUrl, got %s", result.AssetURL)
		}
	})

	t.Run("upload to date", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			if q.Get("date") != "2025-01-15" {
				t.Errorf("Expected date '2025-01-15', got %s", q.Get("date"))
			}
			if q.Get("pageId") != "" {
				t.Errorf("Expected no pageId param, got %s", q.Get("pageId"))
			}

			response := models.UploadResponse{
				BlockID:  "upload-block-2",
				AssetURL: "https://cdn.craft.do/assets/upload-block-2.pdf",
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		result, err := client.UploadFile([]byte("pdf-data"), "", "2025-01-15", "", "start")
		if err != nil {
			t.Fatalf("UploadFile() error = %v", err)
		}
		if result.BlockID != "upload-block-2" {
			t.Errorf("Expected blockId 'upload-block-2', got %s", result.BlockID)
		}
	})

	t.Run("upload next to sibling", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			if q.Get("siblingId") != "sibling-xyz" {
				t.Errorf("Expected siblingId 'sibling-xyz', got %s", q.Get("siblingId"))
			}
			if q.Get("position") != "after" {
				t.Errorf("Expected position 'after', got %s", q.Get("position"))
			}

			response := models.UploadResponse{
				BlockID:  "upload-block-3",
				AssetURL: "https://cdn.craft.do/assets/upload-block-3.jpg",
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		result, err := client.UploadFile([]byte("jpg-data"), "", "", "sibling-xyz", "after")
		if err != nil {
			t.Fatalf("UploadFile() error = %v", err)
		}
		if result.BlockID != "upload-block-3" {
			t.Errorf("Expected blockId 'upload-block-3', got %s", result.BlockID)
		}
	})
}

func TestClient_GetBlockWithOptions(t *testing.T) {
	t.Run("with maxDepth and metadata", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET method, got %s", r.Method)
			}
			if r.URL.Path != "/blocks" {
				t.Errorf("Expected path /blocks, got %s", r.URL.Path)
			}

			q := r.URL.Query()
			if q.Get("id") != "block-456" {
				t.Errorf("Expected id 'block-456', got %s", q.Get("id"))
			}
			if q.Get("maxDepth") != "2" {
				t.Errorf("Expected maxDepth '2', got %s", q.Get("maxDepth"))
			}
			if q.Get("fetchMetadata") != "true" {
				t.Errorf("Expected fetchMetadata 'true', got %s", q.Get("fetchMetadata"))
			}

			response := models.Block{
				ID:       "block-456",
				Type:     "text",
				Markdown: "Block with metadata",
				Metadata: &models.BlockMetadata{
					CreatedAt:      "2025-01-15T10:00:00Z",
					LastModifiedAt: "2025-01-15T11:30:00Z",
				},
				Content: []models.Block{
					{ID: "child-1", Type: "text", Markdown: "Child block"},
				},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		result, err := client.GetBlockWithOptions("block-456", 2, true)

		if err != nil {
			t.Fatalf("GetBlockWithOptions() error = %v", err)
		}

		if result.ID != "block-456" {
			t.Errorf("Expected block ID 'block-456', got %s", result.ID)
		}
		if result.Markdown != "Block with metadata" {
			t.Errorf("Expected markdown 'Block with metadata', got %s", result.Markdown)
		}
		if result.Metadata == nil {
			t.Fatal("Expected metadata, got nil")
		}
		if result.Metadata.CreatedAt != "2025-01-15T10:00:00Z" {
			t.Errorf("Expected createdAt '2025-01-15T10:00:00Z', got %s", result.Metadata.CreatedAt)
		}
		if len(result.Content) != 1 {
			t.Fatalf("Expected 1 child, got %d", len(result.Content))
		}
	})

	t.Run("without optional params", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			if q.Get("id") != "block-789" {
				t.Errorf("Expected id 'block-789', got %s", q.Get("id"))
			}
			if q.Get("maxDepth") != "" {
				t.Errorf("Expected no maxDepth param, got %s", q.Get("maxDepth"))
			}
			if q.Get("fetchMetadata") != "" {
				t.Errorf("Expected no fetchMetadata param, got %s", q.Get("fetchMetadata"))
			}

			response := models.Block{
				ID:       "block-789",
				Type:     "text",
				Markdown: "Simple block",
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		result, err := client.GetBlockWithOptions("block-789", -1, false)
		if err != nil {
			t.Fatalf("GetBlockWithOptions() error = %v", err)
		}
		if result.ID != "block-789" {
			t.Errorf("Expected block ID 'block-789', got %s", result.ID)
		}
		if result.Markdown != "Simple block" {
			t.Errorf("Expected markdown 'Simple block', got %s", result.Markdown)
		}
	})
}
