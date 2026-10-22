package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMoveFolderRootUsesDestinationObject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		destination, ok := body["destination"].(map[string]interface{})
		if !ok || destination["destination"] != "root" {
			t.Errorf("invalid root destination %#v", body)
		}
		w.Write([]byte(`{"items":["folder-a"]}`))
	}))
	defer server.Close()
	if err := NewClient(server.URL).MoveFolder("folder-a", ""); err != nil {
		t.Fatal(err)
	}
}
