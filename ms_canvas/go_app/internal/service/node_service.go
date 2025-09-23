package service

import (
	"context"
	"fmt"
	"log"
	"sync"

	v1 "demo/ms_canvas/go_app/api/proto/public/v1"
	"demo/ms_canvas/go_app/internal/repository/neo4j"

	"github.com/google/uuid"
)

// NodeService handles node CRUD operations

// nodeServiceImpl implements NodeService
type nodeServiceImpl struct {
	nodeRepo *neo4j.NodeRepo
	linkRepo *neo4j.LinkRepo
}

// NewNodeService creates a new NodeService with repository dependencies
func NewNodeService(nodeRepo *neo4j.NodeRepo, linkRepo *neo4j.LinkRepo) NodeService {
	return &nodeServiceImpl{
		nodeRepo: nodeRepo,
		linkRepo: linkRepo,
	}
}

func (s *nodeServiceImpl) GetNodes(ctx context.Context, ids []string) ([]*v1.Node, error) {
	// Parse all IDs first
	var parseErrors []string

	for _, idStr := range ids {
		_, err := uuid.Parse(idStr)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("invalid node ID %s: %v", idStr, err))
			continue
		}
	}

	// Get nodes from repository (need to convert from v1.Node to v1.Node)
	v1Nodes, err := s.nodeRepo.GetNodes(ctx, ids, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %v", err)
	}

	// Convert v1.Node to v1.Node concurrently
	type conversionResult struct {
		index      int
		publicNode *v1.Node
		err        error
	}

	results := make(chan conversionResult, len(v1Nodes))
	var wg sync.WaitGroup

	// Start concurrent conversions
	for i, v1Node := range v1Nodes {
		wg.Add(1)
		go func(index int, node *v1.Node) {
			defer wg.Done()
			publicNode, err := s.convertV1NodeToPublic(node)
			results <- conversionResult{index: index, publicNode: publicNode, err: err}
		}(i, v1Node)
	}

	// Wait for all conversions to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results maintaining original order
	publicNodes := make([]*v1.Node, len(v1Nodes))
	conversionErrors := 0

	for result := range results {
		if result.err != nil {
			log.Printf("Failed to convert node %s to public format: %v", getNodeID(v1Nodes[result.index]), result.err)
			conversionErrors++
			continue
		}
		publicNodes[result.index] = result.publicNode
	}

	// Log aggregated errors for debugging
	if len(parseErrors) > 0 {
		log.Printf("GetNodes completed with %d parse errors", len(parseErrors))
	}
	if conversionErrors > 0 {
		log.Printf("GetNodes completed with %d conversion errors", conversionErrors)
	}

	return publicNodes, nil
}

