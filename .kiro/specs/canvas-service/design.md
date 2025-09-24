# Canvas Service Design Document

## 1. Introduction

This document translates the approved requirements in `requirements.md` into a concrete technical design for the unified `ms_canvas` service. The service uses a hybrid-process model: a Go primary process for API, orchestration, event consumption, and Neo4j persistence; and a co-located Python process for chunking, embedding, and ML/NLP tasks. Both processes run in a single container (demo phase) and communicate via local gRPC. Only the Go process communicates with external microservices; the Python process is internal-only.

Proto-first: `proto/private/v1/canvas_private.proto` is the source of truth for all internal Go↔Python RPC contracts in this phase. Public API proto is deferred.

## 2. Architecture Overview

- Go (primary): public API (gRPC/HTTP), event consumers, orchestration, Neo4j access.
- Python (secondary): internal gRPC service for chunking, embeddings; later summarization and clustering.
- Communication: Go -> Python via localhost unary gRPC with increased `max_receive_message_length`.
- Storage: Neo4j for graph (nodes: ContentNode, ChunkNode; rels: HIERARCHICAL_PARENT, SEMANTIC_LINK, STRUCTURAL_LINK).
- Messaging: Kafka is used only to receive the upstream `document.processed` event from `ms_document_process`. No internal Kafka topics are used within `ms_canvas`.

### Architectural Flow

```
External Events:          Public API:
[Kafka] ----+             [gRPC Clients] ----+
            |                                |
            v                                v
      [events/kafka] ---------> [server/grpc]
            |                                |
            +---> [service/*] <--------------+
                  (business logic)
                        |
                        +---> [gateway/python] -> [Python gRPC]
                        +---> [repository/neo4j] -> [Neo4j]
```

**Simple, Conventional Structure** (aligned with `ms_knowledge`):
- **Transport layers** (`events`, `server`): Handle Kafka/gRPC protocol concerns
- **Business logic** (`service`): Domain workflows and coordination
- **Infrastructure** (`repository`, `gateway`): External system integrations

## 3. Directory Layout

```
ms_canvas/
├── go_app/
│   ├── cmd/
│   │   └── main.go
│   ├── internal/
│   │   ├── service/
│   │   │   ├── search_service.go // SearchService interface and its implementation
│   │   │   ├── node_service.go // NodeService interface and its implementation
│   │   │   └── link_service.go // LinkService interface and its implementation
│   │   ├── gateway/
│   │   │   └── python/
│   │   │       ├── factory.go
│   │   │       └── chunking_and_embedding.go
│   │   ├── server/
│   │   │   └── grpc.go
│   │   ├── repository/
│   │   │   └── neo4j/
│   │   │       ├── driver.go
│   │   │       ├── node_repository.go
│   │   │       └── link_repository.go
│   │   ├── events/
│   │   │   ├── types.go
│   │   │   └── kafka/
│   │   │       ├── consumer.go
│   │   │       └── handler.go
│   │   └── config/
│   │       └── config.go
│   ├── api/
│   │   └── proto/
│   │       └── private/v1/
│   │           └── (generated gRPC client code)
│   └── README.md
├── python_app/
│   ├── app/
│   │   ├── services/
│   │   │   ├── chunking.py
│   │   │   └── embedding.py
│   │   ├── pipelines/
│   │   │   └── chunk_and_embed.py
│   │   ├── server.py
│   │   └── main.py
│   └── pyproject.toml
├── proto/
│   ├── private/
│   │   └── v1/
│   │       └── canvas_private.proto
│   └── public/
│       └── v1/
│           └── canvas.proto
├── scripts/
│   └── start.sh
├── supervisord.conf
└── Dockerfile
```

## 4. Components

