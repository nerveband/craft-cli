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
	Items []Document `json:"items"`
	Total int        `json:"total"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}