func (s *nodeServiceImpl) GetNeighbors(ctx context.Context, ids []string, direction v1.Direction, query *v1.LinkQuery, limitPerNode int32, includeProperties bool) (map[string]*v1.NeighborsList, error) {
	// Step 1.1: Fetch all links for all requested nodes in one call using LinkQuery
	allLinks, err := s.linkRepo.GetLinksByNodes(ctx, ids, direction, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get links by nodes: %v", err)
	}

	// Step 1.2: Gather all unique neighbor node IDs
	neighborIDs := make(map[string]bool)
	for _, link := range allLinks {
		// Access base link through the specific link type
		var sourceID, targetID string
		if hierarchicalLink := link.GetHierarchical(); hierarchicalLink != nil {
			if base := hierarchicalLink.GetBase(); base != nil {
				sourceID = base.GetSourceId()
				targetID = base.GetTargetId()
			}
		} else if semanticLink := link.GetSemantic(); semanticLink != nil {
			if base := semanticLink.GetBase(); base != nil {
				sourceID = base.GetSourceId()
				targetID = base.GetTargetId()
			}
		} else if structuralLink := link.GetStructural(); structuralLink != nil {
			if base := structuralLink.GetBase(); base != nil {
				sourceID = base.GetSourceId()
				targetID = base.GetTargetId()
			}
		}

		if sourceID != "" {
			neighborIDs[sourceID] = true
		}
		if targetID != "" {
			neighborIDs[targetID] = true
		}
	}

	// Remove the original source nodes from the map of neighbors to fetch
	for _, id := range ids {
		delete(neighborIDs, id)
	}

	// Convert neighborIDs map to slice
	neighborIDSlice := make([]string, 0, len(neighborIDs))
	for id := range neighborIDs {
		neighborIDSlice = append(neighborIDSlice, id)
	}

	// Step 1.3: Batch fetch all neighbor node data
	neighborNodes, err := s.nodeRepo.GetNodes(ctx, neighborIDSlice, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get neighbor nodes: %v", err)
	}

	// Step 1.4: Create a map for efficient node lookup
	nodeMap := make(map[string]*v1.Node, len(neighborNodes))
	for _, node := range neighborNodes {
		if node != nil {
			nodeMap[getNodeID(node)] = node
		}
	}

	// Step 1.5: Reconstruct the response
	results := make(map[string]*v1.NeighborsList)
	for _, id := range ids {
		results[id] = &v1.NeighborsList{Neighbors: []*v1.Neighbor{}}
	}

	// Process each link and add neighbors to the appropriate source node
	for _, link := range allLinks {
		// Access base link through the specific link type
		var sourceID, targetID string
		if hierarchicalLink := link.GetHierarchical(); hierarchicalLink != nil {
			if base := hierarchicalLink.GetBase(); base != nil {
				sourceID = base.GetSourceId()
				targetID = base.GetTargetId()
			}
		} else if semanticLink := link.GetSemantic(); semanticLink != nil {
			if base := semanticLink.GetBase(); base != nil {
				sourceID = base.GetSourceId()
				targetID = base.GetTargetId()
			}
		} else if structuralLink := link.GetStructural(); structuralLink != nil {
			if base := structuralLink.GetBase(); base != nil {
				sourceID = base.GetSourceId()
				targetID = base.GetTargetId()
			}
		}

		if sourceID == "" || targetID == "" {
			continue
		}

		// Determine which node is the neighbor based on direction
		var neighborID string
		if sourceID != "" && (sourceID != targetID) {
			// For outgoing links, the target is the neighbor
			neighborID = targetID
		} else if targetID != "" {
			// For incoming links, the source is the neighbor
			neighborID = sourceID
		}

		if neighborNode, exists := nodeMap[neighborID]; exists {
			// Get link type from the link
			var linkType v1.LinkType
			if link.GetHierarchical() != nil {
				linkType = v1.LinkType_LINK_TYPE_HIERARCHICAL
			} else if link.GetSemantic() != nil {
				linkType = v1.LinkType_LINK_TYPE_SEMANTIC
			} else if link.GetStructural() != nil {
				linkType = v1.LinkType_LINK_TYPE_STRUCTURAL
			}

			// Add neighbor to the appropriate source node
			if _, ok := results[sourceID]; ok {
				neighbor := &v1.Neighbor{
					Node:     neighborNode,
					LinkType: linkType,
					Link:     link,
				}
				results[sourceID].Neighbors = append(results[sourceID].Neighbors, neighbor)
			}
		}
	}

	return results, nil
}

