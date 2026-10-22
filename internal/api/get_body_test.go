package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestJSONGETDoesNotSendTypedNilBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if len(body) > 0 || r.Header.Get("Content-Type") != "" {
			t.Errorf("bodyless request carried JSON body: %q", body)
		}
		w.Write([]byte(`{"items":[]}`))
	}))
	defer server.Close()
	var payload map[string]interface{}
	for _, request := range []struct{ method, path string }{{"GET", "/reminders"}, {"DELETE", "/collections/c/views/v"}} {
		if _, err := NewClient(server.URL).RequestJSON(request.method, request.path, payload); err != nil {
			t.Fatal(err)
		}
	}
}
