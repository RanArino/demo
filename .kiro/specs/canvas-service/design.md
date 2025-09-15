# Canvas Service Design Document

## 1. Introduction

This document translates the approved requirements in `requirements.md` into a concrete technical design for the unified `ms_canvas` service. The service uses a hybrid-process model: a Go primary process for API, orchestration, event consumption, and Neo4j persistence; and a co-located Python process for chunking, embedding, and ML/NLP tasks. Both processes run in a single container and communicate via local gRPC.

Proto-first: `proto/canvas.proto` is the source of truth for all internal Go↔Python RPC contracts in this phase. Public API proto is deferred.

## 2. Architecture Overview

- Go (primary): public API (gRPC/HTTP), event consumers, orchestration, Neo4j access.
- Python (secondary): internal gRPC service for chunking and embeddings.
- Communication: Go -> Python via localhost unary gRPC with increased `max_receive_message_length`.
- Storage: Neo4j for graph (nodes: ContentNode, ChunkNode; rels: HIERARCHICAL_PARENT, SEMANTIC_LINK, STRUCTURAL_LINK).
- Messaging: Kafka (example) for document processing completion events from `ms_document_process`.

### High-level Diagram

```
[Kafka] -> [Go Consumer] -> [Go Orchestrator] -> [Python gRPC] -> [Go Neo4j Repo] -> [Neo4j]
                                   ^                                               |
                                   |---------------- Public API (gRPC/HTTP) -------|
```

## 3. Directory Layout

```
ms_canvas/
├── go_app/
│   ├── cmd/
│   │   └── main.go
│   ├── internal/
│   │   ├── application/
│   │   ├── domain/
│   │   ├── infrastructure/
│   │   │   ├── consumer/
│   │   │   │   └── kafka/
│   │   │   │       └── handler.go
│   │   │   ├── client/
│   │   │   ├── handler/
│   │   │   └── repository/
│   │   └── port/
│   ├── go.mod
│   └── go.sum
├── python_app/
│   ├── app/
│   │   ├── services/
│   │   │   ├── chunking.py
│   │   │   └── embedding.py
│   │   ├── server.py
│   │   └── main.py
│   └── pyproject.toml
├── proto/
│   └── canvas.proto
├── scripts/
│   └── start.sh
├── supervisord.conf
└── Dockerfile
```

## 4. Components

### 4.1 Go Application
- `cmd/main.go`: wires dependencies; starts public API server and the Kafka consumer; manages lifecycle.
- `internal/domain`: domain models (`ContentNode`, `ChunkNode`, `Embedding`, relationship types), repository interfaces.
- `internal/application`: use cases (ingest event, chunking orchestration, embedding orchestration, graph linking, query/search).
- `internal/infrastructure/repository`: Neo4j implementation of repositories; index creation helpers.
- `internal/infrastructure/consumer/kafka`: consumer and `handler.go` for `document.processed` events.
- `internal/infrastructure/client`: gRPC client to Python internal service.
- `internal/infrastructure/handler`: public API handlers (gRPC and optional HTTP gateway).
- `internal/port`: server setup (gRPC server, HTTP gateway, health endpoints).

### 4.2 Python Application
- `app/services/chunking.py`: sentence-based text chunking with configurable target_size≈tokens (default 300) and overlap% (default 10%), using spaCy for sentence segmentation and tiktoken for token estimation.
- `app/services/embedding.py`: embedding generation with configurable provider/model.
- `app/server.py`: gRPC definitions implementation for internal API.
- `app/main.py`: boots the Python gRPC server.

#### 4.2.1 Neo4j KG Builder (neo4j_graphrag) — optional
The `neo4j_graphrag` (Neo4j GraphRAG / KG Builder) Python package may be used optionally for knowledge-graph construction (entity/relation extraction and Neo4j writing). However, for the chunking stage do not rely on `neo4j_graphrag`'s text-splitting components. Instead, implement sentence-based splitting explicitly in `app/services/chunking.py` using spaCy for robust, language-aware sentence segmentation. Use `neo4j_graphrag` only for downstream KG-building stages if desired; chunking responsibilities remain within the Python chunking service.

Follow the library's configuration options for KG-building (schema guidance, entity resolution, batch sizing) if and when it is used. See the Neo4j documentation for details: `https://neo4j.com/docs/neo4j-graphrag-python/current/user_guide_kg_builder.html`.

### 4.3 Proto Contracts
- `proto/canvas.proto`: internal RPCs for chunking and embedding; messages include oneof input for inline text vs blob URL.

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
3. Create `ContentNode` with provenance.
4. Enqueue chunking orchestration.

