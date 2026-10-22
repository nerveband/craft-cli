package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUploadDefaultTransportPreservesInlineMediaAndUsesBinaryForFiles(t *testing.T) {
	for _, tc := range []struct {
		name           string
		body           []byte
		explicit, want string
	}{{"text file", []byte("plain notes"), "", "application/octet-stream"}, {"explicit", []byte("plain notes"), "text/plain", "text/plain"}, {"png", []byte{137, 80, 78, 71, 13, 10, 26, 10}, "", "image/png"}} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Content-Type") != tc.want {
					t.Errorf("wrong MIME transport %q", r.Header.Get("Content-Type"))
				}
				w.Write([]byte(`{"blockId":"file-a","assetUrl":"https://example.org/file"}`))
			}))
			defer server.Close()
			if _, err := NewClient(server.URL).UploadFileNamed(tc.body, "page-a", "", "", "end", "test.txt", tc.explicit); err != nil {
				t.Fatal(err)
			}
		})
	}
}
