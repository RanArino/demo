package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"

	canvasv1 "demo/ms_canvas/go_app/api/proto/public/v1"

	"google.golang.org/protobuf/proto"
)

// NewTraversalRepository builds a TraversalRepository backed by the existing node/link repos.
func NewTraversalRepository(nodeRepo NodeRepository, linkRepo LinkRepository) TraversalRepository {
	return &traversalRepository{
		nodeRepo: nodeRepo,
		linkRepo: linkRepo,
	}
}

type traversalRepository struct {
	nodeRepo NodeRepository
	linkRepo LinkRepository
}

func (r *traversalRepository) ListNodesByLink(
	ctx context.Context,
	parent *canvasv1.NodeReference,
	traversal *canvasv1.LinkTraversalSpec,
	childFilter *canvasv1.NodeFilter,
	offset, limit int32,
) ([]*LinkNeighbor, error) {
	if parent == nil {
		return nil, errors.New("parent reference is required")
	}
	if traversal == nil || traversal.Query == nil {
		return nil, errors.New("traversal query is required")
	}
	if traversal.GetMaxHops() > 1 {
		return nil, fmt.Errorf("max_hops=%d not supported", traversal.GetMaxHops())
	}

	parentIDs, err := r.resolveParentIDs(ctx, parent)
	if err != nil {
		return nil, err
	}
	if len(parentIDs) == 0 {
		return []*LinkNeighbor{}, nil
	}

	links, err := r.linkRepo.GetLinksByNodes(ctx, parentIDs, traversal.GetDirection(), traversal.Query)
	if err != nil {
		return nil, err
	}

	entries := r.buildNeighborEntries(parentIDs, traversal.GetDirection(), links)
	if len(entries) == 0 {
		return []*LinkNeighbor{}, nil
	}

	childIDs := uniqueChildIDs(entries)
	nodes, err := r.nodeRepo.GetNodes(ctx, childIDs, childFilter)
	if err != nil {
		return nil, err
	}
	nodeMap := make(map[string]*canvasv1.Node, len(nodes))
	for _, node := range nodes {
		id := getNodeID(node)
		if id != "" {
			nodeMap[id] = node
		}
	}

	filtered := make([]neighborEntry, 0, len(entries))
	for _, entry := range entries {
		if node := nodeMap[entry.childID]; node != nil {
			entry.node = node
			filtered = append(filtered, entry)
		}
	}
	if len(filtered) == 0 {
		return []*LinkNeighbor{}, nil
	}

	start := int(offset)
	if start < 0 {
		start = 0
	}
	if start >= len(filtered) {
		return []*LinkNeighbor{}, nil
	}

	end := len(filtered)
	if limit > 0 && start+int(limit) < end {
		end = start + int(limit)
	}

	neighbors := make([]*LinkNeighbor, 0, end-start)
	for _, entry := range filtered[start:end] {
		neighbors = append(neighbors, &LinkNeighbor{
			Node:     entry.node,
			Link:     entry.link,
			LinkType: entry.linkType,
		})
	}

	return neighbors, nil
}

func (r *traversalRepository) resolveParentIDs(ctx context.Context, ref *canvasv1.NodeReference) ([]string, error) {
	if ref == nil {
		return nil, errors.New("node reference cannot be nil")
	}
	if id := ref.GetNodeId(); id != "" {
		return []string{id}, nil
	}
	if ref.GetExternalId() != "" {
		return nil, errors.New("external_id lookups are not supported yet")
	}
	if contentSrc := ref.GetContentSourceId(); contentSrc != "" {
		filter := &canvasv1.NodeFilter{}
		if space := ref.GetSpaceId(); space != "" {
			filter.SpaceId = proto.String(space)
		}
		switch ref.GetNodeType() {
		case canvasv1.NodeType_NODE_TYPE_CHUNK:
			filter.ChunkFilter = &canvasv1.ChunkNodeFilter{ContentSourceId: proto.String(contentSrc)}
		case canvasv1.NodeType_NODE_TYPE_CLUSTER:
			return nil, errors.New("clusters do not support content_source_id lookups")
		default:
			filter.ContentFilter = &canvasv1.ContentNodeFilter{ContentSourceId: proto.String(contentSrc)}
		}
		nodes, err := r.nodeRepo.SearchNodes(ctx, filter, nil, 0)
		if err != nil {
			return nil, err
		}
		ids := make([]string, 0, len(nodes))
		for _, node := range nodes {
			if id := getNodeID(node); id != "" {
				ids = append(ids, id)
			}
		}
		return dedupeStrings(ids), nil
	}
	return nil, errors.New("node reference requires node_id or content_source_id")
}

