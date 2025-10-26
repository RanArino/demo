import { NextResponse } from 'next/server';

export const dynamic = 'force-dynamic';
export const revalidate = 0;
export const runtime = 'nodejs';

import { listNodesByLink } from '@/api/actions/canvasActions';
import {
  Direction,
  LinkFilter,
  LinkType,
  HierarchicalLinkFilter,
  NodeType,
} from '@/api/generated/v1/canvas_pb';
import { convertProtoNode, type CanvasRenderableNode, type CanvasNodeKind } from '@/lib/canvas';

const NODE_KIND_TO_ENUM: Record<CanvasNodeKind, NodeType> = {
  cluster: NodeType.CLUSTER,
  content: NodeType.CONTENT,
  chunk: NodeType.CHUNK,
};

const CHILD_LIMITS: Record<CanvasNodeKind, number> = {
  cluster: 80,
  content: 160,
  chunk: 0,
};

type ChildrenRequestPayload = {
  parentId?: string;
  contentSourceId?: string;
  parentType: CanvasNodeKind;
  spaceId: string;
  pageToken?: string | null;
};

export async function POST(request: Request, { params }: { params: { spaceId: string } }) {
  try {
    const payload = (await request.json()) as ChildrenRequestPayload;
    const urlSpaceId = params.spaceId;

    if (!payload.spaceId || payload.spaceId !== urlSpaceId) {
      return NextResponse.json({ error: 'spaceId mismatch' }, { status: 400 });
    }

    if (!payload.parentType || (payload.parentType !== 'cluster' && payload.parentType !== 'content')) {
      return NextResponse.json({ error: 'Unsupported parent type' }, { status: 400 });
    }

    if (!payload.parentId && !payload.contentSourceId) {
      return NextResponse.json({ error: 'parentId or contentSourceId is required' }, { status: 400 });
    }

    const parentReference = {
      nodeId: payload.parentId,
      contentSourceId: payload.contentSourceId,
      spaceId: payload.spaceId,
      nodeType: NODE_KIND_TO_ENUM[payload.parentType],
    };

    const perParentLimit = CHILD_LIMITS[payload.parentType] ?? 60;

    const traversalFilter = new LinkFilter({
      hierarchical: new HierarchicalLinkFilter({ hierarchyDepth: 1 }),
    });

    const result = await listNodesByLink({
      parents: [parentReference],
      direction: Direction.OUTGOING,
      linkTypes: [LinkType.HIERARCHICAL],
      linkFilter: traversalFilter,
      limitPerParent: perParentLimit,
      maxTotal: perParentLimit,
      pageToken: payload.pageToken ?? undefined,
      includeLinkMetadata: false,
    });

    if (!result.success) {
      const status = result.error?.code === 'UNAUTHORIZED' ? 401 : 500;
      return NextResponse.json({ error: result.error?.message ?? 'Failed to fetch children' }, { status });
    }

    const batches = result.data?.batches ?? [];
    const nodes: CanvasRenderableNode[] = [];

    for (const batch of batches) {
      for (const neighbor of batch.neighbors ?? []) {
        const converted = convertProtoNode(neighbor.node);
        if (converted) {
          nodes.push(converted);
        }
      }
    }

    const nextPageToken = result.data?.nextPageToken ?? batches[0]?.pageToken ?? null;

    return NextResponse.json({
      nodes,
      nextPageToken,
    });
  } catch (error) {
    console.error('[children route] failed', error);
    return NextResponse.json({ error: 'Failed to fetch child nodes' }, { status: 500 });
  }
}
