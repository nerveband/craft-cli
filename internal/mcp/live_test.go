package mcp

import (
	"encoding/json"
	"os"
	"testing"
)

func liveMCPClient(t *testing.T) *Client {
	t.Helper()
	if os.Getenv("CRAFT_LIVE_TESTS") != "1" {
		t.Skip("set CRAFT_LIVE_TESTS=1 to run live MCP tests")
	}
	url := firstEnv("CRAFT_LIVE_MCP_URL", "CRAFT_MCP_URL")
	if url == "" {
		t.Skip("set CRAFT_LIVE_MCP_URL or CRAFT_MCP_URL")
	}
	return NewClient(url)
}

func firstEnv(names ...string) string {
	for _, name := range names {
		if value := os.Getenv(name); value != "" {
			return value
		}
	}
	return ""
}

func TestLiveMCPReadMatrix(t *testing.T) {
	client := liveMCPClient(t)

	if _, err := client.Initialize(); err != nil {
		t.Fatalf("Initialize() live error = %v", err)
	}

	tools, err := client.ListTools()
	if err != nil {
		t.Fatalf("ListTools() live error = %v", err)
	}
	var toolBody struct {
		Tools []struct {
			Name string `json:"name"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(tools, &toolBody); err != nil {
		t.Fatalf("failed to decode live tools: %v", err)
	}
	if !hasLiveTool(toolBody.Tools, "craft_read") {
		t.Fatalf("live MCP tools missing craft_read: %#v", toolBody.Tools)
	}

	if _, err := client.ListResources(); err != nil {
		t.Fatalf("ListResources() live error = %v", err)
	}

	if _, err := client.CallTool("craft_read", map[string]interface{}{"command": "connection info"}); err != nil {
		t.Fatalf("craft_read connection info live error = %v", err)
	}
	if _, err := client.CallTool("craft_read", map[string]interface{}{"command": "documents list --help"}); err != nil {
		t.Fatalf("craft_read documents list --help live error = %v", err)
	}
	if _, err := client.CallTool("craft_read", map[string]interface{}{"command": "blocks explore-themes --help"}); err != nil {
		t.Fatalf("craft_read blocks explore-themes --help live error = %v", err)
	}
}

func hasLiveTool(tools []struct {
	Name string `json:"name"`
}, name string) bool {
	for _, tool := range tools {
		if tool.Name == name {
			return true
		}
	}
	return false
}
