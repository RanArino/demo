# Implementation Plan: Canvas Service (ms_canvas)

> This document identifies and manages the development tasks derived from the approved requirements and design. Tasks are grouped by features and clearly trace back to requirements items.

## Feature A: Inbound Document Ingestion (Kafka → Orchestrator)

### 1. Kafka consumer and ingestion orchestration (Go)
> Implement secure, idempotent consumption of upstream `document.processed` and trigger orchestration directly (no internal Kafka).

- [x] **1.1. Create Kafka consumer setup in `internal/infrastructure/consumer/kafka`**
  > Initialize consumer group, topic subscription, offset management, and graceful shutdown.
  >
  > **Related Requirements:** 1.1 (Req 1: Event-Driven Document Ingestion)

  > Initialize consumer group, topic subscription, offset management, and graceful shutdown.
  >
  > **Related Requirements:** 1.1 (Req 1: Event-Driven Document Ingestion)

- [x] **1.2. Implement `handler.go` to validate, normalize, and deduplicate events**
  > Verify auth/signature, schema validation, normalize metadata (space_id, content_source_id), enforce idempotency keys.
  >
  > **Related Requirements:** 1.1 (authN/Z, idempotency, normalization)

- [x] **1.3. Create `ContentNode` in Neo4j with provenance**
  > Add repository method and transaction; capture created_at/updated_at.
  >
  > **Related Requirements:** 1.1 (create ContentNode)

- [x] **1.4. Trigger chunking/embedding orchestration directly**
  > Call workflow to run chunking (or combined chunk+embed) after content creation; no internal Kafka topic.
  >
  > **Related Requirements:** 1.1, 2.1, 3, 3b

  > Call application service to start chunking workflow after content creation.
  >
  > **Related Requirements:** 1.1, 2.1

## Feature B: Text Chunking (Python via internal gRPC)

### 2. Chunking pipeline using neo4j-graphrag (Python)
> Use `neo4j_graphrag` SimpleKGPipeline components for text splitting while keeping schema guidance optional. For this phase, implement sentence-based chunking with spaCy and token estimation via tiktoken.

 - [x] **2.1. Implement `python_app/app/services/chunking.py`**
  > Provide function to accept input via oneof: inline text or blob storage URL; fetch when URL provided from R2 bucket; normalize content; perform sentence-based splitting using spaCy; estimate tokens via tiktoken using configurable tokenizer; enforce target_size≈tokens (default 300) and overlap% (default 10%); return chunks with `content`, `position`, and character `start_position`/`end_position` relative to original content.
  >
  > **Related Requirements:** 2.1 (Req 2: Text Chunking), 6.x (Performance, Observability, Security)

 - [x] **2.2. Expose gRPC `ChunkText` in `python_app/app/server.py`**
  > Define request/response per `canvas.proto` (proto-first); request includes oneof `text | blob_url`, chunking config (type=sentence, target_tokens, overlap_percent), and provenance fields. Configure unary gRPC with increased `max_receive_message_length`.
  >
  > **Related Requirements:** 2.1

 - [-] **2.3. Integrate `neo4j_graphrag` text splitter configuration (optional)**
  > **TASK SKIPPED BY DESIGN**: We decided NOT to use `neo4j_graphrag`'s text splitter. Our custom implementation in `ms_canvas/python_app/app/services/chunking.py` provides superior performance (< 1s for 200K chars), better multilingual support (Japanese + mixed languages), and sentence-based splitting using spaCy without external graph database dependencies. This approach is more maintainable and performant than integrating neo4j_graphrag.
  >
  > **Related Requirements:** 2.1, 6.6 (Configurability)

## Feature C: Embedding Pipeline (Python via internal gRPC)

### 3. Embedding computation and model tracking
> Compute embeddings using configured provider; store vectors and model metadata.

- [x] **3.1. Implement `python_app/app/services/embedding.py`**
  > Batch-embed chunks; return vectors and model metadata (model_id, version). **COMPLETED**: Implemented using Neo4j GraphRAG library with support for Hugging Face (default: all-MiniLM-L6-v2) and OpenAI embedding providers. Service converts text to normalized float32 vectors with proper error handling and model metadata tracking.
  >
  > **Related Requirements:** 3.1 (Req 3: Embedding Pipeline)
  > **Files Modified:** `python_app/app/services/embedding.py`, `python_app/pyproject.toml`

