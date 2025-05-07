package sqs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/ran/demo/backend-go/internal/core/models/document"
)

// SQSQueueService implements document.QueuePort using AWS SQS
type SQSQueueService struct {
	client       *sqs.Client
	queueURL     string
	dlqURL       string
	logger       *slog.Logger
	skipSendReal bool // For testing
}

// NewSQSQueueService creates a new SQS queue service
func NewSQSQueueService(
	client *sqs.Client,
	queueURL string,
	dlqURL string,
	logger *slog.Logger,
) *SQSQueueService {
	return &SQSQueueService{
		client:   client,
		queueURL: queueURL,
		dlqURL:   dlqURL,
		logger:   logger,
	}
}

// SendMessage sends a message to the queue
func (s *SQSQueueService) SendMessage(ctx context.Context, message *document.QueueMessage) error {
	// Marshal message to JSON
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message to JSON: %w", err)
	}

	// Log message
	s.logger.Debug("Sending message to SQS",
		"message_type", message.Type,
		"message_id", message.ID,
		"document_id", message.DocumentID)

	// Skip actual sending if configured for testing
	if s.skipSendReal {
		return nil
	}

	// Send message to SQS
	input := &sqs.SendMessageInput{
		MessageBody: aws.String(string(messageJSON)),
		QueueUrl:    aws.String(s.queueURL),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"Type": {
				DataType:    aws.String("String"),
				StringValue: aws.String(string(message.Type)),
			},
			"DocumentID": {
				DataType:    aws.String("String"),
				StringValue: aws.String(string(message.DocumentID)),
			},
		},
	}

	// Add delay if specified
	if message.DelaySeconds > 0 {
		input.DelaySeconds = int32(message.DelaySeconds)
	}

	// Send message
	result, err := s.client.SendMessage(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to send message to SQS: %w", err)
	}

	// Update message ID with SQS message ID
	message.ID = document.QueueMessageID(*result.MessageId)

	return nil
}

// SendDocumentParseTask sends a document parsing task to the queue
func (s *SQSQueueService) SendDocumentParseTask(ctx context.Context, task *document.DocumentParseTask) error {
	// Create queue message
	message := &document.QueueMessage{
		ID:         document.QueueMessageID(task.DocumentID),
		Type:       document.QueueMessageTypeParse,
		DocumentID: task.DocumentID,
		Payload:    task,
		CreatedAt:  time.Now(),
	}

	// Send message
	return s.SendMessage(ctx, message)
}

// SendDocumentProcessTask sends a document processing task to the queue
func (s *SQSQueueService) SendDocumentProcessTask(ctx context.Context, task *document.DocumentProcessTask) error {
	// Create queue message
	message := &document.QueueMessage{
		ID:         document.QueueMessageID(task.DocumentID),
		Type:       document.QueueMessageTypeProcess,
		DocumentID: task.DocumentID,
		Payload:    task,
		CreatedAt:  time.Now(),
	}

	// Send message
	return s.SendMessage(ctx, message)
}

// ReceiveMessages retrieves messages from the queue
func (s *SQSQueueService) ReceiveMessages(
	ctx context.Context,
	maxMessages int,
	waitTimeSeconds int,
) ([]*document.QueueMessage, error) {
	// Validate input
	if maxMessages <= 0 || maxMessages > 10 {
		maxMessages = 10 // Default and max limit for SQS ReceiveMessage
	}

	if waitTimeSeconds < 0 || waitTimeSeconds > 20 {
		waitTimeSeconds = 20 // Default to max for SQS long polling
	}

	// Receive messages from SQS
	input := &sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(s.queueURL),
		MaxNumberOfMessages:   int32(maxMessages),
		WaitTimeSeconds:       int32(waitTimeSeconds),
		MessageAttributeNames: []string{"All"},
	}

	result, err := s.client.ReceiveMessage(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to receive messages from SQS: %w", err)
	}

	// Convert SQS messages to our model
	messages := make([]*document.QueueMessage, 0, len(result.Messages))
	for _, msg := range result.Messages {
		if msg.Body == nil {
			s.logger.Warn("Received SQS message with nil body, skipping")
			continue
		}

		// Parse message
		var queueMessage document.QueueMessage
		err := json.Unmarshal([]byte(*msg.Body), &queueMessage)
		if err != nil {
			s.logger.Error("Failed to unmarshal message",
				"error", err,
				"message_id", *msg.MessageId,
				"body", *msg.Body)
			// Skip invalid messages but don't fail the whole batch
			continue
		}

		// Set receipt handle for later deletion
		queueMessage.ReceiptHandle = *msg.ReceiptHandle

		// Add to result
		messages = append(messages, &queueMessage)
	}

	return messages, nil
}

