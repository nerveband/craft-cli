package models

import "time"

// Document represents a Craft document
type Document struct {
	ID          string    `json:"id"`
	SpaceID     string    `json:"spaceId"`
	Title       string    `json:"title"`
	Content     string    `json:"content,omitempty"`
	Markdown    string    `json:"markdown,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	ParentID    string    `json:"parentId,omitempty"`
	HasChildren bool      `json:"hasChildren"`
}

// Block represents a content block from the Craft blocks API
type Block struct {
	ID        string  `json:"id"`
	Type      string  `json:"type"`
	TextStyle string  `json:"textStyle,omitempty"`
	Markdown  string  `json:"markdown,omitempty"`
	Content   []Block `json:"content,omitempty"`
}

// BlocksResponse represents the response from the blocks API
type BlocksResponse struct {
	ID        string  `json:"id"`
	Type      string  `json:"type"`
	TextStyle string  `json:"textStyle,omitempty"`
	Markdown  string  `json:"markdown"`
	Content   []Block `json:"content,omitempty"`
}

// DocumentList represents the response from listing documents
type DocumentList struct {
	Items []Document `json:"items"`
	Total int        `json:"total"`
}

// CreateDocumentRequest represents the request to create a document
type CreateDocumentRequest struct {
	Title    string `json:"title"`
	Content  string `json:"content,omitempty"`
	Markdown string `json:"markdown,omitempty"`
	ParentID string `json:"parentId,omitempty"`
}

// UpdateDocumentRequest represents the request to update a document
type UpdateDocumentRequest struct {
	Title    string `json:"title,omitempty"`
	Content  string `json:"content,omitempty"`
	Markdown string `json:"markdown,omitempty"`
}

// SearchResult represents a search result
type SearchResult struct {
	Items []SearchItem `json:"items"`
	Total int          `json:"total"`
}

// SearchItem represents a single search result item from the Craft API
type SearchItem struct {
	DocumentID string `json:"documentId"`
	Markdown   string `json:"markdown"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}
