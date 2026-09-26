package api

import (
	"encoding/json"
	"github.com/ashrafali/craft-cli/docs/contracts"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCurrentRequestEncoding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if r.Body != nil {
			json.NewDecoder(r.Body).Decode(&body)
		}
		switch r.URL.Path {
		case "/documents/search":
			q := r.URL.Query()
			if len(q["folderIds"]) != 2 || len(q["documentIds"]) != 2 || q.Has("folderIDs") || q.Has("include") {
				t.Errorf("incorrect search query %v", q)
			}
			w.Write([]byte(`{"items":[]}`))
		case "/blocks/move":
			if r.Method != "PUT" || body["blockIds"] == nil || body["position"] == nil || body["blocks"] != nil {
				t.Errorf("incorrect move %s %#v", r.Method, body)
			}
			w.Write([]byte(`{"items":[{"id":"block-a"}]}`))
		case "/folders":
			folder := body["folders"].([]interface{})[0].(map[string]interface{})
			if folder["parentFolderId"] != "parent-a" || folder["parentId"] != nil {
				t.Errorf("incorrect folder %#v", folder)
			}
			w.Write([]byte(`{"items":[{"id":"folder-a","name":"Folder"}]}`))
		case "/tasks":
			if r.URL.Query().Get("scope") != "document" || r.URL.Query().Get("documentId") != "doc-a" {
				t.Errorf("incorrect task scope %s", r.URL)
			}
			w.Write([]byte(`{"items":[]}`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	c := NewClient(server.URL)
	if _, err := c.SearchDocumentsAdvanced("", SearchOptions{Regexps: "abc", FolderIDs: "one,two", DocumentIDs: "a,b"}); err != nil {
		t.Fatal(err)
	}
	if err := c.MoveBlock("block-a", "page-a", "end"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.CreateFolder("Folder", "parent-a"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.GetDocumentTasks("doc-a"); err != nil {
		t.Fatal(err)
	}
}

func TestRetryMetadataAndWriteOutcomes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Header().Set("Retry-After", "12")
			w.Header().Set("X-RateLimit-Scope", "space")
			w.WriteHeader(429)
			w.Write([]byte(`{"message":"limited"}`))
			return
		}
		w.Write([]byte(`{"items":[{"id":"a","status":"deleted"},{"id":"b","error":"denied"}]}`))
	}))
	defer server.Close()
	c := NewClient(server.URL)
	_, err := c.GetDocuments()
	e, ok := err.(*APIError)
	if !ok || e.Headers["Retry-After"] != "12" {
		t.Fatalf("lost retry metadata: %#v", err)
	}
	var result []byte
	c.ObserveWrite = func(b []byte) { result = b }
	if err := c.DeleteDocuments([]string{"a", "b"}); err == nil || !json.Valid(result) {
		t.Fatalf("lost partial failure: %s %v", result, err)
	}
}

func TestPinnedContractMatchesEncodingFixtures(t *testing.T) {
	spec, err := contracts.OpenAPI()
	if err != nil {
		t.Fatal(err)
	}
	paths := spec["paths"].(map[string]interface{})
	operations := 0
	for _, raw := range paths {
		for method := range raw.(map[string]interface{}) {
			switch method {
			case "get", "post", "put", "delete", "patch":
				operations++
			}
		}
	}
	if operations != 44 {
		t.Fatalf("contract inventory changed to %d operations; review coverage", operations)
	}
	search := paths["/documents/search"].(map[string]interface{})["get"].(map[string]interface{})
	names := map[string]bool{}
	for _, raw := range search["parameters"].([]interface{}) {
		names[raw.(map[string]interface{})["name"].(string)] = true
	}
	for _, name := range []string{"folderIds", "documentIds", "fetchBlocks"} {
		if !names[name] {
			t.Errorf("fixture parameter %s absent from pin", name)
		}
	}
	if names["folderIDs"] || names["documentIDs"] {
		t.Fatal("legacy casing appeared; review encoding")
	}
	move := paths["/blocks/move"].(map[string]interface{})["put"].(map[string]interface{})
	props := move["requestBody"].(map[string]interface{})["content"].(map[string]interface{})["application/json"].(map[string]interface{})["schema"].(map[string]interface{})["properties"].(map[string]interface{})
	if props["blockIds"] == nil || props["position"] == nil {
		t.Fatal("block move contract changed")
	}
}
