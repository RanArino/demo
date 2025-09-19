# Canvas Service Design Document

## 1. Introduction

This document translates the approved requirements in `requirements.md` into a concrete technical design for the unified `ms_canvas` service. The service uses a hybrid-process model: a Go primary process for API, orchestration, event consumption, and Neo4j persistence; and a co-located Python process for chunking, embedding, and ML/NLP tasks. Both processes run in a single container (demo phase) and communicate via local gRPC. Only the Go process communicates with external microservices; the Python process is internal-only.

Proto-first: `proto/canvas.proto` is the source of truth for all internal Go↔Python RPC contracts in this phase. Public API proto is deferred.

## 2. Architecture Overview

- Go (primary): public API (gRPC/HTTP), event consumers, orchestration, Neo4j access.
- Python (secondary): internal gRPC service for chunking, embeddings; later summarization and clustering.
- Communication: Go -> Python via localhost unary gRPC with increased `max_receive_message_length`.
- Storage: Neo4j for graph (nodes: ContentNode, ChunkNode; rels: HIERARCHICAL_PARENT, SEMANTIC_LINK, STRUCTURAL_LINK).
- Messaging: Kafka is used only to receive the upstream `document.processed` event from `ms_document_process`. No internal Kafka topics are used within `ms_canvas`.

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
│   │   ├── workflows/
│   │   │   ├── document.go
│   │   │   └── adapter.go
│   │   ├── gateway/
│   │   │   └── python/
│   │   │       └── chunking_and_embedding.go
│   │   ├── handler/
│   │   │   └── canvas_public.go
│   │   ├── domain/
│   │   │   ├── node_models.go
│   │   │   └── link_models.go
│   │   ├── infrastructure/
│   │   │   ├── repository/
│   │   │   │   └── neo4j/
│   │   │   │       ├── driver.go
│   │   │   │       ├── node_repository.go
│   │   │   │       └── link_repository.go
│   │   │   └── consumer/
│   │   │       └── kafka/
│   │   │           ├── consumer.go
│   │   │           └── handler.go
│   │   ├── events/
│   │   │   └── types.go
│   │   └── config/
│   │       └── config.go
│   ├── api/
│   │   └── proto/
│   │       └── v1/
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
│   └── v1/
│       └── canvas.proto
├── scripts/
│   └── start.sh
├── supervisord.conf
└── Dockerfile
```

## 4. Components

### 4.1 Go Application
- `cmd/main.go`: wires dependencies; starts public API server and the Kafka consumer; manages lifecycle.
- `internal/domain`: domain models (`ContentNode`, `ChunkNode`, `Embedding`, relationship types), repository interfaces.
- `internal/workflows`: orchestration for business use-cases (e.g., `ProcessDocument` coordinating Python pipeline + Neo4j persistence + metrics/idempotency). Invoked by Kafka handlers and/or public API handlers.
 - `internal/workflows`: orchestration for business use-cases (e.g., `ProcessDocument` coordinating Python pipeline + Neo4j persistence + metrics/idempotency). Invoked by Kafka handlers and/or public API handlers. Inside `ms_canvas`, do not use "Event" types; workflows accept non-event request structs (e.g., `ProcessInput`). Kafka is only an inbound boundary from `ms_document_process`.
- `internal/infrastructure/repository`: Neo4j repositories split into `NodeRepository` (nodes: upserts, embeddings, soft delete) and `LinkRepository` (relationships: hierarchical, semantic, structural); index creation helpers.
- `internal/infrastructure/consumer/kafka`: consumer and `handler.go` for inbound `document.processed` events (calls `internal/workflows`).
- `internal/gateway/python`: gRPC client to Python internal service (e.g., `chunking_and_embedding.go`).
- `internal/handler`: public API handlers (gRPC and optional HTTP gateway). Handlers map external/public protobuf services to workflow calls.
- `internal/port`: server setup (gRPC server, HTTP gateway, health endpoints).
- `api/proto/v1`: generated Go code for internal protobufs.

### 4.2 Python Application
- `app/services/chunking.py`: sentence-based text chunking with configurable target_size≈tokens (default 300) and overlap% (default 10%), using spaCy for sentence segmentation and tiktoken for token estimation.
- `app/services/embedding.py`: embedding generation with configurable provider/model.
- `app/pipelines/chunk_and_embed.py`: coarse-grained orchestration that composes `services/chunking` and `services/embedding` into a single function for the `ChunkAndEmbed` RPC. Handles batching and basic error propagation. (New)
- Future services: summarization and clustering (placeholders to be added, callable via internal gRPC).
- `app/server.py`: gRPC definitions implementation for internal API; delegates `ChunkAndEmbed` to `app/pipelines/chunk_and_embed.py`.
- `app/main.py`: boots the Python gRPC server.

#### 4.2.1 Neo4j KG Builder (neo4j_graphrag) — optional
The `neo4j_graphrag` (Neo4j GraphRAG / KG Builder) Python package may be used optionally for knowledge-graph construction (entity/relation extraction and Neo4j writing). However, for the chunking stage do not rely on `neo4j_graphrag`'s text-splitting components. Instead, implement sentence-based splitting explicitly in `app/services/chunking.py` using spaCy for robust, language-aware sentence segmentation. Use `neo4j_graphrag` only for downstream KG-building stages if desired; chunking responsibilities remain within the Python chunking service.

Follow the library's configuration options for KG-building (schema guidance, entity resolution, batch sizing) if and when it is used. See the Neo4j documentation for details: `https://neo4j.com/docs/neo4j-graphrag-python/current/user_guide_kg_builder.html`.

