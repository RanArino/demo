package repository

import (
	"context"
	"testing"

	canvasv1 "demo/ms_canvas/go_app/api/proto/public/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestTraversalRepository_ListNodesByLink_ByNodeID(t *testing.T) {
	ctx := context.Background()
	nodeRepo := new(mockNodeRepo)
	linkRepo := new(mockLinkRepo)
	repo := NewTraversalRepository(nodeRepo, linkRepo)

	parent := &canvasv1.NodeReference{Identifier: &canvasv1.NodeReference_NodeId{NodeId: "parent-1"}}
	traversal := &canvasv1.LinkTraversalSpec{
		Direction: canvasv1.Direction_DIRECTION_OUTGOING,
		Query: &canvasv1.LinkQuery{
			LinkTypes: []canvasv1.LinkType{canvasv1.LinkType_LINK_TYPE_HIERARCHICAL},
		},
	}

	link := &canvasv1.Link{
		Link: &canvasv1.Link_Hierarchical{
			Hierarchical: &canvasv1.HierarchicalLink{
				Base: &canvasv1.BaseLink{Id: "link-1", SourceId: "parent-1", TargetId: "child-1"},
			},
		},
	}

	linkRepo.On("GetLinksByNodes", mock.Anything, []string{"parent-1"}, canvasv1.Direction_DIRECTION_OUTGOING, traversal.Query).
		Return([]*canvasv1.Link{link}, nil).Once()
	nodeRepo.On("GetNodes", mock.Anything, []string{"child-1"}, (*canvasv1.NodeFilter)(nil)).
		Return([]*canvasv1.Node{
			{
				Node: &canvasv1.Node_Content{Content: &canvasv1.ContentNode{Base: &canvasv1.BaseNode{Id: "child-1"}}},
			},
		}, nil).Once()

	neighbors, err := repo.ListNodesByLink(ctx, parent, traversal, nil, 0, 10)
	require.NoError(t, err)
	require.Len(t, neighbors, 1)
	assert.Equal(t, "child-1", getNodeID(neighbors[0].Node))
	assert.Equal(t, canvasv1.LinkType_LINK_TYPE_HIERARCHICAL, neighbors[0].LinkType)
	assert.NotNil(t, neighbors[0].Link)

	linkRepo.AssertExpectations(t)
	nodeRepo.AssertExpectations(t)
}

func TestTraversalRepository_ListNodesByLink_ByContentSource(t *testing.T) {
	ctx := context.Background()
	nodeRepo := new(mockNodeRepo)
	linkRepo := new(mockLinkRepo)
	repo := NewTraversalRepository(nodeRepo, linkRepo)

	parentNodeType := canvasv1.NodeType_NODE_TYPE_CONTENT
	parent := &canvasv1.NodeReference{
		SpaceId:    proto.String("space-9"),
		NodeType:   &parentNodeType,
		Identifier: &canvasv1.NodeReference_ContentSourceId{ContentSourceId: "doc-7"},
	}
	traversal := &canvasv1.LinkTraversalSpec{
		Direction: canvasv1.Direction_DIRECTION_OUTGOING,
		Query:     &canvasv1.LinkQuery{},
	}

	parentNode := &canvasv1.Node{
		Node: &canvasv1.Node_Content{Content: &canvasv1.ContentNode{Base: &canvasv1.BaseNode{Id: "parent-9"}}},
	}
	nodeRepo.On("SearchNodes", mock.Anything, mock.MatchedBy(func(filter *canvasv1.NodeFilter) bool {
		return filter != nil && filter.GetSpaceId() == "space-9" &&
			filter.GetContentFilter() != nil && filter.GetContentFilter().GetContentSourceId() == "doc-7"
	}), (*canvasv1.SpatialBoundingBox)(nil), int32(0)).
		Return([]*canvasv1.Node{parentNode}, nil).Once()

	link := &canvasv1.Link{
		Link: &canvasv1.Link_Hierarchical{
			Hierarchical: &canvasv1.HierarchicalLink{
				Base: &canvasv1.BaseLink{Id: "link-42", SourceId: "parent-9", TargetId: "child-42"},
			},
		},
	}
	linkRepo.On("GetLinksByNodes", mock.Anything, []string{"parent-9"}, canvasv1.Direction_DIRECTION_OUTGOING, traversal.Query).
		Return([]*canvasv1.Link{link}, nil).Once()

	nodeRepo.On("GetNodes", mock.Anything, []string{"child-42"}, (*canvasv1.NodeFilter)(nil)).
		Return([]*canvasv1.Node{
			{
				Node: &canvasv1.Node_Content{Content: &canvasv1.ContentNode{Base: &canvasv1.BaseNode{Id: "child-42"}}},
			},
		}, nil).Once()

	neighbors, err := repo.ListNodesByLink(ctx, parent, traversal, nil, 0, 5)
	require.NoError(t, err)
	require.Len(t, neighbors, 1)
	assert.Equal(t, "child-42", getNodeID(neighbors[0].Node))

	linkRepo.AssertExpectations(t)
	nodeRepo.AssertExpectations(t)
}

