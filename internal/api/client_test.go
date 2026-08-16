package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ashrafali/craft-cli/internal/models"
)

func TestNewClient(t *testing.T) {
	client := NewClient("https://api.example.com")

	if client.baseURL != "https://api.example.com" {
		t.Errorf("baseURL = %v, want https://api.example.com", client.baseURL)
	}

	if client.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
}

func TestClient_GetDocuments(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/documents" {
			t.Errorf("Expected path /documents, got %s", r.URL.Path)
		}

		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		response := models.DocumentList{
			Items: []models.Document{
				{
					ID:             "doc1",
					Title:          "Test Document",
					CreatedAt:      time.Now(),
					LastModifiedAt: time.Now(),
				},
			},
			Total: 1,
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	docs, err := client.GetDocuments()

	if err != nil {
		t.Fatalf("GetDocuments() error = %v", err)
	}

	if len(docs.Items) != 1 {
		t.Errorf("Expected 1 document, got %d", len(docs.Items))
	}

	if docs.Items[0].Title != "Test Document" {
		t.Errorf("Document title = %v, want Test Document", docs.Items[0].Title)
	}
}

func TestClient_GetDocument(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/blocks" {
			t.Errorf("Expected path /blocks, got %s", r.URL.Path)
		}

		if r.URL.Query().Get("id") != "doc1" {
			t.Errorf("Expected query param id=doc1, got %s", r.URL.Query().Get("id"))
		}

		// Return BlocksResponse format (matching the real Craft API)
		response := models.BlocksResponse{
			ID:        "doc1",
			Type:      "page",
			TextStyle: "page",
			Markdown:  "Test Document",
			Content: []models.Block{
				{
					ID:       "block1",
					Type:     "text",
					Markdown: "Test content paragraph 1",
				},
				{
					ID:       "block2",
					Type:     "text",
					Markdown: "Test content paragraph 2",
				},
			},
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	doc, err := client.GetDocument("doc1")

	if err != nil {
		t.Fatalf("GetDocument() error = %v", err)
	}

	if doc.ID != "doc1" {
		t.Errorf("Document ID = %v, want doc1", doc.ID)
	}

	if doc.Title != "Test Document" {
		t.Errorf("Document title = %v, want Test Document", doc.Title)
	}

	// Verify markdown is combined from all blocks
	if doc.Markdown == "" {
		t.Error("Document markdown should not be empty")
	}
}

func TestClient_SearchDocuments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/documents/search" {
			t.Errorf("Expected path /documents/search, got %s", r.URL.Path)
		}

		// Craft API uses 'include' parameter
		include := r.URL.Query().Get("include")
		if include != "test query" {
			t.Errorf("Expected include 'test query', got %s", include)
		}

		response := models.SearchResult{
			Items: []models.SearchItem{
				{
					DocumentID: "doc1",
					Markdown:   "Test **query** result",
				},
			},
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	results, err := client.SearchDocuments("test query")

	if err != nil {
		t.Fatalf("SearchDocuments() error = %v", err)
	}

	if len(results.Items) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results.Items))
	}

	if results.Items[0].DocumentID != "doc1" {
		t.Errorf("Expected DocumentID 'doc1', got %s", results.Items[0].DocumentID)
	}
}

