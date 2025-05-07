package document

import (
	"errors"
	"io"
	"time"
)

// DocumentContent represents the raw content of a document being processed
type DocumentContent struct {
	OriginalFilename string
	MIMEType         string
	Stream           io.Reader
	Size             int64
}

// DocumentID is a unique identifier for a document
type DocumentID string

// DocumentStatus represents the current processing state of a document
type DocumentStatus string

const (
	StatusUploading         DocumentStatus = "UPLOADING"
	StatusPendingProcessing DocumentStatus = "PENDING_PROCESSING"
	StatusParsing           DocumentStatus = "PARSING"
	StatusProcessing        DocumentStatus = "PROCESSING"
	StatusReady             DocumentStatus = "READY"
	StatusError             DocumentStatus = "ERROR"
)

// DocumentMetadata contains information about a document
type DocumentMetadata struct {
	ID                        DocumentID     `json:"id" db:"id"`
	OwnerID                   string         `json:"owner_id" db:"owner_id"` // User who uploaded/owns the document
	OriginalFilename          string         `json:"original_filename" db:"original_filename"`
	Source                    string         `json:"source" db:"source"` // e.g., "upload", "gdrive"
	MIMEType                  string         `json:"mime_type" db:"mime_type"`
	Size                      int64          `json:"size" db:"size"`
	Status                    DocumentStatus `json:"status" db:"status"`
	CreatedAt                 time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt                 time.Time      `json:"updated_at" db:"updated_at"`
	ProcessedContentReference string         `json:"processed_content_reference,omitempty" db:"processed_content_reference"` // path/ID to stored Markdown/StructuredOutput
	OriginalFileReference     string         `json:"original_file_reference,omitempty" db:"original_file_reference"`         // path/ID to stored original file
	Summary                   string         `json:"summary,omitempty" db:"summary"`                                         // Populated by TextAnalyzer
	Keywords                  []string       `json:"keywords,omitempty" db:"keywords"`                                       // Populated by TextAnalyzer
	DomainMetadata            interface{}    `json:"domain_metadata,omitempty" db:"domain_metadata"`                         // Optional: Domain-specific metadata
	IsDeleted                 bool           `json:"is_deleted" db:"is_deleted"`                                             // Soft delete flag
	DeletedAt                 *time.Time     `json:"deleted_at,omitempty" db:"deleted_at"`                                   // When the document was deleted
	DeletedBy                 string         `json:"deleted_by,omitempty" db:"deleted_by"`                                   // Who deleted the document
	ErrorDetails              string         `json:"error_details,omitempty" db:"error_details"`                             // Details of error if status is ERROR
	ProcessingAttempts        int            `json:"processing_attempts" db:"processing_attempts"`                           // Number of processing attempts
	LastProcessedAt           *time.Time     `json:"last_processed_at,omitempty" db:"last_processed_at"`                     // When the document was last processed
}

// Validate checks if a document metadata is valid
func (dm *DocumentMetadata) Validate() error {
	if dm.ID == "" {
		return errors.New("id cannot be empty")
	}
	if dm.OriginalFilename == "" {
		return errors.New("original_filename cannot be empty")
	}
	if dm.OwnerID == "" {
		return errors.New("owner_id cannot be empty")
	}
	return nil
}

// DocumentSpaceAssignment represents the association between a Document and a Space
type DocumentSpaceAssignment struct {
	DocumentID DocumentID `json:"document_id" db:"document_id"` // Foreign Key to DocumentMetadata.ID
	SpaceID    string     `json:"space_id" db:"space_id"`       // Foreign Key to the Space entity's ID
	AssignedAt time.Time  `json:"assigned_at" db:"assigned_at"` // Timestamp of when the assignment was made
	AssignedBy string     `json:"assigned_by" db:"assigned_by"` // UserID of who made the assignment
}

// TableData represents a table extracted from a document
type TableData struct {
	ID          string     `json:"id"`
	DocumentID  DocumentID `json:"document_id"`
	Headers     []string   `json:"headers,omitempty"`
	Rows        [][]string `json:"rows"`
	Description string     `json:"description,omitempty"`
}

// FormData represents a form or structured data extracted from a document
type FormData struct {
	ID         string                 `json:"id"`
	DocumentID DocumentID             `json:"document_id"`
	Fields     map[string]interface{} `json:"fields"`
}

