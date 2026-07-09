package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ashrafali/craft-cli/internal/models"
)

const (
	defaultTimeout = 30 * time.Second
	// defaultInsertChunkBytes is a conservative default to avoid Craft API payload limits.
	defaultInsertChunkBytes = 30000
)

// APIError represents an error response from Craft.
// It preserves status code for machine handling while keeping the human message concise.
type APIError struct {
	StatusCode int
	Err        string
	Message    string
	RawBody    string
}

func (e *APIError) Error() string {
	msg := strings.TrimSpace(e.Message)
	if msg == "" {
		msg = strings.TrimSpace(e.Err)
	}
	if msg == "" {
		msg = strings.TrimSpace(e.RawBody)
	}
	if msg == "" {
		msg = "unknown error"
	}
	return msg
}

// Client represents the Craft API client
type Client struct {
	baseURL    string
	apiKey     string
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

// NewClientWithKey creates a new API client with an API key
func NewClientWithKey(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
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

	// Add API key authentication if configured
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
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
		return &APIError{
			StatusCode: statusCode,
			RawBody:    string(body),
			Message:    fmt.Sprintf("API error (status %d): %s", statusCode, string(body)),
		}
	}

	// Some Craft errors use `error`, some use `message`.
	msg := errResp.Message
	if msg == "" {
		msg = errResp.Error
	}
	if msg == "" {
		msg = string(body)
	}

	// Provide helpful context for permission-related errors
	switch statusCode {
	case 401:
		if c.apiKey != "" {
			msg = "authentication failed: invalid or expired API key"
		} else {
			msg = "authentication required. Use --api-key or configure a profile with an API key"
		}
	case 403:
		// Check for specific permission messages in the response
		lowerMsg := strings.ToLower(msg)
		if strings.Contains(lowerMsg, "read") {
			msg = "permission denied: this API key does not have read access"
		} else if strings.Contains(lowerMsg, "write") || strings.Contains(lowerMsg, "create") || strings.Contains(lowerMsg, "update") {
			msg = "permission denied: this API key does not have write access (read-only)"
		} else if strings.Contains(lowerMsg, "delete") {
			msg = "permission denied: this API key does not have delete access"
		} else {
			msg = "permission denied: " + msg
		}
	case 404:
		msg = "resource not found"
	case 429:
		msg = "rate limit exceeded. Retry later"
	case 500, 502, 503, 504:
		msg = "Craft API error: " + msg
	}

	return &APIError{
		StatusCode: statusCode,
		Err:        errResp.Error,
		Message:    msg,
		RawBody:    string(body),
	}
}

// GetDocuments retrieves all documents
func (c *Client) GetDocuments() (*models.DocumentList, error) {
	return c.GetDocumentsFiltered("", "")
}

// GetDocumentsFiltered retrieves documents with optional folder or location filter
func (c *Client) GetDocumentsFiltered(folderID, location string) (*models.DocumentList, error) {
	path := "/documents"

	var params []string
	if folderID != "" {
		params = append(params, "folderId="+url.QueryEscape(folderID))
	}
	if location != "" {
		params = append(params, "location="+url.QueryEscape(location))
	}

	if len(params) > 0 {
		path = path + "?" + strings.Join(params, "&")
	}

	data, err := c.doRequest("GET", path, nil)
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

	// Combine markdown from all blocks (include title as H1 for readability)
	markdown := CombineBlocksMarkdown(blocksResp, true)

	doc := &models.Document{
		ID:       blocksResp.ID,
		Title:    blocksResp.Markdown,
		Markdown: markdown,
	}

	return doc, nil
}

// GetDocumentContentMarkdown returns only the document content markdown (excluding the title/header).
func (c *Client) GetDocumentContentMarkdown(id string) (string, error) {
	blocksResp, err := c.GetDocumentBlocks(id)
	if err != nil {
		return "", err
	}
	return CombineBlocksMarkdown(blocksResp, false), nil
}

// GetDocumentBlocks retrieves the raw blocks response for a document.
func (c *Client) GetDocumentBlocks(id string) (models.BlocksResponse, error) {
	return c.GetDocumentBlocksWithDepth(id, -1)
}

// GetDocumentBlocksWithDepth retrieves blocks with optional depth control.
func (c *Client) GetDocumentBlocksWithDepth(id string, maxDepth int) (models.BlocksResponse, error) {
	params := url.Values{}
	params.Set("id", id)
	if maxDepth != -1 {
		params.Set("maxDepth", strconv.Itoa(maxDepth))
	}
	path := "/blocks?" + params.Encode()

	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return models.BlocksResponse{}, err
	}

	var blocksResp models.BlocksResponse
	if err := json.Unmarshal(data, &blocksResp); err != nil {
		return models.BlocksResponse{}, fmt.Errorf("invalid response from API: %w", err)
	}

	return blocksResp, nil
}

