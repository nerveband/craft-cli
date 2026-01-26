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
	path := fmt.Sprintf("/blocks?id=%s", url.QueryEscape(id))
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
	Folders []struct {
		ID       string `json:"id"`
		ParentID string `json:"parentId"`
	} `json:"folders"`
}

// MoveFolder moves a folder to a new parent
func (c *Client) MoveFolder(folderID, targetParentID string) error {
	req := moveFolderRequest{
		Folders: []struct {
			ID       string `json:"id"`
			ParentID string `json:"parentId"`
		}{{ID: folderID, ParentID: targetParentID}},
	}

	_, err := c.doRequest("PUT", "/folders", req)
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
	Documents []struct {
		ID       string `json:"id"`
		FolderID string `json:"folderId,omitempty"`
		Location string `json:"location,omitempty"` // unsorted, trash, etc.
	} `json:"documents"`
}

// MoveDocument moves a document to a folder or location
func (c *Client) MoveDocument(docID, folderID, location string) error {
	docMove := struct {
		ID       string `json:"id"`
		FolderID string `json:"folderId,omitempty"`
		Location string `json:"location,omitempty"`
	}{ID: docID}

	if folderID != "" {
		docMove.FolderID = folderID
	}
	if location != "" {
		docMove.Location = location
	}

	req := moveDocumentRequest{
		Documents: []struct {
			ID       string `json:"id"`
			FolderID string `json:"folderId,omitempty"`
			Location string `json:"location,omitempty"`
		}{docMove},
	}

	_, err := c.doRequest("PUT", "/documents", req)
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
		Markdown     string `json:"markdown"`
		Location     string `json:"location,omitempty"` // inbox, document
		DocumentID   string `json:"documentId,omitempty"`
		ScheduleDate string `json:"scheduleDate,omitempty"`
		DeadlineDate string `json:"deadlineDate,omitempty"`
	} `json:"tasks"`
}

// addTaskResponse is the response from adding tasks
type addTaskResponse struct {
	Items []struct {
		ID         string `json:"id"`
		BlockID    string `json:"blockId"`
		DocumentID string `json:"documentId"`
		Markdown   string `json:"markdown"`
		State      string `json:"state"`
	} `json:"items"`
}

// AddTask creates a new task
func (c *Client) AddTask(markdown, location, docID, scheduleDate, deadlineDate string) (*models.Task, error) {
	req := addTaskRequest{
		Tasks: []struct {
			Markdown     string `json:"markdown"`
			Location     string `json:"location,omitempty"`
			DocumentID   string `json:"documentId,omitempty"`
			ScheduleDate string `json:"scheduleDate,omitempty"`
			DeadlineDate string `json:"deadlineDate,omitempty"`
		}{{
			Markdown:     markdown,
			Location:     location,
			DocumentID:   docID,
			ScheduleDate: scheduleDate,
			DeadlineDate: deadlineDate,
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
	return &models.Task{
		ID:         item.ID,
		BlockID:    item.BlockID,
		DocumentID: item.DocumentID,
		Markdown:   item.Markdown,
		State:      item.State,
	}, nil
}

// updateTaskRequest is the request body for updating tasks
type updateTaskRequest struct {
	Tasks []struct {
		ID           string `json:"id"`
		State        string `json:"state,omitempty"`
		ScheduleDate string `json:"scheduleDate,omitempty"`
		DeadlineDate string `json:"deadlineDate,omitempty"`
	} `json:"tasks"`
}

// UpdateTask updates a task's state or dates
func (c *Client) UpdateTask(taskID, state, scheduleDate, deadlineDate string) error {
	req := updateTaskRequest{
		Tasks: []struct {
			ID           string `json:"id"`
			State        string `json:"state,omitempty"`
			ScheduleDate string `json:"scheduleDate,omitempty"`
			DeadlineDate string `json:"deadlineDate,omitempty"`
		}{{
			ID:           taskID,
			State:        state,
			ScheduleDate: scheduleDate,
			DeadlineDate: deadlineDate,
		}},
	}

	_, err := c.doRequest("PUT", "/tasks", req)
	return err
}

// deleteTaskRequest is the request body for deleting tasks
type deleteTaskRequest struct {
	TaskIDs []string `json:"taskIds"`
}

// DeleteTask deletes a task
func (c *Client) DeleteTask(taskID string) error {
	req := deleteTaskRequest{
		TaskIDs: []string{taskID},
	}

	_, err := c.doRequest("DELETE", "/tasks", req)
	return err
}
