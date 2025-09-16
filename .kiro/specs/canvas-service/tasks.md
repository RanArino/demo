# Implementation Plan: Canvas Service (ms_canvas)

> This document identifies and manages the development tasks derived from the approved requirements and design. Tasks are grouped by features and clearly trace back to requirements items.

## Feature A: Event-Driven Document Ingestion

### 1. Kafka consumer and ingestion orchestration (Go)
> Implement secure, idempotent consumption of `document.processed` events and bootstrap orchestration.

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

- [x] **1.4. Enqueue chunking pipeline trigger**
  > Call application service to start chunking workflow after content creation.
  >
  > **Related Requirements:** 1.1, 2.1

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

- [ ] **3.1. Implement `python_app/app/services/embedding.py`**
  > Batch-embed chunks; return vectors and model metadata (model_id, version).
  >
  > **Related Requirements:** 3.1 (Req 3: Embedding Pipeline)

- [ ] **3.2. Expose gRPC `EmbedChunks` in `python_app/app/server.py`**
  > Per `canvas.proto` definitions; support provider/model selection.
  >
  > **Related Requirements:** 3.1, 6.6

- [ ] **3.3. Add `EmbedQuery` RPC for semantic search**
  > For runtime query embedding used by public search.
  >
  > **Related Requirements:** 5.1 (Req 5: Querying and Search)

## Feature D: Graph Construction and Persistence (Go)

### 4. Neo4j repositories and graph linking
> Persist chunks, embeddings; link relationships; create vector index.

- [ ] **4.1. Implement repositories in `internal/infrastructure/repository`**
  > Upserts for ContentNode/ChunkNode; store embeddings and model metadata; soft delete support; persist `start_position`/`end_position` character offsets and normalized content; index usage.
  >
  > **Related Requirements:** 2.1, 3.1, 4.1, 6.3 (Data Management)

- [ ] **4.2. Create `:HIERARCHICAL_PARENT` links**
  > Link chunks to their parent content.
  >
  > **Related Requirements:** 4.1

- [ ] **4.3. Create `:SEMANTIC_LINK` edges above threshold**
  > Similarity computation and edge creation with `score` property.
  >
  > **Related Requirements:** 4.1, 6.6 (threshold configurable)

- [ ] **4.4. Support explicit `:STRUCTURAL_LINK` writes**
  > Expose in public API and repo methods.
  >
  > **Related Requirements:** 4.1

- [ ] **4.5. Add indexes (BTREE + vector) and migrations**
  > Space/content_source indexes; vector index on `ChunkNode.embedding`.
  >
  > **Related Requirements:** 6.1 (Performance), 6.6

## Feature E: Public API (Go)

### 5. gRPC/HTTP handlers for read and search
> Implement read APIs and semantic search endpoint.

- [ ] **5.1. Define `proto/canvas_public.proto` messages and services**
  > `GetNode`, `GetNeighbors`, `SemanticSearch`, `CreateStructuralLink`.
  >
  > **Related Requirements:** 5.1

- [ ] **5.2. Implement handlers in `internal/infrastructure/handler`**
  > Wire to application services and repositories; add validation.
  >
  > **Related Requirements:** 5.1

- [ ] **5.3. Optional gRPC-Gateway HTTP endpoints**
  > Provide HTTP access via gateway; auth middleware.
  >
  > **Related Requirements:** 6.5 (Security)

## Feature F: Application Orchestration (Go)

### 6. Use cases and workflow coordination
> Coordinate ingestion → chunking → embedding → graph build.

- [ ] **6.1. Implement application services in `internal/application`**
  > Orchestrate calls to Python RPCs and Neo4j repositories; emit events/metrics. Emit `chunking.completed` only after chunks are persisted.
  >
  > **Related Requirements:** 1–5, 6.4 (Observability)

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

## General Tasks

### 8. Testing and Quality Assurance

- [ ] **8.1. Unit tests for Go repositories, handlers, and application services**
  > Include idempotency, soft delete, and index usage tests.
  >
  > **Related Requirements:** 6.2–6.4

- [ ] **8.2. Python unit tests for chunking/embedding services**
  > Include splitter/embedding configs; error handling; performance bounds; canonical test case for 200,000-character input; validate multilingual sentence segmentation; verify `start_position`/`end_position` correctness.
  >
  > **Related Requirements:** 2.1, 3.1

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
  > Env-based config for Kafka/Neo4j/models; Python reads chunking/tokenizer/grpc-size env vars; secrets via manager; secure internal gRPC (localhost).
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