func (s *nodeServiceImpl) UpdateNodes(ctx context.Context, nodes []*v1.Node) ([]*v1.Node, error) {
	// Separate nodes by type for specific repository methods
	var contentNodes []*v1.ContentNode
	var chunkNodes []*v1.ChunkNode
	var clusterNodes []*v1.ClusterNode

	for _, publicNode := range nodes {
		switch n := publicNode.Node.(type) {
		case *v1.Node_Content:
			if n.Content != nil {
				contentNodes = append(contentNodes, n.Content)
			}
		case *v1.Node_Chunk:
			if n.Chunk != nil {
				chunkNodes = append(chunkNodes, n.Chunk)
			}
		case *v1.Node_Cluster:
			if n.Cluster != nil {
				clusterNodes = append(clusterNodes, n.Cluster)
			}
		default:
			log.Printf("Unsupported node type for update")
			continue
		}
	}

	// Update nodes using specific repository methods concurrently
	var updateErrors []error
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Helper function to update nodes of a specific type concurrently
	updateNodeType := func(nodeType string, nodeList interface{}, updateFunc func(ctx context.Context, node interface{}) error) {
		defer wg.Done()

		switch nodes := nodeList.(type) {
		case []*v1.ContentNode:
			for _, node := range nodes {
				wg.Add(1)
				go func(n *v1.ContentNode) {
					defer wg.Done()
					if err := updateFunc(ctx, n); err != nil {
						mu.Lock()
						updateErrors = append(updateErrors, fmt.Errorf("failed to update %s node %s: %v", nodeType, n.Base.Id, err))
						mu.Unlock()
					}
				}(node)
			}
		case []*v1.ChunkNode:
			for _, node := range nodes {
				wg.Add(1)
				go func(n *v1.ChunkNode) {
					defer wg.Done()
					if err := updateFunc(ctx, n); err != nil {
						mu.Lock()
						updateErrors = append(updateErrors, fmt.Errorf("failed to update %s node %s: %v", nodeType, n.Base.Id, err))
						mu.Unlock()
					}
				}(node)
			}
		case []*v1.ClusterNode:
			for _, node := range nodes {
				wg.Add(1)
				go func(n *v1.ClusterNode) {
					defer wg.Done()
					if err := updateFunc(ctx, n); err != nil {
						mu.Lock()
						updateErrors = append(updateErrors, fmt.Errorf("failed to update %s node %s: %v", nodeType, n.Base.Id, err))
						mu.Unlock()
					}
				}(node)
			}
		}
	}

	// Start concurrent updates for each node type
	if len(contentNodes) > 0 {
		wg.Add(1)
		go updateNodeType("content", contentNodes, func(ctx context.Context, node interface{}) error {
			return s.nodeRepo.UpdateContentNode(ctx, node.(*v1.ContentNode))
		})
	}

	if len(chunkNodes) > 0 {
		wg.Add(1)
		go updateNodeType("chunk", chunkNodes, func(ctx context.Context, node interface{}) error {
			return s.nodeRepo.UpdateChunkNode(ctx, node.(*v1.ChunkNode))
		})
	}

	if len(clusterNodes) > 0 {
		wg.Add(1)
		go updateNodeType("cluster", clusterNodes, func(ctx context.Context, node interface{}) error {
			return s.nodeRepo.UpdateClusterNode(ctx, node.(*v1.ClusterNode))
		})
	}

	// Wait for all updates to complete
	wg.Wait()

	// If there were errors, return them
	if len(updateErrors) > 0 {
		return nil, fmt.Errorf("multiple update errors: %v", updateErrors)
	}

	// Return the updated nodes (same reference since they're updated in place)
	updatedPublicNodes := make([]*v1.Node, 0, len(nodes))
	for _, publicNode := range nodes {
		updatedPublicNodes = append(updatedPublicNodes, publicNode)
	}

	return updatedPublicNodes, nil
}

func (s *nodeServiceImpl) SearchNodes(ctx context.Context, filter *v1.NodeFilter, spatialBBox *v1.SpatialBoundingBox, limit int32) ([]*v1.Node, error) {
	// Call the repository's SearchNodes method
	nodes, err := s.nodeRepo.SearchNodes(ctx, filter, spatialBBox, limit)
	if err != nil {
		log.Printf("SearchNodes failed: %v", err)
		return nil, fmt.Errorf("failed to search nodes: %w", err)
	}

	// Log the search operation for debugging
	log.Printf("SearchNodes completed successfully. Found %d nodes", len(nodes))

	return nodes, nil
}

// Helper functions for node conversion

// convertV1NodeToPublic converts a v1.Node to v1.Node
func (s *nodeServiceImpl) convertV1NodeToPublic(v1Node *v1.Node) (*v1.Node, error) {
	if v1Node == nil {
		return nil, fmt.Errorf("v1 node is nil")
	}

	switch n := v1Node.Node.(type) {
	case *v1.Node_Content:
		return s.convertV1ContentNodeToPublic(n.Content), nil
	case *v1.Node_Chunk:
		return s.convertV1ChunkNodeToPublic(n.Chunk), nil
	case *v1.Node_Cluster:
		return s.convertV1ClusterNodeToPublic(n.Cluster), nil
	default:
		return nil, fmt.Errorf("unsupported node type")
	}
}