func TestClient_CreateDocument(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// First call: POST /documents to create the document
		if r.URL.Path == "/documents" {
			// Craft API expects {"documents": [...]} wrapper
			var wrapper struct {
				Documents []models.CreateDocumentRequest `json:"documents"`
			}
			if err := json.NewDecoder(r.Body).Decode(&wrapper); err != nil {
				t.Fatalf("Failed to decode request body: %v", err)
			}

			if len(wrapper.Documents) != 1 {
				t.Errorf("Expected 1 document, got %d", len(wrapper.Documents))
			}

			if wrapper.Documents[0].Title != "New Document" {
				t.Errorf("Expected title 'New Document', got %s", wrapper.Documents[0].Title)
			}

			// Return the Craft API response format
			response := struct {
				Items []struct {
					ID    string `json:"id"`
					Title string `json:"title"`
				} `json:"items"`
			}{
				Items: []struct {
					ID    string `json:"id"`
					Title string `json:"title"`
				}{
					{ID: "doc1", Title: wrapper.Documents[0].Title},
				},
			}

			json.NewEncoder(w).Encode(response)
			return
		}

		// Second call: POST /blocks to insert markdown content
		if r.URL.Path == "/blocks" {
			var req struct {
				Markdown string `json:"markdown"`
				Position struct {
					PageID   string `json:"pageId"`
					Position string `json:"position"`
				} `json:"position"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("Failed to decode request body: %v", err)
			}

			if req.Position.PageID != "doc1" {
				t.Errorf("Expected pageId 'doc1', got %s", req.Position.PageID)
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
				}{
					{ID: "block1", Type: "text", Markdown: req.Markdown},
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		t.Errorf("Unexpected path: %s", r.URL.Path)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	req := &models.CreateDocumentRequest{
		Title:    "New Document",
		Markdown: "Test content",
	}

	doc, err := client.CreateDocument(req)

	if err != nil {
		t.Fatalf("CreateDocument() error = %v", err)
	}

	if doc.Title != "New Document" {
		t.Errorf("Document title = %v, want New Document", doc.Title)
	}
}

func TestClient_CreateDocumentsRaw(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/documents" {
			t.Errorf("Expected path /documents, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		docs, ok := body["documents"].([]interface{})
		if !ok || len(docs) != 1 {
			t.Fatalf("documents payload = %#v", body["documents"])
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]string{{"id": "doc1", "title": "Raw Doc"}},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	docs, err := client.CreateDocumentsRaw(map[string]interface{}{
		"documents": []interface{}{map[string]interface{}{"title": "Raw Doc"}},
	})
	if err != nil {
		t.Fatalf("CreateDocumentsRaw() error = %v", err)
	}
	if len(docs) != 1 || docs[0].ID != "doc1" {
		t.Fatalf("CreateDocumentsRaw() docs = %#v", docs)
	}
}

func TestClient_UpdateDocument(t *testing.T) {
	t.Run("title only update succeeds", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Errorf("Expected PUT method, got %s", r.Method)
			}
			if r.URL.Path != "/blocks" {
				t.Errorf("Expected path /blocks, got %s", r.URL.Path)
			}
			var body struct {
				Blocks []struct {
					ID       string `json:"id"`
					Markdown string `json:"markdown"`
				} `json:"blocks"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("Failed to decode request body: %v", err)
			}
			if len(body.Blocks) != 1 {
				t.Fatalf("Expected 1 block, got %d", len(body.Blocks))
			}
			if body.Blocks[0].ID != "doc1" {
				t.Errorf("Expected id 'doc1', got %s", body.Blocks[0].ID)
			}
			if body.Blocks[0].Markdown != "Updated Document" {
				t.Errorf("Expected markdown 'Updated Document', got %s", body.Blocks[0].Markdown)
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"items": []interface{}{}})
		}))
		defer server.Close()

		client := NewClient(server.URL)
		req := &models.UpdateDocumentRequest{Title: "Updated Document"}
		_, err := client.UpdateDocument("doc1", req)
		if err != nil {
			t.Fatalf("UpdateDocument() error = %v", err)
		}
	})

	t.Run("markdown content update succeeds", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}

			if r.URL.Path != "/blocks" {
				t.Errorf("Expected path /blocks, got %s", r.URL.Path)
			}

			var req struct {
				Markdown string `json:"markdown"`
				Position struct {
					PageID   string `json:"pageId"`
					Position string `json:"position"`
				} `json:"position"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("Failed to decode request body: %v", err)
			}

			if req.Position.PageID != "doc1" {
				t.Errorf("Expected pageId 'doc1', got %s", req.Position.PageID)
			}

			if req.Position.Position != "end" {
				t.Errorf("Expected position 'end', got %s", req.Position.Position)
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
				}{
					{ID: "block1", Type: "text", Markdown: "Updated content"},
				},
			}

			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		req := &models.UpdateDocumentRequest{
			Markdown: "Updated content",
		}

		doc, err := client.UpdateDocument("doc1", req)

		if err != nil {
			t.Fatalf("UpdateDocument() error = %v", err)
		}

		if doc.ID != "doc1" {
			t.Errorf("Document ID = %v, want doc1", doc.ID)
		}
	})
}

func TestClient_DeleteDocument(t *testing.T) {
	t.Run("soft-deletes documents", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			if len(body.DocumentIDs) != 1 || body.DocumentIDs[0] != "doc1" {
				t.Errorf("Expected documentIds ['doc1'], got %#v", body.DocumentIDs)
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"items": []string{"doc1"}})
		}))
		defer server.Close()

		client := NewClient(server.URL)
		err := client.DeleteDocument("doc1")
		if err != nil {
			t.Fatalf("DeleteDocument() error = %v", err)
		}
	})
}

func TestClient_MoveDocument(t *testing.T) {
	t.Run("moves to folder using documented endpoint and payload", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Errorf("Expected PUT method, got %s", r.Method)
			}
			if r.URL.Path != "/documents/move" {
				t.Errorf("Expected path /documents/move, got %s", r.URL.Path)
			}

			var body struct {
				DocumentIDs []string `json:"documentIds"`
				Destination struct {
					FolderID    string `json:"folderId"`
					Destination string `json:"destination"`
				} `json:"destination"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("Failed to decode request body: %v", err)
			}
			if len(body.DocumentIDs) != 1 || body.DocumentIDs[0] != "doc1" {
				t.Fatalf("Expected documentIds [doc1], got %#v", body.DocumentIDs)
			}
			if body.Destination.FolderID != "folder1" {
				t.Errorf("Expected destination.folderId folder1, got %s", body.Destination.FolderID)
			}
			if body.Destination.Destination != "" {
				t.Errorf("Expected no virtual destination, got %s", body.Destination.Destination)
			}

			json.NewEncoder(w).Encode(map[string]interface{}{"items": []map[string]string{{"id": "doc1"}}})
		}))
		defer server.Close()

		client := NewClient(server.URL)
		if err := client.MoveDocument("doc1", "folder1", ""); err != nil {
			t.Fatalf("MoveDocument() error = %v", err)
		}
	})

	t.Run("moves to virtual location using documented endpoint and payload", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/documents/move" {
				t.Errorf("Expected path /documents/move, got %s", r.URL.Path)
			}

			var body struct {
				DocumentIDs []string `json:"documentIds"`
				Destination struct {
					Destination string `json:"destination"`
				} `json:"destination"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("Failed to decode request body: %v", err)
			}
			if len(body.DocumentIDs) != 1 || body.DocumentIDs[0] != "doc1" {
				t.Fatalf("Expected documentIds [doc1], got %#v", body.DocumentIDs)
			}
			if body.Destination.Destination != "unsorted" {
				t.Errorf("Expected destination.destination unsorted, got %s", body.Destination.Destination)
			}

			json.NewEncoder(w).Encode(map[string]interface{}{"items": []map[string]string{{"id": "doc1"}}})
		}))
		defer server.Close()

		client := NewClient(server.URL)
		if err := client.MoveDocument("doc1", "", "unsorted"); err != nil {
			t.Fatalf("MoveDocument() error = %v", err)
		}
	})
}

func TestClient_MoveFolder(t *testing.T) {
	t.Run("moves to parent using documented endpoint and payload", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Errorf("Expected PUT method, got %s", r.Method)
			}
			if r.URL.Path != "/folders/move" {
				t.Errorf("Expected path /folders/move, got %s", r.URL.Path)
			}

			var body struct {
				FolderIDs   []string `json:"folderIds"`
				Destination struct {
					ParentFolderID string `json:"parentFolderId"`
				} `json:"destination"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("Failed to decode request body: %v", err)
			}
			if len(body.FolderIDs) != 1 || body.FolderIDs[0] != "folder1" {
				t.Fatalf("Expected folderIds [folder1], got %#v", body.FolderIDs)
			}
			if body.Destination.ParentFolderID != "parent1" {
				t.Errorf("Expected parentFolderId parent1, got %s", body.Destination.ParentFolderID)
			}

			json.NewEncoder(w).Encode(map[string]interface{}{"items": []map[string]string{{"id": "folder1"}}})
		}))
		defer server.Close()

		client := NewClient(server.URL)
		if err := client.MoveFolder("folder1", "parent1"); err != nil {
			t.Fatalf("MoveFolder() error = %v", err)
		}
	})

	t.Run("moves to root using documented endpoint and payload", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/folders/move" {
				t.Errorf("Expected path /folders/move, got %s", r.URL.Path)
			}

			var body struct {
				FolderIDs   []string    `json:"folderIds"`
				Destination interface{} `json:"destination"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("Failed to decode request body: %v", err)
			}
			if len(body.FolderIDs) != 1 || body.FolderIDs[0] != "folder1" {
				t.Fatalf("Expected folderIds [folder1], got %#v", body.FolderIDs)
			}
			if body.Destination != "root" {
				t.Errorf("Expected destination root, got %#v", body.Destination)
			}

			json.NewEncoder(w).Encode(map[string]interface{}{"items": []map[string]string{{"id": "folder1"}}})
		}))
		defer server.Close()

		client := NewClient(server.URL)
		if err := client.MoveFolder("folder1", ""); err != nil {
			t.Fatalf("MoveFolder() error = %v", err)
		}
	})
}

func TestClient_ClearDocumentContent(t *testing.T) {
	t.Run("deletes content blocks", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// First call: GET /blocks to fetch document structure
			if r.Method == "GET" && r.URL.Path == "/blocks" {
				response := models.BlocksResponse{
					ID:        "doc1",
					Type:      "page",
					TextStyle: "page",
					Markdown:  "Test Document",
					Content: []models.Block{
						{ID: "block1", Type: "text", Markdown: "Content block 1"},
						{ID: "block2", Type: "text", Markdown: "Content block 2"},
					},
				}
				json.NewEncoder(w).Encode(response)
				return
			}

			// Second call: DELETE /blocks with block IDs
			if r.Method == "DELETE" && r.URL.Path == "/blocks" {
				var req struct {
					BlockIDs []string `json:"blockIds"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Failed to decode request body: %v", err)
				}

				if len(req.BlockIDs) != 2 {
					t.Errorf("Expected 2 block IDs, got %d", len(req.BlockIDs))
				}

				response := struct {
					Items []struct {
						ID string `json:"id"`
					} `json:"items"`
				}{
					Items: []struct {
						ID string `json:"id"`
					}{
						{ID: "block1"},
						{ID: "block2"},
					},
				}
				json.NewEncoder(w).Encode(response)
				return
			}

			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		count, err := client.ClearDocumentContent("doc1")
		if err != nil {
			t.Fatalf("ClearDocumentContent() error = %v", err)
		}
		if count != 2 {
			t.Errorf("Expected 2 deleted blocks, got %d", count)
		}
	})

	t.Run("fails for empty document", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := models.BlocksResponse{ID: "doc1", Type: "page", Markdown: "Empty Document", Content: []models.Block{}}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		count, err := client.ClearDocumentContent("doc1")
		if err != nil {
			t.Fatalf("Expected no error for empty document, got: %v", err)
		}
		if count != 0 {
			t.Fatalf("Expected 0 deleted blocks, got %d", count)
		}
	})
}