// CombineBlocksMarkdown extracts and combines markdown from all blocks.
func CombineBlocksMarkdown(resp models.BlocksResponse, includeTitle bool) string {
	var parts []string

	// Add the document title/header
	if includeTitle && resp.Markdown != "" {
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
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}

	// The Space API does not reliably accept content in POST /documents.
	// To keep behavior consistent (and avoid duplicates if the API changes), we always insert
	// content via POST /blocks after document creation.
	createReq := *req
	createReq.Markdown = ""
	createReq.Content = ""

	// Craft API expects {"documents": [...]} wrapper
	wrapper := createDocumentsRequest{
		Documents: []models.CreateDocumentRequest{createReq},
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

	content := req.Markdown
	if strings.TrimSpace(content) == "" {
		content = req.Content
	}
	if strings.TrimSpace(content) != "" {
		_, err := c.AppendMarkdown(doc.ID, content, defaultInsertChunkBytes)
		if err != nil {
			return nil, err
		}
	}

	return doc, nil
}

// CreateDocumentsRaw creates one or more documents using the documented REST payload shape.
func (c *Client) CreateDocumentsRaw(req map[string]interface{}) ([]models.Document, error) {
	data, err := c.doRequest("POST", "/documents", req)
	if err != nil {
		return nil, err
	}

	var resp createDocumentsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	docs := make([]models.Document, 0, len(resp.Items))
	for _, item := range resp.Items {
		docs = append(docs, models.Document{ID: item.ID, Title: item.Title})
	}
	return docs, nil
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

type addMarkdownAtPositionRequest struct {
	Markdown string                 `json:"markdown"`
	Position map[string]interface{} `json:"position"`
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
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}

	// Title updates are supported by updating the root page block via PUT /blocks.
	if req.Title != "" {
		if err := c.UpdateBlockMarkdown(id, req.Title); err != nil {
			return nil, err
		}
	}

	if strings.TrimSpace(req.Markdown) == "" {
		// Title-only update is allowed.
		return &models.Document{ID: id, Title: req.Title}, nil
	}

	lastInserted, err := c.AppendMarkdown(id, req.Markdown, defaultInsertChunkBytes)
	if err != nil {
		return nil, err
	}

	return &models.Document{ID: id, Title: req.Title, Markdown: lastInserted}, nil
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

// DeleteDocument soft-deletes a document by moving it to trash.
func (c *Client) DeleteDocument(id string) error {
	req := struct {
		DocumentIDs []string `json:"documentIds"`
	}{
		DocumentIDs: []string{id},
	}

	_, err := c.doRequest("DELETE", "/documents", req)
	return err
}

// ClearDocumentContent deletes all content blocks within a document (does not delete the document itself).
func (c *Client) ClearDocumentContent(id string) (int, error) {
	blocksResp, err := c.GetDocumentBlocks(id)
	if err != nil {
		return 0, fmt.Errorf("failed to get document blocks: %w", err)
	}

	var blockIDs []string
	for _, block := range blocksResp.Content {
		collectBlockIDs(&block, &blockIDs)
	}

	if len(blockIDs) == 0 {
		return 0, nil
	}

	deleteReq := deleteBlocksRequest{BlockIDs: blockIDs}
	_, err = c.doRequest("DELETE", "/blocks", deleteReq)
	if err != nil {
		return 0, fmt.Errorf("failed to delete blocks: %w", err)
	}

	return len(blockIDs), nil
}

// DeleteBlocks deletes specific blocks by ID.
func (c *Client) DeleteBlocks(blockIDs []string) error {
	if len(blockIDs) == 0 {
		return nil
	}
	deleteReq := deleteBlocksRequest{BlockIDs: blockIDs}
	_, err := c.doRequest("DELETE", "/blocks", deleteReq)
	if err != nil {
		return fmt.Errorf("failed to delete blocks: %w", err)
	}
	return nil
}

// UpdateBlockMarkdown updates a block (including the document root page) using PUT /blocks.
func (c *Client) UpdateBlockMarkdown(blockID, markdown string) error {
	req := struct {
		Blocks []struct {
			ID       string `json:"id"`
			Markdown string `json:"markdown"`
		} `json:"blocks"`
	}{
		Blocks: []struct {
			ID       string `json:"id"`
			Markdown string `json:"markdown"`
		}{{ID: blockID, Markdown: markdown}},
	}

	_, err := c.doRequest("PUT", "/blocks", req)
	return err
}

// AppendMarkdown appends markdown to a document by inserting blocks at the end.
// It automatically chunks large markdown to avoid API payload limits.
func (c *Client) AppendMarkdown(docID, markdown string, chunkBytes int) (string, error) {
	if strings.TrimSpace(markdown) == "" {
		return "", nil
	}
	if chunkBytes <= 0 {
		chunkBytes = defaultInsertChunkBytes
	}

	chunks := SplitMarkdownIntoChunks(markdown, chunkBytes)
	var last string
	for _, chunk := range chunks {
		if strings.TrimSpace(chunk) == "" {
			continue
		}
		addReq := addBlockRequest{
			Markdown: chunk,
			Position: blockPosition{PageID: docID, Position: "end"},
		}

		data, err := c.doRequest("POST", "/blocks", addReq)
		if err != nil {
			return "", err
		}

		var resp addBlockResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return "", fmt.Errorf("invalid response from API: %w", err)
		}

		if len(resp.Items) > 0 {
			last = resp.Items[len(resp.Items)-1].Markdown
		}
	}

	return last, nil
}

// AddMarkdownAtPosition inserts markdown at a page/date/sibling position.
// Position accepts pageId+start/end, date+start/end, or siblingId+before/after.
func (c *Client) AddMarkdownAtPosition(markdown string, position map[string]interface{}, chunkBytes int) (string, error) {
	if strings.TrimSpace(markdown) == "" {
		return "", nil
	}
	if chunkBytes <= 0 {
		chunkBytes = defaultInsertChunkBytes
	}
	if len(position) == 0 {
		return "", fmt.Errorf("position is required")
	}

	chunks := SplitMarkdownIntoChunks(markdown, chunkBytes)
	var last string
	for _, chunk := range chunks {
		if strings.TrimSpace(chunk) == "" {
			continue
		}
		req := addMarkdownAtPositionRequest{
			Markdown: chunk,
			Position: position,
		}
		data, err := c.doRequest("POST", "/blocks", req)
		if err != nil {
			return "", err
		}
		var resp addBlockResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return "", fmt.Errorf("invalid response from API: %w", err)
		}
		if len(resp.Items) > 0 {
			last = resp.Items[len(resp.Items)-1].Markdown
		}
	}
	return last, nil
}