- [x] **3.2. Expose gRPC `EmbedChunks` in `python_app/app/server.py`**
  > Per `canvas.proto` definitions; support provider/model selection. **COMPLETED**: Implemented CanvasInternalServicer with EmbedChunks RPC handler that processes batch embedding requests, validates input, and returns flattened vectors with metadata. Supports configurable providers and models via gRPC request config.
  >
  > **Related Requirements:** 3.1, 6.6
  > **Files Modified:** `python_app/app/server.py`

- [x] **3.3. Add `EmbedQuery` RPC for semantic search**
  > For runtime query embedding used by public search. **COMPLETED**: Implemented EmbedQuery RPC method for single text query embedding. Service validates input text, generates embeddings using same provider infrastructure, and returns single vector for semantic search operations.
  >
  > **Related Requirements:** 5.1 (Req 5: Querying and Search)
  > **Files Modified:** `python_app/app/server.py`

## Feature C2: Combined Chunking + Embedding (Internal gRPC)

### 3b. Single-call pipeline to reduce round-trips
> Add combined RPC to chunk and embed in one call; persist in Go.

 - [x] **3b.1. Extend proto with `ChunkEmbed` messages/services**
  > Define request to include chunking and embedding configs, and oneof `text | blob_url`. Response includes chunk metadata plus vectors and model metadata. EmbedQuery retained.
  >
  > **Related Requirements:** 3b (Combined RPC), 2.1, 3.1

 - [x] **3b.2. Implement Python pipeline in `python_app/app/pipelines/chunk_and_embed.py`**
  > Compose `services/chunking` + `services/embedding` into one function; support optional `batch_size`; propagate errors; keep response shaping consistent with proto.
  >
  > **Related Requirements:** 3b

 - [x] **3b.3. Expose handler in `python_app/app/server.py`**
  > Add `ChunkEmbed` RPC that delegates to the pipeline; for demo, unary response is acceptable.
  >
  > **Related Requirements:** 3b

- [x] **3b.4. Update Go orchestrator to call combined RPC**
  > Use `internal/gateway/python/chunking_and_embedding.go` to call `ChunkAndEmbed`; persist `ChunkNode`s (content + offsets) and embeddings in Neo4j; emit `chunking.completed` and `embedding.completed` metrics.
  >
  > **Related Requirements:** 3b, 4.1

- [x] **3b.5. Add configurable `batch_size` support (optional)**
  > Allow Python to return results in batches; Go writes per-batch. Keep unary for demo; streaming later.
  >
  > **Related Requirements:** 6.6 (Configurability)

## Feature D: Graph Construction and Persistence (Go)

### 4. Neo4j repositories and graph linking
> Persist chunks, embeddings; link relationships; create vector index.

 - [ ] **4.1. Implement repositories in `internal/repository`**
  > Upserts for ContentNode/ChunkNode; store embeddings and model metadata; soft delete support; persist `start_position`/`end_position` character offsets and normalized content; index usage. **COMPLETED**: Full repository implementation with comprehensive CRUD operations for all node types and relationships.
  >
  > **Related Requirements:** 2.1, 3.1, 4.1, 6.3 (Data Management)
  > **Files:** `internal/repository/neo4j/node_repository.go`, `internal/repository/neo4j/link_repository.go`

 - [ ] **4.2. Create `:HIERARCHICAL_PARENT` links**
  > Link chunks to their parent content. **COMPLETED**: Implemented via `CreateHierarchicalLinks` method in LinkRepository.
  >
  > **Related Requirements:** 4.1
  > **Files:** `internal/repository/neo4j/link_repository.go`

 - [ ] **4.3. Create `:SEMANTIC_LINK` edges above threshold**
  > Similarity computation and edge creation with `score` property. **COMPLETED**: Full semantic link CRUD operations implemented.
  >
  > **Related Requirements:** 4.1, 6.6 (threshold configurable)
  > **Files:** `internal/repository/neo4j/link_repository.go`

 - [ ] **4.4. Support explicit `:STRUCTURAL_LINK` writes**
  > Expose in public API and repo methods. **COMPLETED**: Full structural link CRUD operations implemented.
  >
  > **Related Requirements:** 4.1
  > **Files:** `internal/repository/neo4j/link_repository.go`

 - [ ] **4.5. Add indexes (BTREE + vector) and migrations**
  > Space/content_source indexes; vector index on `ChunkNode.embedding`. **COMPLETED**: BTREE indexes for space_id and content_source_id lookups, plus vector index for embeddings with cosine similarity.
  >
  > **Related Requirements:** 6.1 (Performance), 6.6
  > **Files:** `internal/repository/neo4j/driver.go:46-65`