### 6.2 Chunking
1. Go calls Python internal `ChunkText` with parameters (target_size≈tokens, overlap%, type=sentence) and either inline text or blob URL.
2. Python fetches text when blob URL is provided, normalizes content for splitting, performs sentence-based chunking using spaCy, estimates tokens via tiktoken, and returns chunk contents with `start_position` and `end_position` character offsets relative to the original text.
3. Go persists `ChunkNode`s and emits `chunking.completed` metric/event.

### 6.3 Embedding
1. Go calls Python internal `EmbedChunks` with model configuration.
2. Python returns vectors and model metadata.
3. Go updates `ChunkNode.embedding` and adds model metadata.
4. Emit `embedding.completed` metric/event.

### 6.4 Graph Construction
1. Link chunks to content via `:HIERARCHICAL_PARENT`.
2. Compute similarities; create `:SEMANTIC_LINK` above threshold.
3. Support explicit `:STRUCTURAL_LINK` writes via public API.

### 6.5 Querying & Search
- Get node by id; get neighbors with relationship filters.
- Semantic search: input query -> temporary embedding (Python) -> similarity on vector index -> return top-N nodes.

## 7. API Design (Proto Sketches)

### 7.1 Public API (`canvas_public.proto`)
- `rpc GetNode(GetNodeRequest) returns (GetNodeResponse)`
- `rpc GetNeighbors(GetNeighborsRequest) returns (GetNeighborsResponse)`
- `rpc SemanticSearch(SemanticSearchRequest) returns (SemanticSearchResponse)`
- `rpc CreateStructuralLink(CreateStructuralLinkRequest) returns (CreateStructuralLinkResponse)`

### 7.2 Internal API (`canvas.proto`)
- `rpc ChunkText(ChunkTextRequest) returns (ChunkTextResponse)`
- `rpc EmbedChunks(EmbedChunksRequest) returns (EmbedChunksResponse)`
- `rpc EmbedQuery(EmbedQueryRequest) returns (EmbedQueryResponse)`

Note: Exact message fields will map to `requirements.md` metadata (space_id, content_source_id, etc.).

## 8. Configuration

- Chunking (env vars, read by Python app):
  - `CANVAS_CHUNK_TYPE=sentence` (fixed for this phase)
  - `CANVAS_CHUNK_TARGET_TOKENS=300`
  - `CANVAS_CHUNK_OVERLAP_PERCENT=10`
  - `CANVAS_TOKENIZER=tiktoken:cl100k_base` (pluggable)
  - `CANVAS_SPACY_MODEL` (optional; future)
  - `CANVAS_MAX_GRPC_MSG_BYTES` (increased limit for unary RPC)
- Embedding: provider (OpenAI, local), model_id, version.
- Similarity threshold for `SEMANTIC_LINK` creation.
- Kafka: brokers, topic, consumer group, auth.
- Neo4j: uri, user, password; vector index configuration.
- Security: event authentication/authorization; secrets management; PII scrubbing in logs.

## 9. Observability & Reliability

- Structured logging with PII scrubbing; request and event correlation IDs.
- Metrics: `document.received`, `chunking.completed` (emitted by Go after persistence), `embedding.completed`, consumer lag, RPC latency; Python processing P99.
- Tracing across Go and Python via OpenTelemetry exporters.
- Idempotency keys for event processing; retry with backoff; poison-queue handling.

## 10. Deployment & Runtime

- Single Dockerfile builds Go binary and installs Python deps.
- `supervisord` or `start.sh` to run both processes; health checks for both.
- Readiness: Go API ready, Python RPC reachable; liveness for both.
- Config via env vars; secrets via mounted files or secret manager.

## 11. Security

- Validate and authorize incoming events; signed messages or mTLS for Kafka.
- Least-privilege Neo4j credentials; encrypt sensitive data at rest.
- Internal gRPC bound to localhost; not exposed externally.

## 12. NFR Mapping

- Performance: vector index lookups < 100ms for S/M workspaces; Python chunking+embedding P99 < 5s for 200k characters.
- Reliability: idempotent handlers; retries; backoff; DLQ.
- Data Management: soft delete strategy.
- Observability: metrics and structured logs at key stages.
- Security: event validation/authn/authz; encryption at rest.
- Configurability: chunking, embedding, thresholds via config.
- Tooling: migration scripts for indexes and sample data.

## 13. Risks & Open Questions

- Vector indexing in Neo4j vs external vector store; fallback plan.
- Embedding provider quotas/costs; local model alternative.
- Message schema evolution and compatibility.
- Backpressure handling for large documents and bulk events.
