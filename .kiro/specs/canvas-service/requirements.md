# Requirements Document: Canvas Service

## 1. Overview

This document outlines the functional and non-functional requirements for the Canvas Service. The service's primary purpose to consolidate vector embeddings and graph-structured content in a unified Neo4j-backed store. It will process documents from `ms_document_process`, chunk them, generate embeddings, construct a knowledge graph, and provide APIs for querying this graph.

Input data: plain text provided inline or via blob storage URL (the `ms_document_process` service sends either text payloads or references for processing).

Runtime/deployment (demo phase): `ms_canvas/go_app` and `ms_canvas/python_app` run together in a single container using `supervisord`. The Go app is the only component that communicates with other microservices; the Python app is internal-only and reachable via localhost gRPC. Within `ms_canvas`, only gRPC is used for distributed operations; Kafka is used solely to receive the upstream `document.processed` event from `ms_document_process`.

Communication contracts are defined in `ms_canvas/proto/v1/canvas.proto`, which is the source of truth for all internal Go↔Python communication in this phase. External/public proto is out of scope for now.

## 2. Requirements List

### Requirement 1: Event-Driven Document Ingestion

**User Story:**
> As a system, I want to consume document processing completion events so that I can initiate the graph construction pipeline for new content.

**Acceptance Criteria:**
```gherkin
GIVEN a document processing completion event is received from the message queue
WHEN the event is processed
THEN the system must validate the event's authenticity and authorization.
AND the system must validate and normalize the incoming document metadata (space_id, content_source_id, etc.).
AND a `ContentNode` representing the source document must be created in the database with provenance information.
AND the document must be scheduled for the text chunking process within the orchestrator (no internal Kafka topic).
AND the event handler must be idempotent to prevent duplicate processing.
```

### Requirement 2: Text Chunking

**User Story:**
> As a system, I want to automatically chunk document content into smaller, manageable units so that they can be vectorized and linked in the graph.

**Acceptance Criteria:**
```gherkin
GIVEN a document has been validated and enqueued for chunking
WHEN the chunking process is triggered
THEN the system must obtain the source text either from an inline request field or by fetching from a provided blob storage URL.
AND the content must be split into chunks using sentence-based chunking only (for this phase) with configurable parameters: target_size≈tokens (default 300) and overlap percentage (default 10%).
AND token estimation/counting must use a configurable tokenizer (default: tiktoken `cl100k_base`).
AND the implementation must use spaCy for sentence segmentation; model selection is configurable for future phases.
AND multi-language inputs (e.g., English, Spanish, Chinese, Japanese) must be supported.
AND chunk content should be normalized prior to splitting, but each chunk must include `start_position` and `end_position` character offsets referring to the original (pre-normalized) source content.
AND `ChunkNode` objects must be created with required properties (id, content_source_id, position, start_position, end_position, content, etc.).
AND internal RPCs use gRPC over localhost with an increased `max_receive_message_length`. For the demo phase, unary calls are sufficient.
AND a `chunking.completed` metric/event must be emitted by the Go orchestrator after chunks are persisted.
```

### Requirement 3: Embedding Pipeline (inline in the same orchestration or combined with chunking)

**User Story:**
> As a system, I want to compute vector embeddings for text chunks so that I can perform semantic similarity searches.

**Acceptance Criteria:**
```gherkin
GIVEN the orchestrator has completed chunking OR it invokes a combined chunking+embedding operation
WHEN the embedding step runs as part of the same orchestrated flow (no intermediate Kafka event)
THEN vector embeddings must be computed for each chunk using a configurable embedding provider.
AND the computed embedding vector must be stored on the corresponding `ChunkNode` in Neo4j.
AND metadata about the embedding model (model_id, version) must be stored alongside the embedding.
AND an `embedding.completed` metric must be emitted.
```

### Requirement 3b: Combined Chunking + Embedding RPC (internal)

**User Story:**
> As a system, I want to minimize internal round-trips by invoking a single RPC that chunks text and generates embeddings so that I can persist both chunks and vectors efficiently.