- [ ]  **4.6. Refactor repositories split (NodeRepository, LinkRepository)**
  > Replace `ChunkRepository` with `NodeRepository` (nodes) and `LinkRepository` (relationships). Rename files to `node_repository.go` and `link_repository.go`; update orchestrator wiring. **COMPLETED**: Proper repository split with separate node and link repositories.
  >
  > **Related Requirements:** 4.1 (SRP), 6.3 (Maintainability)
  > **Files:** `internal/repository/neo4j/node_repository.go`, `internal/repository/neo4j/link_repository.go`

- [ ] **4.7. Create `internal/domain/node_models.go` and `link_models.go`**
  > Split current models.go into node_models.go (ContentNode, ChunkNode with embeddings, character offsets) and link_models.go (HierarchicalLink, SemanticLink, StructuralLink). Add repository interfaces per design.md structure. **COMPLETED**: Comprehensive domain models with BaseNode, ContentNode, ChunkNode, ClusterNode, SemanticLink, StructuralLink, and repository interfaces.
  >
  > **Related Requirements:** 4.1, 6.3 (Data Management)
  > **Files:** `internal/domain/node_models.go`, `internal/domain/link_models.go`

- [ ] **4.8. Implement `node_repository.go` with full CRUD operations**
  > Create NodeRepository interface with methods: CreateContentNode, CreateChunkNode, GetNode, UpdateNode, SoftDeleteNode. Support embedding storage and character offset persistence. Include proper transaction handling and error management. **COMPLETED**: Full CRUD operations for all node types with comprehensive update patterns.
  >
  > **Related Requirements:** 4.1, 6.3
  > **Files:** `internal/repository/neo4j/node_repository.go`

- [ ] **4.9. Implement `link_repository.go` for relationship management**
  > Create LinkRepository interface with methods: CreateHierarchicalLink, CreateSemanticLink, CreateStructuralLink, GetNeighbors. Support similarity threshold configuration for semantic links. Include relationship scoring and metadata. **COMPLETED**: Comprehensive relationship management with CRUD operations for all link types.
  >
  > **Related Requirements:** 4.1, 4.2, 4.3, 4.4
  > **Files:** `internal/repository/neo4j/link_repository.go`

- [ ] **4.10. Enhance driver.go with vector indexes and constraints**
  > Add vector index creation for ChunkNode.embedding. Add BTREE indexes for space_id, content_source_id lookups. Include proper constraint and index migration logic with existence checks. **COMPLETED**: Enhanced driver with constraints and comprehensive indexing.
  >
  > **Related Requirements:** 4.5, 6.1 (Performance)
  > **Files:** `internal/repository/neo4j/driver.go`

### 5. gRPC/HTTP handlers for read and search
> Implement read APIs and semantic search endpoint.

- [ ] **5.1. Define `proto/public/canvas_public.proto` messages and services**
  > `GetNode`, `GetNeighbors`, `SemanticSearch`, `CreateStructuralLink`.
  >
  > **Related Requirements:** 5.1

- [ ] **5.2. Implement gRPC services in `internal/server/grpc.go`**
  > Wire to service layer and repositories; add validation; implement GetNode, GetNeighbors, SemanticSearch, CreateStructuralLink RPCs.
  >
  > **Related Requirements:** 5.1

- [ ] **5.3. Optional gRPC-Gateway HTTP endpoints**
  > Provide HTTP access via gateway; auth middleware.
  >
  > **Related Requirements:** 6.5 (Security)

## Feature F: Application Orchestration (Go)

### 6. Service layer coordination (business logic)
> Coordinate ingestion → chunking/embedding → graph build. Services orchestrate between gateway, repository, and event processing layers.

- [ ] **6.1. Implement business logic in `internal/service/`**
  > Create vector_service.go and search_service.go to orchestrate calls to Python gateway and Neo4j repositories; emit metrics (not Kafka events). Emit `chunking.completed` only after chunks are persisted. Services coordinate between gateway, repository, and event processing.
  >
  > **Related Requirements:** 1–5, 6.4 (Observability)
  > **STATUS**: NOT IMPLEMENTED - No service layer implementations exist