func TestClient_DeleteBlocks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}
		if r.URL.Path != "/blocks" {
			t.Errorf("Expected path /blocks, got %s", r.URL.Path)
		}
		var body map[string][]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}
		got := body["blockIds"]
		if len(got) != 2 || got[0] != "a" || got[1] != "b" {
			t.Fatalf("blockIds = %#v, want [a b]", got)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"items": []interface{}{}})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	if err := client.DeleteBlocks([]string{"a", "b"}); err != nil {
		t.Fatalf("DeleteBlocks() error = %v", err)
	}
}

func TestClient_AddMarkdownAtPosition(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/blocks" {
			t.Errorf("Expected path /blocks, got %s", r.URL.Path)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}
		if body["markdown"] != "## Intro\n\nNew" {
			t.Fatalf("markdown = %q", body["markdown"])
		}
		pos := body["position"].(map[string]interface{})
		if pos["siblingId"] != "next" || pos["position"] != "before" {
			t.Fatalf("position = %#v", pos)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{{"id": "new", "markdown": "## Intro"}},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	last, err := client.AddMarkdownAtPosition("## Intro\n\nNew", map[string]interface{}{"siblingId": "next", "position": "before"}, 30000)
	if err != nil {
		t.Fatalf("AddMarkdownAtPosition() error = %v", err)
	}
	if last != "## Intro" {
		t.Fatalf("last = %q, want heading markdown", last)
	}
}