// ReplaceDocumentContent replaces a document's content by clearing existing blocks and inserting the new markdown.
func (c *Client) ReplaceDocumentContent(docID, markdown string, chunkBytes int) error {
	if strings.TrimSpace(markdown) == "" {
		return fmt.Errorf("markdown content is required")
	}
	_, err := c.ClearDocumentContent(docID)
	if err != nil {
		return err
	}
	_, err = c.AppendMarkdown(docID, markdown, chunkBytes)
	return err
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

// ========== Folder Operations ==========

// GetFolders retrieves all folders
func (c *Client) GetFolders() (*models.FolderList, error) {
	data, err := c.doRequest("GET", "/folders", nil)
	if err != nil {
		return nil, err
	}

	var result models.FolderList
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &result, nil
}

// createFolderRequest wraps the create folder request
type createFolderRequest struct {
	Folders []struct {
		Name     string `json:"name"`
		ParentID string `json:"parentId,omitempty"`
	} `json:"folders"`
}

// createFolderResponse represents the response from creating a folder
type createFolderResponse struct {
	Items []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"items"`
}

// CreateFolder creates a new folder
func (c *Client) CreateFolder(name string, parentID string) (*models.Folder, error) {
	req := createFolderRequest{
		Folders: []struct {
			Name     string `json:"name"`
			ParentID string `json:"parentId,omitempty"`
		}{{Name: name, ParentID: parentID}},
	}

	data, err := c.doRequest("POST", "/folders", req)
	if err != nil {
		return nil, err
	}

	var resp createFolderResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("no folder returned from API")
	}

	return &models.Folder{
		ID:       resp.Items[0].ID,
		Name:     resp.Items[0].Name,
		ParentID: parentID,
	}, nil
}

// moveFolderRequest wraps the move folder request
type moveFolderRequest struct {
	FolderIDs   []string    `json:"folderIds"`
	Destination interface{} `json:"destination"`
}

// MoveFolder moves a folder to a new parent
func (c *Client) MoveFolder(folderID, targetParentID string) error {
	req := moveFolderRequest{
		FolderIDs: []string{folderID},
	}
	if targetParentID == "" {
		req.Destination = "root"
	} else {
		req.Destination = map[string]string{"parentFolderId": targetParentID}
	}

	_, err := c.doRequest("PUT", "/folders/move", req)
	return err
}

// deleteFolderRequest wraps the delete folder request
type deleteFolderRequest struct {
	FolderIDs []string `json:"folderIds"`
}

// DeleteFolder deletes a folder
func (c *Client) DeleteFolder(folderID string) error {
	req := deleteFolderRequest{
		FolderIDs: []string{folderID},
	}

	_, err := c.doRequest("DELETE", "/folders", req)
	return err
}

// ========== Document Move Operations ==========

// moveDocumentRequest wraps the move document request
type moveDocumentRequest struct {
	DocumentIDs []string    `json:"documentIds"`
	Destination interface{} `json:"destination"`
}

// MoveDocument moves a document to a folder or location
func (c *Client) MoveDocument(docID, folderID, location string) error {
	req := moveDocumentRequest{
		DocumentIDs: []string{docID},
	}
	if folderID != "" {
		req.Destination = map[string]string{"folderId": folderID}
	} else {
		req.Destination = map[string]string{"destination": location}
	}

	_, err := c.doRequest("PUT", "/documents/move", req)
	return err
}

// ========== Block Operations (Enhanced) ==========

// GetBlock retrieves a specific block by ID with optional depth
func (c *Client) GetBlock(blockID string) (*models.Block, error) {
	path := fmt.Sprintf("/blocks?id=%s", url.QueryEscape(blockID))
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var block models.Block
	if err := json.Unmarshal(data, &block); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &block, nil
}

// addBlockExtendedRequest is the request body for adding blocks with more control
type addBlockExtendedRequest struct {
	Markdown string `json:"markdown"`
	Position struct {
		PageID    string `json:"pageId,omitempty"`
		Position  string `json:"position,omitempty"` // start, end, or block ID
		SiblingID string `json:"siblingId,omitempty"`
		Relative  string `json:"relative,omitempty"` // before, after
	} `json:"position"`
}

// AddBlock adds a block with position control
func (c *Client) AddBlock(pageID, markdown, position string) (*models.Block, error) {
	req := addBlockExtendedRequest{
		Markdown: markdown,
	}
	req.Position.PageID = pageID
	req.Position.Position = position

	data, err := c.doRequest("POST", "/blocks", req)
	if err != nil {
		return nil, err
	}

	var resp addBlockResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("no block returned from API")
	}

	return &models.Block{
		ID:       resp.Items[0].ID,
		Type:     resp.Items[0].Type,
		Markdown: resp.Items[0].Markdown,
	}, nil
}

// AddBlockRelative adds a block relative to a sibling
func (c *Client) AddBlockRelative(siblingID, markdown, relative string) (*models.Block, error) {
	req := addBlockExtendedRequest{
		Markdown: markdown,
	}
	req.Position.SiblingID = siblingID
	req.Position.Relative = relative

	data, err := c.doRequest("POST", "/blocks", req)
	if err != nil {
		return nil, err
	}

	var resp addBlockResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("no block returned from API")
	}

	return &models.Block{
		ID:       resp.Items[0].ID,
		Type:     resp.Items[0].Type,
		Markdown: resp.Items[0].Markdown,
	}, nil
}

// moveBlockRequest is the request body for moving blocks
type moveBlockRequest struct {
	Blocks []struct {
		ID       string `json:"id"`
		Position struct {
			PageID   string `json:"pageId,omitempty"`
			Position string `json:"position,omitempty"`
		} `json:"position"`
	} `json:"blocks"`
}

// MoveBlock moves a block to a new position
func (c *Client) MoveBlock(blockID, targetPageID, position string) error {
	req := moveBlockRequest{
		Blocks: []struct {
			ID       string `json:"id"`
			Position struct {
				PageID   string `json:"pageId,omitempty"`
				Position string `json:"position,omitempty"`
			} `json:"position"`
		}{{
			ID: blockID,
			Position: struct {
				PageID   string `json:"pageId,omitempty"`
				Position string `json:"position,omitempty"`
			}{
				PageID:   targetPageID,
				Position: position,
			},
		}},
	}

	_, err := c.doRequest("PUT", "/blocks", req)
	return err
}

// ========== Task Operations ==========

// GetTasks retrieves tasks with optional filters
func (c *Client) GetTasks(scope string) (*models.TaskList, error) {
	path := "/tasks"
	if scope != "" {
		path = fmt.Sprintf("/tasks?scope=%s", url.QueryEscape(scope))
	}

	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result models.TaskList
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &result, nil
}

// GetDocumentTasks retrieves tasks for a specific document
func (c *Client) GetDocumentTasks(docID string) (*models.TaskList, error) {
	path := fmt.Sprintf("/tasks?documentId=%s", url.QueryEscape(docID))

	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result models.TaskList
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &result, nil
}

// addTaskRequest is the request body for adding tasks
type addTaskRequest struct {
	Tasks []struct {
		Markdown string                 `json:"markdown"`
		TaskInfo map[string]interface{} `json:"taskInfo,omitempty"`
		Location map[string]string      `json:"location,omitempty"`
	} `json:"tasks"`
}

// addTaskResponse is the response from adding tasks
type addTaskResponse struct {
	Items []struct {
		ID         string           `json:"id"`
		BlockID    string           `json:"blockId"`
		DocumentID string           `json:"documentId"`
		Markdown   string           `json:"markdown"`
		State      string           `json:"state"`
		TaskInfo   *models.TaskInfo `json:"taskInfo,omitempty"`
	} `json:"items"`
}

// AddTask creates a new task
func (c *Client) AddTask(markdown, location, docID, scheduleDate, deadlineDate string) (*models.Task, error) {
	taskInfo := map[string]interface{}{}
	if scheduleDate != "" {
		taskInfo["scheduleDate"] = scheduleDate
	}
	if deadlineDate != "" {
		taskInfo["deadlineDate"] = deadlineDate
	}
	if len(taskInfo) == 0 {
		taskInfo = nil
	}

	locationObj := map[string]string{"type": location}
	if location == "document" && docID != "" {
		locationObj["documentId"] = docID
	}

	req := addTaskRequest{
		Tasks: []struct {
			Markdown string                 `json:"markdown"`
			TaskInfo map[string]interface{} `json:"taskInfo,omitempty"`
			Location map[string]string      `json:"location,omitempty"`
		}{{
			Markdown: markdown,
			TaskInfo: taskInfo,
			Location: locationObj,
		}},
	}

	data, err := c.doRequest("POST", "/tasks", req)
	if err != nil {
		return nil, err
	}

	var resp addTaskResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("no task returned from API")
	}

	item := resp.Items[0]
	state := item.State
	schedule := ""
	deadline := ""
	completedAt := ""
	canceledAt := ""
	if item.TaskInfo != nil {
		if item.TaskInfo.State != "" {
			state = item.TaskInfo.State
		}
		schedule = item.TaskInfo.ScheduleDate
		deadline = item.TaskInfo.DeadlineDate
		completedAt = item.TaskInfo.CompletedAt
		canceledAt = item.TaskInfo.CanceledAt
	}
	return &models.Task{
		ID:           item.ID,
		BlockID:      item.BlockID,
		DocumentID:   item.DocumentID,
		Markdown:     item.Markdown,
		State:        state,
		ScheduleDate: schedule,
		DeadlineDate: deadline,
		CompletedAt:  completedAt,
		CanceledAt:   canceledAt,
	}, nil
}

// AddTasksRaw creates tasks using the documented REST payload shape.
func (c *Client) AddTasksRaw(req map[string]interface{}) ([]models.Task, error) {
	data, err := c.doRequest("POST", "/tasks", req)
	if err != nil {
		return nil, err
	}

	var resp addTaskResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	tasks := make([]models.Task, 0, len(resp.Items))
	for _, item := range resp.Items {
		task := models.Task{
			ID:         item.ID,
			BlockID:    item.BlockID,
			DocumentID: item.DocumentID,
			Markdown:   item.Markdown,
			State:      item.State,
		}
		if item.TaskInfo != nil {
			if item.TaskInfo.State != "" {
				task.State = item.TaskInfo.State
			}
			task.ScheduleDate = item.TaskInfo.ScheduleDate
			task.DeadlineDate = item.TaskInfo.DeadlineDate
			task.CompletedAt = item.TaskInfo.CompletedAt
			task.CanceledAt = item.TaskInfo.CanceledAt
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// updateTaskRequest is the request body for updating tasks
type updateTaskRequest struct {
	TasksToUpdate []struct {
		ID       string                 `json:"id"`
		TaskInfo map[string]interface{} `json:"taskInfo,omitempty"`
	} `json:"tasksToUpdate"`
}

// UpdateTask updates a task's state or dates
func (c *Client) UpdateTask(taskID, state, scheduleDate, deadlineDate string) error {
	taskInfo := map[string]interface{}{}
	if state != "" {
		taskInfo["state"] = state
	}
	if scheduleDate != "" {
		taskInfo["scheduleDate"] = scheduleDate
	}
	if deadlineDate != "" {
		taskInfo["deadlineDate"] = deadlineDate
	}

	req := updateTaskRequest{
		TasksToUpdate: []struct {
			ID       string                 `json:"id"`
			TaskInfo map[string]interface{} `json:"taskInfo,omitempty"`
		}{{
			ID:       taskID,
			TaskInfo: taskInfo,
		}},
	}

	_, err := c.doRequest("PUT", "/tasks", req)
	return err
}

// UpdateTasksRaw updates tasks using the documented REST payload shape.
func (c *Client) UpdateTasksRaw(req map[string]interface{}) error {
	_, err := c.doRequest("PUT", "/tasks", req)
	return err
}

// deleteTaskRequest is the request body for deleting tasks
type deleteTaskRequest struct {
	IDsToDelete []string `json:"idsToDelete"`
}

// DeleteTask deletes a task
func (c *Client) DeleteTask(taskID string) error {
	req := deleteTaskRequest{
		IDsToDelete: []string{taskID},
	}

	_, err := c.doRequest("DELETE", "/tasks", req)
	return err
}

// DeleteTasksRaw deletes tasks using the documented REST payload shape.
func (c *Client) DeleteTasksRaw(req map[string]interface{}) error {
	_, err := c.doRequest("DELETE", "/tasks", req)
	return err
}

// ========== Collection Operations ==========

// GetCollections retrieves all collections, optionally filtered by document IDs
func (c *Client) GetCollections(documentIDs string) (*models.CollectionList, error) {
	path := "/collections"
	if documentIDs != "" {
		path = fmt.Sprintf("/collections?documentIds=%s", url.QueryEscape(documentIDs))
	}

	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result models.CollectionList
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &result, nil
}

// GetCollectionSchema retrieves the schema for a collection
func (c *Client) GetCollectionSchema(collectionID, format string) (*models.CollectionSchema, error) {
	path := fmt.Sprintf("/collections/%s/schema", url.PathEscape(collectionID))
	if format != "" {
		path = fmt.Sprintf("%s?format=%s", path, url.QueryEscape(format))
	}

	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result models.CollectionSchema
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &result, nil
}

// CreateCollectionRaw creates a collection using the documented/raw REST payload shape.
func (c *Client) CreateCollectionRaw(req map[string]interface{}) (map[string]interface{}, error) {
	data, err := c.doRequest("POST", "/collections", req)
	if err != nil {
		return nil, err
	}
	return decodeObjectResponse(data)
}

// UpdateCollectionSchemaRaw updates a collection schema using the documented/raw REST payload shape.
func (c *Client) UpdateCollectionSchemaRaw(collectionID string, req map[string]interface{}) (map[string]interface{}, error) {
	path := fmt.Sprintf("/collections/%s/schema", url.PathEscape(collectionID))
	data, err := c.doRequest("PUT", path, req)
	if err != nil {
		return nil, err
	}
	return decodeObjectResponse(data)
}

// GetCollectionItems retrieves items from a collection
func (c *Client) GetCollectionItems(collectionID string, maxDepth int) (*models.CollectionItemList, error) {
	path := fmt.Sprintf("/collections/%s/items", url.PathEscape(collectionID))
	if maxDepth > 0 {
		path = fmt.Sprintf("%s?maxDepth=%d", path, maxDepth)
	}

	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result models.CollectionItemList
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &result, nil
}

// AddCollectionItem adds an item to a collection
func (c *Client) AddCollectionItem(collectionID, title string, properties map[string]interface{}, allowNewOptions bool) (*models.CollectionItemList, error) {
	path := fmt.Sprintf("/collections/%s/items", url.PathEscape(collectionID))

	req := struct {
		Items []struct {
			Title      string                 `json:"title"`
			Properties map[string]interface{} `json:"properties,omitempty"`
		} `json:"items"`
		AllowNewSelectOptions bool `json:"allowNewSelectOptions"`
	}{
		Items: []struct {
			Title      string                 `json:"title"`
			Properties map[string]interface{} `json:"properties,omitempty"`
		}{{Title: title, Properties: properties}},
		AllowNewSelectOptions: allowNewOptions,
	}

	data, err := c.doRequest("POST", path, req)
	if err != nil {
		return nil, err
	}

	var result models.CollectionItemList
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &result, nil
}

// AddCollectionItemsRaw adds collection items using the documented REST payload shape.
func (c *Client) AddCollectionItemsRaw(collectionID string, req map[string]interface{}) (*models.CollectionItemList, error) {
	path := fmt.Sprintf("/collections/%s/items", url.PathEscape(collectionID))
	data, err := c.doRequest("POST", path, req)
	if err != nil {
		return nil, err
	}
	var result models.CollectionItemList
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}
	return &result, nil
}

// UpdateCollectionItem updates an item in a collection
func (c *Client) UpdateCollectionItem(collectionID, itemID string, properties map[string]interface{}, allowNewOptions bool) error {
	path := fmt.Sprintf("/collections/%s/items", url.PathEscape(collectionID))

	req := struct {
		ItemsToUpdate []struct {
			ID         string                 `json:"id"`
			Properties map[string]interface{} `json:"properties,omitempty"`
		} `json:"itemsToUpdate"`
		AllowNewSelectOptions bool `json:"allowNewSelectOptions"`
	}{
		ItemsToUpdate: []struct {
			ID         string                 `json:"id"`
			Properties map[string]interface{} `json:"properties,omitempty"`
		}{{ID: itemID, Properties: properties}},
		AllowNewSelectOptions: allowNewOptions,
	}

	_, err := c.doRequest("PUT", path, req)
	return err
}

// UpdateCollectionItemsRaw updates collection items using the documented REST payload shape.
func (c *Client) UpdateCollectionItemsRaw(collectionID string, req map[string]interface{}) error {
	path := fmt.Sprintf("/collections/%s/items", url.PathEscape(collectionID))
	_, err := c.doRequest("PUT", path, req)
	return err
}

// DeleteCollectionItem deletes an item from a collection
func (c *Client) DeleteCollectionItem(collectionID, itemID string) error {
	path := fmt.Sprintf("/collections/%s/items", url.PathEscape(collectionID))

	req := struct {
		IDsToDelete []string `json:"idsToDelete"`
	}{
		IDsToDelete: []string{itemID},
	}

	_, err := c.doRequest("DELETE", path, req)
	return err
}

// DeleteCollectionItemsRaw deletes collection items using the documented REST payload shape.
func (c *Client) DeleteCollectionItemsRaw(collectionID string, req map[string]interface{}) error {
	path := fmt.Sprintf("/collections/%s/items", url.PathEscape(collectionID))
	_, err := c.doRequest("DELETE", path, req)
	return err
}

func decodeObjectResponse(data []byte) (map[string]interface{}, error) {
	if len(strings.TrimSpace(string(data))) == 0 {
		return map[string]interface{}{"ok": true}, nil
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}
	return result, nil
}

// ========== Connection ==========

// GetConnection retrieves connection/space info from the API.
func (c *Client) GetConnection() (*models.ConnectionInfo, error) {
	data, err := c.doRequest("GET", "/connection", nil)
	if err != nil {
		return nil, err
	}

	var result models.ConnectionInfo
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &result, nil
}

// ========== Comments ==========

// addCommentRequest is the request body for adding comments.
type addCommentRequest struct {
	Comments []struct {
		BlockID string `json:"blockId"`
		Content string `json:"content"`
	} `json:"comments"`
}

// AddComment adds a comment to a block.
func (c *Client) AddComment(blockID, content string) (*models.CommentResponse, error) {
	req := addCommentRequest{
		Comments: []struct {
			BlockID string `json:"blockId"`
			Content string `json:"content"`
		}{{BlockID: blockID, Content: content}},
	}

	data, err := c.doRequest("POST", "/comments", req)
	if err != nil {
		return nil, err
	}

	var result models.CommentResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &result, nil
}

// AddCommentsRaw adds comments using the documented REST payload shape.
func (c *Client) AddCommentsRaw(req map[string]interface{}) (*models.CommentResponse, error) {
	data, err := c.doRequest("POST", "/comments", req)
	if err != nil {
		return nil, err
	}
	var result models.CommentResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}
	return &result, nil
}

// ========== Block Search ==========

// SearchBlocks searches for blocks matching a pattern within a document.
func (c *Client) SearchBlocks(blockID, pattern string, caseSensitive bool, beforeCount, afterCount int) (*models.BlockSearchResultList, error) {
	params := url.Values{}
	params.Set("blockId", blockID)
	params.Set("pattern", pattern)
	if caseSensitive {
		params.Set("caseSensitive", "true")
	}
	params.Set("beforeBlockCount", strconv.Itoa(beforeCount))
	params.Set("afterBlockCount", strconv.Itoa(afterCount))

	path := "/blocks/search?" + params.Encode()

	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result models.BlockSearchResultList
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &result, nil
}

// ========== Advanced Document Search ==========

// SearchOptions contains optional parameters for advanced document search.
type SearchOptions struct {
	Regexps             string
	Location            string
	FolderIDs           string
	DocumentIDs         string
	FetchMetadata       bool
	CreatedDateGte      string
	CreatedDateLte      string
	LastModifiedDateGte string
	LastModifiedDateLte string
	DailyNoteDateGte    string
	DailyNoteDateLte    string
}

// SearchDocumentsAdvanced searches for documents with full option support.
func (c *Client) SearchDocumentsAdvanced(query string, opts SearchOptions) (*models.SearchResult, error) {
	params := url.Values{}
	params.Set("include", query)

	if opts.Regexps != "" {
		params.Set("regexps", opts.Regexps)
	}
	if opts.Location != "" {
		params.Set("location", opts.Location)
	}
	if opts.FolderIDs != "" {
		params.Set("folderIDs", opts.FolderIDs)
	}
	if opts.DocumentIDs != "" {
		params.Set("documentIDs", opts.DocumentIDs)
	}
	if opts.FetchMetadata {
		params.Set("fetchMetadata", "true")
	}
	if opts.CreatedDateGte != "" {
		params.Set("createdDateGte", opts.CreatedDateGte)
	}
	if opts.CreatedDateLte != "" {
		params.Set("createdDateLte", opts.CreatedDateLte)
	}
	if opts.LastModifiedDateGte != "" {
		params.Set("lastModifiedDateGte", opts.LastModifiedDateGte)
	}
	if opts.LastModifiedDateLte != "" {
		params.Set("lastModifiedDateLte", opts.LastModifiedDateLte)
	}
	if opts.DailyNoteDateGte != "" {
		params.Set("dailyNoteDateGte", opts.DailyNoteDateGte)
	}
	if opts.DailyNoteDateLte != "" {
		params.Set("dailyNoteDateLte", opts.DailyNoteDateLte)
	}

	path := "/documents/search?" + params.Encode()

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

// ========== Advanced Document Listing ==========

// ListDocumentsOptions contains optional parameters for advanced document listing.
type ListDocumentsOptions struct {
	FolderID            string
	Location            string
	FetchMetadata       bool
	CreatedDateGte      string
	CreatedDateLte      string
	LastModifiedDateGte string
	LastModifiedDateLte string
	DailyNoteDateGte    string
	DailyNoteDateLte    string
}

// GetDocumentsAdvanced retrieves documents with full option support.
func (c *Client) GetDocumentsAdvanced(opts ListDocumentsOptions) (*models.DocumentList, error) {
	params := url.Values{}

	if opts.FolderID != "" {
		params.Set("folderId", opts.FolderID)
	}
	if opts.Location != "" {
		params.Set("location", opts.Location)
	}
	if opts.FetchMetadata {
		params.Set("fetchMetadata", "true")
	}
	if opts.CreatedDateGte != "" {
		params.Set("createdDateGte", opts.CreatedDateGte)
	}
	if opts.CreatedDateLte != "" {
		params.Set("createdDateLte", opts.CreatedDateLte)
	}
	if opts.LastModifiedDateGte != "" {
		params.Set("lastModifiedDateGte", opts.LastModifiedDateGte)
	}
	if opts.LastModifiedDateLte != "" {
		params.Set("lastModifiedDateLte", opts.LastModifiedDateLte)
	}
	if opts.DailyNoteDateGte != "" {
		params.Set("dailyNoteDateGte", opts.DailyNoteDateGte)
	}
	if opts.DailyNoteDateLte != "" {
		params.Set("dailyNoteDateLte", opts.DailyNoteDateLte)
	}

	path := "/documents"
	if encoded := params.Encode(); encoded != "" {
		path = path + "?" + encoded
	}

	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result models.DocumentList
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &result, nil
}

// ========== Block by Date ==========

// GetBlockByDate retrieves blocks for a daily note by date.
func (c *Client) GetBlockByDate(date string, maxDepth int, fetchMetadata bool) (*models.Block, error) {
	params := url.Values{}
	params.Set("date", date)
	if maxDepth != -1 {
		params.Set("maxDepth", strconv.Itoa(maxDepth))
	}
	if fetchMetadata {
		params.Set("fetchMetadata", "true")
	}

	path := "/blocks?" + params.Encode()

	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var block models.Block
	if err := json.Unmarshal(data, &block); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &block, nil
}

// ========== Add Block to Date ==========

// addBlockToDateRequest is the request body for adding blocks to a daily note.
type addBlockToDateRequest struct {
	Markdown string `json:"markdown"`
	Position struct {
		Date     string `json:"date"`
		Position string `json:"position"`
	} `json:"position"`
}

// AddBlockToDate adds a block to a daily note page.
func (c *Client) AddBlockToDate(date, markdown, position string) (*models.Block, error) {
	req := addBlockToDateRequest{
		Markdown: markdown,
	}
	req.Position.Date = date
	req.Position.Position = position

	data, err := c.doRequest("POST", "/blocks", req)
	if err != nil {
		return nil, err
	}

	var resp addBlockResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("no block returned from API")
	}

	return &models.Block{
		ID:       resp.Items[0].ID,
		Type:     resp.Items[0].Type,
		Markdown: resp.Items[0].Markdown,
	}, nil
}

// ========== File Upload ==========

// doRequestRaw sends raw bytes with a custom content type.
func (c *Client) doRequestRaw(method, path string, body []byte, contentType string) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	reqURL := fmt.Sprintf("%s%s", c.baseURL, path)
	req, err := http.NewRequest(method, reqURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
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

// UploadFile uploads a file as raw binary data.
// Exactly one of pageID, date, or siblingID must be provided to indicate placement.
func (c *Client) UploadFile(fileData []byte, pageID, date, siblingID, position string) (*models.UploadResponse, error) {
	params := url.Values{}
	if position != "" {
		params.Set("position", position)
	}
	if pageID != "" {
		params.Set("pageId", pageID)
	}
	if date != "" {
		params.Set("date", date)
	}
	if siblingID != "" {
		params.Set("siblingId", siblingID)
	}

	path := "/upload"
	if encoded := params.Encode(); encoded != "" {
		path = path + "?" + encoded
	}

	data, err := c.doRequestRaw("POST", path, fileData, "application/octet-stream")
	if err != nil {
		return nil, err
	}

	var result models.UploadResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &result, nil
}

// ========== JSON Block Operations ==========

// AddBlocksJSON adds blocks using raw JSON maps for full styling support.
// blocks is an array of block maps (type, markdown, textStyle, color, etc.).
// position specifies where to insert (pageId+position, siblingId+position, or date+position).
func (c *Client) AddBlocksJSON(blocks []map[string]interface{}, position map[string]interface{}) ([]models.Block, error) {
	req := map[string]interface{}{
		"blocks":   blocks,
		"position": position,
	}

	data, err := c.doRequest("POST", "/blocks", req)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Items []models.Block `json:"items"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return resp.Items, nil
}

// UpdateBlocksJSON updates blocks using raw JSON maps for full styling support.
// Each block map must include "id" and any fields to update.
func (c *Client) UpdateBlocksJSON(blocks []map[string]interface{}) error {
	req := map[string]interface{}{
		"blocks": blocks,
	}

	_, err := c.doRequest("PUT", "/blocks", req)
	return err
}

// ========== Enhanced Block Retrieval ==========

// GetBlockWithOptions retrieves a block by ID with optional depth and metadata.
func (c *Client) GetBlockWithOptions(blockID string, maxDepth int, fetchMetadata bool) (*models.Block, error) {
	params := url.Values{}
	params.Set("id", blockID)
	if maxDepth != -1 {
		params.Set("maxDepth", strconv.Itoa(maxDepth))
	}
	if fetchMetadata {
		params.Set("fetchMetadata", "true")
	}

	path := "/blocks?" + params.Encode()

	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var block models.Block
	if err := json.Unmarshal(data, &block); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return &block, nil
}

// ========== Whiteboards ==========

// CreateWhiteboard creates a new whiteboard block inside a page.
func (c *Client) CreateWhiteboard(pageID string) (map[string]interface{}, error) {
	req := map[string]interface{}{
		"position": map[string]interface{}{
			"pageId":   pageID,
			"position": "end",
		},
	}

	data, err := c.doRequest("POST", "/whiteboards", req)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return result, nil
}

// GetWhiteboardElements retrieves elements from a whiteboard.
func (c *Client) GetWhiteboardElements(whiteboardID string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/whiteboards/%s/elements", url.PathEscape(whiteboardID))

	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return result, nil
}

// AddWhiteboardElements appends elements to a whiteboard.
func (c *Client) AddWhiteboardElements(whiteboardID string, elements []map[string]interface{}) (map[string]interface{}, error) {
	req := map[string]interface{}{
		"elements": elements,
	}

	path := fmt.Sprintf("/whiteboards/%s/elements", url.PathEscape(whiteboardID))
	data, err := c.doRequest("POST", path, req)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid response from API: %w", err)
	}

	return result, nil
}

// UpdateWhiteboardElements updates specific whiteboard elements.
func (c *Client) UpdateWhiteboardElements(whiteboardID string, elements []map[string]interface{}) error {
	req := map[string]interface{}{
		"elements": elements,
	}

	path := fmt.Sprintf("/whiteboards/%s/elements", url.PathEscape(whiteboardID))
	_, err := c.doRequest("PUT", path, req)
	return err
}

// DeleteWhiteboardElements removes elements from a whiteboard.
func (c *Client) DeleteWhiteboardElements(whiteboardID string, elementIDs []string) error {
	req := map[string]interface{}{
		"elementIds": elementIDs,
	}

	path := fmt.Sprintf("/whiteboards/%s/elements", url.PathEscape(whiteboardID))
	_, err := c.doRequest("DELETE", path, req)
	return err
}
