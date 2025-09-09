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

- [x] **3.1. Audit auth headers across all cached actions**
  > Verify `createAuthHeaders()` is called in all server actions
  > Ensure consistent UNAUTHORIZED error responses across cached and non-cached paths
  > Validate that cached actions don't bypass authentication
  >
  > **Related Requirements:** 3.1
  > **Files:** `frontend/src/api/actions/spaceActions.ts`, `frontend/src/api/actions/contentActions.ts`

- [x] **3.2. Enhance error sanitization for cached responses**
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

- [x] **6.1. Create unit tests for cache utility functions**
  > ✅ Created comprehensive tests for `sanitizeProtobufForJson`, `generateCacheKey`, and cache metrics
  > ✅ Tested BigInt conversion, nested objects, and edge cases with 11 passing test cases
  > ✅ Set up Jest with TypeScript support and Next.js integration
  >
  > **Files:** `frontend/src/api/actions/__tests__/utils.test.ts`, `frontend/src/api/actions/__tests__/utils-simple.test.ts`

- [x] **6.2. Add unit tests for cached server actions**
  > ✅ Created comprehensive test suites for spaceActions and contentActions with 25+ test cases
  > ✅ Tested cache hit/miss scenarios, revalidation triggers, and error handling
  > ✅ Mocked all external dependencies (Clerk, gRPC clients, Next.js cache functions)
  >
  > **Files:** `frontend/src/api/actions/__tests__/spaceActions.test.ts`, `frontend/src/api/actions/__tests__/contentActions.test.ts`

- [x] **6.3. Add integration tests for server-side rendering**
  > ✅ Created integration tests for server-side data fetching and rendering
  > ✅ Tests verify initial data renders without loading spinners and cache invalidation works
  > ✅ Covers both `/spaces` and `/spaces/[spaceId]` pages with parallel data fetching
  >
  > **Files:** `frontend/src/app/spaces/__tests__/integration.test.ts`

- [x] **6.4. Performance testing**
  > ✅ Created performance tests measuring cache hit rates, response times, and memory usage
  > ✅ Tests cache key collision prevention, concurrent load handling, and cross-service coordination
  > ✅ Validates cache optimization goals (>80% hit rate, <500ms response time)
  >
  > **Files:** `frontend/src/api/actions/__tests__/performance.test.ts`

### 7. Infrastructure and Monitoring

- [x] **7.1. Verify environment configuration**
  > ✅ Created comprehensive environment verification tests for all deployment scenarios
  > ✅ Documented complete setup guide with dev/staging/production configurations
  > ✅ Verified Next.js `serverExternalPackages` configuration and gRPC client compatibility
  >
  > **Files:** `frontend/src/api/actions/__tests__/environment.test.ts`, `frontend/ENVIRONMENT_SETUP.md`

- [x] **7.2. Add cache monitoring and metrics**
  > ✅ Implemented real-time cache metrics tracking (hits, misses, response times, error rates)
  > ✅ Created `/api/cache-metrics` REST endpoint for monitoring access
  > ✅ Built admin dashboard with visual metrics, alerts, and export capabilities
  > ✅ Added automated performance monitoring with configurable alerts
  >
  > **Files:** `frontend/src/api/actions/utils.ts`, `frontend/src/app/api/cache-metrics/route.ts`, `frontend/src/components/admin/CacheMetricsDashboard.tsx`

- [x] **7.3. Documentation and runbook updates**
  > ✅ Created comprehensive environment setup guide with deployment configurations
  > ✅ Documented troubleshooting guide for cache-related issues and common problems
  > ✅ Added TTL tuning recommendations and performance optimization guidelines
  > ✅ Created test results summary with verification of all functionality
  >
  > **Files:** `frontend/ENVIRONMENT_SETUP.md`, `frontend/TEST_RESULTS.md`
