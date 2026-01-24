package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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

// GetDocument retrieves a single document by ID using the blocks endpoint
func (c *Client) GetDocument(id string) (*models.Document, error) {
	path := fmt.Sprintf("/blocks?id=%s", url.QueryEscape(id))
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var blocksResp models.BlocksResponse
	if err := json.Unmarshal(data, &blocksResp); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	// Combine markdown from all blocks
	markdown := combineBlocksMarkdown(blocksResp)

	doc := &models.Document{
		ID:       blocksResp.ID,
		Title:    blocksResp.Markdown,
		Markdown: markdown,
	}

	return doc, nil
}

// combineBlocksMarkdown extracts and combines markdown from all blocks
func combineBlocksMarkdown(resp models.BlocksResponse) string {
	var parts []string

	// Add the document title/header
	if resp.Markdown != "" {
		parts = append(parts, "# "+resp.Markdown)
	}

	// Recursively collect markdown from all content blocks
	for _, block := range resp.Content {
		collectBlockMarkdown(&block, &parts)
	}

	return strings.Join(parts, "\n\n")
}

// collectBlockMarkdown recursively collects markdown from a block and its children
func collectBlockMarkdown(block *models.Block, parts *[]string) {
	if block.Markdown != "" {
		*parts = append(*parts, block.Markdown)
	}
	for _, child := range block.Content {
		collectBlockMarkdown(&child, parts)
	}
}

// SearchDocuments searches for documents matching a query
func (c *Client) SearchDocuments(query string) (*models.SearchResult, error) {
	// Craft API uses 'include' parameter instead of 'query'
	path := fmt.Sprintf("/documents/search?include=%s", url.QueryEscape(query))
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

// createDocumentsRequest wraps documents for the API
type createDocumentsRequest struct {
	Documents []models.CreateDocumentRequest `json:"documents"`
}

// createDocumentsResponse represents the API response for document creation
type createDocumentsResponse struct {
	Items []struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	} `json:"items"`
}

// CreateDocument creates a new document
func (c *Client) CreateDocument(req *models.CreateDocumentRequest) (*models.Document, error) {
	// Craft API expects {"documents": [...]} wrapper
	wrapper := createDocumentsRequest{
		Documents: []models.CreateDocumentRequest{*req},
	}

	data, err := c.doRequest("POST", "/documents", wrapper)
	if err != nil {
		return nil, err
	}

	var resp createDocumentsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("no document returned from API")
	}

	doc := &models.Document{
		ID:    resp.Items[0].ID,
		Title: resp.Items[0].Title,
	}

	return doc, nil
}

// blockPosition specifies where to insert a block
type blockPosition struct {
	PageID   string `json:"pageId"`
	Position string `json:"position"` // "start", "end", or block ID
}

// addBlockRequest is the request body for adding blocks
type addBlockRequest struct {
	Markdown string        `json:"markdown"`
	Position blockPosition `json:"position"`
}

// addBlockResponse is the response from adding blocks
type addBlockResponse struct {
	Items []struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		Markdown string `json:"markdown"`
	} `json:"items"`
}

// UpdateDocument updates an existing document by adding content
// Note: The Craft Connect API only supports adding content blocks, not updating title or replacing content
func (c *Client) UpdateDocument(id string, req *models.UpdateDocumentRequest) (*models.Document, error) {
	// Title updates are not supported via the API
	if req.Title != "" && req.Markdown == "" {
		return nil, fmt.Errorf("the Craft Connect API does not support title updates. Use the Craft app to rename documents")
	}

	if req.Markdown == "" {
		return nil, fmt.Errorf("markdown content is required for updates")
	}

	// Add content blocks to the document
	addReq := addBlockRequest{
		Markdown: req.Markdown,
		Position: blockPosition{
			PageID:   id,
			Position: "end",
		},
	}

	data, err := c.doRequest("POST", "/blocks", addReq)
	if err != nil {
		return nil, err
	}

	var resp addBlockResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	// Return a document with the update info
	doc := &models.Document{
		ID: id,
	}

	// If title was also requested, note it in return
	if req.Title != "" {
		doc.Title = req.Title + " (title not updated - API limitation)"
	}

	// Set markdown to confirm what was added
	if len(resp.Items) > 0 {
		doc.Markdown = resp.Items[0].Markdown
	}

	return doc, nil
}

// deleteBlocksRequest is the request body for deleting blocks
type deleteBlocksRequest struct {
	BlockIDs []string `json:"blockIds"`
}

// deleteBlocksResponse is the response from deleting blocks
type deleteBlocksResponse struct {
	Items []struct {
		ID string `json:"id"`
	} `json:"items"`
}

// DeleteDocument deletes a document by ID
// Note: The Craft Connect API does not support deleting root page blocks (documents)
// This will attempt to delete all content blocks within the document
func (c *Client) DeleteDocument(id string) error {
	// First, get the document to find all its content blocks
	doc, err := c.GetDocument(id)
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}

	// Get the blocks again to get block IDs
	path := fmt.Sprintf("/blocks?id=%s", url.QueryEscape(id))
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return fmt.Errorf("failed to get document blocks: %w", err)
	}

	var blocksResp models.BlocksResponse
	if err := json.Unmarshal(data, &blocksResp); err != nil {
		return fmt.Errorf("invalid response from API: %w", err)
	}

	// Collect all content block IDs (not the root page)
	var blockIDs []string
	for _, block := range blocksResp.Content {
		collectBlockIDs(&block, &blockIDs)
	}

	if len(blockIDs) == 0 {
		return fmt.Errorf("the Craft Connect API cannot delete documents (only content blocks). Document '%s' has no deletable content blocks", doc.Title)
	}

	// Delete the content blocks
	deleteReq := deleteBlocksRequest{
		BlockIDs: blockIDs,
	}

	_, err = c.doRequest("DELETE", "/blocks", deleteReq)
	if err != nil {
		return fmt.Errorf("failed to delete blocks: %w", err)
	}

	return nil
}

// collectBlockIDs recursively collects all block IDs
func collectBlockIDs(block *models.Block, ids *[]string) {
	if block.ID != "" {
		*ids = append(*ids, block.ID)
	}
	for _, child := range block.Content {
		collectBlockIDs(&child, ids)
	}
}

// DeleteBlock deletes a specific block by ID
func (c *Client) DeleteBlock(blockID string) error {
	deleteReq := deleteBlocksRequest{
		BlockIDs: []string{blockID},
	}

	_, err := c.doRequest("DELETE", "/blocks", deleteReq)
	if err != nil {
		return fmt.Errorf("failed to delete block: %w", err)
	}

	return nil
}
