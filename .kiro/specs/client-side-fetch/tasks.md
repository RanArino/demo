# Implementation Plan: Hybrid Server-First Data Fetching (Spaces & Content)

> This document identifies and manages the implementation tasks grounded in the current codebase. It breaks down work to add layered caching to server actions and ensure server-first rendering for `/spaces` and `/spaces/[spaceId]` while preserving Clerk-secured Connect gRPC flows.

## Feature A: Server-first Spaces list

### 1. Server action caching for Spaces
> Add layered caching to `searchSpaces` to avoid duplicate upstream calls and speed up first paint.

- [x] **1.1. Create cache utility functions**
  > ✅ Created `normalizeFilters` function to handle undefined/null values consistently in SpaceFilters.
  > ✅ Created cache key generation helper: `generateSpacesListCacheKey(userId: string, filters: SpaceFilters)`
  >
  > **Related Requirements:** 1.1, 4.1
  > **Files:** `frontend/src/api/actions/utils.ts`

- [x] **1.2. Wrap `searchSpaces` in dual-layer caching**
  > ✅ Wrapped in `React.cache` for per-request memoization
  > ✅ Wrapped in `unstable_cache` with key: `spaces-list-${userId}-${JSON.stringify(normalizedFilters)}`
  > ✅ Set 5-minute TTL and tag: `spaces-list-${userId}`
  > ✅ Fixed dynamic data access issue by moving auth/headers outside cached functions
  >
  > **Related Requirements:** 1.1, 4.1
  > **Files:** `frontend/src/api/actions/spaceActions.ts`

- [x] **1.3. Add revalidation to space mutations**
  > ✅ Updated `createSpace`, `updateSpace`, `deleteSpace` to call `revalidateTag('spaces-list-{userId}')`
  > ✅ Kept existing `revalidatePath('/spaces')` calls
  >
  > **Related Requirements:** 1.1
  > **Files:** `frontend/src/api/actions/spaceActions.ts`

- [x] **1.4. Verify server-side data rendering**
  > ✅ Confirmed `/spaces/page.tsx` renders initial data without client loading spinners
  > ✅ Server-first rendering working with cached data - latency dramatically improved
  >
  > **Related Requirements:** 1.1
  > **Files:** `frontend/src/app/spaces/page.tsx`

## Feature B: Server-first Space detail and Content Sources

### 2. Server action caching for Space detail and Content Sources
> Add layered caching to `getSpace` and `listContentSources(spaceId, 'processed')` to avoid duplicate calls and improve load.

- [x] **2.1. Wrap `getSpace` in dual-layer caching**
  > Wrap in `React.cache` for per-request memoization
  > Wrap in `unstable_cache` with tag: `space-${spaceId}` and 5-minute TTL
  > Handle cache errors gracefully with stale-while-revalidate strategy
  >
  > **Related Requirements:** 2.1, 4.1
  > **Files:** `frontend/src/api/actions/spaceActions.ts`

- [x] **2.2. Wrap `listContentSources` in dual-layer caching**
  > Wrap in `React.cache` for per-request memoization
  > Wrap in `unstable_cache` with tag: `content-sources-${spaceId}` and 60-second TTL
  > Include status parameter in cache key generation
  >
  > **Related Requirements:** 2.1, 4.1
  > **Files:** `frontend/src/api/actions/contentActions.ts`

- [x] **2.3. Add revalidation to content mutations**
  > Update `confirmUpload`, `deleteContentSource` to call `revalidateTag('content-sources-{spaceId}')`
  > Add `revalidatePath('/spaces/[spaceId]')` where appropriate
  >
  > **Related Requirements:** 2.1
  > **Files:** `frontend/src/api/actions/contentActions.ts`

- [x] **2.4. Remove dynamic rendering from space detail page**
  > Remove `dynamic = 'force-dynamic'` and `revalidate = 0` from `/spaces/[spaceId]/page.tsx`
  > Ensure proper error handling for both space and content loading failures
  >
  > **Related Requirements:** 2.1
  > **Files:** `frontend/src/app/spaces/[spaceId]/page.tsx`

- [x] **2.5. Update space mutations to revalidate space cache**
  > Add `revalidateTag('space-{spaceId}')` to `updateSpace`
  > Keep existing path revalidation for broader updates
  >
  > **Related Requirements:** 2.1
  > **Files:** `frontend/src/api/actions/spaceActions.ts`

- [x] **2.6. Prevent duplicate loaders on /spaces initial render**
  > Hydrate with server-fetched data and skip client refetch on first mount
  > Guard initial effects in `SpacesClientPage` to avoid immediate refetch
  > Ensure initial UI uses provided props without setting `loading`
  >
  > **Related Requirements:** 1.1, 4.1
  > **Files:** `frontend/src/app/spaces/SpacesClientPage.tsx`

- [x] **2.7. Rely on route loading.tsx; remove redundant Suspense fallback**
  > Remove extra `<Suspense fallback>` from `/spaces/page.tsx`
  > Let `frontend/src/app/spaces/loading.tsx` handle route-level skeleton
  > Verify there is no double-loading UX during navigation
  >
  > **Related Requirements:** 1.1, 4.1
  > **Files:** `frontend/src/app/spaces/page.tsx`, `frontend/src/app/spaces/loading.tsx`

## Feature C: Security and Auth Consistency

### 3. Clerk token propagation and error hygiene

