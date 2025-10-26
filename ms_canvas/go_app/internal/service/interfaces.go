package service

import (
	"context"

	canvaspublicv1 "demo/ms_canvas/go_app/api/proto/public/v1"
	"demo/ms_canvas/go_app/internal/events"
)

// NodeService defines the interface for node business logic operations
type NodeService interface {
	// GetNodes retrieves nodes by their IDs (direct lookup)
	GetNodes(ctx context.Context, ids []string) ([]*canvaspublicv1.Node, error)

	// SearchNodes searches for nodes based on filter criteria and spatial bounds
	SearchNodes(ctx context.Context, filter *canvaspublicv1.NodeFilter, spatialBBox *canvaspublicv1.SpatialBoundingBox, limit int32) ([]*canvaspublicv1.Node, error)

	// UpdateNodes updates existing nodes
	UpdateNodes(ctx context.Context, nodes []*canvaspublicv1.Node) ([]*canvaspublicv1.Node, error)

	// GetNeighbors retrieves neighboring nodes for given node IDs
	GetNeighbors(ctx context.Context, ids []string, direction canvaspublicv1.Direction, query *canvaspublicv1.LinkQuery, limitPerNode int32, includeProperties bool) (map[string]*canvaspublicv1.NeighborsList, error)

	// ListNodesByLink traverses one hop of a given link type using flexible parent identifiers.
	ListNodesByLink(ctx context.Context, req *canvaspublicv1.ListNodesByLinkRequest) (*canvaspublicv1.ListNodesByLinkResponse, error)
}

// SearchService defines the interface for search operations
type SearchService interface {
	// SemanticSearch performs semantic search on nodes
	SemanticSearch(ctx context.Context, req *canvaspublicv1.SemanticSearchRequest) (*canvaspublicv1.SemanticSearchResponse, error)
}

// LinkService defines the interface for link operations
type LinkService interface {
	// CreateStructuralLinks creates new structural links
	CreateStructuralLinks(ctx context.Context, links []*canvaspublicv1.StructuralLinkCreate) ([]*canvaspublicv1.StructuralLink, error)

	// UpdateStructuralLinks updates existing structural links
	UpdateStructuralLinks(ctx context.Context, updates []*canvaspublicv1.StructuralLinkUpdate) ([]*canvaspublicv1.StructuralLink, error)

	// DeleteStructuralLinks deletes structural links
	DeleteStructuralLinks(ctx context.Context, linkIDs []*canvaspublicv1.StructuralLinkIdentifier) (int32, error)
}

// EventHandler defines the interface for processing events from external sources
type EventHandler interface {
	// HandleDocumentProcessed processes incoming document.processed events
	HandleDocumentProcessed(event events.DocumentProcessedEvent) error
}
