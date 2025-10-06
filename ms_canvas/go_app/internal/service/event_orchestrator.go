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
	"demo/ms_canvas/go_app/internal/metrics"
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
	// Update status to PROCESSING before starting
	if err := e.updateChunkingStatus(ctx, contentNodeID, v1.ChunkingStatus_CHUNKING_STATUS_PROCESSING, ""); err != nil {
		log.Printf("[EventOrchestrator] Warning: failed to update chunking status to PROCESSING: %v", err)
	}

	if err := e.triggerChunkingEmbedding(ctx, contentNodeID, event); err != nil {
		log.Printf("[EventOrchestrator] Error: chunking/embedding workflow failed: %v", err)
		// Update status to FAILED with error message
		if updateErr := e.updateChunkingStatus(ctx, contentNodeID, v1.ChunkingStatus_CHUNKING_STATUS_FAILED, err.Error()); updateErr != nil {
			log.Printf("[EventOrchestrator] Warning: failed to update chunking status to FAILED: %v", updateErr)
		}

		// Schedule automatic retry if under max retry limit
		retryCount := e.getRetryCountForEvent(event)
		if retryCount < 3 { // Max 3 retries
			log.Printf("[EventOrchestrator] Scheduling retry %d/3 for ContentNode: %s", retryCount+1, contentNodeID)
			go e.retryChunkingAfterDelay(ctx, contentNodeID, event, retryCount)
		} else {
			log.Printf("[EventOrchestrator] Max retries (3) exceeded for ContentNode: %s", contentNodeID)
		}
		// Don't fail the entire process - the content node was successfully created
	} else {
		// Verify hierarchical links were created successfully before marking as completed
		if err := e.verifyHierarchicalLinksIntegrity(ctx, contentNodeID); err != nil {
			log.Printf("[EventOrchestrator] Warning: hierarchical links verification failed for ContentNode %s: %v", contentNodeID, err)
			// Don't fail the process, but log the issue for monitoring
		}

		// Update status to COMPLETED on success
		if err := e.updateChunkingStatus(ctx, contentNodeID, v1.ChunkingStatus_CHUNKING_STATUS_COMPLETED, ""); err != nil {
			log.Printf("[EventOrchestrator] Warning: failed to update chunking status to COMPLETED: %v", err)
		}
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

	// Set initial chunking status to PENDING
	chunkingStatus := v1.ChunkingStatus_CHUNKING_STATUS_PENDING

	contentNode := &v1.ContentNode{
		Base:            baseNode,
		ContentSourceId: event.ContentSourceID.String(),
		ChunkingStatus:  &chunkingStatus,
	}

	if event.Title != "" {
		contentNode.Title = &event.Title
	}

	// if event.Source != "" {
	// 	contentNode.Source = &event.Source
	// }

	if len(event.Keywords) > 0 {
		keywords := make([]interface{}, len(event.Keywords))
		for i, kw := range event.Keywords {
			keywords[i] = kw
		}
		actionData, err := structpb.NewStruct(map[string]interface{}{
			"keywords": keywords,
		})
		if err == nil {
			contentNode.ActionData = actionData
		} else {
			log.Printf("[EventOrchestrator] Warning: failed to build action data: %v", err)
		}
	}

	if event.ProcessedObjectKey != nil {
		if contentNode.ActionData == nil {
			contentNode.ActionData, _ = structpb.NewStruct(map[string]interface{}{})
		}
		if contentNode.ActionData != nil {
			contentNode.ActionData.Fields["processed_object_key"] = structpb.NewStringValue(*event.ProcessedObjectKey)
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
	start := time.Now()
	defer func() {
		metrics.ChunkingDuration.Observe(time.Since(start).Seconds())
	}()

	log.Printf("[EventOrchestrator] Triggering chunking/embedding workflow for ContentNode: %s", contentNodeID)

	if e.r2Client == nil {
		log.Printf("[EventOrchestrator] Warning: R2 client not available, skipping chunking workflow")
		return nil
	}

	if e.pythonGateway == nil {
		log.Printf("[EventOrchestrator] Warning: Python gateway not available, skipping chunking workflow")
		return nil
	}

	// Validate that ProcessedObjectKey is available for R2 download
	// Note: ProcessedBlobHash is a SHA256 hash, not an R2 object key, so we cannot use it for downloads
	if event.ProcessedObjectKey == nil || *event.ProcessedObjectKey == "" {
		log.Printf("[EventOrchestrator] Error: Missing processed_object_key for content_source_id=%s, cannot proceed with chunking workflow",
			event.ContentSourceID)
		return fmt.Errorf("processed_object_key is required for chunking workflow but was missing for content_source_id=%s",
			event.ContentSourceID)
	}

	downloadKey := *event.ProcessedObjectKey

	log.Printf("[EventOrchestrator] Fetching processed document from R2: %s", downloadKey)
	documentContent, err := e.r2Client.DownloadFile(downloadKey)
	if err != nil {
		metrics.ChunkingTotal.WithLabelValues("failed").Inc()
		return fmt.Errorf("failed to fetch processed document from R2: %w", err)
	}

	log.Printf("[EventOrchestrator] Successfully fetched document (%d bytes)", len(documentContent))

	log.Printf("[EventOrchestrator] Calling Python chunking service...")
	chunkingResponse, err := e.pythonGateway.ChunkDocument(ctx, documentContent, contentNodeID)
	if err != nil {
		metrics.ChunkingTotal.WithLabelValues("failed").Inc()
		return fmt.Errorf("failed to chunk document: %w", err)
	}

	log.Printf("[EventOrchestrator] Python service returned %d chunks", len(chunkingResponse.Chunks))

	if len(chunkingResponse.Chunks) > 0 {
		err = e.createChunkNodes(ctx, chunkingResponse.Chunks, contentNodeID, event.ContentSourceID.String(), event)
		if err != nil {
			metrics.ChunkingTotal.WithLabelValues("failed").Inc()
			return fmt.Errorf("failed to create chunk nodes: %w", err)
		}
		log.Printf("[EventOrchestrator] Successfully created %d chunk nodes", len(chunkingResponse.Chunks))
	}

	log.Printf("[EventOrchestrator] Chunking/embedding workflow completed successfully")
	metrics.ChunkingTotal.WithLabelValues("completed").Inc()
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
			Id:          chunkNodeID,
			SpaceId:     event.SpaceID.String(),
			CreatedAt:   now,
			UpdatedAt:   now,
			Keywords:    event.Keywords,
			ContextType: "chunk",
		}

		// store chunk text on BaseNode.chat_content
		if chunk.Content != "" {
			baseNode.ChatContent = &chunk.Content
		}

		chunkNodes[i] = &v1.ChunkNode{
			Base:            baseNode,
			ContentSourceId: contentSourceID,
			SequenceIndex:   chunk.SequenceIndex,
			StartPosition:   &chunk.StartPosition,
			EndPosition:     &chunk.EndPosition,
		}
	}

	if err := e.nodeRepo.CreateChunkNodes(ctx, chunkNodes); err != nil {
		return fmt.Errorf("failed to create chunk nodes: %w", err)
	}

	// Extract chunk node IDs for hierarchical linking
	chunkNodeIDs := make([]string, len(chunkNodes))
	for i, chunkNode := range chunkNodes {
		chunkNodeIDs[i] = chunkNode.Base.Id
	}

	// Create hierarchical links between ContentNode and ChunkNodes
	if err := e.createHierarchicalLinks(ctx, contentNodeID, chunkNodeIDs); err != nil {
		log.Printf("[EventOrchestrator] Warning: failed to create hierarchical links for ContentNode %s: %v", contentNodeID, err)
		// Don't fail the entire process - chunks were successfully created
	}

	log.Printf("[EventOrchestrator] Successfully created %d chunk nodes", len(chunkNodes))
	return nil
}

