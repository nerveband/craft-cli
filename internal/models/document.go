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
// Enhanced for MCP parity with full styling and metadata support
type Block struct {
	ID        string  `json:"id"`
	Type      string  `json:"type"`                   // text, page, table, code, line, image, file, richUrl
	TextStyle string  `json:"textStyle,omitempty"`    // h1, h2, h3, h4, page, card, caption
	Markdown  string  `json:"markdown,omitempty"`
	Content   []Block `json:"content,omitempty"`

	// Styling properties (MCP parity)
	ListStyle        string   `json:"listStyle,omitempty"`        // bullet, numbered, task, toggle
	Decorations      []string `json:"decorations,omitempty"`      // callout, quote (can combine)
	Color            string   `json:"color,omitempty"`            // #RRGGBB hex
	CardLayout       string   `json:"cardLayout,omitempty"`       // small, regular, large
	IndentationLevel int      `json:"indentationLevel,omitempty"` // 0-5
	LineStyle        string   `json:"lineStyle,omitempty"`        // strong, regular, light, extraLight, pageBreak
	Font             string   `json:"font,omitempty"`             // system, serif, mono, rounded
	TextAlignment    string   `json:"textAlignment,omitempty"`    // left, center, right

	// Task-specific
	TaskInfo *TaskInfo `json:"taskInfo,omitempty"`

	// Media blocks (image, file)
	URL      string `json:"url,omitempty"`
	AltText  string `json:"altText,omitempty"`
	FileName string `json:"fileName,omitempty"`

	// Rich URL blocks
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Layout      string `json:"layout,omitempty"`

	// Code blocks
	Language string `json:"language,omitempty"`
	RawCode  string `json:"rawCode,omitempty"`

	// Table blocks
	Rows [][]TableCell `json:"rows,omitempty"`

	// Metadata
	Metadata *BlockMetadata `json:"metadata,omitempty"`
}

// TaskInfo represents task-specific metadata
type TaskInfo struct {
	State        string `json:"state,omitempty"`        // todo, done, canceled
	CompletedAt  string `json:"completedAt,omitempty"`  // ISO 8601 timestamp
	CanceledAt   string `json:"canceledAt,omitempty"`   // ISO 8601 timestamp
	ScheduleDate string `json:"scheduleDate,omitempty"` // YYYY-MM-DD
	DeadlineDate string `json:"deadlineDate,omitempty"` // YYYY-MM-DD
	Repeat       *RepeatConfig `json:"repeat,omitempty"`
}

// RepeatConfig represents task repeat configuration
type RepeatConfig struct {
	Type     string `json:"type,omitempty"`     // daily, weekly, monthly, yearly
	Interval int    `json:"interval,omitempty"` // every N days/weeks/etc
	Weekdays []int  `json:"weekdays,omitempty"` // 0=Sunday, 6=Saturday
	EndDate  string `json:"endDate,omitempty"`  // YYYY-MM-DD
}

// TableCell represents a cell in a table block
type TableCell struct {
	Value      string     `json:"value"`
	Attributes []TextAttr `json:"attributes,omitempty"`
}

// TextAttr represents text formatting attributes
type TextAttr struct {
	Type  string `json:"type"`            // bold, italic, highlight, strikethrough, code, link
	Start int    `json:"start"`
	End   int    `json:"end"`
	Color string `json:"color,omitempty"` // for highlights (e.g., gradient-blue)
	URL   string `json:"url,omitempty"`   // for links
}

// BlockMetadata contains block timing and authorship information
type BlockMetadata struct {
	CreatedAt      string `json:"createdAt,omitempty"`
	LastModifiedAt string `json:"lastModifiedAt,omitempty"`
	CreatedBy      string `json:"createdBy,omitempty"`
	LastModifiedBy string `json:"lastModifiedBy,omitempty"`
	ClickableLink  string `json:"clickableLink,omitempty"`
}

// BlocksResponse represents the response from the blocks API
// Now supports all block properties for MCP parity
type BlocksResponse struct {
	ID        string  `json:"id"`
	Type      string  `json:"type"`
	TextStyle string  `json:"textStyle,omitempty"`
	Markdown  string  `json:"markdown"`
	Content   []Block `json:"content,omitempty"`

	// Additional properties that may appear on root block
	CardLayout string         `json:"cardLayout,omitempty"`
	Metadata   *BlockMetadata `json:"metadata,omitempty"`
}

// Folder represents a Craft folder
type Folder struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ParentID      string `json:"parentId,omitempty"`
	DocumentCount int    `json:"documentCount,omitempty"`
}