func TestTraversalRepository_ListNodesByLink_OffsetLimit(t *testing.T) {
	ctx := context.Background()
	nodeRepo := new(mockNodeRepo)
	linkRepo := new(mockLinkRepo)
	repo := NewTraversalRepository(nodeRepo, linkRepo)

	parent := &canvasv1.NodeReference{Identifier: &canvasv1.NodeReference_NodeId{NodeId: "parent-1"}}
	traversal := &canvasv1.LinkTraversalSpec{
		Direction: canvasv1.Direction_DIRECTION_OUTGOING,
		Query:     &canvasv1.LinkQuery{},
	}

	links := []*canvasv1.Link{
		{Link: &canvasv1.Link_Hierarchical{Hierarchical: &canvasv1.HierarchicalLink{Base: &canvasv1.BaseLink{SourceId: "parent-1", TargetId: "child-1"}}}},
		{Link: &canvasv1.Link_Hierarchical{Hierarchical: &canvasv1.HierarchicalLink{Base: &canvasv1.BaseLink{SourceId: "parent-1", TargetId: "child-2"}}}},
		{Link: &canvasv1.Link_Hierarchical{Hierarchical: &canvasv1.HierarchicalLink{Base: &canvasv1.BaseLink{SourceId: "parent-1", TargetId: "child-3"}}}},
	}
	linkRepo.On("GetLinksByNodes", mock.Anything, []string{"parent-1"}, canvasv1.Direction_DIRECTION_OUTGOING, traversal.Query).
		Return(links, nil).Once()
	nodeRepo.On("GetNodes", mock.Anything, []string{"child-1", "child-2", "child-3"}, (*canvasv1.NodeFilter)(nil)).
		Return([]*canvasv1.Node{
			{Node: &canvasv1.Node_Content{Content: &canvasv1.ContentNode{Base: &canvasv1.BaseNode{Id: "child-1"}}}},
			{Node: &canvasv1.Node_Content{Content: &canvasv1.ContentNode{Base: &canvasv1.BaseNode{Id: "child-2"}}}},
			{Node: &canvasv1.Node_Content{Content: &canvasv1.ContentNode{Base: &canvasv1.BaseNode{Id: "child-3"}}}},
		}, nil).Once()

	neighbors, err := repo.ListNodesByLink(ctx, parent, traversal, nil, 1, 1)
	require.NoError(t, err)
	require.Len(t, neighbors, 1)
	assert.Equal(t, "child-2", getNodeID(neighbors[0].Node))

	linkRepo.AssertExpectations(t)
	nodeRepo.AssertExpectations(t)
}

