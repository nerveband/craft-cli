package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/ashrafali/craft-cli/internal/models"
)

const (
	defaultTimeout = 30 * time.Second
)

// Client represents the Craft API client
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new API client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// doRequest performs an HTTP request and handles errors
func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	reqURL := fmt.Sprintf("%s%s", c.baseURL, path)
	req, err := http.NewRequest(method, reqURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, c.handleErrorResponse(resp.StatusCode, respBody)
	}

	return respBody, nil
}

// handleErrorResponse converts HTTP errors to user-friendly messages
func (c *Client) handleErrorResponse(statusCode int, body []byte) error {
	var errResp models.ErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		// If we can't parse the error response, return a generic message
		return fmt.Errorf("API error (status %d): %s", statusCode, string(body))
	}

	switch statusCode {
	case 401, 403:
		return fmt.Errorf("authentication failed. Check API URL")
	case 404:
		return fmt.Errorf("resource not found")
	case 429:
		return fmt.Errorf("rate limit exceeded. Retry later")
	case 500, 502, 503, 504:
		return fmt.Errorf("Craft API error: %s", errResp.Message)
	default:
		return fmt.Errorf("API error (%d): %s", statusCode, errResp.Message)
	}
}

// GetDocuments retrieves all documents
func (c *Client) GetDocuments() (*models.DocumentList, error) {
	data, err := c.doRequest("GET", "/documents", nil)
	if err != nil {
		return nil, err
	}

	var result models.DocumentList
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &result, nil
}

// GetDocument retrieves a single document by ID
func (c *Client) GetDocument(id string) (*models.Document, error) {
	path := fmt.Sprintf("/documents/%s", id)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var doc models.Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &doc, nil
}

// SearchDocuments searches for documents matching a query
func (c *Client) SearchDocuments(query string) (*models.SearchResult, error) {
	path := fmt.Sprintf("/documents/search?query=%s", url.QueryEscape(query))
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result models.SearchResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &result, nil
}

// CreateDocument creates a new document
func (c *Client) CreateDocument(req *models.CreateDocumentRequest) (*models.Document, error) {
	data, err := c.doRequest("POST", "/documents", req)
	if err != nil {
		return nil, err
	}

	var doc models.Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &doc, nil
}

// UpdateDocument updates an existing document
func (c *Client) UpdateDocument(id string, req *models.UpdateDocumentRequest) (*models.Document, error) {
	path := fmt.Sprintf("/documents/%s", id)
	data, err := c.doRequest("PUT", path, req)
	if err != nil {
		return nil, err
	}

	var doc models.Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &doc, nil
}

// DeleteDocument deletes a document by ID
func (c *Client) DeleteDocument(id string) error {
	path := fmt.Sprintf("/documents/%s", id)
	_, err := c.doRequest("DELETE", path, nil)
	return err
}