func TestClient_AddBlocksJSON(t *testing.T) {
	t.Run("sends blocks array with position", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}
			if r.URL.Path != "/blocks" {
				t.Errorf("Expected path /blocks, got %s", r.URL.Path)
			}

			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("Failed to decode request body: %v", err)
			}

			// Verify blocks array
			blocks, ok := body["blocks"].([]interface{})
			if !ok {
				t.Fatal("Expected 'blocks' array in request body")
			}
			if len(blocks) != 2 {
				t.Errorf("Expected 2 blocks, got %d", len(blocks))
			}

			// Verify first block has styling
			b1 := blocks[0].(map[string]interface{})
			if b1["type"] != "text" {
				t.Errorf("Expected type 'text', got %v", b1["type"])
			}
			if b1["textStyle"] != "h1" {
				t.Errorf("Expected textStyle 'h1', got %v", b1["textStyle"])
			}
			if b1["color"] != "#ef052a" {
				t.Errorf("Expected color '#ef052a', got %v", b1["color"])
			}

			// Verify second block is a line
			b2 := blocks[1].(map[string]interface{})
			if b2["type"] != "line" {
				t.Errorf("Expected type 'line', got %v", b2["type"])
			}
			if b2["lineStyle"] != "strong" {
				t.Errorf("Expected lineStyle 'strong', got %v", b2["lineStyle"])
			}

			// Verify position
			pos, ok := body["position"].(map[string]interface{})
			if !ok {
				t.Fatal("Expected 'position' map in request body")
			}
			if pos["pageId"] != "page1" {
				t.Errorf("Expected pageId 'page1', got %v", pos["pageId"])
			}
			if pos["position"] != "end" {
				t.Errorf("Expected position 'end', got %v", pos["position"])
			}

			response := struct {
				Items []models.Block `json:"items"`
			}{
				Items: []models.Block{
					{ID: "block1", Type: "text", TextStyle: "h1", Color: "#ef052a"},
					{ID: "block2", Type: "line", LineStyle: "strong"},
				},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		blocks := []map[string]interface{}{
			{"type": "text", "textStyle": "h1", "color": "#ef052a", "markdown": "# Title"},
			{"type": "line", "lineStyle": "strong"},
		}
		position := map[string]interface{}{
			"pageId":   "page1",
			"position": "end",
		}

		result, err := client.AddBlocksJSON(blocks, position)
		if err != nil {
			t.Fatalf("AddBlocksJSON() error = %v", err)
		}
		if len(result) != 2 {
			t.Errorf("Expected 2 blocks returned, got %d", len(result))
		}
		if result[0].ID != "block1" {
			t.Errorf("Expected first block ID 'block1', got %s", result[0].ID)
		}
	})

	t.Run("sends single block with sibling position", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("Failed to decode: %v", err)
			}

			pos := body["position"].(map[string]interface{})
			if pos["siblingId"] != "sib1" {
				t.Errorf("Expected siblingId 'sib1', got %v", pos["siblingId"])
			}
			if pos["position"] != "before" {
				t.Errorf("Expected position 'before', got %v", pos["position"])
			}

			response := struct {
				Items []models.Block `json:"items"`
			}{Items: []models.Block{{ID: "new1", Type: "text"}}}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		blocks := []map[string]interface{}{{"type": "text", "markdown": "Hello"}}
		position := map[string]interface{}{"siblingId": "sib1", "position": "before"}

		result, err := client.AddBlocksJSON(blocks, position)
		if err != nil {
			t.Fatalf("AddBlocksJSON() error = %v", err)
		}
		if len(result) != 1 {
			t.Errorf("Expected 1 block, got %d", len(result))
		}
	})
}