- [ ] **3.1. Audit auth headers across all cached actions**
  > Verify `createAuthHeaders()` is called in all server actions
  > Ensure consistent UNAUTHORIZED error responses across cached and non-cached paths
  > Validate that cached actions don't bypass authentication
  >
  > **Related Requirements:** 3.1
  > **Files:** `frontend/src/api/actions/spaceActions.ts`, `frontend/src/api/actions/contentActions.ts`

- [ ] **3.2. Enhance error sanitization for cached responses**
  > Apply `sanitizeError`/`sanitizeErrorString` consistently
  > Ensure cached error scenarios don't leak sensitive information
  > Add error logging for cache-related authentication failures
  >
  > **Related Requirements:** 3.1
  > **Files:** `frontend/src/api/actions/utils.ts`

## Feature D: BigInt Sanitization Standardization

### 4. Generic protobuf sanitization infrastructure
> Create reusable utilities to handle BigInt/int64 fields across all protobuf-generated types for consistent JSON serialization and caching.

- [x] **4.1. Create generic sanitizeProtobufForJson utility**
  > ✅ Implemented recursive BigInt→number conversion that works with any protobuf type
  > ✅ Handles nested objects, arrays, and optional fields automatically
  > ✅ Preserves type safety with TypeScript generics: `sanitizeProtobufForJson<T>(obj: T): T`
  >
  > **Related Requirements:** 4.1 (performance), security (no data leaks)
  > **Files:** `frontend/src/api/actions/utils.ts`

- [x] **4.2. Replace manual Space sanitization**
  > ✅ Removed hardcoded `sanitizeSpaceForJson` function 
  > ✅ Used generic sanitizer for Space types in `searchSpaces` and `getSpace`
  > ✅ Return types use `number` instead of `bigint` for consistency
  >
  > **Files:** `frontend/src/api/actions/spaceActions.ts`

- [x] **4.3. Add ContentSource sanitization**
  > ✅ Applied generic sanitizer to ContentSource types (handles `size_bytes` int64 field)
  > ✅ Updated `listContentSources`, `createUploadURL`, `getContentSource`, `confirmUpload` actions
  > ✅ All cached ContentSource data is properly serializable
  >
  > **Files:** `frontend/src/api/actions/contentActions.ts`

- [x] **4.4. Standardize User type sanitization**
  > ✅ Replaced `toPlainUserObject` approach with generic sanitizer
  > ✅ Created `SanitizedUser` type with `storage_used_bytes` and `storage_quota_bytes` as numbers
  > ✅ Updated all user-related actions to use consistent sanitization
  >
  > **Files:** `frontend/src/api/actions/userActions.ts`

- [x] **4.5. Enhanced cache key generation**
  > ✅ Created generic `generateCacheKey()` helper for consistent serialization
  > ✅ Updated spaces list cache key generation to use sanitized objects
  > ✅ Prevented cache key issues caused by BigInt serialization
  >
  > **Files:** `frontend/src/api/actions/utils.ts`

## General Tasks

### 6. Testing and Quality Assurance

- [ ] **6.1. Create unit tests for cache utility functions**
  > Test `normalizeFilters` function with various input scenarios
  > Test cache key generation for consistent results
  > Mock `unstable_cache` and `React.cache` for isolation testing
  >
  > **Files:** `frontend/src/api/actions/__tests__/utils.test.ts`

- [ ] **6.2. Add unit tests for cached server actions**
  > Test cache hit/miss scenarios for `searchSpaces`, `getSpace`, `listContentSources`
  > Verify revalidation calls are triggered on mutations
  > Test error handling with stale-while-revalidate behavior
  >
  > **Files:** `frontend/src/api/actions/__tests__/spaceActions.test.ts`, `frontend/src/api/actions/__tests__/contentActions.test.ts`

- [ ] **6.3. Add integration tests for server-side rendering**
  > Test `/spaces` page renders with initial data (no loading spinner)
  > Test `/spaces/[spaceId]` page renders space and content data
  > Test cache invalidation after mutations
  >
  > **Files:** `frontend/src/app/spaces/__tests__/integration.test.ts`

- [ ] **6.4. Performance testing**
  > Measure cache hit rates and response times
  > Test with various filter combinations for cache key effectiveness
  > Load test to verify cache doesn't cause memory issues
  >
  > **Files:** `frontend/src/api/actions/__tests__/performance.test.ts`

### 7. Infrastructure and Monitoring

- [ ] **7.1. Verify environment configuration**
  > Ensure `MS_KNOWLEDGE_GRPC_URL_INTERNAL` and `MS_USER_GRPC_URL_INTERNAL` are set correctly
  > Confirm Next.js `serverExternalPackages` configuration remains valid
  > Test in all environments (dev/staging/prod)
  >
  > **Files:** Check deployment configurations

- [ ] **7.2. Add cache monitoring and metrics**
  > Log cache hit/miss rates for performance analysis
  > Add monitoring for cache invalidation frequency
  > Set up alerts for authentication failures in cached actions
  >
  > **Files:** `frontend/src/api/actions/utils.ts` (logging utilities)

- [ ] **7.3. Documentation and runbook updates**
  > Update deployment documentation with caching considerations
  > Create troubleshooting guide for cache-related issues
  > Document TTL tuning recommendations per environment
  >
  > **Files:** Documentation updates