// DeleteMessage removes a message from the queue after successful processing
func (s *SQSQueueService) DeleteMessage(
	ctx context.Context,
	messageID document.QueueMessageID,
	receiptHandle string,
) error {
	// Skip actual deletion if configured for testing
	if s.skipSendReal {
		return nil
	}

	// Delete message from SQS
	input := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(s.queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	}

	_, err := s.client.DeleteMessage(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete message from SQS: %w", err)
	}

	return nil
}

// ExtendVisibilityTimeout extends the timeout for a message processing
func (s *SQSQueueService) ExtendVisibilityTimeout(
	ctx context.Context,
	messageID document.QueueMessageID,
	receiptHandle string,
	visibilityTimeoutSeconds int,
) error {
	// Skip actual extension if configured for testing
	if s.skipSendReal {
		return nil
	}

	// Extend visibility timeout
	input := &sqs.ChangeMessageVisibilityInput{
		QueueUrl:          aws.String(s.queueURL),
		ReceiptHandle:     aws.String(receiptHandle),
		VisibilityTimeout: int32(visibilityTimeoutSeconds),
	}

	_, err := s.client.ChangeMessageVisibility(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to extend message visibility timeout: %w", err)
	}

	return nil
}

// SendToDeadLetterQueue explicitly moves a message to DLQ
func (s *SQSQueueService) SendToDeadLetterQueue(
	ctx context.Context,
	message *document.QueueMessage,
	reason string,
) error {
	// If no DLQ URL is configured, just log and return
	if s.dlqURL == "" {
		s.logger.Warn("No DLQ URL configured, can't send message to DLQ",
			"message_id", message.ID,
			"reason", reason)
		return nil
	}

	// Add failure reason to message
	message.FailureReason = reason

	// Marshal message to JSON
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message to JSON: %w", err)
	}

	// Skip actual sending if configured for testing
	if s.skipSendReal {
		return nil
	}

	// Send message to DLQ
	input := &sqs.SendMessageInput{
		MessageBody: aws.String(string(messageJSON)),
		QueueUrl:    aws.String(s.dlqURL),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"Type": {
				DataType:    aws.String("String"),
				StringValue: aws.String(string(message.Type)),
			},
			"DocumentID": {
				DataType:    aws.String("String"),
				StringValue: aws.String(string(message.DocumentID)),
			},
			"FailureReason": {
				DataType:    aws.String("String"),
				StringValue: aws.String(message.FailureReason),
			},
		},
	}

	_, err = s.client.SendMessage(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to send message to DLQ: %w", err)
	}

	return nil
}

// GetQueueStats retrieves queue statistics
func (s *SQSQueueService) GetQueueStats(ctx context.Context) (map[string]interface{}, error) {
	// Get queue attributes
	input := &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(s.queueURL),
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeNameApproximateNumberOfMessages,
			types.QueueAttributeNameApproximateNumberOfMessagesNotVisible,
			types.QueueAttributeNameApproximateNumberOfMessagesDelayed,
		},
	}

	result, err := s.client.GetQueueAttributes(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue attributes: %w", err)
	}

	// Convert to map
	stats := make(map[string]interface{})
	if result.Attributes != nil {
		for key, value := range result.Attributes {
			stats[string(key)] = value
		}
	}

	return stats, nil
}

// PurgeQueue removes all messages from the queue
func (s *SQSQueueService) PurgeQueue(ctx context.Context) error {
	// Skip actual purge if configured for testing
	if s.skipSendReal {
		return nil
	}

	// Purge queue
	input := &sqs.PurgeQueueInput{
		QueueUrl: aws.String(s.queueURL),
	}

	_, err := s.client.PurgeQueue(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to purge queue: %w", err)
	}

	return nil
}