// createHierarchicalLinks creates hierarchical links between a ContentNode and its ChunkNodes
// This method is idempotent and handles retry scenarios gracefully by using MERGE operations
func (e *EventOrchestrator) createHierarchicalLinks(ctx context.Context, contentNodeID string, chunkNodeIDs []string) error {
	if len(chunkNodeIDs) == 0 {
		log.Printf("[EventOrchestrator] No chunk nodes to link for ContentNode: %s", contentNodeID)
		return nil
	}

	log.Printf("[EventOrchestrator] Creating hierarchical links for ContentNode %s with %d chunk nodes", contentNodeID, len(chunkNodeIDs))

	now := timestamppb.New(time.Now())
	links := make([]*v1.HierarchicalLink, len(chunkNodeIDs))

	for i, chunkNodeID := range chunkNodeIDs {
		linkID := uuid.New().String()

		links[i] = &v1.HierarchicalLink{
			Base: &v1.BaseLink{
				Id:        linkID,
				SourceId:  contentNodeID,
				TargetId:  chunkNodeID,
				CreatedAt: now,
				UpdatedAt: now,
			},
			ConnectionType: v1.HierarchicalConnectionType_HIERARCHICAL_CONNECTION_TYPE_ABSTRACTION,
			HierarchyDepth: 1,
		}
	}

	// Use repository's MERGE-based operation for idempotent link creation
	// This ensures that duplicate links are handled gracefully during retries
	if err := e.linkRepo.CreateHierarchicalLinks(ctx, links); err != nil {
		// Log detailed error information for troubleshooting
		log.Printf("[EventOrchestrator] Error creating hierarchical links for ContentNode %s: %v", contentNodeID, err)
		return fmt.Errorf("failed to create hierarchical links for ContentNode %s: %w", contentNodeID, err)
	}

	log.Printf("[EventOrchestrator] Successfully created/updated %d hierarchical links for ContentNode: %s", len(links), contentNodeID)
	return nil
}

