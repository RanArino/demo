package document

import (
	"context"

	"github.com/ran/demo/backend-go/internal/core/models/document"
)

// QueuePort defines the interface for document processing queue operations
type QueuePort interface {
	// SendMessage sends a message to the queue
	// This is used to queue document processing tasks
	SendMessage(ctx context.Context, message *document.QueueMessage) error

	// SendDocumentParseTask sends a document parsing task to the queue
	// This is a convenience method that builds the QueueMessage from a DocumentParseTask
	SendDocumentParseTask(ctx context.Context, task *document.DocumentParseTask) error

	// SendDocumentProcessTask sends a document processing task to the queue
	// This is a convenience method that builds the QueueMessage from a DocumentProcessTask
	SendDocumentProcessTask(ctx context.Context, task *document.DocumentProcessTask) error

	// ReceiveMessages retrieves messages from the queue
	// Used by worker processes to get tasks to process
	ReceiveMessages(ctx context.Context, maxMessages int, waitTimeSeconds int) ([]*document.QueueMessage, error)

	// DeleteMessage removes a message from the queue after successful processing
	// This prevents the message from being processed again
	DeleteMessage(ctx context.Context, messageID document.QueueMessageID, receiptHandle string) error

	// ExtendVisibilityTimeout extends the timeout for a message processing
	// Used when processing is taking longer than expected
	ExtendVisibilityTimeout(ctx context.Context, messageID document.QueueMessageID, receiptHandle string, visibilityTimeoutSeconds int) error

	// SendToDeadLetterQueue explicitly moves a message to DLQ
	// Used when a message cannot be processed after multiple attempts
	SendToDeadLetterQueue(ctx context.Context, message *document.QueueMessage, reason string) error

	// GetQueueStats retrieves queue statistics
	// Used for monitoring and health checks
	GetQueueStats(ctx context.Context) (map[string]interface{}, error)

	// PurgeQueue removes all messages from the queue
	// Use with caution - typically only for testing/dev environments
	PurgeQueue(ctx context.Context) error
}