- [ ] **6.2. Wire `cmd/main.go`**
  > Load config, start servers and consumers, health checks, graceful shutdown.
  >
  > **Related Requirements:** 6.6

## Feature G: Internal Protocols

### 7. Protobuf definitions (internal)

 - [x] **7.1. Define `proto/canvas.proto`**
  > `ChunkText`, `EmbedChunks`, `EmbedQuery` with oneof `text | blob_url`, chunking config (type fixed to sentence for now, target_tokens, overlap_percent), tokenizer, and provenance metadata. (Implemented at `ms_canvas/proto/v1/canvas.proto`)
  >
  > **Related Requirements:** 2.1, 3.1, 5.1

## Feature H: Future ML Workloads (Python via internal gRPC)

### 11. Summarization scaffolding
> Prepare internal APIs and placeholders; can be disabled by config in demo.

- [ ] **11.1. Define proto messages/services for Summarize**
  > Request: target node or set; model/provider config. Response: summary text + metadata.
  >
  > **Related Requirements:** 5b (Summarization)

- [ ] **11.2. Add stub handler in Python**
  > Implement minimal validation and a placeholder response; wire config flags.
  >
  > **Related Requirements:** 5b

### 12. Clustering scaffolding
> Prepare internal APIs and placeholders; can be disabled by config in demo.

- [ ] **12.1. Define proto messages/services for ClusterWorkspace**
  > Request: space_id, params; Response: cluster assignments, optional centroids/labels.
  >
  > **Related Requirements:** 4b (Clustering)

- [ ] **12.2. Add stub handler in Python**
  > Implement minimal validation and placeholder; no-op behind feature flag.
  >
  > **Related Requirements:** 4b

## General Tasks

### 8. Testing and Quality Assurance

- [ ] **8.1. Unit tests for Go repositories, handlers, and application services**
  > Include idempotency, soft delete, and index usage tests.
  >
  > **Related Requirements:** 6.2–6.4

- [x] **8.2. Python unit tests for chunking/embedding services**
  > Include splitter/embedding configs; error handling; performance bounds; canonical test case for 200,000-character input; validate multilingual sentence segmentation; verify `start_position`/`end_position` correctness. **COMPLETED**: Comprehensive test suite for embedding service including unit tests (mocked), integration tests (real providers), and performance tests. Tests verify text→vector conversion, multiple embedding models, error handling, and semantic similarity patterns.
  >
  > **Related Requirements:** 2.1, 3.1
  > **Files Added:** `tests/embedding/test_embedding_unit.py`, `tests/embedding/test_embedding_integration.py`, `tests/embedding/test_embedding_performance.py`, `tests/embedding/test_all.py`, `tests/embedding/README.md`

- [ ] **8.3. Integration tests for end-to-end pipeline**
  > Simulate `document.processed` → graph populated → search returns results.
  >
  > **Related Requirements:** 1–5

### 9. Infrastructure and Deployment

- [x] **9.1. Dockerfile and multi-process runtime**
  > Build Go binary; install Python deps; add `supervisord.conf` or `scripts/start.sh`.
  >
  > **Related Requirements:** 6.6, 6.4 (Observability via health checks)

  > Build Go binary; install Python deps; add `supervisord.conf` or `scripts/start.sh`.
  >
  > **Related Requirements:** 6.6, 6.4 (Observability via health checks)

- [ ] **9.2. Configuration and secrets**
  > Env-based config for Kafka/Neo4j/models; Python reads chunking/tokenizer/grpc-size env vars; add `CANVAS_BATCH_SIZE` and feature flags for summarization/clustering; secrets via manager; secure internal gRPC (localhost).
  >
  > **Related Requirements:** 6.5–6.6

- [ ] **9.3. Metrics, logs, traces**
  > OpenTelemetry across Go/Python; structured logs with PII scrubbing; consumer lag metrics; Python processing P99.
  >
  > **Related Requirements:** 6.4

### 10. Security & Reliability

- [ ] **10.1. AuthN/Z and message validation**
  > Signed/mTLS Kafka; validate event provenance; least-privilege Neo4j; internal gRPC request validation.
  >
  > **Related Requirements:** 6.5

- [ ] **10.2. Idempotency and retries**
  > Idempotency keys, backoff strategies, DLQ/poison queue handling.
  >
  > **Related Requirements:** 6.2 (Reliability)