### 4.1 Go Application
- `cmd/main.go`: wires dependencies; starts gRPC public API server, HTTP gateway (optional), and the Kafka consumer; manages lifecycle.
- `internal/server`: gRPC server setup, service registration, health checks, graceful shutdown (follows `ms_knowledge` pattern).
- `internal/service`: Business logic layer containing workflows and domain services (e.g., `node_service.go`, `link_service.go`, `search_service.go`, `event_orchestrator.go`, `task_executor.go` for orchestration and extensible task execution). Services coordinate gateway calls and repository operations. Features worker pools, priority-based task execution, and plugin architecture for task types.
- `internal/repository`: Data access layer with Neo4j repositories (`NodeRepository`, `LinkRepository`) and driver setup.
- `internal/events`: Event-related code including event types (`types.go`) and Kafka consumer implementation that delegates to `internal/service`.
- `internal/service/event_orchestrator.go`: Core event orchestration service implementing the EventHandler interface to coordinate document ingestion workflows, ContentNode creation, and extensible task execution.
- `internal/gateway/python`: gRPC client to Python internal service for chunking/embedding operations.
- `api/proto/private/v1`: generated Go code for private/internal protobufs.
- `api/proto/public/v1`: generated Go code for public protobufs.

### 4.2 Python Application
- `app/services/chunking.py`: sentence-based text chunking with configurable target_size≈tokens (default 300) and overlap% (default 10%), using spaCy for sentence segmentation and tiktoken for token estimation.
- `app/services/embedding.py`: embedding generation with configurable provider/model.
- `app/pipelines/chunk_and_embed.py`: coarse-grained orchestration that composes `services/chunking` and `services/embedding` into a single function for the `ChunkEmbed` RPC. Handles batching and basic error propagation. (New)
- Future services: summarization and clustering (placeholders to be added, callable via internal gRPC).
- `app/server.py`: gRPC definitions implementation for internal API; delegates `ChunkEmbed` to `app/pipelines/chunk_and_embed.py`.
- `app/main.py`: boots the Python gRPC server.

#### 4.2.1 Neo4j KG Builder (neo4j_graphrag) — optional
The `neo4j_graphrag` (Neo4j GraphRAG / KG Builder) Python package may be used optionally for knowledge-graph construction (entity/relation extraction and Neo4j writing). However, for the chunking stage do not rely on `neo4j_graphrag`'s text-splitting components. Instead, implement sentence-based splitting explicitly in `app/services/chunking.py` using spaCy for robust, language-aware sentence segmentation. Use `neo4j_graphrag` only for downstream KG-building stages if desired; chunking responsibilities remain within the Python chunking service.

Follow the library's configuration options for KG-building (schema guidance, entity resolution, batch sizing) if and when it is used. See the Neo4j documentation for details: `https://neo4j.com/docs/neo4j-graphrag-python/current/user_guide_kg_builder.html`.

### 4.3 Proto Contracts
- `proto/private/v1/canvas_private.proto`: internal RPCs for chunking/embedding and the combined `ChunkEmbed`; messages include oneof input for inline text vs blob URL.
- Combined RPC: introduce `ChunkEmbed` to reduce round trips; keep existing `ChunkText` and `EmbedChunks` for modularity.

Public/external API (planned):
- `proto/public/v1/canvas.proto` (planned): external/public gRPC API definitions (e.g., create/read/query endpoints like `CreateContentNodes`).
- `go_app/api/proto/public/v1` (planned): generated Go stubs for public API.
- Handlers for public API live in `internal/server` and delegate to `internal/service`.

## 5. Data Model (Neo4j)

Nodes:
- `ContentNode { id, space_id, content_source_id, title?, created_at, updated_at, deleted_at? }`
- `ChunkNode { id, content_source_id, position, content, embedding: float[], model_id, model_version, created_at, updated_at, deleted_at? }`

Relationships:
- `(:ContentNode)-[:HIERARCHICAL_PARENT]->(:ChunkNode)`
- `(:ChunkNode|:ContentNode)-[:SEMANTIC_LINK { score }]->(:ChunkNode|:ContentNode)`
- `[:STRUCTURAL_LINK { type, created_by }]`

Indexes:
- BTREE on `ContentNode(space_id)`, `ChunkNode(content_source_id)`
- Vector index on `ChunkNode.embedding` (via Neo4j vector index capability or external vector store if required)

