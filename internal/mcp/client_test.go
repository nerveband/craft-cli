package mcp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_ListToolsParsesEventStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Accept") != "application/json, text/event-stream" {
			t.Errorf("expected MCP Accept header, got %s", r.Header.Get("Accept"))
		}

		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Method != "tools/list" {
			t.Errorf("expected tools/list, got %s", req.Method)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(`event: message
data: {"jsonrpc":"2.0","id":1,"result":{"tools":[{"name":"craft_read"}]}}

`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.ListTools()
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}

	var body struct {
		Tools []struct {
			Name string `json:"name"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(result, &body); err != nil {
		t.Fatalf("failed to decode result: %v", err)
	}
	if len(body.Tools) != 1 || body.Tools[0].Name != "craft_read" {
		t.Fatalf("unexpected tools result: %#v", body.Tools)
	}
}

func TestClient_CallToolSendsArguments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Method != "tools/call" {
			t.Errorf("expected tools/call, got %s", req.Method)
		}

		params := req.Params.(map[string]interface{})
		if params["name"] != "craft_read" {
			t.Errorf("expected tool craft_read, got %v", params["name"])
		}
		args := params["arguments"].(map[string]interface{})
		if args["command"] != "connection info" {
			t.Errorf("expected command connection info, got %v", args["command"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"content":[{"type":"text","text":"ok"}]}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.CallTool("craft_read", map[string]interface{}{"command": "connection info"})
	if err != nil {
		t.Fatalf("CallTool() error = %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected result")
	}
}

func TestClient_ReadResourceSendsURI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Method != "resources/read" {
			t.Errorf("expected resources/read, got %s", req.Method)
		}
		params := req.Params.(map[string]interface{})
		if params["uri"] != "ui://craft/edit-review" {
			t.Errorf("expected edit-review URI, got %v", params["uri"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"contents":[]}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.ReadResource("ui://craft/edit-review")
	if err != nil {
		t.Fatalf("ReadResource() error = %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected result")
	}
}

func TestClient_ReturnsJSONRPCError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"error":{"code":-32601,"message":"missing"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.ListResources()
	if err == nil {
		t.Fatal("expected error")
	}
}