func TestTraversalRepository_ListNodesByLink_MissingIdentifier(t *testing.T) {
	repo := NewTraversalRepository(new(mockNodeRepo), new(mockLinkRepo))
	_, err := repo.ListNodesByLink(context.Background(), &canvasv1.NodeReference{}, &canvasv1.LinkTraversalSpec{Query: &canvasv1.LinkQuery{}}, nil, 0, 10)
	require.Error(t, err)
}

// ---- test doubles ----

type mockNodeRepo struct{ mock.Mock }

func (m *mockNodeRepo) GetNodes(ctx context.Context, ids []string, filter *canvasv1.NodeFilter) ([]*canvasv1.Node, error) {
	args := m.Called(ctx, ids, filter)
	return args.Get(0).([]*canvasv1.Node), args.Error(1)
}

func (m *mockNodeRepo) SearchNodes(ctx context.Context, filter *canvasv1.NodeFilter, spatialBBox *canvasv1.SpatialBoundingBox, limit int32) ([]*canvasv1.Node, error) {
	args := m.Called(ctx, filter, spatialBBox, limit)
	return args.Get(0).([]*canvasv1.Node), args.Error(1)
}

func (m *mockNodeRepo) CreateChunkNodes(ctx context.Context, chunks []*canvasv1.ChunkNode) error {
	panic("not implemented")
}

func (m *mockNodeRepo) CreateContentNodes(ctx context.Context, contents []*canvasv1.ContentNode) error {
	panic("not implemented")
}

func (m *mockNodeRepo) CreateClusterNodes(ctx context.Context, clusters []*canvasv1.ClusterNode) error {
	panic("not implemented")
}

func (m *mockNodeRepo) UpdateClusterNode(ctx context.Context, update *canvasv1.ClusterNode) error {
	panic("not implemented")
}

func (m *mockNodeRepo) UpdateContentNode(ctx context.Context, update *canvasv1.ContentNode) error {
	panic("not implemented")
}

func (m *mockNodeRepo) UpdateChunkNode(ctx context.Context, update *canvasv1.ChunkNode) error {
	panic("not implemented")
}

func (m *mockNodeRepo) SoftDeleteNodes(ctx context.Context, nodeIDs []string) error {
	panic("not implemented")
}

type mockLinkRepo struct{ mock.Mock }

func (m *mockLinkRepo) CreateHierarchicalLinks(ctx context.Context, links []*canvasv1.HierarchicalLink) error {
	panic("not implemented")
}

func (m *mockLinkRepo) CreateSemanticLinks(ctx context.Context, links []*canvasv1.SemanticLink) error {
	panic("not implemented")
}

func (m *mockLinkRepo) CreateStructuralLinks(ctx context.Context, links []*canvasv1.StructuralLink) error {
	panic("not implemented")
}

func (m *mockLinkRepo) GetLinks(ctx context.Context, ids []string, filter *canvasv1.BaseLinkFilter) ([]*canvasv1.Link, error) {
	panic("not implemented")
}

func (m *mockLinkRepo) GetLinksByNodes(ctx context.Context, nodeIDs []string, direction canvasv1.Direction, query *canvasv1.LinkQuery) ([]*canvasv1.Link, error) {
	args := m.Called(ctx, nodeIDs, direction, query)
	return args.Get(0).([]*canvasv1.Link), args.Error(1)
}

func (m *mockLinkRepo) UpdateSemanticLinks(ctx context.Context, links []*canvasv1.SemanticLink) error {
	panic("not implemented")
}

func (m *mockLinkRepo) UpdateStructuralLinks(ctx context.Context, links []*canvasv1.StructuralLink) error {
	panic("not implemented")
}

func (m *mockLinkRepo) UpdateHierarchicalLinks(ctx context.Context, links []*canvasv1.HierarchicalLink) error {
	panic("not implemented")
}

func (m *mockLinkRepo) DeleteLinks(ctx context.Context, linkIDs []string) error {
	panic("not implemented")
}

func (m *mockLinkRepo) DeleteLinksForNodes(ctx context.Context, nodeIDs []string) error {
	panic("not implemented")
}
