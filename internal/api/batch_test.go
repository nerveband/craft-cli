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