// convertPublicNodeToV1 converts a v1.Node to v1.Node
func (s *nodeServiceImpl) convertPublicNodeToV1(publicNode *v1.Node) (*v1.Node, error) {
	if publicNode == nil {
		return nil, fmt.Errorf("public node is nil")
	}

	switch n := publicNode.Node.(type) {
	case *v1.Node_Content:
		return s.convertPublicContentNodeToV1(n.Content), nil
	case *v1.Node_Chunk:
		return s.convertPublicChunkNodeToV1(n.Chunk), nil
	case *v1.Node_Cluster:
		return s.convertPublicClusterNodeToV1(n.Cluster), nil
	default:
		return nil, fmt.Errorf("unsupported node type")
	}
}

// Helper function to get node ID from v1 node
func getNodeID(node *v1.Node) string {
	if node == nil {
		return "nil"
	}

	switch n := node.Node.(type) {
	case *v1.Node_Content:
		if n.Content != nil && n.Content.Base != nil {
			return n.Content.Base.Id
		}
	case *v1.Node_Chunk:
		if n.Chunk != nil && n.Chunk.Base != nil {
			return n.Chunk.Base.Id
		}
	case *v1.Node_Cluster:
		if n.Cluster != nil && n.Cluster.Base != nil {
			return n.Cluster.Base.Id
		}
	}
	return "unknown"
}

// Helper function to convert LinkType enum to string
func linkTypeToString(linkType v1.LinkType) string {
	switch linkType {
	case v1.LinkType_LINK_TYPE_HIERARCHICAL:
		return "hierarchical"
	case v1.LinkType_LINK_TYPE_SEMANTIC:
		return "semantic"
	case v1.LinkType_LINK_TYPE_STRUCTURAL:
		return "structural"
	default:
		return ""
	}
}

// Conversion functions for ContentNode
func (s *nodeServiceImpl) convertV1ContentNodeToPublic(v1Node *v1.ContentNode) *v1.Node {
	if v1Node == nil || v1Node.Base == nil {
		return nil
	}

	var position3D *v1.SpatialCoordinates
	if v1Node.Base.Position_3D != nil {
		position3D = &v1.SpatialCoordinates{
			X: v1Node.Base.Position_3D.X,
			Y: v1Node.Base.Position_3D.Y,
			Z: v1Node.Base.Position_3D.Z,
		}
	}

	var displayProps *v1.DisplayProps
	if v1Node.Base.DisplayProps != nil {
		displayProps = &v1.DisplayProps{
			Size:    v1Node.Base.DisplayProps.Size,
			Opacity: v1Node.Base.DisplayProps.Opacity,
			Shape:   v1Node.Base.DisplayProps.Shape,
			Color:   v1Node.Base.DisplayProps.Color,
		}
	}

	var engagementScore *v1.EngagementScore
	if v1Node.Base.EngagementScore != nil {
		engagementScore = &v1.EngagementScore{
			CanvasScore:  v1Node.Base.EngagementScore.CanvasScore,
			ChatScore:    v1Node.Base.EngagementScore.ChatScore,
			OverallScore: v1Node.Base.EngagementScore.OverallScore,
		}
	}

	return &v1.Node{
		Node: &v1.Node_Content{
			Content: &v1.ContentNode{
				Base: &v1.BaseNode{
					Id:               v1Node.Base.Id,
					SpaceId:          v1Node.Base.SpaceId,
					AbstractionLevel: v1Node.Base.AbstractionLevel,
					ContextType:      v1Node.Base.ContextType,
					Keywords:         v1Node.Base.Keywords,
					DisplayContent:   v1Node.Base.DisplayContent,
					SemanticDensity:  v1Node.Base.SemanticDensity,
					Position_3D:      position3D,
					IsPositionLocked: v1Node.Base.IsPositionLocked,
					Visibility:       v1Node.Base.Visibility,
					DisplayProps:     displayProps,
					EngagementScore:  engagementScore,
					CreatedAt:        v1Node.Base.CreatedAt,
					UpdatedAt:        v1Node.Base.UpdatedAt,
				},
				ContentSourceId: v1Node.ContentSourceId,
				Title:           v1Node.Title,
				MediaType:       v1Node.MediaType,
				Source:          v1Node.Source,
				TokenCount:      v1Node.TokenCount,
			},
		},
	}
}

