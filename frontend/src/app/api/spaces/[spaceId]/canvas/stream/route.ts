import { NextResponse } from 'next/server';

// Ensure this route never gets statically cached or buffered by the
// framework so the NDJSON chunks flush to the client immediately.
export const dynamic = 'force-dynamic';
export const revalidate = 0;
export const fetchCache = 'force-no-store';
export const runtime = 'nodejs';
import { getNeighbors, searchNodesUncached } from '@/api/actions/canvasActions';
import { convertProtoNodes, convertProtoNode } from '@/lib/canvas';
import type { CanvasRenderableNode, CanvasNodeKind } from '@/lib/canvas';
import { Direction, LinkType } from '@/api/generated/v1/canvas_pb';

const CLUSTER_LIMIT_DEFAULT = 60;
const CONTENT_LIMIT_PER_CLUSTER = 40;
const CHUNK_LIMIT_PER_CONTENT = 120;
const CONTENT_NEIGHBOR_BATCH_SIZE = 12;
const ROOT_CONTENT_LIMIT = 160;
const ROOT_CHUNK_LIMIT = 320;

type StreamEvent =
  | { event: 'meta'; message?: string; limits?: Record<string, number> }
  | { event: 'nodes'; level: CanvasNodeKind; parentId?: string; nodes: CanvasRenderableNode[] }
  | { event: 'error'; message: string }
  | { event: 'end'; counts: Record<CanvasNodeKind, number> };

function encode(controller: ReadableStreamDefaultController<Uint8Array>, payload: StreamEvent) {
  const data = JSON.stringify(payload);
  controller.enqueue(new TextEncoder().encode(`${data}\n`));
}

