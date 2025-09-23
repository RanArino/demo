package service

import (
	"context"
	"fmt"
	"log"
	"time"

	v1 "demo/ms_canvas/go_app/api/proto/public/v1"
	"demo/ms_canvas/go_app/internal/config"
	"demo/ms_canvas/go_app/internal/events"
	"demo/ms_canvas/go_app/internal/gateway/python"
	"demo/ms_canvas/go_app/internal/repository"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// EventOrchestrator implements EventHandler to coordinate document ingestion workflows
type EventOrchestrator struct {
	config        config.Config
	nodeRepo      repository.NodeRepository
	linkRepo      repository.LinkRepository
	pythonGateway *python.Gateway
}

// EventOrchestrator is focused ONLY on consuming events from other microservices
// It does NOT produce events or use complex task management
// All Task/TaskManager functionality is in task_executor.go for future use when needed

// NewEventOrchestrator creates a new event orchestrator with dependencies
func NewEventOrchestrator(
	config config.Config,
	nodeRepo repository.NodeRepository,
	linkRepo repository.LinkRepository,
	pythonGateway *python.Gateway,
) EventHandler {
	return &EventOrchestrator{
		config:        config,
		nodeRepo:      nodeRepo,
		linkRepo:      linkRepo,
		pythonGateway: pythonGateway,
	}
}

// HandleDocumentProcessed processes incoming document.processed events
func (e *EventOrchestrator) HandleDocumentProcessed(event events.DocumentProcessedEvent) error {
	ctx := context.Background()

	log.Printf("[EventOrchestrator] Processing document.processed event for content_source_id=%s, status=%s",
		event.ContentSourceID, event.Status)

	// Validate event status
	if event.Status != "PROCESSED" {
		log.Printf("[EventOrchestrator] Skipping event with status: %s", event.Status)
		return nil
	}

	// Step 1: Create ContentNode in database
	contentNodeID, err := e.createContentNode(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to create content node: %w", err)
	}

	log.Printf("[EventOrchestrator] Created ContentNode with ID: %s", contentNodeID)

	// Step 2: Trigger chunking and embedding workflow
	// Note: This is a placeholder for now - actual implementation will depend on
	// the specific requirements. For now, we create an empty task structure
	// that can be extended later without changing the core orchestration.

	if err := e.triggerChunkingEmbedding(ctx, contentNodeID, event); err != nil {
		log.Printf("[EventOrchestrator] Warning: chunking/embedding workflow failed: %v", err)
		// Don't fail the entire process for workflow issues
		// The content node was successfully created
	}

	// TODO: Future operations can be added here if needed:
	// - Update existing nodes: e.nodeRepo.UpdateContentNode(ctx, updatedNode)
	// - Soft delete nodes: e.nodeRepo.SoftDeleteNode(ctx, nodeID)
	// - Create additional links: e.linkRepo.CreateStructuralLinks(ctx, links)

	log.Printf("[EventOrchestrator] Successfully processed document ingestion event")
	return nil
}

// createContentNode creates a ContentNode in the database with proper metadata
func (e *EventOrchestrator) createContentNode(ctx context.Context, event events.DocumentProcessedEvent) (string, error) {
	// Generate new UUID for the content node
	contentNodeID := uuid.New().String()

	// Create ContentNode with current timestamp
	now := timestamppb.New(time.Now())

	// Create BaseNode with required fields
	baseNode := &v1.BaseNode{
		Id:        contentNodeID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Create ContentNode
	contentNode := &v1.ContentNode{
		Base:            baseNode,
		ContentSourceId: event.ContentSourceID.String(),
	}

	// Add optional fields if provided in the event
	if event.ProcessedBlobHash != nil {
		// TODO: Store blob hash in ActionData field as a simple key-value structure
		// For now, this is a placeholder - in a real implementation, you would:
		// 1. Create a structpb.Struct from the blob hash data
		// 2. Set contentNode.ActionData = &structpb.Struct{...}
		// This can be implemented later when the actual blob storage integration is added
		log.Printf("[EventOrchestrator] Blob hash available: %s (storage integration pending)", *event.ProcessedBlobHash)
	}

	// Persist to database
	contentNodes := []*v1.ContentNode{contentNode}
	if err := e.nodeRepo.CreateContentNodes(ctx, contentNodes); err != nil {
		return "", fmt.Errorf("failed to create content node: %w", err)
	}

	return contentNodeID, nil
}

// triggerChunkingEmbedding triggers the chunking and embedding workflow
// This is a placeholder implementation that establishes the pattern for future task execution
func (e *EventOrchestrator) triggerChunkingEmbedding(ctx context.Context, contentNodeID string, event events.DocumentProcessedEvent) error {
	log.Printf("[EventOrchestrator] Triggering chunking/embedding workflow for ContentNode: %s", contentNodeID)

	// TODO: Simple chunking and embedding workflow:
	// 1. Call Python chunking service with document content
	// 2. Create ChunkNodes in Neo4j
	// 3. Create hierarchical links between ContentNode and ChunkNodes

	// For now, this establishes the extensible framework
	// Future tasks will be added here as empty implementations that can be filled in later

	log.Printf("[EventOrchestrator] Chunking/embedding workflow placeholder executed")
	return nil
}