// updateChunkingStatus updates the chunking status of a ContentNode
func (e *EventOrchestrator) updateChunkingStatus(ctx context.Context, contentNodeID string, status v1.ChunkingStatus, errorMsg string) error {
	// Fetch the existing content node first
	nodes, err := e.nodeRepo.GetNodes(ctx, []string{contentNodeID}, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch content node for status update: %w", err)
	}

	if len(nodes) == 0 {
		return fmt.Errorf("content node not found: %s", contentNodeID)
	}

	node := nodes[0]
	contentNode, ok := node.Node.(*v1.Node_Content)
	if !ok {
		return fmt.Errorf("node is not a content node: %s", contentNodeID)
	}

	// Update chunking status
	contentNode.Content.ChunkingStatus = &status
	if errorMsg != "" {
		contentNode.Content.ChunkingError = &errorMsg
	} else {
		contentNode.Content.ChunkingError = nil
	}

	// Update the node in the database
	if err := e.nodeRepo.UpdateContentNode(ctx, contentNode.Content); err != nil {
		return fmt.Errorf("failed to update content node status: %w", err)
	}

	log.Printf("[EventOrchestrator] Updated ContentNode %s chunking status to %s", contentNodeID, status.String())
	return nil
}

// retryChunkingAfterDelay schedules automatic retry for failed chunking with exponential backoff
func (e *EventOrchestrator) retryChunkingAfterDelay(ctx context.Context, contentNodeID string, event events.DocumentProcessedEvent, retryCount int) {
	// Exponential backoff: 1s, 2s, 4s
	delay := time.Duration(1<<retryCount) * time.Second
	log.Printf("[EventOrchestrator] Waiting %v before retry %d for ContentNode: %s", delay, retryCount+1, contentNodeID)

	// Wait for backoff duration or context cancellation
	select {
	case <-time.After(delay):
		// continue with retry
	case <-ctx.Done():
		log.Printf("[EventOrchestrator] Retry for ContentNode %s cancelled due to context done", contentNodeID)
		return
	}

	// Increment retry metrics
	metrics.ChunkingRetries.Inc()

	// Update status to PROCESSING before retry
	if err := e.updateChunkingStatus(ctx, contentNodeID, v1.ChunkingStatus_CHUNKING_STATUS_PROCESSING, ""); err != nil {
		log.Printf("[EventOrchestrator] Warning: failed to update status before retry: %v", err)
	}

	// Clean up any existing chunk nodes and hierarchical links before retry
	// This ensures idempotency and prevents duplicate nodes/links
	if err := e.cleanupExistingChunkData(ctx, contentNodeID); err != nil {
		log.Printf("[EventOrchestrator] Warning: failed to cleanup existing chunk data before retry: %v", err)
		// Continue with retry even if cleanup fails
	}

	// Retry chunking
	log.Printf("[EventOrchestrator] Attempting retry %d/3 for ContentNode: %s", retryCount+1, contentNodeID)
	if err := e.triggerChunkingEmbedding(ctx, contentNodeID, event); err != nil {
		log.Printf("[EventOrchestrator] Retry %d failed for ContentNode %s: %v", retryCount+1, contentNodeID, err)

		// Update status to FAILED
		if updateErr := e.updateChunkingStatus(ctx, contentNodeID, v1.ChunkingStatus_CHUNKING_STATUS_FAILED, err.Error()); updateErr != nil {
			log.Printf("[EventOrchestrator] Warning: failed to update status after retry failure: %v", updateErr)
		}

		// Schedule next retry if under limit
		if retryCount+1 < 3 {
			log.Printf("[EventOrchestrator] Scheduling retry %d/3 for ContentNode: %s", retryCount+2, contentNodeID)
			go e.retryChunkingAfterDelay(ctx, contentNodeID, event, retryCount+1)
		} else {
			log.Printf("[EventOrchestrator] Max retries (3) reached for ContentNode: %s", contentNodeID)
		}
	} else {
		// Success! Verify hierarchical links before marking as completed
		if err := e.verifyHierarchicalLinksIntegrity(ctx, contentNodeID); err != nil {
			log.Printf("[EventOrchestrator] Warning: hierarchical links verification failed after retry %d for ContentNode %s: %v", retryCount+1, contentNodeID, err)
		}

		log.Printf("[EventOrchestrator] Retry %d succeeded for ContentNode: %s", retryCount+1, contentNodeID)
		if err := e.updateChunkingStatus(ctx, contentNodeID, v1.ChunkingStatus_CHUNKING_STATUS_COMPLETED, ""); err != nil {
			log.Printf("[EventOrchestrator] Warning: failed to update status after retry success: %v", err)
		}
	}
}