// ImageReference represents an image extracted from a document
type ImageReference struct {
	ID          string     `json:"id"`
	DocumentID  DocumentID `json:"document_id"`
	Reference   string     `json:"reference"` // path/ID to stored image
	Description string     `json:"description,omitempty"`
	Position    int        `json:"position"` // Position within the document
}

// StructuredDocumentOutput represents the parsed and structured output from a document
type StructuredDocumentOutput struct {
	MarkdownContent       string                 `json:"markdown_content" db:"markdown_content"`             // Primary textual output
	StorageReference      string                 `json:"storage_reference,omitempty" db:"storage_reference"` // Reference to storage object
	RawText               string                 `json:"raw_text,omitempty" db:"raw_text"`
	Title                 string                 `json:"title,omitempty" db:"title"`
	Author                string                 `json:"author,omitempty" db:"author"`
	LayoutInfo            interface{}            `json:"layout_info,omitempty" db:"layout_info"` // e.g., page structure, sections
	Tables                []TableData            `json:"tables,omitempty"`
	TablesJSON            string                 `json:"tables_json,omitempty" db:"tables_json"` // JSON serialized version of tables
	Forms                 []FormData             `json:"forms,omitempty"`
	Images                []ImageReference       `json:"images,omitempty"`
	ImagesJSON            string                 `json:"images_json,omitempty" db:"images_json"` // JSON serialized version of images
	OriginalFileReference string                 `json:"original_file_reference,omitempty" db:"original_file_reference"`
	OtherMetadata         map[string]interface{} `json:"other_metadata,omitempty"`
	MetadataJSON          string                 `json:"metadata_json,omitempty" db:"metadata_json"` // JSON serialized version of other metadata
}

// UserContext contains information about the user and their context
type UserContext struct {
	UserID string `json:"user_id"`
	Tier   string `json:"tier"`   // e.g., "free", "premium"
	Domain string `json:"domain"` // e.g., "general", "academic", "financial"
}

// FileDownload contains information needed to download a file
type FileDownload struct {
	ContentType     string    `json:"content_type"`
	FileName        string    `json:"file_name"`
	ContentLength   int64     `json:"content_length"`
	DownloadURL     string    `json:"download_url,omitempty"`     // For pre-signed URLs
	ExpiresAt       time.Time `json:"expires_at,omitempty"`       // For pre-signed URLs
	Content         io.Reader `json:"-"`                          // For direct content streaming
	ContentProvider string    `json:"content_provider,omitempty"` // e.g., "s3", "local"
}

// UserRole defines the role of a user in a document
type UserRole string

const (
	RoleOwner     UserRole = "OWNER"
	RoleAdmin     UserRole = "ADMIN"
	RoleEditor    UserRole = "EDITOR"
	RoleCommenter UserRole = "COMMENTER"
	RoleViewer    UserRole = "VIEWER"
	RoleGuest     UserRole = "GUEST"
	RoleAuditor   UserRole = "AUDITOR"
)

// RolePermissions maps roles to their allowed actions
var RolePermissions = map[UserRole]map[string]bool{
	RoleOwner: {
		"view":               true,
		"download":           true,
		"edit":               true,
		"comment":            true,
		"delete":             true,
		"permanent_delete":   true,
		"change_permissions": true,
		"share":              true,
	},
	RoleAdmin: {
		"view":               true,
		"download":           true,
		"edit":               true,
		"comment":            true,
		"delete":             true,
		"permanent_delete":   false,
		"change_permissions": true,
		"share":              true,
	},
	RoleEditor: {
		"view":               true,
		"download":           true,
		"edit":               true,
		"comment":            true,
		"delete":             false,
		"permanent_delete":   false,
		"change_permissions": false,
		"share":              true,
	},
	RoleCommenter: {
		"view":               true,
		"download":           true,
		"edit":               false,
		"comment":            true,
		"delete":             false,
		"permanent_delete":   false,
		"change_permissions": false,
		"share":              false,
	},
	RoleViewer: {
		"view":               true,
		"download":           true,
		"edit":               false,
		"comment":            false,
		"delete":             false,
		"permanent_delete":   false,
		"change_permissions": false,
		"share":              false,
	},
	RoleGuest: {
		"view":               true,
		"download":           false,
		"edit":               false,
		"comment":            false,
		"delete":             false,
		"permanent_delete":   false,
		"change_permissions": false,
		"share":              false,
	},
	RoleAuditor: {
		"view":               true,
		"download":           true,
		"edit":               false,
		"comment":            false,
		"delete":             false,
		"permanent_delete":   false,
		"change_permissions": false,
		"share":              false,
	},
}
