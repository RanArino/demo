// Ensure every ChunkNode has a HIERARCHICAL_PARENT link to its matching ContentNode
// based on shared content_source_id.
//
// - Relationship direction: (content:ContentNode) -> (chunk:ChunkNode)
// - Properties:
//     connection_type = 1   // ABSTRACTION
//     hierarchy_depth = 1
//     id = random UUID
//     created_at / updated_at = datetime()
//
// This script is idempotent thanks to MERGE. Existing links will have
// connection_type, hierarchy_depth, and updated_at refreshed.
MATCH (content:Node:ContentNode)
MATCH (chunk:Node:ChunkNode)
WHERE content.content_source_id IS NOT NULL
  AND content.content_source_id = chunk.content_source_id
MERGE (content)-[rel:HIERARCHICAL_PARENT]->(chunk)
ON CREATE SET
    rel.id = toString(randomUUID()),
    rel.connection_type = 1,
    rel.hierarchy_depth = 1,
    rel.created_at = datetime(),
    rel.updated_at = datetime(),
    rel.deleted_at = NULL
ON MATCH SET
    rel.connection_type = 1,
    rel.hierarchy_depth = 1,
    rel.updated_at = datetime();