**Acceptance Criteria:**
```gherkin
GIVEN a document is ready for processing
WHEN the Go orchestrator calls the Python service via a combined `ChunkAndEmbed` RPC
THEN the Python service may fetch text via `blob_url`, perform sentence-based chunking, and compute embeddings for each returned chunk using the configured model
AND the response must include chunk metadata (sequence_index, start_position, end_position, content) and embedding vectors with model metadata
AND the Go app must persist `ChunkNode`s and embeddings in Neo4j, linking them to the parent `ContentNode`.
AND no internal Kafka message is emitted or consumed between chunking and embedding.
```

### Requirement 4: Knowledge Graph Construction

**User Story:**
> As a system, I want to build a graph structure connecting content and chunks so that users can explore relationships within the knowledge base.

**Acceptance Criteria:**
```gherkin
GIVEN `ContentNode` and `ChunkNode`s with embeddings exist
WHEN the graph construction process runs
THEN `ChunkNode`s must be linked to their parent `ContentNode` with a `:HIERARCHICAL_PARENT` relationship.
AND `:SEMANTIC_LINK` relationships must be created between similar nodes (chunks or content) based on a configurable similarity threshold.
AND the system must support the creation of explicit `:STRUCTURAL_LINK` relationships via user or system actions.
AND all nodes and relationships must include provenance metadata (created_at, updated_at, etc.).
```

### Requirement 5b: LLM Summarization (future)

**User Story:**
> As a system, I want to generate concise summaries of content sources or clusters so that users can understand content at a glance.

**Acceptance Criteria:**
```gherkin
GIVEN `ContentNode` and associated `ChunkNode`s exist
WHEN summarization is triggered for a content source or cluster
THEN the Python service must produce an LLM-generated summary with model metadata (provider, model_id, version)
AND the summary must be stored on the target node (e.g., `ContentNode.summary`) or related summary node, with provenance
AND this feature may be disabled by configuration during the demo phase.
```

### Requirement 4b: Clustering (future)

**User Story:**
> As a system, I want to compute clusters of related content/chunks so that users can explore thematic groups.

**Acceptance Criteria:**
```gherkin
GIVEN a workspace (space) with embedded nodes
WHEN clustering is triggered
THEN the Python service must compute clusters using a configurable method and parameters
AND cluster assignments must be persisted (e.g., `ClusterNode`s with links to members) with provenance and optional cluster-level embeddings/summaries
AND this feature may be disabled by configuration during the demo phase.
```

### Requirement 5: Querying and Search

**User Story:**
> As a developer, I want to query the canvas service to retrieve nodes, relationships, and perform semantic searches so that can build user-facing exploration features.

**Acceptance Criteria:**
```gherkin
GIVEN the knowledge graph is populated
WHEN a read API request is made to fetch a node or its neighbors
THEN the API must return the requested graph data.
WHEN a semantic search query is executed via the API
THEN the API must return a list of nodes that are semantically similar to the query, using the vector index.
```

### Requirement 6: Non-functional Requirements

**Requirements:**
- **Performance:**
  - Vector search operations should complete in under 100ms for small to medium workspaces. Indexes must be created on `space_id` and `content_source_id` for fast lookups.
  - Python chunking and embedding processing must achieve P99 latency under 5 seconds for a 200,000-character input (on target deployment hardware).
- **Reliability:** All event handlers must be idempotent and support retries.
- **Data Management:** The system must support soft deletes for nodes and relationships.
 - **Observability:** The service must emit metrics for key stages (e.g., `document.received`, `chunking.completed`, `embedding.completed`). It must also produce structured logs for failures, and include OpenTelemetry traces/metrics across Go and Python components.
- **Security:** Incoming events must be validated for authentication and authorization. Sensitive data should be encrypted at rest. Logs must scrub/redact PII by default.
- **Configurability:** Chunking parameters (type=sentence only for now, target_size≈tokens, overlap%), embedding models, tokenizer, and similarity thresholds must be configurable via environment variables.
- **Tooling:** The service must include data migration scripts for setting up indexes and sample data.
 - **Communication:** Internal Go↔Python interactions use gRPC over localhost with increased message size limits; input text may be provided inline or by blob storage URL. The Go app is the sole external-facing service; the Python service is internal-only. Within `ms_canvas`, no internal Kafka topics are used; Kafka is only used to receive `document.processed` from `ms_document_process`.
 - **Deployment (Demo):** Both processes run in a single container under `supervisord`; health endpoints must reflect readiness of both Go and Python processes.