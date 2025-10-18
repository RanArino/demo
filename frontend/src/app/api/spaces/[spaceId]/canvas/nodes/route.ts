import { NextResponse } from 'next/server';
import { searchNodes } from '@/api/actions/canvasActions';
import { listContentSources } from '@/api/actions/contentActions';
import { convertProtoNodes } from '@/lib/canvas';

export async function GET(
  _request: Request,
  { params }: { params: { spaceId: string } }
) {
  const { spaceId } = params;

  if (!spaceId) {
    return NextResponse.json({ error: 'spaceId is required' }, { status: 400 });
  }

  const result = await searchNodes({
    filter: {
      // Only filter by space. Do not constrain abstraction_level here because
      // many Content/Chunk nodes don’t persist that property yet, and Neo4j
      // comparisons against missing properties exclude those rows.
      spaceId,
    },
    limit: 500,
  });

  if (!result.success) {
    return NextResponse.json(
      { error: result.error?.message ?? 'Failed to load canvas nodes' },
      { status: 500 }
    );
  }

  const response = result.data;
  const nodes = convertProtoNodes(response.nodes);
  const filteredBySpace = nodes.filter((node) => node.spaceId === spaceId);

  if (filteredBySpace.length !== nodes.length) {
    console.warn(
      `[canvas-nodes] Dropped ${nodes.length - filteredBySpace.length} nodes that did not match space ${spaceId}`
    );
  }

  let allowedSourceIds: Set<string> | null = null;
  try {
    const sourcesResult = await listContentSources(spaceId);
    if (sourcesResult.ok && Array.isArray(sourcesResult.data)) {
      allowedSourceIds = new Set(
        sourcesResult.data
          .map((item) => item?.id)
          .filter((id): id is string => typeof id === 'string' && id.length > 0)
      );
    } else if (!sourcesResult.ok) {
      console.warn('[canvas-nodes] Failed to fetch content sources for filtering:', sourcesResult.error);
    }
  } catch (error) {
    console.warn('[canvas-nodes] Error retrieving content sources:', error);
  }

  let finalNodes = filteredBySpace;
  if (allowedSourceIds && allowedSourceIds.size > 0) {
    const before = finalNodes.length;
    finalNodes = finalNodes.filter((node) => {
      if (node.kind === 'chunk' || node.kind === 'content') {
        if (node.contentSourceId) {
          return allowedSourceIds!.has(node.contentSourceId);
        }
      }
      return true;
    });
    const dropped = before - finalNodes.length;
    if (dropped > 0) {
      console.warn(`[canvas-nodes] Dropped ${dropped} nodes with foreign contentSourceId for space ${spaceId}`);
    }
  }

  return NextResponse.json({ nodes: finalNodes });
}
