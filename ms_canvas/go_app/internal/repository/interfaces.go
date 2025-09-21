package repository

import (
	"context"

	v1 "demo/ms_canvas/go_app/api/proto/private/v1"
	canvasv1 "demo/ms_canvas/go_app/api/proto/public/v1"
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
	// Link creation methods - bulk operations
	CreateHierarchicalLinks(ctx context.Context, links []*v1.HierarchicalLink) error
	CreateSemanticLinks(ctx context.Context, links []*v1.SemanticLink) error
	CreateStructuralLinks(ctx context.Context, links []*v1.StructuralLink) error

	// Link read methods - consolidated
	GetLinks(ctx context.Context, ids []string, filter *v1.BaseLinkFilter) ([]*v1.Link, error)
	GetLinksByNodes(ctx context.Context, nodeIDs []string, direction canvasv1.Direction, filter *v1.BaseLinkFilter) ([]*v1.Link, error)

	// Link update methods - bulk operations
	UpdateSemanticLinks(ctx context.Context, links []*v1.SemanticLink) error
	UpdateStructuralLinks(ctx context.Context, links []*v1.StructuralLink) error
	UpdateHierarchicalLinks(ctx context.Context, links []*v1.HierarchicalLink) error

	// Link deletion methods - consolidated
	DeleteLinks(ctx context.Context, linkIDs []string) error
	DeleteLinksForNodes(ctx context.Context, nodeIDs []string) error
}
