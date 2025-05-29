package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/ran/demo/backend-go/internal/core/models/document"
	ports "github.com/ran/demo/backend-go/internal/core/ports/document"
)

// DocumentParserWorker handles document parsing tasks from the queue
type DocumentParserWorker struct {
	queueService      ports.QueuePort
	repository        ports.DocumentRepositoryPort
	storage           ports.StoragePort
	parser            ports.FileParsingOrchestratorPort
	logger            *slog.Logger
	stopCh            chan struct{}
	wg                sync.WaitGroup
	maxConcurrency    int
	pollInterval      time.Duration
	visibilityTimeout int
	maxRetries        int
}

// NewDocumentParserWorker creates a new document parser worker
func NewDocumentParserWorker(
	queueService ports.QueuePort,
	repository ports.DocumentRepositoryPort,
	storage ports.StoragePort,
	parser ports.FileParsingOrchestratorPort,
	logger *slog.Logger,
) *DocumentParserWorker {
	return &DocumentParserWorker{
		queueService:      queueService,
		repository:        repository,
		storage:           storage,
		parser:            parser,
		logger:            logger,
		stopCh:            make(chan struct{}),
		maxConcurrency:    5,               // Process up to 5 documents concurrently
		pollInterval:      5 * time.Second, // Poll every 5 seconds
		visibilityTimeout: 300,             // 5 minutes to process a message
		maxRetries:        3,               // Retry failed messages up to 3 times
	}
}

// Start begins processing messages from the queue
func (w *DocumentParserWorker) Start(ctx context.Context) {
	w.logger.Info("Starting document parser worker",
		"max_concurrency", w.maxConcurrency,
		"poll_interval", w.pollInterval)

	// Create a semaphore to limit concurrency
	semaphore := make(chan struct{}, w.maxConcurrency)

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		ticker := time.NewTicker(w.pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				w.logger.Info("Context cancelled, stopping worker")
				return
			case <-w.stopCh:
				w.logger.Info("Stop signal received, stopping worker")
				return
			case <-ticker.C:
				w.pollAndProcessMessages(ctx, semaphore)
			}
		}
	}()
}

// Stop gracefully stops the worker
func (w *DocumentParserWorker) Stop() {
	w.logger.Info("Stopping document parser worker")
	close(w.stopCh)
	w.wg.Wait()
	w.logger.Info("Document parser worker stopped")
}

// pollAndProcessMessages polls for messages and processes them
func (w *DocumentParserWorker) pollAndProcessMessages(ctx context.Context, semaphore chan struct{}) {
	// Receive messages from the queue
	messages, err := w.queueService.ReceiveMessages(ctx, w.maxConcurrency, 20) // Long polling
	if err != nil {
		w.logger.Error("Failed to receive messages from queue", "error", err)
		return
	}

	if len(messages) == 0 {
		return
	}

	w.logger.Debug("Received messages from queue", "count", len(messages))

	// Process each message concurrently
	for _, message := range messages {
		select {
		case semaphore <- struct{}{}: // Acquire semaphore
			w.wg.Add(1)
			go func(msg *document.QueueMessage) {
				defer func() {
					<-semaphore // Release semaphore
					w.wg.Done()
				}()
				w.processMessage(ctx, msg)
			}(message)
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		}
	}
}

// processMessage processes a single message
func (w *DocumentParserWorker) processMessage(ctx context.Context, message *document.QueueMessage) {
	startTime := time.Now()
	w.logger.Info("Processing message",
		"message_id", message.ID,
		"type", message.Type,
		"document_id", message.DocumentID)

	var err error
	switch message.Type {
	case document.QueueMessageTypeParse:
		err = w.processParseTask(ctx, message)
	case document.QueueMessageTypeProcess:
		err = w.processDocumentTask(ctx, message)
	default:
		w.logger.Warn("Unknown message type", "type", message.Type, "message_id", message.ID)
		// Delete unknown message types to prevent infinite reprocessing
		if err := w.queueService.DeleteMessage(ctx, message.ID, message.ReceiptHandle); err != nil {
			w.logger.Error("Failed to delete unknown message type", "message_id", message.ID, "error", err)
		}
		return
	}

	processingTime := time.Since(startTime)

	if err != nil {
		w.logger.Error("Failed to process message",
			"message_id", message.ID,
			"error", err,
			"processing_time", processingTime)

		// Handle retry logic - this could be improved with exponential backoff
		w.handleMessageFailure(ctx, message, err)
	} else {
		w.logger.Info("Successfully processed message",
			"message_id", message.ID,
			"processing_time", processingTime)

		// Delete the message from the queue
		err = w.queueService.DeleteMessage(ctx, message.ID, message.ReceiptHandle)
		if err != nil {
			w.logger.Error("Failed to delete processed message",
				"message_id", message.ID,
				"error", err)
		}
	}
}

