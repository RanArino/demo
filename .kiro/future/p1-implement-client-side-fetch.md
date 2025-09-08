Here is the updated development plan, revised to incorporate the more performant and robust server-side caching strategy. This new approach aligns perfectly with the Next.js App Router's paradigm.

-----



The fundamental hybrid model remains, but we will now orchestrate it in a more idiomatic way for Next.js.

  * **Server Component for Initial Data:** We will use a **Server Component** to fetch and cache the initial data. This ensures the data is ready *before* the page is sent to the client, eliminating client-side loading states.
  * **Client Component for Real-Time Updates:** The component handling the live updates will remain a **Client Component**, but it will now receive its initial state via props, simplifying its logic.

Here is the revised, high-performance implementation.

-----

### Revised Architectural Plan: Layered Caching

To fully align with Next.js's caching model, we will combine two complementary layers:

1. **React `cache` (Request Memoization):** Prevents duplicate fetches within a single server render.
2. **`unstable_cache` (Data Cache):** Persists data across requests and deployments with tag/time revalidation.

This layered approach maximizes performance for both complex component trees and repeated navigations.

-----

### 1. Server-Side gRPC Client (Singleton)

This piece of the architecture is perfect as-is. A singleton instance on the server is the correct and efficient way to manage connections.

**File:** `frontend/src/api/server-client.ts`
*(No changes to this file)*

```typescript
import { GrpcTransport } from "@protobuf-ts/grpc-transport";
import { KnowledgeServiceClient } from "./generated/v1/knowledge.client";

let client: KnowledgeServiceClient | null = null;

export function getGrpcClient() {
  if (!client) {
    const transport = new GrpcTransport({
      host: process.env.GRPC_BACKEND_URL!,
    });
    client = new KnowledgeServiceClient(transport);
  }
  return client;
}
```

-----

### 2. Layered & Cached Server Action

We will wrap our data-fetching logic in two layers of caching, as recommended for non-`fetch` clients:

- **Outer `cache` (React Request Memoization):** Ensures only one execution per server render.
- **Inner `unstable_cache` (Next.js Data Cache):** Persists results across requests with tag/time revalidation.

**File:** `frontend/src/api/actions/canvasActions.ts`
*(Updated with Layered Caching)*

```typescript
"use server";

import { getGrpcClient } from "@/api/server-client";
import { unstable_cache } from "next/cache";
import { cache } from "react"; // React's per-request memoization

// React.cache → per-request memoization
export const getInitialCanvasState = cache(
  // Next.js unstable_cache → persistent Data Cache
  unstable_cache(
    async (canvasId: string) => {
      const client = getGrpcClient();
      // Logs only on Data Cache MISS
      console.log(`Fetching data for canvas: ${canvasId}`);
      const response = await client.getCanvas({ canvasId });
      return response.response;
    },
    ["canvas-state"], // Cache key group
    {
      // On-demand revalidation with revalidateTag(`canvas-${id}`)
      tags: [(canvasId) => `canvas-${canvasId}` as unknown as string],
      revalidate: 3600, // Optional: revalidate at most once per hour
    }
  )
);
```

-----

### 3. Server Component for Data Orchestration

This is a new and crucial piece. We'll create a page using a Server Component (the default in the `app` directory). Its job is to call the cached action and pass the data down to our client component. This is the core of the server-first pattern.

**File:** `frontend/src/app/canvas/[id]/page.tsx`
*(New file)*

```typescript
import { getInitialCanvasState } from "@/api/actions/canvasActions";
import { ReadOnlyCanvas } from "@/features/canvas/ReadOnlyCanvas";
import { Suspense } from "react";

// This is an async Server Component
export default async function CanvasPage({ params }: { params: { id: string } }) {
  // 1. Fetch data on the server. The result will be cached.
  const initialData = await getInitialCanvasState(params.id);

  // 2. Pass the data as a prop to the client component.
  // The client component will render instantly with this data.
  return (
    <Suspense fallback={<div>Loading Canvas...</div>}>
      <ReadOnlyCanvas canvasId={params.id} initialData={initialData} />
    </Suspense>
  );
}
```

-----

### 4. Client-Side gRPC-web Client for Streaming

This client-side utility for establishing the real-time connection remains unchanged.

**File:** `frontend/src/api/client.ts`
*(No changes to this file)*

```typescript
import { GrpcWebFetchTransport } from "@protobuf-ts/grpcweb-transport";
import { KnowledgeServiceClient } from "./generated/v1/knowledge.client";

export function getGrpcWebClient() {
  const transport = new GrpcWebFetchTransport({
    baseUrl: process.env.NEXT_PUBLIC_GRPC_WEB_URL!,
  });
  return new KnowledgeServiceClient(transport);
}
```

-----

### 5. Updated Hybrid Component Implementation

Finally, we simplify the `ReadOnlyCanvas` component. It no longer needs to fetch its own data or manage a loading state. It receives the initial data as a prop and is only responsible for rendering and establishing the real-time stream. ✅

**File:** `frontend/src/features/canvas/ReadOnlyCanvas.tsx`
*(Updated)*

```typescript
"use client";

import { useEffect } from "react";
import { getGrpcWebClient } from "@/api/client";
import { ThreeScene } from "./ThreeScene"; // Your 3D component
import { CanvasState } from "@/api/generated/v1/knowledge";

// 1. Receive the server-fetched initialData as a prop
export function ReadOnlyCanvas({ canvasId, initialData }: { canvasId: string, initialData: CanvasState }) {
  // 2. The initial fetch logic (useState, useEffect) is now gone!

  // 3. Real-time updates via gRPC-web Stream
  useEffect(() => {
    // The connection can be established immediately
    const client = getGrpcWebClient();
    const stream = client.streamCanvasUpdates({ canvasId });

    const listen = async () => {
      for await (const update of stream.responses) {
        console.log("Received position update:", update);
        // Update your 3D scene's state with the new data
      }
    };

    listen();

    return () => stream.cancel(); // Clean up on unmount
  }, [canvasId]); // Dependency on initialData is no longer needed

  // 4. No loading state needed, the data is always present on render
  return <ThreeScene data={initialData} />;
}
```

This updated architecture provides a significantly better user experience with instant page loads on navigation, while also simplifying the client-side component and adhering to modern Next.js best practices. 🚀