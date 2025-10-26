package server

import (
	"context"
	canvasv1 "demo/ms_canvas/go_app/api/proto/public/v1"
	"demo/ms_canvas/go_app/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Default and safety limit constants
const (
	// Default values for optional parameters
	DefaultTopK         = 25
	DefaultLimitPerNode = 15

	// Safety caps for batch operations to prevent abuse
	MaxNodeIdsBatch               = 100
	MaxNodeUpdatesBatch           = 50
	MaxNeighborIdsBatch           = 50
	MaxStructuralLinksBatch       = 100
	MaxStructuralLinkUpdatesBatch = 50
	MaxStructuralLinkIdsBatch     = 100
)

// canvasPublicServer implements the CanvasPublicServer interface
type canvasPublicServer struct {
	canvasv1.UnimplementedCanvasPublicServer

	searchService service.SearchService
	nodeService   service.NodeService
	linkService   service.LinkService
}

// NewCanvasPublicServer creates a new CanvasPublicServer
func NewCanvasPublicServer(
	searchService service.SearchService,
	nodeService service.NodeService,
	linkService service.LinkService,
) canvasv1.CanvasPublicServer {
	return &canvasPublicServer{
		searchService: searchService,
		nodeService:   nodeService,
		linkService:   linkService,
	}
}

// GetNodes implements the GetNodes RPC
func (s *canvasPublicServer) GetNodes(ctx context.Context, req *canvasv1.GetNodesRequest) (*canvasv1.GetNodesResponse, error) {
	if len(req.Ids) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no node IDs provided")
	}

	// Add safety cap for batch operations
	if len(req.Ids) > MaxNodeIdsBatch {
		return nil, status.Errorf(codes.InvalidArgument, "too many node IDs in batch (max %d)", MaxNodeIdsBatch)
	}

	// Validate node IDs are provided
	for _, idStr := range req.Ids {
		if idStr == "" {
			return nil, status.Error(codes.InvalidArgument, "node ID cannot be empty")
		}
	}

	nodes, err := s.nodeService.GetNodes(ctx, req.Ids)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get nodes: %v", err)
	}

	return &canvasv1.GetNodesResponse{
		Nodes: nodes,
	}, nil
}

// SearchNodes implements the SearchNodes RPC
func (s *canvasPublicServer) SearchNodes(ctx context.Context, req *canvasv1.SearchNodesRequest) (*canvasv1.SearchNodesResponse, error) {
	// Set default values
	if req.Limit <= 0 {
		req.Limit = 100 // Default limit
	}
	if req.Limit > 1000 {
		req.Limit = 1000 // Safety cap for canvas rendering
	}

	if req.Filter == nil {
		return nil, status.Error(codes.InvalidArgument, "filter is required for SearchNodes")
	}

	// spatial_bbox is optional - when not provided, returns all nodes matching filter without spatial constraints
	// When provided, it's used for viewport-based pagination on the canvas

	nodes, err := s.nodeService.SearchNodes(ctx, req.Filter, req.SpatialBbox, req.Limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to search nodes: %v", err)
	}

	return &canvasv1.SearchNodesResponse{
		Nodes: nodes,
	}, nil
}

