'use server';

import { auth } from '@clerk/nextjs/server';
import { revalidatePath } from 'next/cache';
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
import { ActionResult, safeTimestampToDate } from '@/lib/types'; 
import { Timestamp } from '@bufbuild/protobuf';
import { createAuthHeaders, sanitizeError } from './utils';


function protoTimestampToISOString(ts: Timestamp | undefined): string {
  if (!ts) return '';
  const date = safeTimestampToDate(ts);
  return date ? date.toISOString() : '';
}

/**
 * Search/list spaces with filters
 */
export async function searchSpaces(filters?: SpaceFilters): Promise<ActionResult<{spaces: Space[], totalCount: bigint, page: number, pageSize: number}>> {
  try {
    const { userId } = await auth();
    if (!userId) {
      return { ok: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };
    }

    const client = getKnowledgeServiceClient();
    const headers = await createAuthHeaders();

    // Create a proper ListSpacesRequest from SpaceFilters
    const request = new ListSpacesRequest({
      q: filters?.q || '',
      keywords: filters?.keywords || [],
      // Set ownerId to current user to filter spaces by owner
      ownerId: userId,
    });

    const response = await client.listSpaces(request, { headers });

    const spaces = response.items;
    return {
      ok: true,
      data: {
        spaces,
        totalCount: BigInt(spaces.length),
        page: 1, // Default page since SpaceFilters doesn't have page property
        pageSize: spaces.length, // Default pageSize since SpaceFilters doesn't have pageSize property
      },
    };
  } catch (error) {
    console.error('searchSpaces action error:', error);
    return { ok: false, error: sanitizeError(error) };
  }
}

/**
 * Get a single space by ID
 */
export async function getSpace(spaceId: string): Promise<ActionResult<Space>> {
  try {
    const { userId } = await auth();
    if (!userId) {
      return { ok: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };
    }

    const client = getKnowledgeServiceClient();
    const headers = await createAuthHeaders();
    const request = new GetSpaceRequest({ id: spaceId });

    const response = await client.getSpace(request, { headers });

    return {
      ok: true,
      data: response,
    };
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
      title: input.title,
      description: input.description,
      keywords: input.keywords || [],
      icon: input.icon || '',
    });

    const response = await client.updateSpace(request, { headers });

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

    revalidatePath('/spaces');
    return { ok: true };
  } catch (error) {
    console.error('deleteSpace action error:', error);
    return { ok: false, error: sanitizeError(error) };
  }
}