func TestClient_UpdateBlocksJSON(t *testing.T) {
	t.Run("sends blocks array with styling", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Errorf("Expected PUT method, got %s", r.Method)
			}
			if r.URL.Path != "/blocks" {
				t.Errorf("Expected path /blocks, got %s", r.URL.Path)
			}

			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("Failed to decode request body: %v", err)
			}

			blocks, ok := body["blocks"].([]interface{})
			if !ok {
				t.Fatal("Expected 'blocks' array in request body")
			}
			if len(blocks) != 1 {
				t.Errorf("Expected 1 block, got %d", len(blocks))
			}

			b := blocks[0].(map[string]interface{})
			if b["id"] != "block1" {
				t.Errorf("Expected id 'block1', got %v", b["id"])
			}
			if b["color"] != "#0400ff" {
				t.Errorf("Expected color '#0400ff', got %v", b["color"])
			}
			if b["font"] != "serif" {
				t.Errorf("Expected font 'serif', got %v", b["font"])
			}
			if b["textStyle"] != "h2" {
				t.Errorf("Expected textStyle 'h2', got %v", b["textStyle"])
			}

			json.NewEncoder(w).Encode(map[string]interface{}{"items": []interface{}{}})
		}))
		defer server.Close()

		client := NewClient(server.URL)
		blocks := []map[string]interface{}{
			{"id": "block1", "color": "#0400ff", "font": "serif", "textStyle": "h2"},
		}

		err := client.UpdateBlocksJSON(blocks)
		if err != nil {
			t.Fatalf("UpdateBlocksJSON() error = %v", err)
		}
	})

	t.Run("sends multiple blocks", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("Failed to decode: %v", err)
			}

			blocks := body["blocks"].([]interface{})
			if len(blocks) != 2 {
				t.Errorf("Expected 2 blocks, got %d", len(blocks))
			}

			b1 := blocks[0].(map[string]interface{})
			b2 := blocks[1].(map[string]interface{})
			if b1["id"] != "a" {
				t.Errorf("Expected id 'a', got %v", b1["id"])
			}
			if b2["id"] != "b" {
				t.Errorf("Expected id 'b', got %v", b2["id"])
			}

			json.NewEncoder(w).Encode(map[string]interface{}{"items": []interface{}{}})
		}))
		defer server.Close()

		client := NewClient(server.URL)
		blocks := []map[string]interface{}{
			{"id": "a", "markdown": "Updated A"},
			{"id": "b", "color": "#ff0000"},
		}

		err := client.UpdateBlocksJSON(blocks)
		if err != nil {
			t.Fatalf("UpdateBlocksJSON() error = %v", err)
		}
	})
}

