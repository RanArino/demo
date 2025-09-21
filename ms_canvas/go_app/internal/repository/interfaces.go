package repository

import (
	"context"
	"time"

	v1 "demo/ms_canvas/go_app/api/proto/private/v1"
)

// NodeRepository defines the interface for node data operations
type NodeRepository interface {
	// Node read methods
	GetNodes(ctx context.Context, ids []string, filter *v1.NodeFilter) ([]*v1.Node, error)

	// Node creation methods
	CreateChunkNodes(ctx context.Context, chunks []*v1.ChunkNode) error
	CreateContentNodes(ctx context.Context, contents []*v1.ContentNode) error
	CreateClusterNodes(ctx context.Context, clusters []*v1.ClusterNode) error

	// Node update methods
	UpdateClusterNode(ctx context.Context, update *v1.ClusterNode) error
	UpdateContentNode(ctx context.Context, update *v1.ContentNode) error
	UpdateChunkNode(ctx context.Context, update *v1.ChunkNode) error

	// Node deletion
	SoftDeleteNodes(ctx context.Context, nodeIDs []string) error
}

// LinkRepository defines the interface for link data operations
type LinkRepository interface {
	// Link creation methods
	CreateHierarchicalLinks(ctx context.Context, contentSourceID string, chunkIDs []string, connectionType string, hierarchyDepth int32, createdAt time.Time) error
	CreateSemanticLinks(ctx context.Context, links []*v1.SemanticLink) error
	CreateStructuralLink(ctx context.Context, link *v1.StructuralLink) error
	CreateHierarchicalLink(ctx context.Context, link *v1.HierarchicalLink) error

	// Semantic link operations
	GetSemanticLink(ctx context.Context, sourceID, targetID string) (*v1.SemanticLink, error)
	ListSemanticLinksFrom(ctx context.Context, nodeID string) ([]*v1.SemanticLink, error)
	ListSemanticLinksTo(ctx context.Context, nodeID string) ([]*v1.SemanticLink, error)
	UpdateSemanticLink(ctx context.Context, link *v1.SemanticLink) error
	DeleteSemanticLink(ctx context.Context, sourceID, targetID string) error
	DeleteSemanticLinksForNode(ctx context.Context, nodeID string) error

	// Structural link operations
	GetStructuralLink(ctx context.Context, sourceID, targetID string) (*v1.StructuralLink, error)
	ListStructuralLinksFrom(ctx context.Context, nodeID string) ([]*v1.StructuralLink, error)
	ListStructuralLinksTo(ctx context.Context, nodeID string) ([]*v1.StructuralLink, error)
	UpdateStructuralLink(ctx context.Context, link *v1.StructuralLink) error
	DeleteStructuralLink(ctx context.Context, sourceID, targetID string) error
	DeleteStructuralLinksForNode(ctx context.Context, nodeID string) error

	// Hierarchical link operations
	GetHierarchicalLink(ctx context.Context, sourceID, targetID string) (*v1.HierarchicalLink, error)
	ListHierarchicalLinksFrom(ctx context.Context, nodeID string) ([]*v1.HierarchicalLink, error)
	ListHierarchicalLinksTo(ctx context.Context, nodeID string) ([]*v1.HierarchicalLink, error)
	UpdateHierarchicalLink(ctx context.Context, link *v1.HierarchicalLink) error
	DeleteHierarchicalLink(ctx context.Context, sourceID, targetID string) error
	DeleteHierarchicalLinksForNode(ctx context.Context, nodeID string) error

	// Convenience methods for hierarchical links
	DeleteHierarchicalLinksByContentSource(ctx context.Context, contentSourceID string) error
	DeleteHierarchicalLinks(ctx context.Context, contentSourceID string, chunkIDs []string) error
}
