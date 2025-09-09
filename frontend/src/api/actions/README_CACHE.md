# Server Action Caching Guide

## Architectural Context & Design Rationale

Before implementing caching for a server action, understand the rationale behind our approach.

- Dual-Layer Caching Rationale:
  - **React.cache** is used for intra-request memoization to avoid duplicate calls during a single server render (e.g., parent and child components both calling the same action).
  - **unstable_cache** is used for inter-request persistence to serve subsequent requests without hitting backend services. This provides large latency wins for repeat visitors and navigations.

- Stateless and Deterministic Principle:
  - Functions wrapped by caching layers must be stateless and deterministic: their outputs should depend only on their inputs. Dynamic, per-request data (like Clerk tokens) must be obtained outside cached functions and passed in as `Headers`.
  - This separation prevents cache poisoning and ensures reliable, predictable cache behavior.

This document is a reference for implementing server-action caching consistently across the codebase. Follow these steps when adding a new server action that calls gRPC backends and returns protobuf-generated types.

- Principles
  - Server actions must remain the source of truth for data returned to server-rendered pages.
  - Use a dual-layer cache: `React.cache` for per-request memoization and Next.js `unstable_cache` for cross-request persistence with tag-based revalidation.
  - Always sanitize protobuf objects before serializing / returning to the renderer (BigInt → number/string, Timestamp → ISO). Use `sanitizeProtobufForJson` from `./utils`.
  - Never allow cached actions to bypass authentication — call `createAuthHeaders()` outside cached functions.

- Files & helpers to reuse
  - `frontend/src/api/actions/utils.ts`
    - `createAuthHeaders()` — obtain Clerk token and construct `Headers` for backend calls.
    - `sanitizeProtobufForJson(obj)` — recursive sanitizer for BigInt/Timestamp and protobuf messages.
    - `generateCacheKey(prefix, userId, filters)` — stable cache-key generator for complex filters.
    - `sanitizeError` / `sanitizeErrorString` — return sanitized error objects/strings for logs and responses.

- Implementation recipe (for a read action)

  1. Implement a core fetch function that performs the gRPC call and returns a typed result. It MUST accept `headers: Headers` as an argument and must NOT call `auth()` or `createAuthHeaders()` itself.

     Example signature:

     ```ts
     async function myActionCore(userId: string, headers: Headers, params: MyParams): Promise<ActionResult<MyType>>
     ```

  2. Wrap the core function with `React.cache` to prevent duplicate calls within the same server render. This function still must accept `headers` and other data as arguments.

     ```ts
     const myActionMemoized = cache(async (userId: string, headers: Headers, params: MyParams) => {
       return myActionCore(userId, headers, params);
     });
     ```

  3. Wrap the memoized function with `unstable_cache` to persist across requests and enable tag-based revalidation. Use `generateCacheKey` to build a stable key (include `userId` for user-scoped data and serialized filter params).

     ```ts
     const myActionCached = (userId: string, headers: Headers, params: MyParams) => {
       const key = generateCacheKey('my-action', userId, params);
       return unstable_cache(
         async () => myActionMemoized(userId, headers, params),
         [key],
         { tags: [`my-action-${userId}`], revalidate: 300 }
       )();
     };
     ```

  4. Expose the public action that calls `auth()` and `createAuthHeaders()` outside of the cached functions, then calls `myActionCached(...)`.

     ```ts
     export async function myAction(params: MyParams): Promise<ActionResult<MyType>> {
       const { userId } = await auth();
       if (!userId) return { ok: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };
       const headers = await createAuthHeaders();
       return myActionCached(userId, headers, params);
     }
     ```

  5. Sanitize the response before returning or caching (inside core or after response) using `sanitizeProtobufForJson` so the cached value is JSON-safe.

     ```ts
     const response = await client.someRpc(request, { headers });
     const sanitized = sanitizeProtobufForJson(response);
     return { ok: true, data: sanitized };
     ```

- Revalidation (mutations)
  - Any mutation that modifies data returned by cached actions must call `revalidateTag('tag-name')` using the same tag used by `unstable_cache` for the affected resource.
  - Also call `revalidatePath('/spaces')` or `revalidatePath('/spaces/[spaceId]')` when you need immediate path-level refresh.

  Example in a create/update/delete server action:

  ```ts
  // after successful mutation
  revalidateTag(`spaces-list-${userId}`);
  revalidatePath('/spaces');
  ```

  ### Note on Invalidation Granularity

  Our current strategy uses coarse-grained invalidation (for example: `revalidateTag('spaces-list-{userId}')`). When any item in the collection changes we invalidate the entire collection. This approach is intentionally simple and reliable — it guarantees freshness without adding complexity.

  Fine-grained updates (such as patching a single item within a cached list) are possible but significantly more complex. For the current scale and requirements, coarse-grained revalidation provides the best trade-off between correctness, developer time, and operational simplicity.

- TTL recommendations
  - Spaces list: 5 minutes (300s)
  - Space details: 5 minutes (300s)
  - Content sources: 60 seconds (60s)

- Error handling and auth hygiene
  - Always call `createAuthHeaders()` outside cached functions so cached functions remain deterministic and don't implicitly depend on dynamic auth state.
  - Use `sanitizeError` to create safe error objects for client consumption and `sanitizeErrorString` for logging.
  - Detect and log auth failures with `isUnauthorizedError` and `logAuthFailure` to help troubleshooting without leaking tokens or PII.

  ### Policy on Caching Errors

  The persistent cache (`unstable_cache`) must only store successful responses. If a core fetch function returns an error (e.g., the backend returns a transient 5xx or a 4xx), do not write that error result into the shared cache. Caching errors risks serving a temporary outage to all clients until the TTL expires. Instead, let the next request attempt a fresh fetch so the system can self-heal when the backend recovers.

- Testing checklist
  - Unit test the cache key generation and sanitizer (`sanitizeProtobufForJson`) with representative protobuf objects (bigints, timestamps, nested arrays).
  - Unit test cached actions for hit/miss behavior (mock `unstable_cache` and `React.cache` or use small integration harness).
  - Integration test that server-rendered pages render with initial data (no client loading spinner) and that mutations trigger `revalidateTag`.

- Monitoring & observability
  - Add logging around cache errors and auth failures in `utils.ts` (use existing `logAuthFailure`).
  - Consider emitting metrics for cache hit/miss rates and revalidation counts.

- Notes / gotchas
  - Avoid serializing BigInt directly (Next.js errors on BigInt). Use `sanitizeProtobufForJson`.
  - Keep `unstable_cache` keys stable and deterministic; avoid including non-deterministic values.
  - Be careful with `toJson()`/`toObject()` methods on generated protobuf classes — prefer those when available for safe plain-object conversion.

If you follow this recipe, new server actions will integrate with the existing caching, revalidation, and auth model consistently.