func (s *nodeServiceImpl) convertPublicContentNodeToV1(publicNode *v1.ContentNode) *v1.Node {
	if publicNode == nil || publicNode.Base == nil {
		return nil
	}

	var position3D *v1.SpatialCoordinates
	if publicNode.Base.Position_3D != nil {
		position3D = &v1.SpatialCoordinates{
			X: publicNode.Base.Position_3D.X,
			Y: publicNode.Base.Position_3D.Y,
			Z: publicNode.Base.Position_3D.Z,
		}
	}

	var displayProps *v1.DisplayProps
	if publicNode.Base.DisplayProps != nil {
		displayProps = &v1.DisplayProps{
			Size:    publicNode.Base.DisplayProps.Size,
			Opacity: publicNode.Base.DisplayProps.Opacity,
			Shape:   publicNode.Base.DisplayProps.Shape,
			Color:   publicNode.Base.DisplayProps.Color,
		}
	}

	var engagementScore *v1.EngagementScore
	if publicNode.Base.EngagementScore != nil {
		engagementScore = &v1.EngagementScore{
			CanvasScore:  publicNode.Base.EngagementScore.CanvasScore,
			ChatScore:    publicNode.Base.EngagementScore.ChatScore,
			OverallScore: publicNode.Base.EngagementScore.OverallScore,
		}
	}

	return &v1.Node{
		Node: &v1.Node_Content{
			Content: &v1.ContentNode{
				Base: &v1.BaseNode{
					Id:               publicNode.Base.Id,
					SpaceId:          publicNode.Base.SpaceId,
					AbstractionLevel: publicNode.Base.AbstractionLevel,
					ContextType:      publicNode.Base.ContextType,
					Keywords:         publicNode.Base.Keywords,
					DisplayContent:   publicNode.Base.DisplayContent,
					SemanticDensity:  publicNode.Base.SemanticDensity,
					Position_3D:      position3D,
					IsPositionLocked: publicNode.Base.IsPositionLocked,
					Visibility:       publicNode.Base.Visibility,
					DisplayProps:     displayProps,
					EngagementScore:  engagementScore,
					CreatedAt:        publicNode.Base.CreatedAt,
					UpdatedAt:        publicNode.Base.UpdatedAt,
				},
				ContentSourceId: publicNode.ContentSourceId,
				Title:           publicNode.Title,
				MediaType:       publicNode.MediaType,
				Source:          publicNode.Source,
				TokenCount:      publicNode.TokenCount,
			},
		},
	}
}

// Conversion functions for ChunkNode
func (s *nodeServiceImpl) convertV1ChunkNodeToPublic(v1Node *v1.ChunkNode) *v1.Node {
	if v1Node == nil || v1Node.Base == nil {
		return nil
	}

	var position3D *v1.SpatialCoordinates
	if v1Node.Base.Position_3D != nil {
		position3D = &v1.SpatialCoordinates{
			X: v1Node.Base.Position_3D.X,
			Y: v1Node.Base.Position_3D.Y,
			Z: v1Node.Base.Position_3D.Z,
		}
	}

	var displayProps *v1.DisplayProps
	if v1Node.Base.DisplayProps != nil {
		displayProps = &v1.DisplayProps{
			Size:    v1Node.Base.DisplayProps.Size,
			Opacity: v1Node.Base.DisplayProps.Opacity,
			Shape:   v1Node.Base.DisplayProps.Shape,
			Color:   v1Node.Base.DisplayProps.Color,
		}
	}

	var engagementScore *v1.EngagementScore
	if v1Node.Base.EngagementScore != nil {
		engagementScore = &v1.EngagementScore{
			CanvasScore:  v1Node.Base.EngagementScore.CanvasScore,
			ChatScore:    v1Node.Base.EngagementScore.ChatScore,
			OverallScore: v1Node.Base.EngagementScore.OverallScore,
		}
	}

	return &v1.Node{
		Node: &v1.Node_Chunk{
			Chunk: &v1.ChunkNode{
				Base: &v1.BaseNode{
					Id:               v1Node.Base.Id,
					SpaceId:          v1Node.Base.SpaceId,
					AbstractionLevel: v1Node.Base.AbstractionLevel,
					ContextType:      v1Node.Base.ContextType,
					Keywords:         v1Node.Base.Keywords,
					DisplayContent:   v1Node.Base.DisplayContent,
					SemanticDensity:  v1Node.Base.SemanticDensity,
					Position_3D:      position3D,
					IsPositionLocked: v1Node.Base.IsPositionLocked,
					Visibility:       v1Node.Base.Visibility,
					DisplayProps:     displayProps,
					EngagementScore:  engagementScore,
					CreatedAt:        v1Node.Base.CreatedAt,
					UpdatedAt:        v1Node.Base.UpdatedAt,
				},
				ContentSourceId: v1Node.ContentSourceId,
				SequenceIndex:   v1Node.SequenceIndex,
				ChunkType:       v1Node.ChunkType,
				StartPosition:   v1Node.StartPosition,
				EndPosition:     v1Node.EndPosition,
				Content:         v1Node.Content,
				TokenCount:      v1Node.TokenCount,
			},
		},
	}
}

