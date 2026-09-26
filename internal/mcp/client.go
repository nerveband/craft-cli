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
	url             string
	httpClient      *http.Client
	nextID          int64
	sessionID       string
	protocolVersion string
	initialized     json.RawMessage
	ClientVersion   string
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

type HTTPError struct {
	Status  int
	Message string
	Headers map[string]string
}

func (e *HTTPError) Error() string { return fmt.Sprintf("MCP HTTP %d: %s", e.Status, e.Message) }

// Error is a JSON-RPC error payload.
type Error struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// ToolError preserves an MCP application failure, including its original result.
type ToolError struct {
	Result  json.RawMessage
	Message string
}

func (e *ToolError) Error() string { return "MCP tool failed: " + e.Message }

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("MCP error %d: %s", e.Code, e.Message)
}

// Initialize performs the MCP initialize request.
func (c *Client) Initialize() (json.RawMessage, error) {
	if c.initialized != nil {
		return c.initialized, nil
	}
	clientVersion := c.ClientVersion
	if clientVersion == "" {
		clientVersion = "development"
	}
	params := map[string]interface{}{
		"protocolVersion": "2025-06-18",
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]string{
			"name":    "craft-cli",
			"version": clientVersion,
		},
	}
	result, err := c.Call("initialize", params)
	if err != nil {
		return nil, err
	}
	var negotiation struct {
		ProtocolVersion string `json:"protocolVersion"`
	}
	if err := json.Unmarshal(result, &negotiation); err != nil {
		return nil, err
	}
	if negotiation.ProtocolVersion != "" {
		c.protocolVersion = negotiation.ProtocolVersion
	}
	if err := c.notifyInitialized(); err != nil {
		return nil, err
	}
	c.initialized = result
	return result, nil
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
	result, err := c.Call("tools/call", params)
	if err != nil {
		return nil, err
	}
	var outcome struct {
		IsError bool `json:"isError"`
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(result, &outcome); err != nil {
		return nil, fmt.Errorf("invalid MCP tool result: %w", err)
	}
	if outcome.IsError {
		var messages []string
		for _, item := range outcome.Content {
			if item.Text != "" {
				messages = append(messages, item.Text)
			}
		}
		message := strings.Join(messages, "\n")
		if message == "" {
			message = "server returned isError=true"
		}
		return result, &ToolError{Result: result, Message: message}
	}
	return result, nil
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
	protocol := c.protocolVersion
	if protocol == "" {
		protocol = "2025-06-18"
	}
	req.Header.Set("MCP-Protocol-Version", protocol)
	if c.sessionID != "" {
		req.Header.Set("Mcp-Session-Id", c.sessionID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("MCP request failed: %w", err)
	}
	defer resp.Body.Close()
	if session := resp.Header.Get("Mcp-Session-Id"); session != "" {
		c.sessionID = session
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read MCP response: %w", err)
	}
	if resp.StatusCode >= 400 {
		headers := map[string]string{}
		for k, v := range resp.Header {
			lower := strings.ToLower(k)
			if lower == "retry-after" || strings.Contains(lower, "ratelimit") || strings.Contains(lower, "rate-limit") {
				headers[k] = strings.Join(v, ", ")
			}
		}
		return nil, &HTTPError{Status: resp.StatusCode, Message: strings.TrimSpace(string(body)), Headers: headers}
	}

	responseBody := body
	if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream") {
		responseBody, err = matchingSSEData(body, id)
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
	if rpcResp.ID != 0 && rpcResp.ID != id {
		return nil, fmt.Errorf("MCP response ID %d does not match request %d", rpcResp.ID, id)
	}
	return rpcResp.Result, nil
}

func matchingSSEData(body []byte, id int64) ([]byte, error) {
	// Each SSE event is a separate JSON-RPC message. Ignore notifications.
	scanner := bufio.NewScanner(bytes.NewReader(body))
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	var lines []string
	match := func() []byte {
		data := []byte(strings.Join(lines, "\n"))
		var response Response
		if json.Unmarshal(data, &response) == nil && response.ID == id {
			return data
		}
		return nil
	}
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if data := match(); data != nil {
				return data, nil
			}
			lines = nil
		} else if strings.HasPrefix(line, "data:") {
			lines = append(lines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if data := match(); data != nil {
		return data, nil
	}
	return nil, fmt.Errorf("MCP event stream did not include response %d", id)
}

func firstSSEData(body []byte) ([]byte, error) {
	scanner := bufio.NewScanner(bytes.NewReader(body))
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
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

func (c *Client) SetTimeout(timeout time.Duration) {
	if timeout > 0 {
		c.httpClient.Timeout = timeout
	}
}

func (c *Client) notifyInitialized() error {
	req, err := http.NewRequest("POST", c.url, strings.NewReader(`{"jsonrpc":"2.0","method":"notifications/initialized"}`))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	protocol := c.protocolVersion
	if protocol == "" {
		protocol = "2025-06-18"
	}
	req.Header.Set("MCP-Protocol-Version", protocol)
	if c.sessionID != "" {
		req.Header.Set("Mcp-Session-Id", c.sessionID)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("MCP initialized notification: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("MCP initialized notification HTTP %d", resp.StatusCode)
	}
	return nil
}
