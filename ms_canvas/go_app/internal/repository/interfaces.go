package repository

import (
	"context"

	canvasv1 "demo/ms_canvas/go_app/api/proto/public/v1"
)

// NodeRepository defines the interface for node data operations
type NodeRepository interface {
	// Node read methods
	GetNodes(ctx context.Context, ids []string, filter *canvasv1.NodeFilter) ([]*canvasv1.Node, error)
	SearchNodes(ctx context.Context, filter *canvasv1.NodeFilter, spatialBBox *canvasv1.SpatialBoundingBox, limit int32) ([]*canvasv1.Node, error)
	// Node creation methods
	CreateChunkNodes(ctx context.Context, chunks []*canvasv1.ChunkNode) error
	CreateContentNodes(ctx context.Context, contents []*canvasv1.ContentNode) error
	CreateClusterNodes(ctx context.Context, clusters []*canvasv1.ClusterNode) error

	// Node update methods
	UpdateClusterNode(ctx context.Context, update *canvasv1.ClusterNode) error
	UpdateContentNode(ctx context.Context, update *canvasv1.ContentNode) error
	UpdateChunkNode(ctx context.Context, update *canvasv1.ChunkNode) error

	// Node deletion
	SoftDeleteNodes(ctx context.Context, nodeIDs []string) error
}

// LinkRepository defines the interface for link data operations
type LinkRepository interface {
	// Link creation methods - bulk operations
	CreateHierarchicalLinks(ctx context.Context, links []*canvasv1.HierarchicalLink) error
	CreateSemanticLinks(ctx context.Context, links []*canvasv1.SemanticLink) error
	CreateStructuralLinks(ctx context.Context, links []*canvasv1.StructuralLink) error

	// Link read methods - consolidated
	GetLinks(ctx context.Context, ids []string, query *canvasv1.LinkQuery) ([]*canvasv1.Link, error)
	GetLinksByNodes(ctx context.Context, nodeIDs []string, direction canvasv1.Direction, query *canvasv1.LinkQuery) ([]*canvasv1.Link, error)

	// Link update methods - bulk operations
	UpdateSemanticLinks(ctx context.Context, links []*canvasv1.SemanticLink) error
	UpdateStructuralLinks(ctx context.Context, links []*canvasv1.StructuralLink) error
	UpdateHierarchicalLinks(ctx context.Context, links []*canvasv1.HierarchicalLink) error

	// Link deletion methods - consolidated
	DeleteLinks(ctx context.Context, linkIDs []string) error
	DeleteLinksForNodes(ctx context.Context, nodeIDs []string) error
}
