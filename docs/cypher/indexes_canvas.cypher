// Composite indexes to support abstraction-level traversal and per-space scans
CREATE INDEX node_space_level IF NOT EXISTS
FOR (n:Node)
ON (n.space_id, n.abstraction_level);

// Optional label-specific indexes (uncomment if generic :Node label is absent)
// CREATE INDEX content_space_level IF NOT EXISTS FOR (n:ContentNode) ON (n.space_id, n.abstraction_level);
// CREATE INDEX chunk_space_level IF NOT EXISTS FOR (n:ChunkNode) ON (n.space_id, n.abstraction_level);
// CREATE INDEX cluster_space_level IF NOT EXISTS FOR (n:ClusterNode) ON (n.space_id, n.abstraction_level);
