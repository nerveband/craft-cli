package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

type fixtureFailingTransport struct{}

func (fixtureFailingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, context.DeadlineExceeded
}
func TestTransportErrorsHideConnectionCapabilities(t *testing.T) {
	c := NewClient("https://example.org/links/fixture-capability/api/v1")
	c.httpClient = &http.Client{Transport: fixtureFailingTransport{}}
	for _, call := range []func() error{func() error { _, e := c.GetDocuments(); return e }, func() error { _, e := c.UploadFile([]byte("test"), "page-123", "", "", "end"); return e }} {
		err := call()
		if err == nil || strings.Contains(err.Error(), "fixture-capability") || !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal("unsafe transport error or lost timeout cause")
		}
	}
}

func TestInvalidRequestURLDoesNotExposeCapabilities(t *testing.T) {
	c := NewClient("https://example.org/links/fixture-capability/\n")
	_, err := c.GetDocuments()
	if err == nil || strings.Contains(err.Error(), "fixture-capability") {
		t.Fatal("invalid URL leaked a connection capability")
	}
}
