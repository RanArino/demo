# Frontend Data Fetching Implementation Guide (ms_canvas via Direct gRPC)

This guide provides the specific code changes needed to implement frontend data fetching from ms_canvas, following the same pattern as ms_knowledge and ms_user services.

## Backend Changes (ms_canvas)

### 1. Update cmd/main.go

Add gRPC server setup to the main function:

```go
// Add these imports
import (
    "net"
    canvasv1 "demo/ms_canvas/go_app/api/proto/public/v1"
    "demo/ms_canvas/go_app/internal/server"
    "google.golang.org/grpc"
)

// In main() function, after initializing services:
func main() {
    // ... existing initialization code ...

    // Start gRPC server for CanvasPublic API
    grpcSrv := grpc.NewServer()

    // Initialize services
    searchRepo := neo4j.NewSearchRepo(drv)
    searchSvc := service.NewSearchService(pythonGateway, searchRepo)
    nodeSvc := service.NewNodeService(nodeRepo, linkRepo)
    linkSvc := service.NewLinkService(linkRepo)

    // Register CanvasPublic server
    canvasServer := server.NewCanvasPublicServer(searchSvc, nodeSvc, linkSvc)
    canvasv1.RegisterCanvasPublicServer(grpcSrv, canvasServer)

    // Start gRPC server on port 50054
    lis, err := net.Listen("tcp", ":50054")
    if err != nil {
        log.Fatalf("failed to listen on gRPC port: %v", err)
    }
    go func() {
        log.Printf("[Main] gRPC server listening on :50054")
        if err := grpcSrv.Serve(lis); err != nil {
            log.Printf("[Main] gRPC server stopped: %v", err)
        }
    }()

    // ... rest of existing code ...
}
```

### 2. Update internal/server/grpc.go

Modify SearchNodes to be more flexible for testing:

```go
// In SearchNodes method, remove strict spatial_bbox requirement
func (s *canvasPublicServer) SearchNodes(ctx context.Context, req *canvasv1.SearchNodesRequest) (*canvasv1.SearchNodesResponse, error) {
    // ... existing code ...

    // Remove this check to allow searches without spatial bounds
    // if req.SpatialBbox == nil {
    //     return nil, status.Error(codes.InvalidArgument, "spatial_bbox is required for SearchNodes")
    // }

    // ... rest of existing code ...
}
```

## Frontend Changes (Next.js)

### 3. Create gRPC Client for ms_canvas

Create `frontend/src/api/canvasClient.ts` that connects directly to ms_canvas gRPC service, following the same pattern as ms_knowledge and ms_user:

```typescript
// frontend/src/api/canvasClient.ts
import { createPromiseClient, PromiseClient } from '@bufbuild/connect';
import { createGrpcTransport } from '@bufbuild/connect-node';
import { CanvasPublic } from './generated/v1/canvas_connectweb';

/**
 * Singleton gRPC client manager for ms_canvas operations
 * Following the same pattern as ms_knowledge and ms_user
 */
class CanvasClientManager {
  private static instance: PromiseClient<typeof CanvasPublic> | null = null;

  static getInstance(): PromiseClient<typeof CanvasPublic> {
    if (!this.instance) {
      const grpcUrl = process.env.MS_CANVAS_GRPC_URL || 'http://localhost:50054';
      const fullUrl = grpcUrl.startsWith('http') ? grpcUrl : `http://${grpcUrl}`;

      const transport = createGrpcTransport({
        httpVersion: '2',
        baseUrl: fullUrl,
      });

      this.instance = createPromiseClient(CanvasPublic, transport);
    }

    return this.instance;
  }
}

export const getCanvasClient = (): PromiseClient<typeof CanvasPublic> => {
  return CanvasClientManager.getInstance();
};
```

### 4. Create Server Actions (Direct gRPC)

Create `frontend/src/api/actions/canvasActions.ts` that calls ms_canvas gRPC endpoints directly:

```ts
'use server';

import { auth } from '@clerk/nextjs/server';
import { unstable_cache } from 'next/cache';
import { cache } from 'react';
import { ActionResult } from '@/lib/types';
import { sanitizeError, isUnauthorizedError, logAuthFailure } from './utils';
import { getCanvasClient } from '../canvasClient';
import { GetNodesRequest } from '../generated/v1/canvas_pb';

async function getNodesCore(ids: string[]): Promise<ActionResult<any>> {
  try {
    const client = getCanvasClient();
    const request = new GetNodesRequest({ nodeIds: ids });

    const response = await client.getNodes(request);
    return { ok: true, data: { nodes: response.nodes } };
  } catch (error) {
    return { ok: false, error: sanitizeError(error) };
  }
}

const getNodesMemoized = cache(async (ids: string[]) => getNodesCore(ids));

export async function getNodes(ids: string[]): Promise<ActionResult<any>> {
  try {
    const { userId } = await auth();
    if (!userId) return { ok: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };
    const key = [`canvas-getNodes-${ids.sort().join(',')}`];
    const fetcher = () => getNodesMemoized(ids);
    return await unstable_cache(fetcher, key, { tags: ['canvas-nodes'], revalidate: 300 })();
  } catch (error) {
    if (isUnauthorizedError(error)) logAuthFailure('getNodes', error);
    return { ok: false, error: sanitizeError(error) };
  }
}

