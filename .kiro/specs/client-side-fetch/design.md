# Design Document: Hybrid Server-First Data Fetching (Spaces & Content)

## 1. Overview
This design aligns with the current codebase, focusing on `/spaces` (list) and `/spaces/[spaceId]` (detail). We will render initial data on the server through existing server actions backed by Connect-based gRPC clients and Clerk auth headers. We add layered caching to reduce duplicate upstream calls and speed up first paint.

## 2. Architecture Design

### 2.1. System Architecture Diagram
```mermaid
graph TD
    A[Browser] -->|GET /spaces| B[Next.js App Router]
    A -->|GET /spaces/{id}| B
    B -->|Server Actions| C[Frontend: Connect gRPC Clients]
    C -->|gRPC| D[ms_knowledge]
    C -->|gRPC| E[ms_user]
    B -->|SSR HTML with initial data| A
    A -->|Client interactions| B
```

### 2.2. Technology Stack
- **Frontend:** Next.js App Router, React 18, TypeScript, Zustand, Radix UI
- **RPC:** `@bufbuild/connect` with `createGrpcTransport`
- **Auth:** Clerk (`auth()` + token in headers via `createAuthHeaders`)
- **Caching:** React `cache` + Next.js `unstable_cache` with tags/time

## 3. Data and Models
No schema changes introduced here. We rely on generated protobuf types for `Space`, `ContentSource`, and `User`.

## 4. Server Actions and Caching

### 4.1 Spaces List (`searchSpaces`)
- Location: `frontend/src/api/actions/spaceActions.ts`
- Inputs: `SpaceFilters` (q, keywords); server constrains to `ownerId` from Clerk.
- Behavior: Calls `listSpaces` on Knowledge service via Connect client.
- Caching:
  - Wrap in `React.cache` to avoid duplicate calls within a single render.
  - Wrap with `unstable_cache` keyed by `spaces-list-{userId}` and `filters` signature.
  - Revalidate on `createSpace`, `updateSpace`, `deleteSpace` via `revalidatePath('/spaces')` and/or `revalidateTag('spaces-list-{userId}')`.

### 4.2 Space Detail (`getSpace`) + Content Sources (`listContentSources`)
- Locations: `frontend/src/api/actions/spaceActions.ts`, `frontend/src/api/actions/contentActions.ts`
- Behavior: Fetch space by `spaceId`, and processed content sources for that space.
- Caching:
  - `unstable_cache` with tags: `space-{spaceId}`, `content-sources-{spaceId}`.
  - Revalidate on content upload/confirm/delete.

### 4.3 Auth Headers
- Location: `frontend/src/api/actions/utils.ts`
- `createAuthHeaders()` obtains Clerk token and attaches `Authorization: Bearer <token>` for backend calls.

### 4.4 gRPC Clients (Server-Side)
- Location: `frontend/src/api/server-client.ts`
- Singleton pattern via `GRPCClientManager` for User and Knowledge services using `createGrpcTransport` and envs `MS_*_GRPC_URL_INTERNAL`.

## 5. UI/UX Integration

### 5.1 `/spaces` Page
- Location: `frontend/src/app/spaces/page.tsx`
- Server component calls `searchSpaces` to get `initialSpaces`, passes to `SpacesClientPage`.
- Client component (`SpacesClientPage`) handles filters/search and incremental updates.

### 5.2 `/spaces/[spaceId]` Page
- Location: `frontend/src/app/spaces/[spaceId]/page.tsx`
- Server component calls `getSpace` and `listContentSources(spaceId, 'processed')` in parallel and renders the composite layout.

## 6. API Details (Illustrative)
- No REST endpoints added; uses server actions and Connect gRPC.
- Caching tags proposed:
  - `spaces-list-{userId}`
  - `space-{spaceId}`
  - `content-sources-{spaceId}`

## 7. Security
- All server actions check authentication; unauthenticated requests return a consistent error result.
- Errors sanitized via `sanitizeError`/`sanitizeErrorString`.

## 8. Operations
- Ensure `MS_KNOWLEDGE_GRPC_URL_INTERNAL` and `MS_USER_GRPC_URL_INTERNAL` are configured for all environments.
- Keep Next.js `serverExternalPackages` for grpc deps (see `next.config.ts`).