Soft delete: `deleted_at != null` implies filtered from reads.

## 6. Flows

### 6.1 Event-Driven Ingestion
1. Kafka emits `document.processed` with metadata (space_id, content_source_id, location, etc.).
2. Go consumer validates auth/signature and schema.
3. EventOrchestrator processes the event, creates `ContentNode` with provenance.
4. EventOrchestrator triggers chunking (and embedding) orchestration via extensible task execution framework.

### 6.2 Chunking
1. Go calls Python internal `ChunkText` with parameters (target_size≈tokens, overlap%, type=sentence) and either inline text or blob URL.
2. Python fetches text when blob URL is provided, normalizes content for splitting, performs sentence-based chunking using spaCy, estimates tokens via tiktoken, and returns chunk contents with `start_position` and `end_position` character offsets relative to the original text.
3. Go persists `ChunkNode`s and emits `chunking.completed` metric/event.

### 6.3 Embedding
1. The orchestrator runs embedding inline after chunking (same flow) or uses the combined RPC.
2. If modular: Go calls Python internal `EmbedChunks` with model configuration; Python returns vectors and model metadata; Go updates `ChunkNode.embedding` and adds model metadata.
3. Emit `embedding.completed` metric.

### 6.3b Combined Chunking + Embedding
1. Go calls `ChunkEmbed` with chunking and embedding configs plus oneof `text | blob_url` via `internal/gateway/python`.
2. Python `server.py` delegates to `app/pipelines/chunk_and_embed.py`, which performs sentence chunking and computes embeddings for each returned chunk by calling `services/chunking` and `services/embedding`.
3. Response returns batches (or full list for demo) of chunks with vectors and model metadata.
4. Go persists `ChunkNode`s and embeddings via `internal/infrastructure/repository/neo4j`, and emits both `chunking.completed` and `embedding.completed` metrics after successful writes.
5. Idempotency: upsert by `(content_source_id, sequence_index)` for chunks and overwrite/update embedding vectors for existing chunk nodes to avoid duplicates on reprocessing.
6. Gateway integration: the orchestrator uses `internal/gateway/python/chunking_and_embedding.go` client to invoke `ChunkEmbed` with increased `max_receive_message_length`; for demo, unary responses are used, but responses may be processed in optional batches.
7. Error handling: if Neo4j persistence fails, do not emit metrics; return an error to the caller and allow retry. Partial successes should be retried safely due to idempotent upsert semantics.

### 6.4 Graph Construction
1. Link chunks to content via `:HIERARCHICAL_PARENT`.
2. Compute similarities; create `:SEMANTIC_LINK` above threshold.
3. Support explicit `:STRUCTURAL_LINK` writes via public API.

### 6.5 Querying & Search
- Get node by id; get neighbors with relationship filters.
- Semantic search: input query -> embedding (Python) -> similarity on vector index -> return top-N nodes with database similarity scores.

## 7. API Endpoint Design
[This section describes the specifications for each API endpoint. Clarify requests, responses, authentication requirements, etc.]

### `POST /api/semantic-search`
- **Description:** Performs semantic search on nodes in the knowledge graph.
- **Authentication:** API key required
- **Request Body:**
  ```json
  {
    "space_id": "string",
    "query": "string",
    "top_k": 25,
    "node_types": ["CONTENT", "CHUNK", "CLUSTER"]
  }
  ```
- **Response (200 OK):**
  ```json
  {
    "results": [
      {
        "node": { /* node object */ },
        "score": 0.95
      }
    ]
  }
  ```

### `GET /api/nodes/{id}`
- **Description:** Retrieves a specific node by ID.
- **Authentication:** API key required
- **Response (200 OK):**
  ```json
  {
    "id": "string",
    "type": "CONTENT",
    "content": "string",
    "metadata": { /* additional metadata */ }
  }
  ```

## 8. UI/UX Design
[As this is a backend service, UI/UX design is not applicable for the core service functionality. The service provides gRPC and REST APIs for frontend applications to consume.]

