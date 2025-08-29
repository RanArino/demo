# REQ-37: Dual-Bucket Content Handling – Refactor Specification

## Background
Each `CONTENT_SOURCE` represents a single logical document, with two possible physical objects stored in Cloudflare R2:
- Source/original object in bucket `R2_BUCKET_SOURCE_NAME`
- Processed object in bucket `R2_BUCKET_PROCESSED_NAME`

Original and processed objects are created, confirmed, and fetched independently. The API must allow callers to operate on one kind at a time.

## Goals
- Issue presigned upload URLs for a requested kind (source OR processed) when creating an upload URL.
- Confirm upload for a requested kind by persisting only that kind’s blob hash. Make sure to store blob hash in the correct field of CONTENT_SOURCE table(either original_blob_hash or processed_blob_hash).
- Generate presigned download URLs for a requested kind.
- Remove all hardcoded bucket names; use configuration only.
- Keep object keys deterministic and shared across buckets (same key, different buckets).

## Non-Goals
- Changing the database schema (no new columns/tables required).
- Forcing clients to upload or confirm both kinds at the same time.

## Terminology
- Source (original) object: stored in `R2_BUCKET_SOURCE_NAME`.
- Processed object: stored in `R2_BUCKET_PROCESSED_NAME`.
- Object key: deterministic path like `spaces/{space_id}/content/{content_id}/{filename}` used in both buckets.

## API (Proto) Changes
Introduce an enum to specify which object is targeted by a request:
- `enum DownloadObjectKind { DOWNLOAD_OBJECT_KIND_UNSPECIFIED = 0; DOWNLOAD_OBJECT_KIND_ORIGINAL = 1; DOWNLOAD_OBJECT_KIND_PROCESSED = 2; }`

1) CreateUploadURL
- Request: add `object_kind` (enum) to indicate whether to create an upload URL for ORIGINAL or PROCESSED.
- Response:
  - `upload_url` (string)
  - `object_key` (string) – same key design for both kinds
  - `expires_at` (google.protobuf.Timestamp)
  - `content_source` (ContentSource)

2) ConfirmUpload
- Request (operate on one kind only):
  - `content_source_id` (string) – required
  - `object_kind` (enum) – required
  - `blob_hash` (string) – required (hash of the uploaded object for the specified kind)
- Behavior:
  - If `object_kind == ORIGINAL`: set `original_blob_hash` on the record
  - If `object_kind == PROCESSED`: set `processed_blob_hash` on the record

3) GenerateDownloadURL
- Request: caller MUST specify which object to download.
  - `content_source_id` (string) – required
  - `expires_seconds` (int32) – optional, bounded server-side
  - `object_kind` (enum) – required
- Response: unchanged (`url`, `expires_at`, `object_key`)
- Reject `UNSPECIFIED` with `InvalidArgument`.

## Service Layer Changes (`internal/service/content_service.go`)
- Constructor:
  - Ensure the service holds `*config.Config` and does NOT call `os.Getenv` directly.
- Bucket resolution:
  - Implement `resolveBucket(kind string) (string, error)` with NO hardcoded fallbacks.
  - `kind == "source"` → `cfg.R2.BucketSourceName` (error if unset)
  - `kind == "processed"` → `cfg.R2.BucketProcessedName` (error if unset)
- CreateUploadURL:
  - Accept `kind` parameter; compute `objectKey = spaces/{space}/content/{id}/{filename}`.
  - Resolve the appropriate bucket for the requested kind and return a single presigned upload URL.
- ConfirmUpload:
  - Accept `kind` and `blobHash`; update only the corresponding hash field.
- GenerateDownloadURLWithKind:
  - Require `kind` parameter ("source" or "processed") from server.
  - Resolve bucket via `resolveBucket(kind)` and sign GET for `objectKey`.

## Server Layer Changes (`internal/server/grpc.go`)
- CreateUploadURL handler:
  - Read `object_kind` and pass the mapped kind ("source"/"processed") to the service; return a single URL for that kind.
- ConfirmUpload handler:
  - Require `object_kind` and `blob_hash`; update only the targeted kind.
- GenerateDownloadURL handler:
  - Map `object_kind` enum → `kind` string (`ORIGINAL`→"source", `PROCESSED`→"processed").
  - Reject `UNSPECIFIED` with `InvalidArgument`.

## Storage Adapter (`internal/storage/r2`)
- No change to signatures required.
- Continue to expose: `GeneratePresignedUploadURL(bucket, key, ttl)` and `GeneratePresignedDownloadURL(bucket, key, ttl)`.

## Configuration (`internal/config/config.go`)
- Ensure both env-backed settings exist and are loaded:
  - `R2_BUCKET_SOURCE_NAME`
  - `R2_BUCKET_PROCESSED_NAME`
- Service must fail fast (clear error) if a required bucket is missing for the requested operation.

## Object Key Strategy
- Keep one deterministic key shared across buckets: `spaces/{space_id}/content/{content_id}/{filename}`.
- Do NOT diverge key shapes between buckets.

## Validation Rules
- `CreateUploadURL`:
  - Validate space existence, filename, mimeType.
  - Validate `object_kind` and ensure corresponding bucket is configured.
- `ConfirmUpload`:
  - Validate `object_kind` and `blob_hash`.
- `GenerateDownloadURL`:
  - Reject `UNSPECIFIED` object kind.
  - Enforce TTL bounds (e.g., 30s–1h).

## Security
- Enforce RBAC/ownership checks before issuing any presigned URL.
- Short TTLs (default 15m) with server-side bounds.
- Avoid leaking bucket names in logs.

## Observability
- Log presign operations with request context (space_id, content_id, kind, TTL), without secrets.
- Add counters for per-kind presigns and per-kind confirms.

## Testing
- Unit tests:
  - Service: bucket resolution errors; per-kind URL generation; per-kind confirm; per-kind download.
  - Server: request validation for `object_kind` and required fields.
- Integration tests:
  - Presign per-kind and validate URL shapes.
  - Confirm persists the targeted hash only.
  - Download works per-kind with correct bucket.

## Acceptance Criteria
- `CreateUploadURL` returns one valid presigned URL for the requested kind referencing the correct `object_key` and TTL.
- `ConfirmUpload` updates only the targeted hash (`original_blob_hash` OR `processed_blob_hash`).
- `GenerateDownloadURL` requires `object_kind` and returns a URL from the correct bucket.
- No hardcoded bucket names remain; bucket names are read from config only.
- All tests pass; lints are clean.
