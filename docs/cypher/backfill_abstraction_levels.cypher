// Content nodes: abstraction_level = 0, context_type = 'content'
MATCH (n:ContentNode)
WHERE n.abstraction_level IS NULL
SET n.abstraction_level = 0;

MATCH (n:ContentNode)
WHERE n.context_type IS NULL
SET n.context_type = 'content';

// Chunk nodes: abstraction_level = -1, context_type = 'chunk'
MATCH (n:ChunkNode)
WHERE n.abstraction_level IS NULL
SET n.abstraction_level = -1;

MATCH (n:ChunkNode)
WHERE n.context_type IS NULL
SET n.context_type = 'chunk';

// Optional: ensure visibility defaults to true when missing
// MATCH (n:ContentNode)
// WHERE n.visibility IS NULL
// SET n.visibility = true;