// cleanupExistingChunkData removes existing chunk nodes and their hierarchical links
// for a ContentNode to ensure idempotency during retries
func (e *EventOrchestrator) cleanupExistingChunkData(ctx context.Context, contentNodeID string) error {
	log.Printf("[EventOrchestrator] Cleaning up existing chunk data for ContentNode: %s", contentNodeID)

	// First, get the ContentNode to find its content_source_id
	contentNodes, err := e.nodeRepo.GetNodes(ctx, []string{contentNodeID}, nil)
	if err != nil {
		return fmt.Errorf("failed to get ContentNode for cleanup: %w", err)
	}
	if len(contentNodes) == 0 {
		log.Printf("[EventOrchestrator] ContentNode %s not found during cleanup", contentNodeID)
		return nil
	}

	contentNode, ok := contentNodes[0].Node.(*v1.Node_Content)
	if !ok {
		return fmt.Errorf("node %s is not a ContentNode", contentNodeID)
	}

	contentSourceID := contentNode.Content.ContentSourceId

	// Find existing chunk nodes for this content source
	chunkFilter := &v1.NodeFilter{
		ChunkFilter: &v1.ChunkNodeFilter{
			ContentSourceId: &contentSourceID,
		},
	}

	existingChunkNodes, err := e.nodeRepo.SearchNodes(ctx, chunkFilter, nil, 1000) // Use high limit to get all chunks
	if err != nil {
		return fmt.Errorf("failed to search existing chunk nodes: %w", err)
	}

	if len(existingChunkNodes) == 0 {
		log.Printf("[EventOrchestrator] No existing chunk nodes found for ContentNode: %s", contentNodeID)
		return nil
	}

	// Extract chunk node IDs
	chunkNodeIDs := make([]string, len(existingChunkNodes))
	for i, node := range existingChunkNodes {
		if chunkNode := node.GetChunk(); chunkNode != nil {
			chunkNodeIDs[i] = chunkNode.Base.Id
		}
	}

	log.Printf("[EventOrchestrator] Found %d existing chunk nodes to cleanup for ContentNode: %s", len(chunkNodeIDs), contentNodeID)

	// Delete hierarchical links first (this will also clean up any existing hierarchical links)
	if err := e.linkRepo.DeleteLinksForNodes(ctx, chunkNodeIDs); err != nil {
		log.Printf("[EventOrchestrator] Warning: failed to delete links for chunk nodes: %v", err)
		// Continue with node deletion even if link deletion fails
	}

	// Soft delete the chunk nodes
	if err := e.nodeRepo.SoftDeleteNodes(ctx, chunkNodeIDs); err != nil {
		return fmt.Errorf("failed to soft delete existing chunk nodes: %w", err)
	}

	log.Printf("[EventOrchestrator] Successfully cleaned up %d chunk nodes and their links for ContentNode: %s", len(chunkNodeIDs), contentNodeID)
	return nil
}

