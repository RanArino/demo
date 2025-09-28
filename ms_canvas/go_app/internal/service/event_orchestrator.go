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
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// EventOrchestrator implements EventHandler to coordinate document ingestion workflows
type EventOrchestrator struct {
	config        config.Config
	nodeRepo      repository.NodeRepository
	linkRepo      repository.LinkRepository
	pythonGateway ChunkingGateway
	r2Client      R2Client
}

// EventOrchestrator is focused ONLY on consuming events from other microservices
// It does NOT produce events or use complex task management
// All Task/TaskManager functionality is in task_executor.go for future use when needed

type ChunkingGateway interface {
	ChunkDocument(ctx context.Context, documentContent []byte, contentNodeID string) (*python.ChunkDocumentResponse, error)
}

type R2Client interface {
	DownloadFile(key string) ([]byte, error)
}

// NewEventOrchestrator creates a new event orchestrator with dependencies
func NewEventOrchestrator(
	config config.Config,
	nodeRepo repository.NodeRepository,
	linkRepo repository.LinkRepository,
	pythonGateway ChunkingGateway,
) EventHandler {
	var r2Client R2Client
	var err error
	if r2Config := config.R2Config; r2Config.Endpoint != "" {
		r2Client, err = infra.NewR2Client(r2Config)
		if err != nil {
			log.Printf("[EventOrchestrator] Warning: Failed to initialize R2 client: %v", err)
		} else {
			log.Printf("[EventOrchestrator] R2 client initialized successfully")
		}
	} else {
		log.Printf("[EventOrchestrator] R2 client not initialized - missing configuration")
	}

	if pythonGateway == nil {
		pythonGateway = &python.Gateway{}
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
	contentNodeID := uuid.New().String()
	now := timestamppb.New(time.Now())

	baseNode := &v1.BaseNode{
		Id:        contentNodeID,
		SpaceId:   event.SpaceID.String(),
		CreatedAt: now,
		UpdatedAt: now,
		Keywords:  event.Keywords,
	}

	if event.Summary != "" {
		baseNode.ChatContent = &event.Summary
	}

	contentNode := &v1.ContentNode{
		Base:            baseNode,
		ContentSourceId: event.ContentSourceID.String(),
	}

	if event.Title != "" {
		contentNode.Title = &event.Title
	}

	if event.Summary != "" {
		contentNode.Source = &event.Summary
	}

	if len(event.Keywords) > 0 {
		actionData, err := structpb.NewStruct(map[string]interface{}{
			"keywords": event.Keywords,
		})
		if err == nil {
			contentNode.ActionData = actionData
		} else {
			log.Printf("[EventOrchestrator] Warning: failed to build action data: %v", err)
		}
	}

	if event.ProcessedBlobHash != nil {
		if contentNode.ActionData == nil {
			contentNode.ActionData, _ = structpb.NewStruct(map[string]interface{}{})
		}
		if contentNode.ActionData != nil {
			contentNode.ActionData.Fields["processed_blob_hash"] = structpb.NewStringValue(*event.ProcessedBlobHash)
		}
	}

	if err := e.nodeRepo.CreateContentNodes(ctx, []*v1.ContentNode{contentNode}); err != nil {
		return "", fmt.Errorf("failed to create content node: %w", err)
	}

	return contentNodeID, nil
}

// triggerChunkingEmbedding triggers the chunking and embedding workflow
func (e *EventOrchestrator) triggerChunkingEmbedding(ctx context.Context, contentNodeID string, event events.DocumentProcessedEvent) error {
	log.Printf("[EventOrchestrator] Triggering chunking/embedding workflow for ContentNode: %s", contentNodeID)

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

	log.Printf("[EventOrchestrator] Fetching processed document from R2: %s", *event.ProcessedBlobHash)
	documentContent, err := e.r2Client.DownloadFile(*event.ProcessedBlobHash)
	if err != nil {
		return fmt.Errorf("failed to fetch processed document from R2: %w", err)
	}

	log.Printf("[EventOrchestrator] Successfully fetched document (%d bytes)", len(documentContent))

	log.Printf("[EventOrchestrator] Calling Python chunking service...")
	chunkingResponse, err := e.pythonGateway.ChunkDocument(ctx, documentContent, contentNodeID)
	if err != nil {
		return fmt.Errorf("failed to chunk document: %w", err)
	}

	log.Printf("[EventOrchestrator] Python service returned %d chunks", len(chunkingResponse.Chunks))

	if len(chunkingResponse.Chunks) > 0 {
		err = e.createChunkNodes(ctx, chunkingResponse.Chunks, contentNodeID, event.ContentSourceID.String(), event)
		if err != nil {
			return fmt.Errorf("failed to create chunk nodes: %w", err)
		}
		log.Printf("[EventOrchestrator] Successfully created %d chunk nodes", len(chunkingResponse.Chunks))
	}

	log.Printf("[EventOrchestrator] Chunking/embedding workflow completed successfully")
	return nil
}

// createChunkNodes creates ChunkNodes from chunking results
func (e *EventOrchestrator) createChunkNodes(ctx context.Context, chunks []python.ChunkInfo, contentNodeID string, contentSourceID string, event events.DocumentProcessedEvent) error {
	log.Printf("[EventOrchestrator] Creating %d chunk nodes for ContentNode: %s", len(chunks), contentNodeID)

	chunkNodes := make([]*v1.ChunkNode, len(chunks))
	now := timestamppb.New(time.Now())

	for i, chunk := range chunks {
		chunkNodeID := uuid.New().String()

		baseNode := &v1.BaseNode{
			Id:        chunkNodeID,
			SpaceId:   event.SpaceID.String(),
			CreatedAt: now,
			UpdatedAt: now,
			Keywords:  event.Keywords,
		}

		if event.Title != "" {
			baseNode.ContextType = event.Title
		}

		chunkNodes[i] = &v1.ChunkNode{
			Base:            baseNode,
			ContentSourceId: contentSourceID,
			Content:         chunk.Content,
			SequenceIndex:   chunk.SequenceIndex,
			StartPosition:   &chunk.StartPosition,
			EndPosition:     &chunk.EndPosition,
		}
	}

	if err := e.nodeRepo.CreateChunkNodes(ctx, chunkNodes); err != nil {
		return fmt.Errorf("failed to create chunk nodes: %w", err)
	}

	log.Printf("[EventOrchestrator] Successfully created %d chunk nodes", len(chunkNodes))
	return nil
}