// UpdateNodes implements the UpdateNodes RPC
func (s *canvasPublicServer) UpdateNodes(ctx context.Context, req *canvasv1.UpdateNodesRequest) (*canvasv1.UpdateNodesResponse, error) {
	if len(req.Nodes) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no node updates provided")
	}

	// Add safety cap for batch operations
	if len(req.Nodes) > MaxNodeUpdatesBatch {
		return nil, status.Errorf(codes.InvalidArgument, "too many node updates in batch (max %d)", MaxNodeUpdatesBatch)
	}

	// Validate nodes have IDs
	for i, node := range req.Nodes {
		if node == nil {
			return nil, status.Errorf(codes.InvalidArgument, "node cannot be nil at index %d", i)
		}
		// Extract ID based on node type
		var id string
		switch n := node.Node.(type) {
		case *canvasv1.Node_Content:
			if n.Content != nil && n.Content.Base != nil {
				id = n.Content.Base.Id
			}
		case *canvasv1.Node_Chunk:
			if n.Chunk != nil && n.Chunk.Base != nil {
				id = n.Chunk.Base.Id
			}
		case *canvasv1.Node_Cluster:
			if n.Cluster != nil && n.Cluster.Base != nil {
				id = n.Cluster.Base.Id
			}
		default:
			return nil, status.Errorf(codes.InvalidArgument, "node ID is required at index %d", i)
		}

		if id == "" {
			return nil, status.Errorf(codes.InvalidArgument, "node ID is required at index %d", i)
		}
	}

	updatedNodes, err := s.nodeService.UpdateNodes(ctx, req.Nodes)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update nodes: %v", err)
	}

	return &canvasv1.UpdateNodesResponse{
		Nodes: updatedNodes,
	}, nil
}

// GetNeighbors implements the GetNeighbors RPC
func (s *canvasPublicServer) GetNeighbors(ctx context.Context, req *canvasv1.GetNeighborsRequest) (*canvasv1.GetNeighborsResponse, error) {
	if len(req.Ids) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no node IDs provided")
	}

	// Add safety cap for batch operations
	if len(req.Ids) > MaxNeighborIdsBatch {
		return nil, status.Errorf(codes.InvalidArgument, "too many node IDs in batch (max %d)", MaxNeighborIdsBatch)
	}

	// Validate node IDs are provided
	for _, idStr := range req.Ids {
		if idStr == "" {
			return nil, status.Error(codes.InvalidArgument, "node ID cannot be empty")
		}
	}

	// Validate limit_per_node and set default if needed
	if req.LimitPerNode < 0 {
		return nil, status.Error(codes.InvalidArgument, "limit_per_node must be non-negative")
	}
	if req.LimitPerNode == 0 {
		req.LimitPerNode = DefaultLimitPerNode // Default limit per node
	}

	neighbors, err := s.nodeService.GetNeighbors(ctx, req.Ids, req.Direction, req.Query, req.LimitPerNode, req.IncludeProperties)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get neighbors: %v", err)
	}

	return &canvasv1.GetNeighborsResponse{
		Results: neighbors,
	}, nil
}

// ListNodesByLink implements the ListNodesByLink RPC.
func (s *canvasPublicServer) ListNodesByLink(ctx context.Context, req *canvasv1.ListNodesByLinkRequest) (*canvasv1.ListNodesByLinkResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if len(req.Parents) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one parent is required")
	}
	if req.Traversal == nil || req.Traversal.Query == nil {
		return nil, status.Error(codes.InvalidArgument, "traversal.query is required")
	}

	if req.Traversal.GetMaxHops() > 1 {
		return nil, status.Error(codes.InvalidArgument, "max_hops greater than 1 is not supported")
	}

	resp, err := s.nodeService.ListNodesByLink(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list nodes by link: %v", err)
	}

	return resp, nil
}

// SemanticSearch implements the SemanticSearch RPC
func (s *canvasPublicServer) SemanticSearch(ctx context.Context, req *canvasv1.SemanticSearchRequest) (*canvasv1.SemanticSearchResponse, error) {
	if req.Query == "" {
		return nil, status.Error(codes.InvalidArgument, "query cannot be empty")
	}

	// Set default top_k if not provided or invalid
	if req.TopK <= 0 {
		req.TopK = DefaultTopK
	}

	// Validate space ID if provided
	if req.SpaceId != "" {
		// For now, just check it's not empty - in real implementation would validate format
	}

	results, err := s.searchService.SemanticSearch(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to perform semantic search: %v", err)
	}

	return results, nil
}

