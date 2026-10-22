package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBlockSearchConnectionTargets(t *testing.T) {
	for _, target := range []string{"blockId", "documentId", "date"} {
		t.Run(target, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()
				if q.Get(target) != "fixture" || q.Get("pattern") != "notes" {
					t.Errorf("wrong query %v", q)
				}
				for _, other := range []string{"blockId", "documentId", "date"} {
					if other != target && q.Has(other) {
						t.Errorf("unexpected target %s", other)
					}
				}
				w.Write([]byte(`{"items":[]}`))
			}))
			defer server.Close()
			if _, err := NewClient(server.URL).SearchBlocksWithTarget(target, "fixture", "notes", false, 1, 2, true); err != nil {
				t.Fatal(err)
			}
		})
	}
}
