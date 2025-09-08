# Requirements Document: Hybrid Server-First Data Fetching (Spaces & Content)

## 1. Overview
This initiative grounds the client/server data-fetching model in the current codebase. It optimizes the Spaces list and Space detail flows by rendering initial data on the server via server actions, adding layered caching, and preserving secure gRPC connectivity through Connect with Clerk auth headers. Client components continue to handle interactive updates (filters, navigation, uploads), but first paint should have data without a loading spinner.

- **Target Users:** Authenticated users browsing their Spaces and viewing each Space’s details and documents.
- **Primary Functionality:** Server renders initial data for `/spaces` and `/spaces/[spaceId]`; client interactions fetch updates via existing server actions.
- **Key Problems Addressed:** Duplicate upstream calls, slower first render, and lack of cross-request caching.

## 2. Requirements List

### Requirement 1: Server-fetched Spaces list with layered caching

**User Story:**
> As a user, I want the Spaces list to render with results immediately so that I can browse and filter without an initial loading delay.

**Acceptance Criteria:**
```gherkin
WHEN navigating to "/spaces"
THEN the server renders with initialSpaces present and no client-side initial loading spinner

GIVEN a cached result for the current user’s spaces exists and is fresh
WHEN the list is requested again
THEN the server action returns from cache without a duplicate upstream gRPC call

GIVEN a space is created, updated, or deleted
WHEN revalidation is triggered
THEN the next load of "/spaces" reflects the change
```

### Requirement 2: Server-fetched Space detail and Content Sources with layered caching

**User Story:**
> As a user, I want the Space detail page to render the Space info and its processed documents immediately so that I can see context without waiting.

**Acceptance Criteria:**
```gherkin
WHEN navigating to "/spaces/{spaceId}"
THEN the server renders with Space and processed Content Sources without an initial spinner

GIVEN cache entries for Space and Content Sources exist and are fresh
WHEN revisiting the same page
THEN data returns from cache with no duplicate upstream gRPC call(s)

GIVEN a document is uploaded/confirmed or deleted
WHEN revalidation is triggered
THEN a subsequent visit shows updated document lists
```

### Requirement 3: Security and authentication headers (Clerk)

**Requirements:**
- All server actions must enforce authentication using Clerk and pass a Bearer token in headers created by `createAuthHeaders()`.
- Unauthenticated requests must return a consistent error shape and never leak data.

**Acceptance Criteria:**
```gherkin
GIVEN the user is not authenticated
WHEN calling any server action
THEN the result indicates UNAUTHORIZED and returns no sensitive data

GIVEN a valid session
WHEN server actions call gRPC backends
THEN the Authorization header is set with the Clerk token
```

### Requirement 4: Non-functional Requirements (Performance & Reliability)

**Requirements:**
- Avoid duplicate upstream gRPC calls per server render using `React.cache` (per-request memoization).
- Persist cross-request data with `unstable_cache` plus tag/time revalidation for: `spaces-list-{userId}`, `space-{spaceId}`, `content-sources-{spaceId}`.
- Keep current environment variables working for Connect transports: `MS_KNOWLEDGE_GRPC_URL_INTERNAL`, `MS_USER_GRPC_URL_INTERNAL`.

**Acceptance Criteria:**
```gherkin
GIVEN typical conditions
WHEN rendering "/spaces" or "/spaces/{spaceId}"
THEN first paint includes initial data and avoids duplicate upstream calls per render

GIVEN a revalidation trigger occurs
WHEN the next request arrives
THEN fresh data is fetched and cached again
```