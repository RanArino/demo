package document

import (
	"time"
)

// QueueProvider specifies the type of queue backend
type QueueProvider string

const (
	QueueProviderSQS    QueueProvider = "sqs"
	QueueProviderLocal  QueueProvider = "local"
	QueueProviderMemory QueueProvider = "memory" // For testing
)

// QueueMessageID is a unique identifier for a queue message
type QueueMessageID string

// QueueMessageType defines the type of message
type QueueMessageType string

const (
	// QueueMessageTypeParse represents a document parsing task
	QueueMessageTypeParse QueueMessageType = "PARSE"

	// QueueMessageTypeProcess represents a document processing task
	QueueMessageTypeProcess QueueMessageType = "PROCESS"

	// QueueMessageTypeIndex represents a document indexing task
	QueueMessageTypeIndex QueueMessageType = "INDEX"
)

// MessageStatus represents the current state of a message
type MessageStatus string

const (
	MessageStatusPending    MessageStatus = "PENDING"
	MessageStatusProcessing MessageStatus = "PROCESSING"
	MessageStatusCompleted  MessageStatus = "COMPLETED"
	MessageStatusFailed     MessageStatus = "FAILED"
	MessageStatusRetry      MessageStatus = "RETRY"
)

// MessagePriority represents the priority of a message
type MessagePriority int

const (
	PriorityLow    MessagePriority = 1
	PriorityNormal MessagePriority = 5
	PriorityHigh   MessagePriority = 10
)

// TaskType identifies what kind of task a message represents
type TaskType string

const (
	TaskTypeParseDocument     TaskType = "PARSE_DOCUMENT"
	TaskTypeProcessDocument   TaskType = "PROCESS_DOCUMENT"
	TaskTypeGenerateEmbedding TaskType = "GENERATE_EMBEDDING"
	TaskTypeGenerateSummary   TaskType = "GENERATE_SUMMARY"
	TaskTypeExtractKeywords   TaskType = "EXTRACT_KEYWORDS"
)

// QueueMessage represents a message in the document processing queue
type QueueMessage struct {
	// ID is the unique identifier for the message
	ID QueueMessageID `json:"id"`

	// Type is the type of message
	Type QueueMessageType `json:"type"`

	// DocumentID is the ID of the document this message is about
	DocumentID DocumentID `json:"document_id"`

	// Payload is the message-specific payload
	Payload interface{} `json:"payload"`

	// CreatedAt is when the message was created
	CreatedAt time.Time `json:"created_at"`

	// DelaySeconds is the number of seconds to delay the message
	DelaySeconds int `json:"delay_seconds,omitempty"`

	// ReceiptHandle is the receipt handle for the message (set when received)
	ReceiptHandle string `json:"receipt_handle,omitempty"`

	// FailureReason is the reason for message failure (for DLQ)
	FailureReason string `json:"failure_reason,omitempty"`
}

// DocumentParseTask represents a document parsing task
type DocumentParseTask struct {
	// DocumentID is the ID of the document to parse
	DocumentID DocumentID `json:"document_id"`

	// StorageObjectID is the ID of the storage object containing the document content
	StorageObjectID StorageObjectID `json:"storage_object_id"`

	// UserID is the ID of the user who initiated the task
	UserID string `json:"user_id"`

	// UserTier is the tier of the user who initiated the task
	UserTier string `json:"user_tier"`

	// UserDomain is the domain of the user who initiated the task
	UserDomain string `json:"user_domain"`

	// OriginalFilename is the original filename of the document
	OriginalFilename string `json:"original_filename"`

	// MIMEType is the MIME type of the document
	MIMEType string `json:"mime_type"`

	// Size is the size of the document in bytes
	Size int64 `json:"size"`

	// ParsingOptions contains options for parsing the document
	ParsingOptions map[string]interface{} `json:"parsing_options,omitempty"`
}

// DocumentProcessTask represents a document processing task after parsing
type DocumentProcessTask struct {
	// DocumentID is the ID of the document to process
	DocumentID DocumentID `json:"document_id"`

	// ParsedContentReference is the reference to the parsed content
	ParsedContentReference string `json:"parsed_content_reference"`

	// UserID is the ID of the user who initiated the task
	UserID string `json:"user_id"`

	// ProcessingOptions contains options for processing the document
	ProcessingOptions map[string]interface{} `json:"processing_options,omitempty"`
}

// ToQueueMessage converts a DocumentParseTask to a QueueMessage
func (t *DocumentParseTask) ToQueueMessage() (*QueueMessage, error) {
	return &QueueMessage{
		ID:           QueueMessageID(t.DocumentID),
		Type:         QueueMessageTypeParse,
		DocumentID:   t.DocumentID,
		Payload:      t,
		CreatedAt:    time.Now(),
		DelaySeconds: 0,
	}, nil
}

// ToQueueMessage converts a DocumentProcessTask to a QueueMessage
func (t *DocumentProcessTask) ToQueueMessage() (*QueueMessage, error) {
	return &QueueMessage{
		ID:           QueueMessageID(t.DocumentID),
		Type:         QueueMessageTypeProcess,
		DocumentID:   t.DocumentID,
		Payload:      t,
		CreatedAt:    time.Now(),
		DelaySeconds: 0,
	}, nil
}

// SQSConfig defines configuration for the SQS provider
type SQSConfig struct {
	Region                string `json:"region"`
	QueueURL              string `json:"queue_url"`
	VisibilityTimeoutSecs int    `json:"visibility_timeout_secs"`
	WaitTimeSeconds       int    `json:"wait_time_seconds"`
	MaxNumberOfMessages   int    `json:"max_number_of_messages"`
	AccessKey             string `json:"access_key,omitempty"`     // For local development
	SecretKey             string `json:"secret_key,omitempty"`     // For local development
	UseIAMRoleCredentials bool   `json:"use_iam_role_credentials"` // For production
	IsFIFO                bool   `json:"is_fifo"`                  // Whether the queue is FIFO
	DLQEnabled            bool   `json:"dlq_enabled"`              // Whether to use a dead letter queue
	DLQUrl                string `json:"dlq_url,omitempty"`        // URL of the dead letter queue
}
