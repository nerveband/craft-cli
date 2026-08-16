package api

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ashrafali/craft-cli/internal/models"
)

func liveRESTClient(t *testing.T) *Client {
	t.Helper()
	if os.Getenv("CRAFT_LIVE_TESTS") != "1" {
		t.Skip("set CRAFT_LIVE_TESTS=1 to run live REST tests")
	}
	url := firstEnv("CRAFT_LIVE_REST_URL", "CRAFT_API_URL")
	if url == "" {
		t.Skip("set CRAFT_LIVE_REST_URL or CRAFT_API_URL")
	}
	key := firstEnv("CRAFT_LIVE_REST_KEY", "CRAFT_API_KEY")
	if key != "" {
		return NewClientWithKey(url, key)
	}
	return NewClient(url)
}

func requireLiveMutations(t *testing.T) {
	t.Helper()
	if os.Getenv("CRAFT_LIVE_TESTS") != "1" || os.Getenv("CRAFT_LIVE_MUTATION_TESTS") != "1" {
		t.Skip("set CRAFT_LIVE_TESTS=1 and CRAFT_LIVE_MUTATION_TESTS=1 to run live mutation tests")
	}
}

func liveArtifactPrefix() string {
	return "craft-cli-live-test-" + time.Now().Format("20060102-150405")
}

func firstEnv(names ...string) string {
	for _, name := range names {
		if value := os.Getenv(name); value != "" {
			return value
		}
	}
	return ""
}

func TestLiveRESTRead(t *testing.T) {
	client := liveRESTClient(t)

	if _, err := client.GetConnection(); err != nil {
		t.Fatalf("GetConnection() live error = %v", err)
	}
	if _, err := client.GetDocumentsFiltered("", ""); err != nil {
		t.Fatalf("GetDocumentsFiltered() live error = %v", err)
	}
	if _, err := client.SearchDocumentsAdvanced("test", SearchOptions{}); err != nil {
		t.Fatalf("SearchDocumentsAdvanced() live error = %v", err)
	}
}

func TestLiveRESTWriteOnlyState(t *testing.T) {
	if os.Getenv("CRAFT_LIVE_TESTS") != "1" {
		t.Skip("set CRAFT_LIVE_TESTS=1 to run live write-only state test")
	}
	url := os.Getenv("CRAFT_LIVE_WRITEONLY_URL")
	if url == "" {
		t.Skip("set CRAFT_LIVE_WRITEONLY_URL")
	}
	client := NewClient(url)
	_, err := client.GetDocumentsFiltered("", "")
	if err != nil {
		t.Logf("write-only read failed cleanly: %v", err)
		return
	}
	t.Log("write-only read returned limited data; endpoint allows this state")
}

func TestLiveRESTMutationCreateUpdateDelete(t *testing.T) {
	requireLiveMutations(t)
	client := liveRESTClient(t)
	prefix := liveArtifactPrefix()

	doc, err := client.CreateDocument(&models.CreateDocumentRequest{
		Title:    prefix,
		Markdown: "temporary craft-cli live test document",
	})
	if err != nil {
		t.Fatalf("CreateDocument() live error = %v", err)
	}
	if !strings.HasPrefix(doc.Title, "craft-cli-live-test-") {
		t.Fatalf("refusing to mutate unexpected title %q", doc.Title)
	}
	t.Cleanup(func() {
		if err := client.DeleteDocument(doc.ID); err != nil {
			t.Logf("cleanup DeleteDocument(%s) failed: %v", doc.ID, err)
		}
	})

	if _, err := client.UpdateDocument(doc.ID, &models.UpdateDocumentRequest{Markdown: "live update"}); err != nil {
		t.Fatalf("UpdateDocument() live error = %v", err)
	}

	task, err := client.AddTask(prefix+" task", "inbox", "", "", "", nil)
	if err != nil {
		t.Fatalf("AddTask() live error = %v", err)
	}
	t.Cleanup(func() {
		if err := client.DeleteTask(task.ID); err != nil {
			t.Logf("cleanup DeleteTask(%s) failed: %v", task.ID, err)
		}
	})
	if err := client.UpdateTask(task.ID, "done", "", "", nil); err != nil {
		t.Fatalf("UpdateTask() live error = %v", err)
	}
	if err := client.DeleteTask(task.ID); err != nil {
		t.Fatalf("DeleteTask() live error = %v", err)
	}
}
