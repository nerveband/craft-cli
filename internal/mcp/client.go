package mcp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

const defaultTimeout = 30 * time.Second

// Client is a minimal Streamable HTTP MCP client for Craft's public MCP endpoint.
type Client struct {
	url        string
	httpClient *http.Client
	nextID     int64
}

// NewClient creates a new MCP client.
func NewClient(url string) *Client {
	return &Client{
		url: url,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// Request is a JSON-RPC request.
type Request struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int64       `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// Response is a JSON-RPC response.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

// Error is a JSON-RPC error payload.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("MCP error %d: %s", e.Code, e.Message)
}

// Initialize performs the MCP initialize request.
func (c *Client) Initialize() (json.RawMessage, error) {
	params := map[string]interface{}{
		"protocolVersion": "2025-06-18",
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]string{
			"name":    "craft-cli",
			"version": "1.0.0",
		},
	}
	return c.Call("initialize", params)
}

// ListTools returns the raw tools/list result.
func (c *Client) ListTools() (json.RawMessage, error) {
	return c.Call("tools/list", map[string]interface{}{})
}

// ListResources returns the raw resources/list result.
func (c *Client) ListResources() (json.RawMessage, error) {
	return c.Call("resources/list", map[string]interface{}{})
}

// ReadResource reads an MCP resource by URI.
func (c *Client) ReadResource(uri string) (json.RawMessage, error) {
	return c.Call("resources/read", map[string]interface{}{"uri": uri})
}

// CallTool invokes an MCP tool by name.
func (c *Client) CallTool(name string, arguments map[string]interface{}) (json.RawMessage, error) {
	params := map[string]interface{}{
		"name":      name,
		"arguments": arguments,
	}
	return c.Call("tools/call", params)
}

// Call sends a JSON-RPC request and returns the raw result payload.
func (c *Client) Call(method string, params interface{}) (json.RawMessage, error) {
	id := atomic.AddInt64(&c.nextID, 1)
	reqBody, err := json.Marshal(Request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal MCP request: %w", err)
	}

	req, err := http.NewRequest("POST", c.url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("MCP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read MCP response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("MCP HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	responseBody := body
	if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream") {
		responseBody, err = firstSSEData(body)
		if err != nil {
			return nil, err
		}
	}

	var rpcResp Response
	if err := json.Unmarshal(responseBody, &rpcResp); err != nil {
		return nil, fmt.Errorf("invalid MCP response: %w", err)
	}
	if rpcResp.Error != nil {
		return nil, rpcResp.Error
	}
	return rpcResp.Result, nil
}

func firstSSEData(body []byte) ([]byte, error) {
	scanner := bufio.NewScanner(bytes.NewReader(body))
	var data strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data:") {
			if data.Len() > 0 {
				data.WriteByte('\n')
			}
			data.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan MCP event stream: %w", err)
	}
	if data.Len() == 0 {
		return nil, fmt.Errorf("MCP event stream did not include data")
	}
	return []byte(data.String()), nil
}
