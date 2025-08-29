# REQ-39: ContentSource Create & Update – Implementation Plan

## Scope
- Add/verify Create flow for `ContentSource` with per-kind upload handling (ORIGINAL/PROCESSED) per REQ-37.
- Add Update RPC to allow changing only `title` and `keywords`.
- Ensure server/service align with docs/coding-guide/BACKEND.md and existing per-kind R2 behavior.

## Current State (Audit)
- Proto `KnowledgeService` has content RPCs: `CreateUploadURL`, `ConfirmUpload`, `GetContentSource`, `ListContentSources`, `UpdateContentSourceStatus`, `DeleteContentSource`, `GenerateDownloadURL`. There is NO `CreateContentSource` RPC and NO `UpdateContentSource` RPC.
- Service `ContentService` implements:
  - `CreateUploadURLWithKind` (and `CreateUploadURL` wrapper for original)
  - `ConfirmUploadWithKind` (and `ConfirmUpload` wrapper for original)
  - `GetContentSource`, `ListContentSources`, `UpdateContentSourceStatus`, `GenerateDownloadURLWithKind`, `DeleteContentSource`
- Server handlers exist for: `CreateUploadURL`, `ConfirmUpload`, `GetContentSource`, `ListContentSources`, `UpdateContentSourceStatus`, `DeleteContentSource`, `GenerateDownloadURL`.
- Ent schema `ContentSource` includes fields needed for update (`title`, `keywords`). Repo supports `Update` that can set `Title` and `Keywords`.

## Clarification of Create Flow
- We will NOT add a separate `CreateContentSource` RPC because the existing flow already creates the DB row at `CreateUploadURLWithKind` time. For creation semantics:
  - The server is responsible for creating the `CONTENT_SOURCE` record when issuing an upload URL; clients do not provide the final `title` or other persisted fields at creation time.
  - The initial `title` MUST be derived from the uploaded object's metadata (specifically the filename) and set by the server based on the `filename` field provided in the upload request or as extracted from the uploaded object. Clients do not directly set the persisted title during create.
  - The server populates `media_type` and `source` from the provided `mime_type` and the known upload source; `content_summary` and `keywords` are optional and not required during creation.
  - The original blob hash is only persisted when the client calls `ConfirmUpload` after the upload completes.
  - Behavior today:
  - `CreateUploadURLWithKind(space_id, filename, mime_type, size_bytes, title, kind)`:
    - Validates inputs and space existence.
    - Creates a `ContentSource` DB row with `status=UPLOADING`, owner inferred from context.
    - If `title` is empty, set the `ContentSource.Title` to the `filename` supplied in the request.
    - Computes deterministic `object_key = spaces/{space_id}/content/{content_id}/{filename}`.
    - Resolves bucket based on `kind` via config-only `resolveBucket`.
    - Returns a presigned upload URL for that kind, the object key, expiry, and the created content row.
  - `ConfirmUploadWithKind(content_id, kind, blob_hash)` updates only the targeted blob-hash field:
    - `ORIGINAL` → sets `original_blob_hash` and `status=UPLOADED` (persists original blob hash on confirm).
    - `PROCESSED` → sets `processed_blob_hash`.
- This aligns with REQ-37’s dual-bucket per-kind semantics. Therefore:
  - "CreateContentSource" is effectively accomplished by `CreateUploadURL*` (DB record creation) + `ConfirmUpload*` (blob hash persistence per kind). No additional RPC is required for creation.

## Gaps to Implement
- Add `UpdateContentSource` RPC to update only `title` and `keywords`.
- Add server handler and service method to perform update with RBAC checks.
- Ensure tests cover per-kind flows per REQ-37 and new update RPC.

## API Changes (Proto)
- Add request/response definitions:
  - `rpc UpdateContentSource(UpdateContentSourceRequest) returns (ContentSource);`
  - `message UpdateContentSourceRequest { string id = 1; ContentSource content = 2; google.protobuf.FieldMask update_mask = 3; }`
- Update Mask handling:
  - Allowed paths: `"title"`, `"keywords"` only.
  - If mask empty, update both if provided.
  - Ignore any other fields and reject with `InvalidArgument`.

## Server Layer (`internal/server/grpc.go`)
- Implement `UpdateContentSource` handler:
  - Parse `id` as UUID; return `InvalidArgument` on error.
  - Fetch content via service for RBAC check; if not admin, verify caller owns the content (`owner_id`).
  - Build updates map based on `update_mask`:
    - If `title` in mask → read from `req.content.title` if non-empty.
    - If `keywords` in mask → read from `req.content.keywords` (can be empty slice to clear).
    - If mask is nil → consider both fields.
  - Call service method to apply updates.
  - Return updated `ContentSource`.

## Service Layer (`internal/service/content_service.go`)
- Add `UpdateContentSource(ctx, id, title *string, keywords *[]string) (*domain.ContentSource, error)`:
  - Load current content.
  - Update only provided fields (partial update semantics):
    - If `title != nil`, set trimmed value (server may set `title` to filename at creation time; for updates, provided value overwrites).
    - If `keywords != nil`, set to provided slice (allow empty slice to clear keywords).
  - `nil` pointer arguments MUST be interpreted as "no change" (partial update).
  - Call `contentRepo.Update` and return the updated entity (re-fetch after update).
  - Validation:
    - Title max length/policy as needed (keep consistent with existing constraints if any).
- Creation note: `CreateUploadURLWithKind` MUST set `Title` to the `filename` when `title` is empty in the request; `ConfirmUploadWithKind` persists the chosen blob hash for the specified kind (original or processed).

## Repository Layer
- Already supports setting `Title` and `Keywords` in `Update`. No change required.

## RBAC
- Follow existing pattern used in `DeleteContentSource` and `GenerateDownloadURL`:
  - Admins bypass.
  - Non-admins must be the owner of the content.

## Configuration & Buckets (REQ-37 Compliance)
- Already compliant:
  - `resolveBucket(kind)` reads only from config for both source and processed buckets.
  - Deterministic `object_key` shared across buckets.

## Testing
- Unit tests to add:
  - Server: `UpdateContentSource` should
    - Update title only with update mask.
    - Update keywords only with update mask.
    - Update both when mask omitted.
    - Enforce RBAC (non-owner forbidden; admin allowed).
    - Reject invalid paths in mask with `InvalidArgument`.
  - Service: `UpdateContentSource` sets only provided fields and persists them.
- Existing tests cover create/upload/confirm and status updates; retain/extend where needed.

## Migration
- No schema changes required.

## Backward Compatibility
- Existing clients using `CreateUploadURL`/`ConfirmUpload` flow remain unaffected.
- New `UpdateContentSource` is additive.

## Step-by-Step Work Items
1) Proto: add `UpdateContentSource` RPC and request message; regenerate Go code.
2) Server: implement `UpdateContentSource` handler with RBAC and mask validation.
3) Service: implement `UpdateContentSource` method.
4) Tests: add unit tests for server/service update logic; extend mocks as needed.
5) Docs: ensure `BACKEND.md`/`REQ_37.md` expectations are satisfied (already matched for create/confirm/download).

## Acceptance Criteria
- `UpdateContentSource` updates only `title` and/or `keywords` per `FieldMask` and RBAC.
- Create flow remains: DB row created at upload URL issuance; per-kind confirm sets the correct blob hash.
- No hardcoded bucket names; deterministic key unchanged.
- All tests pass; lints clean.
