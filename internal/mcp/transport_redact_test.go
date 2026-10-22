package mcp

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
	c := NewClient("https://example.org/links/fixture-capability/mcp")
	c.httpClient = &http.Client{Transport: fixtureFailingTransport{}}
	_, err := c.Initialize()
	if err == nil || strings.Contains(err.Error(), "fixture-capability") || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("unsafe transport error or lost timeout cause")
	}
}

func TestInvalidRequestURLDoesNotExposeCapabilities(t *testing.T) {
	c := NewClient("https://example.org/links/fixture-capability/\n")
	_, err := c.Initialize()
	if err == nil || strings.Contains(err.Error(), "fixture-capability") {
		t.Fatal("invalid URL leaked a connection capability")
	}
}