func (s *nodeServiceImpl) convertPublicChunkNodeToV1(publicNode *v1.ChunkNode) *v1.Node {
	if publicNode == nil || publicNode.Base == nil {
		return nil
	}

	var position3D *v1.SpatialCoordinates
	if publicNode.Base.Position_3D != nil {
		position3D = &v1.SpatialCoordinates{
			X: publicNode.Base.Position_3D.X,
			Y: publicNode.Base.Position_3D.Y,
			Z: publicNode.Base.Position_3D.Z,
		}
	}

	var displayProps *v1.DisplayProps
	if publicNode.Base.DisplayProps != nil {
		displayProps = &v1.DisplayProps{
			Size:    publicNode.Base.DisplayProps.Size,
			Opacity: publicNode.Base.DisplayProps.Opacity,
			Shape:   publicNode.Base.DisplayProps.Shape,
			Color:   publicNode.Base.DisplayProps.Color,
		}
	}

	var engagementScore *v1.EngagementScore
	if publicNode.Base.EngagementScore != nil {
		engagementScore = &v1.EngagementScore{
			CanvasScore:  publicNode.Base.EngagementScore.CanvasScore,
			ChatScore:    publicNode.Base.EngagementScore.ChatScore,
			OverallScore: publicNode.Base.EngagementScore.OverallScore,
		}
	}

	return &v1.Node{
		Node: &v1.Node_Chunk{
			Chunk: &v1.ChunkNode{
				Base: &v1.BaseNode{
					Id:               publicNode.Base.Id,
					SpaceId:          publicNode.Base.SpaceId,
					AbstractionLevel: publicNode.Base.AbstractionLevel,
					ContextType:      publicNode.Base.ContextType,
					Keywords:         publicNode.Base.Keywords,
					DisplayContent:   publicNode.Base.DisplayContent,
					SemanticDensity:  publicNode.Base.SemanticDensity,
					Position_3D:      position3D,
					IsPositionLocked: publicNode.Base.IsPositionLocked,
					Visibility:       publicNode.Base.Visibility,
					DisplayProps:     displayProps,
					EngagementScore:  engagementScore,
					CreatedAt:        publicNode.Base.CreatedAt,
					UpdatedAt:        publicNode.Base.UpdatedAt,
				},
				ContentSourceId: publicNode.ContentSourceId,
				SequenceIndex:   publicNode.SequenceIndex,
				ChunkType:       publicNode.ChunkType,
				StartPosition:   publicNode.StartPosition,
				EndPosition:     publicNode.EndPosition,
				Content:         publicNode.Content,
				TokenCount:      publicNode.TokenCount,
			},
		},
	}
}