// verifyHierarchicalLinksIntegrity checks that hierarchical links exist between
// a ContentNode and its ChunkNodes to ensure the chunking process completed successfully
func (e *EventOrchestrator) verifyHierarchicalLinksIntegrity(ctx context.Context, contentNodeID string) error {
	log.Printf("[EventOrchestrator] Verifying hierarchical links integrity for ContentNode: %s", contentNodeID)

	// Get hierarchical links from this ContentNode
	linkQuery := &v1.LinkQuery{
		LinkTypes: []v1.LinkType{v1.LinkType_LINK_TYPE_HIERARCHICAL},
		Filter: &v1.LinkFilter{
			Filter: &v1.LinkFilter_Hierarchical{
				Hierarchical: &v1.HierarchicalLinkFilter{
					Base: &v1.BaseLinkFilter{
						SourceId: &contentNodeID,
					},
				},
			},
		},
	}

	links, err := e.linkRepo.GetLinksByNodes(ctx, []string{contentNodeID}, v1.Direction_DIRECTION_OUTGOING, linkQuery)
	if err != nil {
		return fmt.Errorf("failed to get hierarchical links for verification: %w", err)
	}

	hierarchicalLinkCount := 0
	for _, link := range links {
		if hierarchicalLink := link.GetHierarchical(); hierarchicalLink != nil {
			if hierarchicalLink.ConnectionType == v1.HierarchicalConnectionType_HIERARCHICAL_CONNECTION_TYPE_ABSTRACTION {
				hierarchicalLinkCount++
			}
		}
	}

	log.Printf("[EventOrchestrator] Found %d hierarchical abstraction links for ContentNode: %s", hierarchicalLinkCount, contentNodeID)

	// If no hierarchical links found, this might indicate an issue
	if hierarchicalLinkCount == 0 {
		return fmt.Errorf("no hierarchical abstraction links found for ContentNode %s", contentNodeID)
	}

	return nil
}

// getRetryCountForEvent extracts retry count from event metadata (always 0 for initial events)
func (e *EventOrchestrator) getRetryCountForEvent(event events.DocumentProcessedEvent) int {
	// Initial events don't have retry count, so always return 0
	// Retry count is tracked internally in the goroutine chain
	return 0
}
