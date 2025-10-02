'use server';

/**
 * Available server actions for canvas operations:
 * - getNodes(ids: string[])
 * - semanticSearch({ query: string, spaceId: string, topK?: number, nodeTypes?: number[] })
 * - searchNodes({ filter?: {...}, spatialBbox?: {...}, limit?: number })
 */

import { auth } from '@clerk/nextjs/server';
import { unstable_cache } from 'next/cache';
import { cache } from 'react';
import { getCanvasServiceClient } from '../server-client';
import {
  GetNodesRequest,
  GetNodesResponse,
  SemanticSearchRequest,
  SemanticSearchResponse,
  SearchNodesRequest,
  SearchNodesResponse,
  NodeFilter,
  SpatialBoundingBox,
  Node,
} from '../generated/v1/canvas_pb';
import {
  createAuthHeaders,
  sanitizeError,
  isUnauthorizedError,
  logAuthFailure,
} from './utils';

// ===== Type Definitions =====

type ApiError = {
  code?: string;
  message: string;
};

type ActionResult<T> =
  | { success: true; data: T }
  | { success: false; error: ApiError };

type HeaderResult =
  | { success: true; headers: Headers }
  | { success: false; error: ApiError };

async function buildCanvasHeaders(userId: string): Promise<HeaderResult> {
  const headers = await createAuthHeaders();

  // Ensure auth token is present before proceeding
  const hasAuthToken = headers.has('authorization') || !!headers.get('Authorization');
  if (!hasAuthToken) {
    return {
      success: false,
      error: {
        code: 'UNAUTHORIZED',
        message: 'Authentication token not available for current user.',
      },
    };
  }

  headers.set('x-user-id', userId);

  return {
    success: true,
    headers,
  };
}

// ===== Server Actions =====

/**
 * Core implementation for get nodes (without caching, with headers)
 */
async function getNodesCore(ids: string[], headers: Headers): Promise<ActionResult<{ nodes: Node[] }>> {
  try {
    const client = getCanvasServiceClient();
    const request = new GetNodesRequest({ ids });

    const response = await client.getNodes(request, { headers }) as GetNodesResponse;

    return {
      success: true,
      data: {
        nodes: response.nodes || []
      }
    };
  } catch (error) {
    console.error('[getNodesCore] Error:', error);
    return { success: false, error: sanitizeError(error) };
  }
}

/**
 * Server action to get nodes by their IDs
 */
export async function getNodes(ids: string[]): Promise<ActionResult<{ nodes: Node[] }>> {
  try {
    const { userId } = await auth();
    if (!userId) {
      return { success: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };
    }

    if (!ids || ids.length === 0) {
      return { success: false, error: { code: 'INVALID_ARGUMENT', message: 'Node IDs are required' } };
    }

    // Get auth headers BEFORE caching
    const headerResult = await buildCanvasHeaders(userId);
    if (!headerResult.success) {
      return headerResult;
    }
    const { headers } = headerResult;

    // Create stable cache key from sorted IDs
    const sortedIds = [...ids].sort();
    const key = [`canvas-getNodes-${userId}-${sortedIds.join(',')}`];

    const fetcher = () => getNodesCore(sortedIds, headers);
    return await unstable_cache(fetcher, key, { tags: ['canvas-nodes'], revalidate: 300 })();
  } catch (error) {
    if (isUnauthorizedError(error)) {
      logAuthFailure('getNodes', error);
    }
    return { success: false, error: sanitizeError(error) };
  }
}

/**
 * Core implementation for semantic search (without caching, with headers)
 */
async function semanticSearchCore(input: {
  query: string;
  spaceId: string;
  topK?: number;
  nodeTypes?: number[];
}, headers: Headers): Promise<ActionResult<SemanticSearchResponse>> {
  try {
    const client = getCanvasServiceClient();

    const request = new SemanticSearchRequest({
      query: input.query.trim(),
      spaceId: input.spaceId.trim(),
      topK: input.topK || 25,
      nodeTypes: input.nodeTypes || [],
    });

    const response = await client.semanticSearch(request, { headers }) as SemanticSearchResponse;

    return {
      success: true,
      data: response
    };
  } catch (error) {
    console.error('[semanticSearchCore] Error:', error);
    if (isUnauthorizedError(error)) {
      logAuthFailure('semanticSearch', error);
    }
    return { success: false, error: sanitizeError(error) };
  }
}

/**
 * Server action to perform semantic search
 */