// CreateStructuralLinks implements the CreateStructuralLinks RPC
func (s *canvasPublicServer) CreateStructuralLinks(ctx context.Context, req *canvasv1.CreateStructuralLinksRequest) (*canvasv1.CreateStructuralLinksResponse, error) {
	if len(req.Links) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no structural links provided")
	}

	// Add safety cap for batch operations
	if len(req.Links) > MaxStructuralLinksBatch {
		return nil, status.Errorf(codes.InvalidArgument, "too many links in batch (max %d)", MaxStructuralLinksBatch)
	}

	// Validate links
	for i, link := range req.Links {
		if link.SourceId == "" || link.TargetId == "" {
			return nil, status.Errorf(codes.InvalidArgument, "source_id and target_id are required for link %d", i)
		}
		if link.ConnectionType == canvasv1.StructuralConnectionType_STRUCTURAL_CONNECTION_TYPE_UNSPECIFIED {
			return nil, status.Errorf(codes.InvalidArgument, "connection_type is required for link %d", i)
		}
		if link.CreatedBy == "" {
			return nil, status.Errorf(codes.InvalidArgument, "created_by is required for link %d", i)
		}
		// Validate confidence_score is in valid range [0, 1]
		if link.ConfidenceScore < 0 || link.ConfidenceScore > 1 {
			return nil, status.Errorf(codes.InvalidArgument, "confidence_score must be between 0 and 1 for link %d", i)
		}
	}

	createdLinks, err := s.linkService.CreateStructuralLinks(ctx, req.Links)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create structural links: %v", err)
	}

	return &canvasv1.CreateStructuralLinksResponse{
		Links: createdLinks,
	}, nil
}

// UpdateStructuralLinks implements the UpdateStructuralLinks RPC
func (s *canvasPublicServer) UpdateStructuralLinks(ctx context.Context, req *canvasv1.UpdateStructuralLinksRequest) (*canvasv1.UpdateStructuralLinksResponse, error) {
	if len(req.Updates) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no structural link updates provided")
	}

	// Add safety cap for batch operations
	if len(req.Updates) > MaxStructuralLinkUpdatesBatch {
		return nil, status.Errorf(codes.InvalidArgument, "too many structural link updates in batch (max %d)", MaxStructuralLinkUpdatesBatch)
	}

	// Validate updates
	for i, update := range req.Updates {
		if update.SourceId == "" || update.TargetId == "" {
			return nil, status.Errorf(codes.InvalidArgument, "source_id and target_id are required for update %d", i)
		}

		// Ensure at least one field is set to update
		if update.ConnectionType == nil && update.ConfidenceScore == nil && update.Description == nil &&
			update.ExplorationMetadata == nil && update.StyleMetadata == nil {
			return nil, status.Errorf(codes.InvalidArgument, "no fields set to update for update %d", i)
		}
	}

	updatedLinks, err := s.linkService.UpdateStructuralLinks(ctx, req.Updates)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update structural links: %v", err)
	}

	return &canvasv1.UpdateStructuralLinksResponse{
		Links: updatedLinks,
	}, nil
}

// DeleteStructuralLinks implements the DeleteStructuralLinks RPC
func (s *canvasPublicServer) DeleteStructuralLinks(ctx context.Context, req *canvasv1.DeleteStructuralLinksRequest) (*canvasv1.DeleteStructuralLinksResponse, error) {
	if len(req.LinkIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no structural link IDs provided")
	}

	// Add safety cap for batch operations
	if len(req.LinkIds) > MaxStructuralLinkIdsBatch {
		return nil, status.Errorf(codes.InvalidArgument, "too many structural link IDs in batch (max %d)", MaxStructuralLinkIdsBatch)
	}

	// Validate link IDs
	for i, linkID := range req.LinkIds {
		if linkID.SourceId == "" || linkID.TargetId == "" {
			return nil, status.Errorf(codes.InvalidArgument, "source_id and target_id are required for link ID %d", i)
		}
	}

	deletedCount, err := s.linkService.DeleteStructuralLinks(ctx, req.LinkIds)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete structural links: %v", err)
	}

	return &canvasv1.DeleteStructuralLinksResponse{
		DeletedCount: deletedCount,
	}, nil
}
