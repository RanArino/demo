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
	"demo/ms_canvas/go_app/internal/infra"
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
	r2Client      *infra.R2Client
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
	// Initialize R2 client
	var r2Client *infra.R2Client
	if r2Config := config.R2Config; r2Config.Endpoint != "" {
		var err error
		r2Client, err = infra.NewR2Client(r2Config)
		if err != nil {
			log.Printf("[EventOrchestrator] Warning: Failed to initialize R2 client: %v", err)
		} else {
			log.Printf("[EventOrchestrator] R2 client initialized successfully")
		}
	} else {
		log.Printf("[EventOrchestrator] R2 client not initialized - missing configuration")
	}

	return &EventOrchestrator{
		config:        config,
		nodeRepo:      nodeRepo,
		linkRepo:      linkRepo,
		pythonGateway: pythonGateway,
		r2Client:      r2Client,
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
	// - Soft delete nodes: e.nodeRepo.SoftDeleteNodes(ctx, []string{nodeID})
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
		// For now, we have R2 client available but will implement proper storage later
		log.Printf("[EventOrchestrator] Blob hash available: %s (R2 client ready for document fetching)", *event.ProcessedBlobHash)
	}

	// Persist to database
	contentNodes := []*v1.ContentNode{contentNode}
	if err := e.nodeRepo.CreateContentNodes(ctx, contentNodes); err != nil {
		return "", fmt.Errorf("failed to create content node: %w", err)
	}

	return contentNodeID, nil
}

// triggerChunkingEmbedding triggers the chunking and embedding workflow
func (e *EventOrchestrator) triggerChunkingEmbedding(ctx context.Context, contentNodeID string, event events.DocumentProcessedEvent) error {
	log.Printf("[EventOrchestrator] Triggering chunking/embedding workflow for ContentNode: %s", contentNodeID)

	// Check if we have the required components
	if e.r2Client == nil {
		log.Printf("[EventOrchestrator] Warning: R2 client not available, skipping chunking workflow")
		return nil
	}

	if e.pythonGateway == nil {
		log.Printf("[EventOrchestrator] Warning: Python gateway not available, skipping chunking workflow")
		return nil
	}

	if event.ProcessedBlobHash == nil || *event.ProcessedBlobHash == "" {
		log.Printf("[EventOrchestrator] Warning: No processed blob hash available, skipping chunking workflow")
		return nil
	}

	// Step 1: Fetch processed document from R2 storage
	log.Printf("[EventOrchestrator] Fetching processed document from R2: %s", *event.ProcessedBlobHash)
	documentContent, err := e.r2Client.DownloadFile(*event.ProcessedBlobHash)
	if err != nil {
		return fmt.Errorf("failed to fetch processed document from R2: %w", err)
	}

	log.Printf("[EventOrchestrator] Successfully fetched document (%d bytes)", len(documentContent))

	// Step 2: Call Python chunking service
	log.Printf("[EventOrchestrator] Calling Python chunking service...")
	chunkingResponse, err := e.pythonGateway.ChunkDocument(ctx, documentContent, contentNodeID)
	if err != nil {
		return fmt.Errorf("failed to chunk document: %w", err)
	}

	log.Printf("[EventOrchestrator] Python service returned %d chunks", len(chunkingResponse.Chunks))

	// Step 3: Create ChunkNodes in database
	if len(chunkingResponse.Chunks) > 0 {
		err = e.createChunkNodes(ctx, chunkingResponse.Chunks, contentNodeID, event.ContentSourceID.String())
		if err != nil {
			return fmt.Errorf("failed to create chunk nodes: %w", err)
		}
		log.Printf("[EventOrchestrator] Successfully created %d chunk nodes", len(chunkingResponse.Chunks))
	}

	// Step 4: Create hierarchical links between ContentNode and ChunkNodes
	// Note: This would be implemented when we have the actual chunk node IDs
	// For now, we log that this step is ready to be implemented

	log.Printf("[EventOrchestrator] Chunking/embedding workflow completed successfully")
	return nil
}

// createChunkNodes creates ChunkNodes from chunking results
func (e *EventOrchestrator) createChunkNodes(ctx context.Context, chunks []python.ChunkInfo, contentNodeID string, contentSourceID string) error {
	log.Printf("[EventOrchestrator] Creating %d chunk nodes for ContentNode: %s", len(chunks), contentNodeID)

	// Create ChunkNode instances
	chunkNodes := make([]*v1.ChunkNode, len(chunks))
	now := timestamppb.New(time.Now())

	for i, chunk := range chunks {
		// Generate new UUID for the chunk node
		chunkNodeID := uuid.New().String()

		// Create BaseNode with required fields
		baseNode := &v1.BaseNode{
			Id:        chunkNodeID,
			CreatedAt: now,
			UpdatedAt: now,
		}

		// Create ChunkNode
		chunkNodes[i] = &v1.ChunkNode{
			Base:            baseNode,
			ContentSourceId: contentSourceID,
			Content:         chunk.Content,
			SequenceIndex:   chunk.SequenceIndex,
			StartPosition:   &chunk.StartPosition,
			EndPosition:     &chunk.EndPosition,
		}
	}

	// Persist to database
	if err := e.nodeRepo.CreateChunkNodes(ctx, chunkNodes); err != nil {
		return fmt.Errorf("failed to create chunk nodes: %w", err)
	}

	log.Printf("[EventOrchestrator] Successfully created %d chunk nodes", len(chunkNodes))
	return nil
}
