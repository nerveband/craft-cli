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