export async function semanticSearch(input: {
  query: string;
  spaceId: string;
  topK?: number;
  nodeTypes?: number[];
}): Promise<ActionResult<SemanticSearchResponse>> {
  try {
    const { userId } = await auth();
    if (!userId) {
      return { success: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };
    }

    if (!input.query?.trim()) {
      return { success: false, error: { code: 'INVALID_ARGUMENT', message: 'Query is required' } };
    }

    if (!input.spaceId?.trim()) {
      return { success: false, error: { code: 'INVALID_ARGUMENT', message: 'Space ID is required' } };
    }

    // Get auth headers BEFORE caching
    const headerResult = await buildCanvasHeaders(userId);
    if (!headerResult.success) {
      return headerResult;
    }
    const { headers } = headerResult;

    // Create stable cache key
    const normalizedInput = {
      query: input.query.trim(),
      spaceId: input.spaceId.trim(),
      topK: input.topK || 25,
      nodeTypes: input.nodeTypes || []
    };
    const key = [`canvas-semanticSearch-${userId}-${JSON.stringify(normalizedInput)}`];

    const fetcher = () => semanticSearchCore(normalizedInput, headers);
    return await unstable_cache(fetcher, key, { tags: ['canvas-semantic-search'], revalidate: 300 })();
  } catch (error) {
    if (isUnauthorizedError(error)) {
      logAuthFailure('semanticSearch', error);
    }
    return { success: false, error: sanitizeError(error) };
  }
}

/**
 * Core implementation for search nodes (without caching, with headers)
 */
async function searchNodesCore(input: {
  filter?: {
    spaceId?: string;
    abstractionLevelMin?: number;
    abstractionLevelMax?: number;
    contextType?: string;
    keywords?: string[];
  };
  spatialBbox?: {
    minCoords?: { x?: number; y?: number; z?: number };
    maxCoords?: { x?: number; y?: number; z?: number };
  };
  limit?: number;
}, headers: Headers): Promise<ActionResult<SearchNodesResponse>> {
  try {
    const client = getCanvasServiceClient();

    // Build NodeFilter if provided
    let nodeFilter: NodeFilter | undefined;
    if (input.filter) {
      nodeFilter = new NodeFilter({
        spaceId: input.filter.spaceId,
        abstractionLevelMin: input.filter.abstractionLevelMin,
        abstractionLevelMax: input.filter.abstractionLevelMax,
        contextType: input.filter.contextType,
        keywords: input.filter.keywords || [],
      });
    }

    // Build SpatialBoundingBox if provided
    let spatialBbox: SpatialBoundingBox | undefined;
    if (input.spatialBbox) {
      spatialBbox = new SpatialBoundingBox({
        minCoords: input.spatialBbox.minCoords,
        maxCoords: input.spatialBbox.maxCoords,
      });
    }

    const request = new SearchNodesRequest({
      filter: nodeFilter,
      spatialBbox: spatialBbox,
      limit: input.limit || 100,
    });

    const response = await client.searchNodes(request, { headers }) as SearchNodesResponse;

    return {
      success: true,
      data: response
    };
  } catch (error) {
    console.error('[searchNodesCore] Error:', error);
    if (isUnauthorizedError(error)) {
      logAuthFailure('searchNodes', error);
    }
    return { success: false, error: sanitizeError(error) };
  }
}

/**
 * Server action to search nodes using filters and spatial bounds
 */
export async function searchNodes(input: {
  filter?: {
    spaceId?: string;
    abstractionLevelMin?: number;
    abstractionLevelMax?: number;
    contextType?: string;
    keywords?: string[];
  };
  spatialBbox?: {
    minCoords?: { x?: number; y?: number; z?: number };
    maxCoords?: { x?: number; y?: number; z?: number };
  };
  limit?: number;
}): Promise<ActionResult<SearchNodesResponse>> {
  try {
    const { userId } = await auth();
    if (!userId) {
      return { success: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };
    }

    // Get auth headers BEFORE caching
    const headerResult = await buildCanvasHeaders(userId);
    if (!headerResult.success) {
      return headerResult;
    }
    const { headers } = headerResult;

    // Create stable cache key
    const normalizedInput = {
      filter: input.filter || {},
      spatialBbox: input.spatialBbox || {},
      limit: input.limit || 100
    };
    const key = [`canvas-searchNodes-${userId}-${JSON.stringify(normalizedInput)}`];

    const fetcher = () => searchNodesCore(normalizedInput, headers);
    return await unstable_cache(fetcher, key, { tags: ['canvas-nodes'], revalidate: 300 })();
  } catch (error) {
    if (isUnauthorizedError(error)) {
      logAuthFailure('searchNodes', error);
    }
    return { success: false, error: sanitizeError(error) };
  }
}