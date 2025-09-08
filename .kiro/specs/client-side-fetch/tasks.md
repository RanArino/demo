# Implementation Plan: Hybrid Server-First Data Fetching (Spaces & Content)

> This document identifies and manages the implementation tasks grounded in the current codebase. It breaks down work to add layered caching to server actions and ensure server-first rendering for `/spaces` and `/spaces/[spaceId]` while preserving Clerk-secured Connect gRPC flows.

## Feature A: Server-first Spaces list

### 1. Server action caching for Spaces
> Add layered caching to `searchSpaces` to avoid duplicate upstream calls and speed up first paint.

- [ ] **1.1. Wrap `searchSpaces` in React `cache` and Next.js `unstable_cache`**
  > Key by `spaces-list-{userId}` and filter signature; set revalidate window; consider `revalidateTag`.
  >
  > **Related Requirements:** 1.1, 4.1

- [ ] **1.2. Revalidation on mutations**
  > Ensure `createSpace`, `updateSpace`, `deleteSpace` revalidate `'/spaces'` and/or relevant tags.
  >
  > **Related Requirements:** 1.1

- [ ] **1.3. Ensure initial render uses server-fetched data**
  > `/spaces/page.tsx` already calls `searchSpaces`; confirm client has no initial loading spinner for first paint.
  >
  > **Related Requirements:** 1.1

## Feature B: Server-first Space detail and Content Sources

### 2. Server action caching for Space detail and Content Sources
> Add layered caching to `getSpace` and `listContentSources(spaceId, 'processed')` to avoid duplicate calls and improve load.

- [ ] **2.1. Wrap `getSpace` in React `cache` and `unstable_cache`**
  > Tag as `space-{spaceId}` with reasonable revalidate.
  >
  > **Related Requirements:** 2.1, 4.1

- [ ] **2.2. Wrap `listContentSources` similarly**
  > Tag as `content-sources-{spaceId}`; revalidate on upload/confirm/delete.
  >
  > **Related Requirements:** 2.1, 4.1

- [ ] **2.3. Verify `/spaces/[spaceId]/page.tsx` renders without initial spinner**
  > Ensure server data is passed to client components with no blocking client fetch.
  >
  > **Related Requirements:** 2.1

## Feature C: Security and Auth Consistency

### 3. Clerk token propagation and error hygiene

- [ ] **3.1. Confirm `createAuthHeaders` used across actions**
  > Validate presence of Bearer token; unify unauthorized responses.
  >
  > **Related Requirements:** 3.1

- [ ] **3.2. Sanitize errors consistently**
  > Use `sanitizeError`/`sanitizeErrorString` where applicable.
  >
  > **Related Requirements:** 3.1

## General Tasks

### 4. Testing and Quality Assurance

- [ ] **4.1. Add unit tests for cached server actions**
- [ ] **4.2. Add integration tests for list/detail initial render**

### 5. Infrastructure Setup and Deployment

- [ ] **5.1. Verify env vars for gRPC transports**
- [ ] **5.2. Confirm Next.js config for server external packages remains valid**
