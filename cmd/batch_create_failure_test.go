package cmd

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestBatchCreateStopsAndReportsKnownPartialFailure(t *testing.T) {
	oldURL, oldKey, oldEnv, oldQuiet, oldDry, oldWrites, oldInput := apiURL, apiKey, apiKeyEnv, quietMode, dryRun, writeResponses, os.Stdin
	defer func() {
		apiURL, apiKey, apiKeyEnv, quietMode, dryRun, writeResponses, os.Stdin = oldURL, oldKey, oldEnv, oldQuiet, oldDry, oldWrites, oldInput
	}()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Write([]byte(`{"items":[{"id":"created-a","title":"First"}]}`))
		} else {
			w.WriteHeader(500)
			w.Write([]byte(`{"error":"fixture failure"}`))
		}
	}))
	defer server.Close()
	apiURL, apiKey, apiKeyEnv, quietMode, dryRun, writeResponses = server.URL, "", "", true, false, nil
	input, err := os.CreateTemp(t.TempDir(), "batch-input")
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	input.WriteString(`[{"title":"First"},{"title":"Second"},{"title":"Third"}]`)
	input.Seek(0, 0)
	os.Stdin = input
	err = runBatchCreate()
	var coded *cliCodeError
	if !errors.As(err, &coded) || coded.Code != "PARTIAL_FAILURE" || calls != 2 || len(writeResponses) != 1 {
		t.Fatalf("partial batch not preserved: calls=%d known=%d err=%v", calls, len(writeResponses), err)
	}
}
