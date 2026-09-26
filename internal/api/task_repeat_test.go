package api

import (
	"encoding/json"
	"github.com/ashrafali/craft-cli/internal/models"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTaskRepeatContract(t *testing.T) {
	for _, method := range []string{"POST", "PUT"} {
		t.Run(method, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != method || r.URL.Path != "/tasks" {
					t.Errorf("unexpected request %s %s", r.Method, r.URL)
				}
				var body map[string][]map[string]json.RawMessage
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
					return
				}
				key := "tasks"
				if method == "PUT" {
					key = "tasksToUpdate"
				}
				if len(body[key]) != 1 {
					t.Errorf("invalid body: %v", body)
					return
				}
				item := body[key][0]
				var repeat map[string]interface{}
				json.Unmarshal(item["repeat"], &repeat)
				if repeat["type"] != "fixed" || repeat["frequency"] != "weekly" {
					t.Errorf("wrong repeat discriminator: %s", item["repeat"])
				}
				var info map[string]interface{}
				json.Unmarshal(item["taskInfo"], &info)
				if _, ok := info["repeat"]; ok {
					t.Error("repeat belongs at task level, not taskInfo")
				}
				w.Write([]byte(`{"items":[{"id":"task1","markdown":"Water","taskInfo":{"state":"todo"},"location":{"type":"inbox"}}]}`))
			}))
			defer server.Close()
			client := NewClient(server.URL)
			repeat := &models.RepeatConfig{Type: "fixed", Frequency: "weekly", Weekly: map[string]interface{}{"days": []string{"monday"}}}
			var err error
			if method == "POST" {
				_, err = client.AddTask("Water", "inbox", "", "", "", repeat)
			} else {
				err = client.UpdateTask("task1", "todo", "", "", repeat)
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
