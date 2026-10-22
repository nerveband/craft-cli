package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddCollectionItemSendsEmptyPropertiesAndOmitsUnrequestedOptionMutation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if _, ok := body["allowNewSelectOptions"]; ok {
			t.Error("sent unrequested option mutation flag")
		}
		item := body["items"].([]interface{})[0].(map[string]interface{})
		if properties, ok := item["properties"].(map[string]interface{}); !ok || len(properties) != 0 {
			t.Errorf("missing empty properties: %#v", item)
		}
		w.Write([]byte(`{"items":[{"id":"item-a"}]}`))
	}))
	defer server.Close()
	if _, err := NewClient(server.URL).AddCollectionItem("collection-a", "Item", nil, false); err != nil {
		t.Fatal(err)
	}
}
