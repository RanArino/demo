'use server';

import { auth } from '@clerk/nextjs/server';
import { revalidatePath, revalidateTag, unstable_cache } from 'next/cache';
import { cache } from 'react';
import { getKnowledgeServiceClient } from '../server-client';
import {
  Space,
  SpaceFilters,
  ListSpacesRequest,
  GetSpaceRequest,
  CreateSpaceRequest,
  UpdateSpaceRequest,
  DeleteSpaceRequest,
} from '../generated/v1/knowledge_pb';
import { ActionResult } from '@/lib/types'; 
import { createAuthHeaders, sanitizeError, generateSpacesListCacheKey, sanitizeProtobufForJson } from './utils';



/**
 * Core implementation for searching spaces (without caching)
 * This is the actual gRPC call that will be cached
 * Auth headers must be passed in to avoid dynamic data access inside cache
 */
async function searchSpacesCore(
  userId: string, 
  headers: Headers, 
  filters?: SpaceFilters
): Promise<ActionResult<{spaces: Space[], totalCount: number, page: number, pageSize: number}>> {
  try {
    const client = getKnowledgeServiceClient();

    // Create a proper ListSpacesRequest from SpaceFilters
    const request = new ListSpacesRequest({
      q: filters?.q || '',
      keywords: filters?.keywords || [],
      // Always filter by current user's spaces for security
      ownerId: userId,
    });

    const response = await client.listSpaces(request, { headers });

    // Ensure JSON-serializable response: convert any BigInt fields within Space to numbers
    const spaces = response.items.map((space) => sanitizeProtobufForJson(space));
    return {
      ok: true,
      data: {
        spaces,
        totalCount: spaces.length,
        page: 1, // Default page since SpaceFilters doesn't have page property
        pageSize: spaces.length, // Default pageSize since SpaceFilters doesn't have pageSize property
      },
    };
  } catch (error) {
    console.error('searchSpacesCore error:', error);
    return { ok: false, error: sanitizeError(error) };
  }
}

/**
 * Per-request memoized version using React.cache
 * Prevents duplicate calls within a single render
 */
const searchSpacesMemoized = cache(async (userId: string, headers: Headers, filters?: SpaceFilters) => {
  return searchSpacesCore(userId, headers, filters);
});

/**
 * Cross-request cached version using unstable_cache
 * Persists data across requests with tag-based revalidation
 */
const searchSpacesCached = (userId: string, headers: Headers, filters?: SpaceFilters) => {
  const cacheKey = generateSpacesListCacheKey(userId, filters);
  
  return unstable_cache(
    async () => searchSpacesMemoized(userId, headers, filters),
    [cacheKey],
    {
      tags: [`spaces-list-${userId}`],
      revalidate: 300, // 5 minutes TTL as per design.md
    }
  )();
};

/**
 * Search/list spaces with filters (with dual-layer caching)
 */
export async function searchSpaces(filters?: SpaceFilters): Promise<ActionResult<{spaces: Space[], totalCount: number, page: number, pageSize: number}>> {
  try {
    const { userId } = await auth();
    if (!userId) {
      return { ok: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };
    }

    // Create auth headers outside the cached function
    const headers = await createAuthHeaders();
    
    return await searchSpacesCached(userId, headers, filters);
  } catch (error) {
    console.error('searchSpaces action error:', error);
    return { ok: false, error: sanitizeError(error) };
  }
}

/**
 * Get a single space by ID
 */
// Core fetcher for getSpace
async function getSpaceCore(
  _userId: string,
  headers: Headers,
  spaceId: string
): Promise<ActionResult<Space>> {
  try {
    const client = getKnowledgeServiceClient();
    const request = new GetSpaceRequest({ id: spaceId });
    const response = await client.getSpace(request, { headers });
    const sanitized = sanitizeProtobufForJson(response);
    return { ok: true, data: sanitized };
  } catch (error) {
    console.error('getSpaceCore error:', error);
    return { ok: false, error: sanitizeError(error) };
  }
}

// Per-request memoization
const getSpaceMemoized = cache(async (_userId: string, headers: Headers, spaceId: string) => {
  return getSpaceCore(_userId, headers, spaceId);
});

// Cross-request cache with tag
const getSpaceCached = (_userId: string, headers: Headers, spaceId: string) => {
  const key = [`getSpace-${spaceId}`];
  return unstable_cache(
    async () => getSpaceMemoized(_userId, headers, spaceId),
    key,
    {
      tags: [`space-${spaceId}`],
      revalidate: 300,
    }
  )();
};

export async function getSpace(spaceId: string): Promise<ActionResult<Space>> {
  try {
    const { userId } = await auth();
    if (!userId) {
      return { ok: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };
    }
    const headers = await createAuthHeaders();
    return await getSpaceCached(userId, headers, spaceId);
  } catch (error) {
    console.error('getSpace action error:', error);
    return { ok: false, error: sanitizeError(error) };
  }
}

/**
 * Create a new space
 */
export async function createSpace(input: CreateSpaceRequest): Promise<ActionResult<Space>> {
  try {
    const { userId } = await auth();
    if (!userId) {
      return { ok: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };
    }

    const client = getKnowledgeServiceClient();
    const headers = await createAuthHeaders();
    const request = new CreateSpaceRequest({
      title: input.title,
      description: input.description,
    });

    const response = await client.createSpace(request, { headers });

    // Revalidate cache tags and paths
    revalidateTag(`spaces-list-${userId}`);
    revalidatePath('/spaces');
    return {
      ok: true,
      data: response,
    };
  } catch (error) {
    console.error('createSpace action error:', error);
    return { ok: false, error: sanitizeError(error) };
  }
}

/**
 * Update an existing space
 */
export async function updateSpace(spaceId: string, input: UpdateSpaceRequest): Promise<ActionResult<Space>> {
  try {
    const { userId } = await auth();
    if (!userId) {
      return { ok: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };
    }

    const client = getKnowledgeServiceClient();
    const headers = await createAuthHeaders();

    const request = new UpdateSpaceRequest({
      id: spaceId,
      title: input.title,
      description: input.description,
      keywords: input.keywords || [],
      icon: input.icon || '',
      accessLevel: input.accessLevel || '',
    });

    const response = await client.updateSpace(request, { headers });

    // Revalidate cache tags and paths
    revalidateTag(`spaces-list-${userId}`);
    revalidateTag(`space-${spaceId}`);
    revalidatePath('/spaces');
    revalidatePath(`/spaces/${spaceId}`);
    return {
      ok: true,
      data: response,
    };
  } catch (error) {
    console.error('updateSpace action error:', error);
    return { ok: false, error: sanitizeError(error) };
  }
}

/**
 * Delete a space
 */
export async function deleteSpace(spaceId: string): Promise<ActionResult<void>> {
  try {
    const { userId } = await auth();
    if (!userId) {
      return { ok: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };
    }

    const client = getKnowledgeServiceClient();
    const headers = await createAuthHeaders();
    const request = new DeleteSpaceRequest({ id: spaceId });

    await client.deleteSpace(request, { headers });

    // Revalidate cache tags and paths
    revalidateTag(`spaces-list-${userId}`);
    revalidatePath('/spaces');
    return { ok: true };
  } catch (error) {
    console.error('deleteSpace action error:', error);
    return { ok: false, error: sanitizeError(error) };
  }
}