- **Note:** UI/UX design would be handled by consuming frontend applications
- **API Documentation:** OpenAPI/Swagger specifications will be generated for all endpoints
- **Client Libraries:** Generated gRPC client libraries will be provided for different programming languages

## 9. API Design (Proto Sketches)

### 9.1 Public API (`canvas_public.proto`)
- `rpc GetNode(GetNodeRequest) returns (GetNodeResponse)`
- `rpc GetNeighbors(GetNeighborsRequest) returns (GetNeighborsResponse)`
- `rpc SemanticSearch(SemanticSearchRequest) returns (SemanticSearchResponse)`
- `rpc CreateStructuralLink(CreateStructuralLinkRequest) returns (CreateStructuralLinkResponse)`

### 9.2 Internal API (`proto/private/v1/canvas_private.proto`)
- `rpc ChunkText(ChunkTextRequest) returns (ChunkTextResponse)`
- `rpc EmbedChunks(EmbedChunksRequest) returns (EmbedChunksResponse)`
- `rpc EmbedQuery(EmbedQueryRequest) returns (EmbedQueryResponse)`
- `rpc ChunkEmbed(ChunkEmbedRequest) returns (ChunkEmbedResponse)` (new; response may be updated to streaming in a later phase)

Note: Exact message fields will map to `requirements.md` metadata (space_id, content_source_id, etc.).

## 10. Configuration

- Chunking (env vars, read by Python app):
  - `CANVAS_CHUNK_TYPE=sentence` (fixed for this phase)
  - `CANVAS_CHUNK_TARGET_TOKENS=300`
  - `CANVAS_CHUNK_OVERLAP_PERCENT=10`
  - `CANVAS_TOKENIZER=tiktoken:cl100k_base` (pluggable)
  - `CANVAS_SPACY_MODEL` (optional; future)
  - `CANVAS_MAX_GRPC_MSG_BYTES` (increased limit for unary RPC)
  - `CANVAS_BATCH_SIZE` (optional; batch size for combined responses)
- Embedding: provider (OpenAI, local), model_id, version.
- Similarity threshold for `SEMANTIC_LINK` creation.
- Kafka: brokers, topic, consumer group, auth.
- Neo4j: uri, user, password; vector index configuration.
- Security: event authentication/authorization; secrets management; PII scrubbing in logs.

## 11. Observability & Reliability

- Structured logging with PII scrubbing; request and event correlation IDs.
- Metrics: `document.received`, `chunking.completed` (emitted by Go after persistence), `embedding.completed`, consumer lag, RPC latency; Python processing P99.
- Tracing across Go and Python via OpenTelemetry exporters.
- Idempotency keys for event processing; retry with backoff; poison-queue handling.

## 12. Deployment & Runtime

- Single Dockerfile builds Go binary and installs Python deps.
- `supervisord` or `start.sh` to run both processes; health checks for both. Go exposes primary health endpoints; Python exposes internal-only healthz for readiness checks used by Go.
- Readiness: Go API ready, Python RPC reachable; liveness for both.
- Config via env vars; secrets via mounted files or secret manager.

## 13. Security

- Validate and authorize incoming events; signed messages or mTLS for Kafka.
- Least-privilege Neo4j credentials; encrypt sensitive data at rest.
- Internal gRPC bound to localhost; not exposed externally.

## 14. NFR Mapping

- Performance: vector index lookups < 100ms for S/M workspaces; Python chunking+embedding P99 < 5s for 200k characters.
- Reliability: idempotent handlers; retries; backoff; DLQ.
- Data Management: soft delete strategy.
- Observability: metrics and structured logs at key stages.
- Security: event validation/authn/authz; encryption at rest.
- Configurability: chunking, embedding, thresholds via config.
- Tooling: migration scripts for indexes and sample data.

## 15. Risks & Open Questions

- Vector indexing in Neo4j vs external vector store; fallback plan.
- Embedding provider quotas/costs; local model alternative.
- Message schema evolution and compatibility.
- Backpressure handling for large documents and bulk events.