// FolderList represents a list of folders
type FolderList struct {
	Items []Folder `json:"items"`
	Total int      `json:"total"`
}

// MoveRequest represents a request to move a document or folder
type MoveRequest struct {
	TargetFolderID string `json:"targetFolderId,omitempty"`
	TargetLocation string `json:"targetLocation,omitempty"` // unsorted, trash, etc.
}

// Task represents a task from the tasks API
type Task struct {
	ID           string    `json:"id"`
	BlockID      string    `json:"blockId"`
	DocumentID   string    `json:"documentId"`
	Markdown     string    `json:"markdown"`
	State        string    `json:"state"` // todo, done, canceled
	CompletedAt  string    `json:"completedAt,omitempty"`
	CanceledAt   string    `json:"canceledAt,omitempty"`
	ScheduleDate string    `json:"scheduleDate,omitempty"`
	DeadlineDate string    `json:"deadlineDate,omitempty"`
	Repeat       *RepeatConfig `json:"repeat,omitempty"`
}

// TaskList represents a list of tasks
type TaskList struct {
	Items []Task `json:"items"`
	Total int    `json:"total"`
}

// AddBlockRequest represents a request to add a block
type AddBlockRequest struct {
	Markdown       string `json:"markdown,omitempty"`
	Position       string `json:"position,omitempty"`       // start, end
	SiblingBlockID string `json:"siblingBlockId,omitempty"` // for before/after positioning
	RelativePos    string `json:"relativePos,omitempty"`    // before, after
}

// UpdateBlockRequest represents a request to update a block
type UpdateBlockRequest struct {
	Markdown  string `json:"markdown,omitempty"`
	TextStyle string `json:"textStyle,omitempty"`
	ListStyle string `json:"listStyle,omitempty"`
	Color     string `json:"color,omitempty"`
	Font      string `json:"font,omitempty"`
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

// Collection represents a Craft collection (database)
type Collection struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	ItemCount  int    `json:"itemCount"`
	DocumentID string `json:"documentId"`
}

// CollectionList represents a list of collections
type CollectionList struct {
	Items []Collection `json:"items"`
}

// CollectionSchema represents a collection's schema
type CollectionSchema struct {
	Key                string                 `json:"key"`
	Name               string                 `json:"name"`
	ContentPropDetails *CollectionPropDetails `json:"contentPropDetails,omitempty"`
	Properties         []CollectionProperty   `json:"properties"`
}

// CollectionPropDetails describes the content/title property
type CollectionPropDetails struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// CollectionProperty represents a property in a collection schema
type CollectionProperty struct {
	Key     string   `json:"key"`
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Options []string `json:"options,omitempty"`
}

// CollectionItem represents an item in a collection
type CollectionItem struct {
	ID         string                 `json:"id"`
	Title      string                 `json:"title"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Content    []Block                `json:"content,omitempty"`
}

// CollectionItemList represents a list of collection items
type CollectionItemList struct {
	Items []CollectionItem `json:"items"`
}

// ConnectionInfo represents the response from GET /connection
type ConnectionInfo struct {
	Space struct {
		ID           string `json:"id"`
		Timezone     string `json:"timezone"`
		Time         string `json:"time"`
		FriendlyDate string `json:"friendlyDate"`
	} `json:"space"`
	UTC struct {
		Time string `json:"time"`
	} `json:"utc"`
	URLTemplates struct {
		App string `json:"app"`
	} `json:"urlTemplates"`
}

// UploadResponse represents the response from POST /upload
type UploadResponse struct {
	BlockID  string `json:"blockId"`
	AssetURL string `json:"assetUrl"`
}

// CommentResponse represents the response from POST /comments
type CommentResponse struct {
	Items []struct {
		CommentID string `json:"commentId"`
	} `json:"items"`
}

// BlockSearchResult represents a single block search match
type BlockSearchResult struct {
	BlockID       string           `json:"blockId"`
	Markdown      string           `json:"markdown"`
	PageBlockPath []PageBlockEntry `json:"pageBlockPath,omitempty"`
	BeforeBlocks  []BlockContext   `json:"beforeBlocks,omitempty"`
	AfterBlocks   []BlockContext   `json:"afterBlocks,omitempty"`
}

// PageBlockEntry represents a path entry in block search results
type PageBlockEntry struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

// BlockContext represents a surrounding block in search results
type BlockContext struct {
	BlockID  string `json:"blockId"`
	Markdown string `json:"markdown"`
}

// BlockSearchResultList represents block search results
type BlockSearchResultList struct {
	Items []BlockSearchResult `json:"items"`
}
