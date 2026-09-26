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

func TestSessionNegotiationAndToolError(t *testing.T) {
	notified := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request map[string]interface{}
		json.NewDecoder(r.Body).Decode(&request)
		method, _ := request["method"].(string)
		if method == "initialize" {
			w.Header().Set("Mcp-Session-Id", "fixture-session")
			json.NewEncoder(w).Encode(map[string]interface{}{"jsonrpc": "2.0", "id": request["id"], "result": map[string]interface{}{"protocolVersion": "2025-06-18", "capabilities": map[string]interface{}{}, "serverInfo": map[string]string{"name": "fixture", "version": "1"}}})
			return
		}
		if r.Header.Get("Mcp-Session-Id") != "fixture-session" {
			t.Error("session not propagated")
		}
		if method == "notifications/initialized" {
			notified = true
			w.WriteHeader(202)
			return
		}
		if !notified {
			t.Error("called tool before initialization notification")
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"jsonrpc": "2.0", "id": request["id"], "result": map[string]interface{}{"isError": true, "structuredContent": map[string]string{"reason": "fixture"}, "content": []map[string]string{{"type": "text", "text": "refused"}}}})
	}))
	defer server.Close()
	client := NewClient(server.URL)
	if _, err := client.Initialize(); err != nil {
		t.Fatal(err)
	}
	result, err := client.CallTool("craft_read", map[string]interface{}{"command": "documents list"})
	if _, ok := err.(*ToolError); !ok || len(result) == 0 {
		t.Fatalf("tool error lost: %s %v", result, err)
	}
}

func TestSSESeparatesNotifications(t *testing.T) {
	data := []byte("event: message\ndata: {\"jsonrpc\":\"2.0\",\"method\":\"notifications/progress\"}\n\nevent: message\ndata: {\"jsonrpc\":\"2.0\",\"id\":8,\"result\":{\"ok\":true}}\n\n")
	response, err := matchingSSEData(data, 8)
	if err != nil || !json.Valid(response) {
		t.Fatalf("SSE failed: %s %v", response, err)
	}
	if _, err := matchingSSEData(data, 9); err == nil {
		t.Fatal("accepted wrong response ID")
	}
}
