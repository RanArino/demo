package repository

import (
	"context"

	v1 "demo/ms_canvas/go_app/api/proto/private/v1"
)

// NodeRepository defines the interface for node data operations
type NodeRepository interface {
	// Node read methods
	GetNodes(ctx context.Context, ids []string, filter *v1.NodeFilter) ([]*v1.Node, error)

	// Node creation methods
	CreateChunkNodes(ctx context.Context, chunks []*v1.ChunkNode) error
	CreateContentNode(ctx context.Context, node *v1.ContentNode) error
	CreateClusterNode(ctx context.Context, id string, spaceID string, abstractionLevel int32, clusterScope string, title *string, embedding *[]float32) error

	// Node update methods
	UpdateClusterNode(ctx context.Context, update *v1.ClusterNode) error
	UpdateContentNode(ctx context.Context, update *v1.ContentNode) error
	UpdateChunkNode(ctx context.Context, update *v1.ChunkNode) error

	// Node deletion
	SoftDeleteNode(ctx context.Context, nodeID string) error
}