export async function GET(_request: Request, { params }: { params: { spaceId: string } }) {
  const { spaceId } = params;
  if (!spaceId) {
    return NextResponse.json({ error: 'spaceId is required' }, { status: 400 });
  }

  const stream = new ReadableStream<Uint8Array>({
    async start(controller) {
      encode(controller, {
        event: 'meta',
        message: 'Starting canvas graph stream',
        limits: {
          cluster: CLUSTER_LIMIT_DEFAULT,
          contentPerCluster: CONTENT_LIMIT_PER_CLUSTER,
          chunkPerContent: CHUNK_LIMIT_PER_CONTENT,
        },
      });

      const counts: Record<CanvasNodeKind, number> = {
        cluster: 0,
        content: 0,
        chunk: 0,
      };

      const seenIds = new Set<string>();

      try {
        const initialSearch = await searchNodesUncached({
          filter: { spaceId },
          limit: ROOT_CHUNK_LIMIT,
        });

        if (!initialSearch.success) {
          encode(controller, { event: 'error', message: initialSearch.error?.message ?? 'Failed to load nodes' });
          controller.close();
          return;
        }

        const initialNodes = convertProtoNodes(initialSearch.data.nodes).filter((node) => node.spaceId === spaceId);
        const clusters = initialNodes.filter((node): node is CanvasRenderableNode & { kind: 'cluster' } => node.kind === 'cluster');
        const contents = initialNodes.filter((node): node is CanvasRenderableNode & { kind: 'content' } => node.kind === 'content');
        const chunks = initialNodes.filter((node): node is CanvasRenderableNode & { kind: 'chunk' } => node.kind === 'chunk');

        const contentIds: string[] = [];

        const topLayerKind: CanvasNodeKind | null = clusters.length > 0
          ? 'cluster'
          : contents.length > 0
            ? 'content'
            : chunks.length > 0
              ? 'chunk'
              : null;

        if (!topLayerKind) {
          encode(controller, { event: 'end', counts });
          controller.close();
          return;
        }

        if (topLayerKind === 'cluster') {
          const topClusters = clusters.slice(0, CLUSTER_LIMIT_DEFAULT);
          counts.cluster += topClusters.length;
          encode(controller, { event: 'nodes', level: 'cluster', nodes: topClusters });
          topClusters.forEach((node) => {
            seenIds.add(node.id);
          });

          const neighborResult = await getNeighbors({
            ids: topClusters.map((node) => node.id),
            direction: Direction.OUTGOING,
            limitPerNode: CONTENT_LIMIT_PER_CLUSTER,
            linkTypes: [LinkType.HIERARCHICAL],
          });

          if (!neighborResult.success) {
            encode(controller, { event: 'error', message: neighborResult.error?.message ?? 'Failed to load content neighbors' });
            controller.close();
            return;
          }

          const entries = Object.entries(neighborResult.data.results ?? {});
          for (const [clusterId, list] of entries) {
            const neighbors = list?.neighbors ?? [];
            const converted: CanvasRenderableNode[] = [];
            for (const neighbor of neighbors) {
              const node = convertProtoNode(neighbor.node);
              if (!node || node.kind !== 'content' || node.spaceId !== spaceId) {
                continue;
              }
              if (seenIds.has(node.id)) {
                continue;
              }
              seenIds.add(node.id);
              contentIds.push(node.id);
              converted.push(node);
            }
            if (converted.length > 0) {
              counts.content += converted.length;
              encode(controller, { event: 'nodes', level: 'content', parentId: clusterId, nodes: converted });
            }
          }
          if (contentIds.length === 0 && contents.length > 0) {
            const fresh = contents.filter((node) => !seenIds.has(node.id)).slice(0, ROOT_CONTENT_LIMIT);
            fresh.forEach((node) => {
              seenIds.add(node.id);
              contentIds.push(node.id);
            });
            if (fresh.length > 0) {
              counts.content += fresh.length;
              encode(controller, { event: 'nodes', level: 'content', nodes: fresh });
            }
          }
        } else if (topLayerKind === 'content') {
          const topContents = contents.slice(0, ROOT_CONTENT_LIMIT);
          counts.content += topContents.length;
          encode(controller, { event: 'nodes', level: 'content', nodes: topContents });
          topContents.forEach((node) => {
            seenIds.add(node.id);
            contentIds.push(node.id);
          });
        } else if (topLayerKind === 'chunk') {
          const topChunks = chunks.slice(0, ROOT_CHUNK_LIMIT);
          counts.chunk += topChunks.length;
          encode(controller, { event: 'nodes', level: 'chunk', nodes: topChunks });
          topChunks.forEach((node) => seenIds.add(node.id));
        }

        if (contentIds.length > 0) {
          const chunkIds = new Set<string>();
          for (let i = 0; i < contentIds.length; i += CONTENT_NEIGHBOR_BATCH_SIZE) {
            const slice = contentIds.slice(i, i + CONTENT_NEIGHBOR_BATCH_SIZE);
            const chunkNeighbors = await getNeighbors({
              ids: slice,
              direction: Direction.OUTGOING,
              limitPerNode: CHUNK_LIMIT_PER_CONTENT,
              linkTypes: [LinkType.HIERARCHICAL],
            });

            if (!chunkNeighbors.success) {
              encode(controller, { event: 'error', message: chunkNeighbors.error?.message ?? 'Failed to load chunk neighbors' });
              controller.close();
              return;
            }

            const entries = Object.entries(chunkNeighbors.data.results ?? {});
            for (const [contentId, list] of entries) {
              const neighbors = list?.neighbors ?? [];
              const converted: CanvasRenderableNode[] = [];
              for (const neighbor of neighbors) {
                const node = convertProtoNode(neighbor.node);
                if (!node || node.kind !== 'chunk' || node.spaceId !== spaceId) {
                  continue;
                }
                if (chunkIds.has(node.id) || seenIds.has(node.id)) {
                  continue;
                }
                chunkIds.add(node.id);
                seenIds.add(node.id);
                converted.push(node);
              }
              if (converted.length > 0) {
                counts.chunk += converted.length;
                encode(controller, { event: 'nodes', level: 'chunk', parentId: contentId, nodes: converted });
              }
            }
          }

          if (chunkIds.size === 0 && chunks.length > 0) {
            const freshChunks = chunks.filter((node) => !seenIds.has(node.id)).slice(0, ROOT_CHUNK_LIMIT);
            if (freshChunks.length > 0) {
              freshChunks.forEach((node) => seenIds.add(node.id));
              counts.chunk += freshChunks.length;
              encode(controller, { event: 'nodes', level: 'chunk', nodes: freshChunks });
            }
          }
        }

        encode(controller, { event: 'end', counts });
        controller.close();
      } catch (error) {
        globalThis.console?.error?.('[canvas-stream] Unexpected error', error);
        const message = error instanceof Error ? error.message : 'Unexpected error';
        encode(controller, { event: 'error', message });
        controller.close();
      }
    },
  });

  return new Response(stream, {
    headers: {
      'Content-Type': 'application/x-ndjson; charset=utf-8',
      'Cache-Control': 'no-store',
    },
  });
}
