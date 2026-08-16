package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ashrafali/craft-cli/internal/models"
)

func TestClient_AddTaskWithRepeat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Tasks []struct {
				Markdown string `json:"markdown"`
				TaskInfo struct {
					Repeat *models.RepeatConfig `json:"repeat"`
				} `json:"taskInfo"`
			} `json:"tasks"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}
		if len(body.Tasks) != 1 {
			t.Fatalf("Expected 1 task, got %d", len(body.Tasks))
		}
		repeat := body.Tasks[0].TaskInfo.Repeat
		if repeat == nil {
			t.Fatal("Expected taskInfo.repeat in payload")
		}
		if repeat.Type != "weekly" || repeat.Interval != 2 || !repeat.SkipWeekends || repeat.Reminder != "09:00" {
			t.Errorf("Unexpected repeat payload: %+v", repeat)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{{"id": "task1", "markdown": "Water plants", "state": "todo"}},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	repeat := &models.RepeatConfig{Type: "weekly", Interval: 2, SkipWeekends: true, Reminder: "09:00"}
	task, err := client.AddTask("Water plants", "inbox", "", "", "", repeat)
	if err != nil {
		t.Fatalf("AddTask() error = %v", err)
	}
	if task.ID != "task1" {
		t.Errorf("task.ID = %q, want task1", task.ID)
	}
}

func TestClient_UpdateTaskWithRepeat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			TasksToUpdate []struct {
				ID       string `json:"id"`
				TaskInfo struct {
					Repeat *models.RepeatConfig `json:"repeat"`
				} `json:"taskInfo"`
			} `json:"tasksToUpdate"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}
		if len(body.TasksToUpdate) != 1 || body.TasksToUpdate[0].ID != "task1" {
			t.Fatalf("Unexpected tasksToUpdate: %+v", body.TasksToUpdate)
		}
		repeat := body.TasksToUpdate[0].TaskInfo.Repeat
		if repeat == nil || repeat.Frequency != "daily" || !repeat.DynamicDays {
			t.Errorf("Unexpected repeat payload: %+v", repeat)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"items": []string{"task1"}})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	repeat := &models.RepeatConfig{Frequency: "daily", DynamicDays: true}
	if err := client.UpdateTask("task1", "", "", "", repeat); err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}
}