### 4.3 Proto Contracts
- `proto/v1/canvas.proto`: internal RPCs for chunking/embedding and the combined `ChunkAndEmbed`; messages include oneof input for inline text vs blob URL.
- Combined RPC: introduce `ChunkAndEmbed` to reduce round trips; keep existing `ChunkText` and `EmbedChunks` for modularity.

Public/external API (planned):
- `proto/public/canvas_public.proto` (planned): external/public gRPC API definitions (e.g., create/read/query endpoints like `CreateContentNodes`).
- `go_app/api/proto/public/v1` (planned): generated Go stubs for public API.
- Handlers for public API live in `internal/infrastructure/handler` and delegate to `internal/workflows`.

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
4. Trigger chunking (and embedding) orchestration directly (no internal Kafka).

### 6.2 Chunking
1. Go calls Python internal `ChunkText` with parameters (target_size≈tokens, overlap%, type=sentence) and either inline text or blob URL.
2. Python fetches text when blob URL is provided, normalizes content for splitting, performs sentence-based chunking using spaCy, estimates tokens via tiktoken, and returns chunk contents with `start_position` and `end_position` character offsets relative to the original text.
3. Go persists `ChunkNode`s and emits `chunking.completed` metric/event.

### 6.3 Embedding
1. The orchestrator runs embedding inline after chunking (same flow) or uses the combined RPC.
2. If modular: Go calls Python internal `EmbedChunks` with model configuration; Python returns vectors and model metadata; Go updates `ChunkNode.embedding` and adds model metadata.
3. Emit `embedding.completed` metric.

### 6.3b Combined Chunking + Embedding
1. Go calls `ChunkAndEmbed` with chunking and embedding configs plus oneof `text | blob_url` via `internal/gateway/python`.
2. Python `server.py` delegates to `app/pipelines/chunk_and_embed.py`, which performs sentence chunking and computes embeddings for each returned chunk by calling `services/chunking` and `services/embedding`.
3. Response returns batches (or full list for demo) of chunks with vectors and model metadata.
4. Go persists `ChunkNode`s and embeddings via `internal/infrastructure/repository/neo4j`, and emits both `chunking.completed` and `embedding.completed` metrics after successful writes.
5. Idempotency: upsert by `(content_source_id, sequence_index)` for chunks and overwrite/update embedding vectors for existing chunk nodes to avoid duplicates on reprocessing.
6. Gateway integration: the orchestrator uses `internal/gateway/python/chunking_and_embedding.go` client to invoke `ChunkAndEmbed` with increased `max_receive_message_length`; for demo, unary responses are used, but responses may be processed in optional batches.
7. Error handling: if Neo4j persistence fails, do not emit metrics; return an error to the caller and allow retry. Partial successes should be retried safely due to idempotent upsert semantics.

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
- `rpc ChunkAndEmbed(ChunkAndEmbedRequest) returns (ChunkAndEmbedResponse)` (new; response may be updated to streaming in a later phase)

Note: Exact message fields will map to `requirements.md` metadata (space_id, content_source_id, etc.).

## 8. Configuration

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

## 9. Observability & Reliability

- Structured logging with PII scrubbing; request and event correlation IDs.
- Metrics: `document.received`, `chunking.completed` (emitted by Go after persistence), `embedding.completed`, consumer lag, RPC latency; Python processing P99.
- Tracing across Go and Python via OpenTelemetry exporters.
- Idempotency keys for event processing; retry with backoff; poison-queue handling.

## 10. Deployment & Runtime

- Single Dockerfile builds Go binary and installs Python deps.
- `supervisord` or `start.sh` to run both processes; health checks for both. Go exposes primary health endpoints; Python exposes internal-only healthz for readiness checks used by Go.
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