func (r *traversalRepository) buildNeighborEntries(parentIDs []string, direction canvasv1.Direction, links []*canvasv1.Link) []neighborEntry {
	parentSet := make(map[string]struct{}, len(parentIDs))
	for _, id := range parentIDs {
		parentSet[id] = struct{}{}
	}
	seenChildren := make(map[string]struct{})
	entries := make([]neighborEntry, 0)

	appendEntry := func(ownerID, neighborID string, linkType canvasv1.LinkType, link *canvasv1.Link) {
		if ownerID == "" || neighborID == "" {
			return
		}
		if _, ok := parentSet[ownerID]; !ok {
			return
		}
		if _, exists := seenChildren[neighborID]; exists {
			return
		}
		seenChildren[neighborID] = struct{}{}
		entries = append(entries, neighborEntry{
			childID:  neighborID,
			linkType: linkType,
			link:     cloneLink(link),
		})
	}

	for _, link := range links {
		linkType, sourceID, targetID := extractLinkEndpoints(link)
		if sourceID == "" || targetID == "" {
			continue
		}
		switch direction {
		case canvasv1.Direction_DIRECTION_INCOMING:
			appendEntry(targetID, sourceID, linkType, link)
		case canvasv1.Direction_DIRECTION_BOTH:
			appendEntry(sourceID, targetID, linkType, link)
			appendEntry(targetID, sourceID, linkType, link)
		default:
			appendEntry(sourceID, targetID, linkType, link)
		}
	}

	return entries
}

type neighborEntry struct {
	childID  string
	linkType canvasv1.LinkType
	link     *canvasv1.Link
	node     *canvasv1.Node
}

func extractLinkEndpoints(link *canvasv1.Link) (canvasv1.LinkType, string, string) {
	if link == nil {
		return canvasv1.LinkType_LINK_TYPE_UNSPECIFIED, "", ""
	}
	switch l := link.Link.(type) {
	case *canvasv1.Link_Hierarchical:
		if l.Hierarchical != nil && l.Hierarchical.Base != nil {
			return canvasv1.LinkType_LINK_TYPE_HIERARCHICAL, l.Hierarchical.Base.SourceId, l.Hierarchical.Base.TargetId
		}
	case *canvasv1.Link_Semantic:
		if l.Semantic != nil && l.Semantic.Base != nil {
			return canvasv1.LinkType_LINK_TYPE_SEMANTIC, l.Semantic.Base.SourceId, l.Semantic.Base.TargetId
		}
	case *canvasv1.Link_Structural:
		if l.Structural != nil && l.Structural.Base != nil {
			return canvasv1.LinkType_LINK_TYPE_STRUCTURAL, l.Structural.Base.SourceId, l.Structural.Base.TargetId
		}
	}
	return canvasv1.LinkType_LINK_TYPE_UNSPECIFIED, "", ""
}

func uniqueChildIDs(entries []neighborEntry) []string {
	ids := make([]string, 0, len(entries))
	seen := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if entry.childID == "" {
			continue
		}
		if _, ok := seen[entry.childID]; ok {
			continue
		}
		seen[entry.childID] = struct{}{}
		ids = append(ids, entry.childID)
	}
	sort.Strings(ids)
	return ids
}

func dedupeStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	result := make([]string, 0, len(in))
	for _, val := range in {
		if val == "" {
			continue
		}
		if _, ok := seen[val]; ok {
			continue
		}
		seen[val] = struct{}{}
		result = append(result, val)
	}
	return result
}

func cloneLink(link *canvasv1.Link) *canvasv1.Link {
	if link == nil {
		return nil
	}
	return proto.Clone(link).(*canvasv1.Link)
}

func getNodeID(node *canvasv1.Node) string {
	if node == nil {
		return ""
	}
	switch n := node.Node.(type) {
	case *canvasv1.Node_Content:
		if n.Content != nil && n.Content.Base != nil {
			return n.Content.Base.Id
		}
	case *canvasv1.Node_Chunk:
		if n.Chunk != nil && n.Chunk.Base != nil {
			return n.Chunk.Base.Id
		}
	case *canvasv1.Node_Cluster:
		if n.Cluster != nil && n.Cluster.Base != nil {
			return n.Cluster.Base.Id
		}
	}
	return ""
}
