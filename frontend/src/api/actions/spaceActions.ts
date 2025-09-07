'use server';

import { auth } from '@clerk/nextjs/server';
import { revalidatePath } from 'next/cache';
import { getKnowledgeServiceClient } from '../server-client';
import {
  ListSpacesRequest,
  GetSpaceRequest,
  CreateSpaceRequest,
  UpdateSpaceRequest,
  DeleteSpaceRequest,
  Space as ProtoSpace,
  Pagination,
} from '../generated/v1/knowledge_pb';
import {
  Space,
  SpaceFilters,
  SearchSpacesResponse,
  CreateSpaceInput,
  UpdateSpaceInput,
  ActionResult,
} from '@/app/spaces/types/spaces';
import { AccessLevel } from '@/app/spaces/types/shared';
import { ConnectError } from '@bufbuild/connect';
import { Timestamp } from '@bufbuild/protobuf';

/**
 * Helper function to create headers with JWT token
 */
async function createAuthHeaders(): Promise<Headers> {
  const { getToken } = await auth();
  const token = await getToken({ template: 'ms-user-auth' });

  const headers = new Headers();
  if (token) {
    headers.append('authorization', `Bearer ${token}`);
  }

  return headers;
}

/**
 * Helper function to sanitize error messages for security
 */
function sanitizeError(error: unknown): { code: string; message: string } {
  if (error instanceof ConnectError) {
    return { code: error.code.toString(), message: error.message };
  }

  const isDevelopment = process.env.NODE_ENV === 'development';
  const message = error instanceof Error ? error.message : 'An unexpected error occurred';

  return {
    code: 'INTERNAL',
    message: isDevelopment ? message : 'An unexpected error occurred. Please try again.',
  };
}

function protoTimestampToISOString(ts: Timestamp | undefined): string {
  if (!ts) return '';
  return ts.toDate().toISOString();
}

// Helper function to convert proto Space to our TypeScript Space type
function protoSpaceToSpace(protoSpace: ProtoSpace): Space {
  return {
    id: protoSpace.id,
    userId: protoSpace.ownerId,
    title: protoSpace.title,
    description: protoSpace.description,
    keywords: [], // keywords are not part of the proto Space message
    accessLevel: AccessLevel.PRIVATE, // accessLevel is not part of the proto Space message
    createdAt: protoTimestampToISOString(protoSpace.createdAt),
    updatedAt: protoTimestampToISOString(protoSpace.updatedAt),
    lastUpdatedAt: protoTimestampToISOString(protoSpace.updatedAt),
    documentCount: Number(protoSpace.stats?.contentCount ?? 0),
    totalSizeBytes: 0, // totalSizeBytes is not part of the proto Space message
    contentCount: Number(protoSpace.stats?.contentCount ?? 0),
    userCount: Number(protoSpace.stats?.linkCount ?? 0),
  };
}

/**
 * Search/list spaces with filters
 */
export async function searchSpaces(filters?: SpaceFilters): Promise<ActionResult<SearchSpacesResponse>> {
  try {
    const { userId } = await auth();
    if (!userId) {
      return { ok: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };
    }

    const client = getKnowledgeServiceClient();
    const headers = await createAuthHeaders();

    const request: Partial<ListSpacesRequest> = {};

    if (filters?.q) {
      request.q = filters.q;
    }
    if (filters?.keywords && filters.keywords.length > 0) {
      request.keywords = filters.keywords;
    }
    if (filters?.pageSize) {
      request.page = new Pagination({ pageSize: filters.pageSize });
    }

    const response = await client.listSpaces(request, { headers });

    const spaces = response.items.map(protoSpaceToSpace);
    return {
      ok: true,
      data: {
        spaces,
        totalCount: spaces.length,
        page: filters?.page || 1,
        pageSize: filters?.pageSize || spaces.length,
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
      data: protoSpaceToSpace(response),
    };
  } catch (error) {
    console.error('getSpace action error:', error);
    return { ok: false, error: sanitizeError(error) };
  }
}

/**
 * Create a new space
 */
export async function createSpace(input: CreateSpaceInput): Promise<ActionResult<Space>> {
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
      data: protoSpaceToSpace(response),
    };
  } catch (error) {
    console.error('createSpace action error:', error);
    return { ok: false, error: sanitizeError(error) };
  }
}

/**
 * Update an existing space
 */
export async function updateSpace(spaceId: string, input: UpdateSpaceInput): Promise<ActionResult<Space>> {
  try {
    const { userId } = await auth();
    if (!userId) {
      return { ok: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };
    }

    const client = getKnowledgeServiceClient();
    const headers = await createAuthHeaders();

    const spaceToUpdate: Partial<ProtoSpace> = {};
    if (input.title !== undefined) {
      spaceToUpdate.title = input.title;
    }
    if (input.description !== undefined) {
      spaceToUpdate.description = input.description;
    }

    const request = new UpdateSpaceRequest({
      id: spaceId,
      space: new ProtoSpace(spaceToUpdate),
    });

    const response = await client.updateSpace(request, { headers });

    revalidatePath('/spaces');
    revalidatePath(`/spaces/${spaceId}`);
    return {
      ok: true,
      data: protoSpaceToSpace(response),
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