// Conversion functions for ClusterNode
func (s *nodeServiceImpl) convertV1ClusterNodeToPublic(v1Node *v1.ClusterNode) *v1.Node {
	if v1Node == nil || v1Node.Base == nil {
		return nil
	}

	var position3D *v1.SpatialCoordinates
	if v1Node.Base.Position_3D != nil {
		position3D = &v1.SpatialCoordinates{
			X: v1Node.Base.Position_3D.X,
			Y: v1Node.Base.Position_3D.Y,
			Z: v1Node.Base.Position_3D.Z,
		}
	}

	var displayProps *v1.DisplayProps
	if v1Node.Base.DisplayProps != nil {
		displayProps = &v1.DisplayProps{
			Size:    v1Node.Base.DisplayProps.Size,
			Opacity: v1Node.Base.DisplayProps.Opacity,
			Shape:   v1Node.Base.DisplayProps.Shape,
			Color:   v1Node.Base.DisplayProps.Color,
		}
	}

	var engagementScore *v1.EngagementScore
	if v1Node.Base.EngagementScore != nil {
		engagementScore = &v1.EngagementScore{
			CanvasScore:  v1Node.Base.EngagementScore.CanvasScore,
			ChatScore:    v1Node.Base.EngagementScore.ChatScore,
			OverallScore: v1Node.Base.EngagementScore.OverallScore,
		}
	}

	return &v1.Node{
		Node: &v1.Node_Cluster{
			Cluster: &v1.ClusterNode{
				Base: &v1.BaseNode{
					Id:               v1Node.Base.Id,
					SpaceId:          v1Node.Base.SpaceId,
					AbstractionLevel: v1Node.Base.AbstractionLevel,
					ContextType:      v1Node.Base.ContextType,
					Keywords:         v1Node.Base.Keywords,
					DisplayContent:   v1Node.Base.DisplayContent,
					SemanticDensity:  v1Node.Base.SemanticDensity,
					Position_3D:      position3D,
					IsPositionLocked: v1Node.Base.IsPositionLocked,
					Visibility:       v1Node.Base.Visibility,
					DisplayProps:     displayProps,
					EngagementScore:  engagementScore,
					CreatedAt:        v1Node.Base.CreatedAt,
					UpdatedAt:        v1Node.Base.UpdatedAt,
				},
				ClusterScope:  v1Node.ClusterScope,
				Title:         v1Node.Title,
				MemberCount:   v1Node.MemberCount,
				CoverageScore: v1Node.CoverageScore,
			},
		},
	}
}

func (s *nodeServiceImpl) convertPublicClusterNodeToV1(publicNode *v1.ClusterNode) *v1.Node {
	if publicNode == nil || publicNode.Base == nil {
		return nil
	}

	var position3D *v1.SpatialCoordinates
	if publicNode.Base.Position_3D != nil {
		position3D = &v1.SpatialCoordinates{
			X: publicNode.Base.Position_3D.X,
			Y: publicNode.Base.Position_3D.Y,
			Z: publicNode.Base.Position_3D.Z,
		}
	}

	var displayProps *v1.DisplayProps
	if publicNode.Base.DisplayProps != nil {
		displayProps = &v1.DisplayProps{
			Size:    publicNode.Base.DisplayProps.Size,
			Opacity: publicNode.Base.DisplayProps.Opacity,
			Shape:   publicNode.Base.DisplayProps.Shape,
			Color:   publicNode.Base.DisplayProps.Color,
		}
	}

	var engagementScore *v1.EngagementScore
	if publicNode.Base.EngagementScore != nil {
		engagementScore = &v1.EngagementScore{
			CanvasScore:  publicNode.Base.EngagementScore.CanvasScore,
			ChatScore:    publicNode.Base.EngagementScore.ChatScore,
			OverallScore: publicNode.Base.EngagementScore.OverallScore,
		}
	}

	return &v1.Node{
		Node: &v1.Node_Cluster{
			Cluster: &v1.ClusterNode{
				Base: &v1.BaseNode{
					Id:               publicNode.Base.Id,
					SpaceId:          publicNode.Base.SpaceId,
					AbstractionLevel: publicNode.Base.AbstractionLevel,
					ContextType:      publicNode.Base.ContextType,
					Keywords:         publicNode.Base.Keywords,
					DisplayContent:   publicNode.Base.DisplayContent,
					SemanticDensity:  publicNode.Base.SemanticDensity,
					Position_3D:      position3D,
					IsPositionLocked: publicNode.Base.IsPositionLocked,
					Visibility:       publicNode.Base.Visibility,
					DisplayProps:     displayProps,
					EngagementScore:  engagementScore,
					CreatedAt:        publicNode.Base.CreatedAt,
					UpdatedAt:        publicNode.Base.UpdatedAt,
				},
				ClusterScope:  publicNode.ClusterScope,
				Title:         publicNode.Title,
				MemberCount:   publicNode.MemberCount,
				CoverageScore: publicNode.CoverageScore,
			},
		},
	}
}
