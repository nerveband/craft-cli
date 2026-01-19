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
					ID:        "doc1",
					Title:     "Test Document",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
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
		if r.URL.Path != "/documents/doc1" {
			t.Errorf("Expected path /documents/doc1, got %s", r.URL.Path)
		}
		
		response := models.Document{
			ID:        "doc1",
			Title:     "Test Document",
			Content:   "Test content",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
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
}

func TestClient_SearchDocuments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/documents/search" {
			t.Errorf("Expected path /documents/search, got %s", r.URL.Path)
		}
		
		query := r.URL.Query().Get("query")
		if query != "test query" {
			t.Errorf("Expected query 'test query', got %s", query)
		}
		
		response := models.SearchResult{
			Items: []models.Document{
				{
					ID:        "doc1",
					Title:     "Test Document",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			Total: 1,
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
		t.Errorf("Expected 1 document, got %d", len(results.Items))
	}
}

func TestClient_CreateDocument(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		
		var req models.CreateDocumentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}
		
		if req.Title != "New Document" {
			t.Errorf("Expected title 'New Document', got %s", req.Title)
		}
		
		response := models.Document{
			ID:        "doc1",
			Title:     req.Title,
			Content:   req.Content,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	req := &models.CreateDocumentRequest{
		Title:   "New Document",
		Content: "Test content",
	}
	
	doc, err := client.CreateDocument(req)
	
	if err != nil {
		t.Fatalf("CreateDocument() error = %v", err)
	}
	
	if doc.Title != "New Document" {
		t.Errorf("Document title = %v, want New Document", doc.Title)
	}
}

func TestClient_UpdateDocument(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT method, got %s", r.Method)
		}
		
		if r.URL.Path != "/documents/doc1" {
			t.Errorf("Expected path /documents/doc1, got %s", r.URL.Path)
		}
		
		var req models.UpdateDocumentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}
		
		response := models.Document{
			ID:        "doc1",
			Title:     req.Title,
			Content:   req.Content,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	req := &models.UpdateDocumentRequest{
		Title:   "Updated Document",
		Content: "Updated content",
	}
	
	doc, err := client.UpdateDocument("doc1", req)
	
	if err != nil {
		t.Fatalf("UpdateDocument() error = %v", err)
	}
	
	if doc.Title != "Updated Document" {
		t.Errorf("Document title = %v, want Updated Document", doc.Title)
	}
}

func TestClient_DeleteDocument(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}
		
		if r.URL.Path != "/documents/doc1" {
			t.Errorf("Expected path /documents/doc1, got %s", r.URL.Path)
		}
		
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.DeleteDocument("doc1")
	
	if err != nil {
		t.Fatalf("DeleteDocument() error = %v", err)
	}
}

func TestClient_ErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    string
	}{
		{"Unauthorized", 401, "authentication failed"},
		{"Forbidden", 403, "authentication failed"},
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