export async function semanticSearch(input: { query: string; spaceId: string; topK?: number; nodeTypes?: number[]; }): Promise<ActionResult<any>> {
  try {
    const { userId } = await auth();
    if (!userId) return { ok: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };
    const headers = await createAuthHeader();
    const res = await fetch(`${base}/v1/canvas/semanticSearch`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...headers },
      body: JSON.stringify({ query: input.query, space_id: input.spaceId, top_k: input.topK ?? 25, node_types: input.nodeTypes ?? [] })
    });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const data = await res.json();
    return { ok: true, data };
  } catch (error) {
    if (isUnauthorizedError(error)) logAuthFailure('semanticSearch', error);
    return { ok: false, error: sanitizeError(error) };
  }
}
```

## Environment Configuration

### 5. Update docker-compose.dev.yml

Add environment variable for canvas service via API Gateway:

```yaml
services:
  frontend:
    environment:
      - NEXT_PUBLIC_CANVAS_HTTP_BASE=https://<your-api-gateway-host>

  ms_canvas:
    # No change required for gRPC port exposure
```

## Testing the Implementation

### 6. Simple Test Page (using Server Actions)

Create `frontend/src/app/canvas-test/page.tsx`:

```tsx
'use client';

import { useState } from 'react';

export default function CanvasTestPage() {
    const [nodeIds, setNodeIds] = useState('');
    const [searchQuery, setSearchQuery] = useState('');
    const [spaceId, setSpaceId] = useState('');
    const [results, setResults] = useState<any>(null);
    const [loading, setLoading] = useState(false);

    const fetchNodes = async () => {
        setLoading(true);
        try {
            const action = (await import('@/src/api/actions/canvasActions')).getNodes;
            const data = await action(nodeIds.split(',').map(id => id.trim()));
            setResults(data);
        } catch (error) {
            setResults({ error: 'Failed to fetch nodes' });
        }
        setLoading(false);
    };

    const semanticSearch = async () => {
        setLoading(true);
        try {
            const action = (await import('@/src/api/actions/canvasActions')).semanticSearch;
            const data = await action({ query: searchQuery, spaceId, topK: 10 });
            setResults(data);
        } catch (error) {
            setResults({ error: 'Failed to perform search' });
        }
        setLoading(false);
    };

    return (
        <div className="p-6">
            <h1 className="text-2xl font-bold mb-4">Canvas Service Test</h1>

            <div className="mb-6">
                <h2 className="text-lg font-semibold mb-2">Get Nodes by IDs</h2>
                <input
                    type="text"
                    placeholder="Enter node IDs (comma-separated)"
                    value={nodeIds}
                    onChange={(e) => setNodeIds(e.target.value)}
                    className="border p-2 mr-2"
                />
                <button
                    onClick={fetchNodes}
                    disabled={loading}
                    className="bg-blue-500 text-white px-4 py-2 rounded"
                >
                    {loading ? 'Loading...' : 'Fetch Nodes'}
                </button>
            </div>

            <div className="mb-6">
                <h2 className="text-lg font-semibold mb-2">Semantic Search</h2>
                <input
                    type="text"
                    placeholder="Search query"
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="border p-2 mr-2"
                />
                <input
                    type="text"
                    placeholder="Space ID"
                    value={spaceId}
                    onChange={(e) => setSpaceId(e.target.value)}
                    className="border p-2 mr-2"
                />
                <button
                    onClick={semanticSearch}
                    disabled={loading}
                    className="bg-green-500 text-white px-4 py-2 rounded"
                >
                    {loading ? 'Loading...' : 'Search'}
                </button>
            </div>

            {results && (
                <div className="mt-6">
                    <h2 className="text-lg font-semibold mb-2">Results</h2>
                    <pre className="bg-gray-100 p-4 rounded overflow-auto">
                        {JSON.stringify(results, null, 2)}
                    </pre>
                </div>
            )}
        </div>
    );
}
```

## Deployment Steps

1. **Start Backend Services:**
   ```bash
   cd ms_canvas/go_app
   go run cmd/main.go
   ```

2. **Start Frontend:**
   ```bash
   cd frontend
   npm run dev
   ```

3. **Test the Integration:**
   - Navigate to `/canvas-test`
   - Enter node IDs or search terms
   - Verify data is fetched from ms_canvas

4. **Monitor Logs:**
   - Check ms_canvas logs for gRPC calls
   - Check frontend console for errors
   - Verify database queries in Neo4j logs

## Troubleshooting

- **gRPC Connection Issues:** Check MS_CANVAS_GRPC_URL environment variable
- **Database Errors:** Ensure Neo4j is running and accessible
- **Python Service Errors:** Check if Python ML service is available for semantic search
- **Port Conflicts:** Ensure port 50054 is not already in use

This implementation provides a complete end-to-end solution for frontend data fetching from the ms_canvas backend service.