// processParseTask processes a document parsing task
func (w *DocumentParserWorker) processParseTask(ctx context.Context, message *document.QueueMessage) error {
	// Parse the payload
	var task document.DocumentParseTask
	payloadBytes, err := json.Marshal(message.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	err = json.Unmarshal(payloadBytes, &task)
	if err != nil {
		return fmt.Errorf("failed to unmarshal parse task: %w", err)
	}

	// Update document status to parsing
	err = w.repository.UpdateDocumentStatus(ctx, task.DocumentID, document.StatusParsing)
	if err != nil {
		return fmt.Errorf("failed to update document status to parsing: %w", err)
	}

	// Download the document content from storage
	content, err := w.storage.Download(ctx, task.StorageObjectID)
	if err != nil {
		w.repository.UpdateDocumentStatus(ctx, task.DocumentID, document.StatusError)
		return fmt.Errorf("failed to download document content: %w", err)
	}

	// Parse the document
	parsingOptions := &ports.FileParsingOptions{
		ExtractTables: true,
		ExtractForms:  true,
		ExtractImages: true,
		Language:      "auto",
	}

	// Add any task-specific parsing options
	if task.ParsingOptions != nil {
		if extractTables, ok := task.ParsingOptions["extract_tables"].(bool); ok {
			parsingOptions.ExtractTables = extractTables
		}
		if extractForms, ok := task.ParsingOptions["extract_forms"].(bool); ok {
			parsingOptions.ExtractForms = extractForms
		}
		if extractImages, ok := task.ParsingOptions["extract_images"].(bool); ok {
			parsingOptions.ExtractImages = extractImages
		}
		if language, ok := task.ParsingOptions["language"].(string); ok {
			parsingOptions.Language = language
		}
	}

	structuredOutput, err := w.parser.ParseDocument(ctx, content, task.UserID, parsingOptions)
	if err != nil {
		w.repository.UpdateDocumentStatus(ctx, task.DocumentID, document.StatusError)
		return fmt.Errorf("failed to parse document: %w", err)
	}

	// Store the parsed content
	err = w.repository.StoreProcessedContent(ctx, task.DocumentID, structuredOutput)
	if err != nil {
		w.repository.UpdateDocumentStatus(ctx, task.DocumentID, document.StatusError)
		return fmt.Errorf("failed to store processed content: %w", err)
	}

	// Update document status to ready
	err = w.repository.UpdateDocumentStatus(ctx, task.DocumentID, document.StatusReady)
	if err != nil {
		return fmt.Errorf("failed to update document status to ready: %w", err)
	}

	// If there are additional processing steps needed, queue a process task
	if len(structuredOutput.Tables) > 0 || len(structuredOutput.Images) > 0 {
		processTask := &document.DocumentProcessTask{
			DocumentID:             task.DocumentID,
			ParsedContentReference: fmt.Sprintf("document:%s:parsed", task.DocumentID),
			UserID:                 task.UserID,
		}

		err = w.queueService.SendDocumentProcessTask(ctx, processTask)
		if err != nil {
			w.logger.Warn("Failed to queue document process task",
				"document_id", task.DocumentID,
				"error", err)
			// Don't fail the parsing task if we can't queue processing
		}
	}

	return nil
}

// processDocumentTask processes a document processing task (post-parsing)
func (w *DocumentParserWorker) processDocumentTask(ctx context.Context, message *document.QueueMessage) error {
	// Parse the payload
	var task document.DocumentProcessTask
	payloadBytes, err := json.Marshal(message.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	err = json.Unmarshal(payloadBytes, &task)
	if err != nil {
		return fmt.Errorf("failed to unmarshal process task: %w", err)
	}

	// Update document status to processing
	err = w.repository.UpdateDocumentStatus(ctx, task.DocumentID, document.StatusProcessing)
	if err != nil {
		return fmt.Errorf("failed to update document status to processing: %w", err)
	}

	// For now, this is a placeholder for additional processing steps
	// In the future, this could include:
	// - Vector embedding generation
	// - Content indexing for search
	// - Image processing and OCR
	// - Table data extraction and structuring
	// - Content summarization

	w.logger.Info("Processing document task",
		"document_id", task.DocumentID,
		"parsed_content_ref", task.ParsedContentReference)

	// Simulate processing time
	time.Sleep(1 * time.Second)

	// Update document status to ready (final state after processing)
	err = w.repository.UpdateDocumentStatus(ctx, task.DocumentID, document.StatusReady)
	if err != nil {
		return fmt.Errorf("failed to update document status to ready: %w", err)
	}

	return nil
}

// handleMessageFailure handles message processing failures
func (w *DocumentParserWorker) handleMessageFailure(ctx context.Context, message *document.QueueMessage, err error) {
	// For now, simply send to DLQ if configured
	// In a more sophisticated implementation, you might:
	// - Implement exponential backoff
	// - Track retry count in message attributes
	// - Apply different retry strategies based on error type

	dlqErr := w.queueService.SendToDeadLetterQueue(ctx, message, err.Error())
	if dlqErr != nil {
		w.logger.Error("Failed to send message to DLQ",
			"message_id", message.ID,
			"original_error", err,
			"dlq_error", dlqErr)
	}

	// Try to delete the message to prevent infinite reprocessing
	deleteErr := w.queueService.DeleteMessage(ctx, message.ID, message.ReceiptHandle)
	if deleteErr != nil {
		w.logger.Error("Failed to delete failed message",
			"message_id", message.ID,
			"error", deleteErr)
	}
}
