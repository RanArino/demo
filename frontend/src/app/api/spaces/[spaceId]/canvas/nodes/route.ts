import { NextResponse } from 'next/server';
import { searchNodes } from '@/api/actions/canvasActions';
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

  return NextResponse.json({ nodes });
}