func TestClient_TaskPayloads(t *testing.T) {
	t.Run("adds task with nested taskInfo and location", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}
			if r.URL.Path != "/tasks" {
				t.Errorf("Expected path /tasks, got %s", r.URL.Path)
			}

			var body struct {
				Tasks []struct {
					Markdown string `json:"markdown"`
					TaskInfo struct {
						ScheduleDate string `json:"scheduleDate"`
						DeadlineDate string `json:"deadlineDate"`
					} `json:"taskInfo"`
					Location struct {
						Type       string `json:"type"`
						DocumentID string `json:"documentId"`
					} `json:"location"`
				} `json:"tasks"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("Failed to decode request body: %v", err)
			}
			if len(body.Tasks) != 1 {
				t.Fatalf("Expected 1 task, got %d", len(body.Tasks))
			}
			task := body.Tasks[0]
			if task.Markdown != "Review PR" {
				t.Errorf("Expected markdown Review PR, got %s", task.Markdown)
			}
			if task.TaskInfo.ScheduleDate != "2026-02-01" {
				t.Errorf("Expected scheduleDate 2026-02-01, got %s", task.TaskInfo.ScheduleDate)
			}
			if task.TaskInfo.DeadlineDate != "2026-02-15" {
				t.Errorf("Expected deadlineDate 2026-02-15, got %s", task.TaskInfo.DeadlineDate)
			}
			if task.Location.Type != "document" {
				t.Errorf("Expected location.type document, got %s", task.Location.Type)
			}
			if task.Location.DocumentID != "doc1" {
				t.Errorf("Expected location.documentId doc1, got %s", task.Location.DocumentID)
			}

			json.NewEncoder(w).Encode(map[string]interface{}{
				"items": []map[string]interface{}{
					{
						"id":       "task1",
						"markdown": "Review PR",
						"taskInfo": map[string]interface{}{
							"state":        "todo",
							"scheduleDate": "2026-02-01",
							"deadlineDate": "2026-02-15",
						},
					},
				},
			})
		}))
		defer server.Close()

		client := NewClient(server.URL)
		task, err := client.AddTask("Review PR", "document", "doc1", "2026-02-01", "2026-02-15", nil)
		if err != nil {
			t.Fatalf("AddTask() error = %v", err)
		}
		if task.State != "todo" {
			t.Errorf("Expected state todo, got %s", task.State)
		}
		if task.ScheduleDate != "2026-02-01" {
			t.Errorf("Expected scheduleDate 2026-02-01, got %s", task.ScheduleDate)
		}
	})

	t.Run("updates task with tasksToUpdate and nested taskInfo", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Errorf("Expected PUT method, got %s", r.Method)
			}
			if r.URL.Path != "/tasks" {
				t.Errorf("Expected path /tasks, got %s", r.URL.Path)
			}

			var body struct {
				TasksToUpdate []struct {
					ID       string `json:"id"`
					TaskInfo struct {
						State        string `json:"state"`
						ScheduleDate string `json:"scheduleDate"`
					} `json:"taskInfo"`
				} `json:"tasksToUpdate"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("Failed to decode request body: %v", err)
			}
			if len(body.TasksToUpdate) != 1 {
				t.Fatalf("Expected 1 task update, got %d", len(body.TasksToUpdate))
			}
			update := body.TasksToUpdate[0]
			if update.ID != "task1" {
				t.Errorf("Expected id task1, got %s", update.ID)
			}
			if update.TaskInfo.State != "done" {
				t.Errorf("Expected state done, got %s", update.TaskInfo.State)
			}
			if update.TaskInfo.ScheduleDate != "2026-02-01" {
				t.Errorf("Expected scheduleDate 2026-02-01, got %s", update.TaskInfo.ScheduleDate)
			}

			json.NewEncoder(w).Encode(map[string]interface{}{"items": []map[string]string{{"id": "task1"}}})
		}))
		defer server.Close()

		client := NewClient(server.URL)
		if err := client.UpdateTask("task1", "done", "2026-02-01", "", nil); err != nil {
			t.Fatalf("UpdateTask() error = %v", err)
		}
	})

	t.Run("deletes task with idsToDelete", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			if len(body.IDsToDelete) != 1 || body.IDsToDelete[0] != "task1" {
				t.Fatalf("Expected idsToDelete [task1], got %#v", body.IDsToDelete)
			}

			json.NewEncoder(w).Encode(map[string]interface{}{"items": []map[string]string{{"id": "task1"}}})
		}))
		defer server.Close()

		client := NewClient(server.URL)
		if err := client.DeleteTask("task1"); err != nil {
			t.Fatalf("DeleteTask() error = %v", err)
		}
	})

	t.Run("raw task methods pass documented envelopes", func(t *testing.T) {
		seen := map[string]bool{}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/tasks" {
				t.Errorf("Expected path /tasks, got %s", r.URL.Path)
			}
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("Failed to decode request body: %v", err)
			}
			switch r.Method {
			case "POST":
				seen["post"] = true
				if _, ok := body["tasks"].([]interface{}); !ok {
					t.Fatalf("Expected raw tasks envelope, got %#v", body)
				}
				json.NewEncoder(w).Encode(map[string]interface{}{
					"items": []map[string]interface{}{{"id": "task1", "markdown": "Raw task"}},
				})
			case "PUT":
				seen["put"] = true
				if _, ok := body["tasksToUpdate"].([]interface{}); !ok {
					t.Fatalf("Expected raw tasksToUpdate envelope, got %#v", body)
				}
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
			case "DELETE":
				seen["delete"] = true
				if _, ok := body["idsToDelete"].([]interface{}); !ok {
					t.Fatalf("Expected raw idsToDelete envelope, got %#v", body)
				}
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
			default:
				t.Fatalf("Unexpected method %s", r.Method)
			}
		}))
		defer server.Close()

		client := NewClient(server.URL)
		if _, err := client.AddTasksRaw(map[string]interface{}{"tasks": []interface{}{map[string]interface{}{"markdown": "Raw task"}}}); err != nil {
			t.Fatalf("AddTasksRaw() error = %v", err)
		}
		if err := client.UpdateTasksRaw(map[string]interface{}{"tasksToUpdate": []interface{}{map[string]interface{}{"id": "task1"}}}); err != nil {
			t.Fatalf("UpdateTasksRaw() error = %v", err)
		}
		if err := client.DeleteTasksRaw(map[string]interface{}{"idsToDelete": []interface{}{"task1"}}); err != nil {
			t.Fatalf("DeleteTasksRaw() error = %v", err)
		}
		for _, key := range []string{"post", "put", "delete"} {
			if !seen[key] {
				t.Fatalf("Expected raw %s request", key)
			}
		}
	})
}

func TestClient_ErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    string
	}{
		{"Unauthorized", 401, "authentication required"},
		{"Forbidden", 403, "permission denied"},
		{"NotFound", 404, "resource not found"},
		{"RateLimit", 429, "rate limit exceeded"},
		{"ServerError", 500, "Craft API error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				response := models.ErrorResponse{
					Error:   "error",
					Message: "test error message",
					Code:    tt.statusCode,
				}
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client := NewClient(server.URL)
			_, err := client.GetDocuments()

			if err == nil {
				t.Error("Expected error, got nil")
			}

			if err != nil && len(tt.wantErr) > 0 {
				errStr := err.Error()
				if len(errStr) < len(tt.wantErr) || errStr[:len(tt.wantErr)] != tt.wantErr {
					t.Errorf("Error = %v, want to contain %v", err, tt.wantErr)
				}
			}
		})
	}
}
