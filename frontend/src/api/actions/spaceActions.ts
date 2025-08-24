'use server';

import { auth } from '@clerk/nextjs/server';
import { revalidatePath } from 'next/cache';
import { getKnowledgeServiceClient } from '../server-client';
import * as grpc from '@grpc/grpc-js';
// knowledge_pb is generated JS (google-protobuf); use require to avoid TS type issues
// eslint-disable-next-line @typescript-eslint/no-var-requires
const knowledge: any = require('../generated/v1/knowledge_pb');
import { 
  Space, 
  SpaceFilters, 
  SearchSpacesResponse, 
  CreateSpaceInput, 
  UpdateSpaceInput,
  ActionResult,
  CreateUploadURLRequest,
  CreateUploadURLResponse,
  ContentSource
} from '@/app/spaces/types/spaces';

// Create gRPC metadata with Clerk JWT
async function createMetadataWithAuth(): Promise<grpc.Metadata> {
  const { getToken } = await auth();
  const token = await getToken({ template: 'ms-user-auth' });
  const md = new grpc.Metadata();
  if (token) md.add('authorization', `Bearer ${token}`);
  return md;
}

// Helper function to convert proto Space to our TypeScript Space type
function timestampToISOString(ts: any): string {
  if (!ts) return '';
  try {
    if (typeof ts.getSeconds === 'function') {
      const seconds = ts.getSeconds();
      const nanos = typeof ts.getNanos === 'function' ? ts.getNanos() : 0;
      const ms = seconds * 1000 + Math.floor((nanos || 0) / 1_000_000);
      return new Date(ms).toISOString();
    }
    if (typeof ts.toDate === 'function') {
      return ts.toDate().toISOString();
    }
  } catch {
    // fall through to empty string
  }
  return '';
}

function protoSpaceToSpace(protoSpace: any): Space {
  return {
    id: protoSpace.getId(),
    userId: protoSpace.getOwnerId?.() || protoSpace.getUserId?.() || '',
    title: protoSpace.getTitle(),
    description: protoSpace.getDescription(),
    icon: protoSpace.getIcon?.() || undefined,
    keywords: protoSpace.getKeywordsList?.() || [],
    createdAt: timestampToISOString(protoSpace.getCreatedAt?.()),
    updatedAt: timestampToISOString(protoSpace.getUpdatedAt?.()),
    contentCount: protoSpace.getStats?.()?.getContentCount?.() || 0,
    userCount: protoSpace.getStats?.()?.getLinkCount?.() || 0,
  };
}

// Helper function to convert proto ContentSource to our TypeScript ContentSource type
function protoContentSourceToContentSource(protoSource: any): ContentSource {
  return {
    id: protoSource.getId(),
    spaceId: protoSource.getSpaceId(),
    name: protoSource.getName(),
    type: protoSource.getType(),
    url: protoSource.getUrl() || undefined,
    size: protoSource.getSize() || undefined,
    mimeType: protoSource.getMimeType() || undefined,
    status: protoSource.getStatus(),
    createdAt: timestampToISOString(protoSource.getCreatedAt?.()),
    updatedAt: timestampToISOString(protoSource.getUpdatedAt?.()),
  };
}

/**
 * Search/list spaces with filters
 * Used for both initial load and filtered searches
 * Note: Using ListSpaces RPC as per the proto definition
 */
export async function searchSpaces(filters?: SpaceFilters): Promise<ActionResult<SearchSpacesResponse>> {
  try {
    const { userId } = await auth();
    if (!userId) {
      return {
        ok: false,
        error: { code: 'UNAUTHORIZED', message: 'User not authenticated' }
      };
    }

    const client = getKnowledgeServiceClient();
    const request = new knowledge.ListSpacesRequest();
    
    // Force-clear owner_id to avoid INVALID_ARGUMENT from backend (expects UUID)
    if (request.setOwnerId) {
      request.setOwnerId('');
    }

    // Search query
    if (filters?.q) {
      request.setQ(filters.q);
    }

    // Keywords filter
    if (filters?.keywords && filters.keywords.length > 0) {
      request.setKeywordsList(filters.keywords);
    }
    
    // Pagination (proto supports page_size + page_token)
    if (filters?.pageSize) {
      const pagination = new knowledge.Pagination();
      pagination.setPageSize(filters.pageSize);
      request.setPage(pagination);
    }

    const metadata = await createMetadataWithAuth();
    return new Promise((resolve) => {
      client.listSpaces(request, metadata, (error: any, response: any) => {
        if (error) {
          console.error('ListSpaces error:', error);
          resolve({
            ok: false,
            error: { code: error.code || 'UNKNOWN', message: error.message }
          });
        } else {
          const spaces = response.getItemsList().map(protoSpaceToSpace);
          const nextToken = response.getNextPageToken?.() || '';
          resolve({
            ok: true,
            data: {
              spaces,
              totalCount: spaces.length,
              page: filters?.page || 1,
              pageSize: filters?.pageSize || spaces.length,
            }
          });
        }
      });
    });
  } catch (error) {
    console.error('ListSpaces exception:', error);
    return {
      ok: false,
      error: { code: 'INTERNAL', message: 'An unexpected error occurred' }
    };
  }
}

/**
 * Get a single space by ID
 */
export async function getSpace(spaceId: string): Promise<ActionResult<Space>> {
  try {
    const { userId } = await auth();
    if (!userId) {
      return {
        ok: false,
        error: { code: 'UNAUTHORIZED', message: 'User not authenticated' }
      };
    }

    const client = getKnowledgeServiceClient();
    const request = new knowledge.GetSpaceRequest();
    request.setId(spaceId);

    const metadata = await createMetadataWithAuth();
    return new Promise((resolve) => {
      client.getSpace(request, metadata, (error: any, response: any) => {
        if (error) {
          console.error('GetSpace error:', error);
          resolve({
            ok: false,
            error: { code: error.code || 'UNKNOWN', message: error.message }
          });
        } else {
          // Verify the space belongs to the current user
          if (response.getOwnerId?.() !== userId) {
            resolve({
              ok: false,
              error: { code: 'FORBIDDEN', message: 'Access denied' }
            });
          } else {
            resolve({
              ok: true,
              data: protoSpaceToSpace(response)
            });
          }
        }
      });
    });
  } catch (error) {
    console.error('GetSpace exception:', error);
    return {
      ok: false,
      error: { code: 'INTERNAL', message: 'An unexpected error occurred' }
    };
  }
}

/**
 * Create a new space
 */
export async function createSpace(input: CreateSpaceInput): Promise<ActionResult<Space>> {
  try {
    const { userId, getToken } = await auth();
    if (!userId) {
      return {
        ok: false,
        error: { code: 'UNAUTHORIZED', message: 'User not authenticated' }
      };
    }

    const client = getKnowledgeServiceClient();
    const request = new knowledge.CreateSpaceRequest();
    
    // CreateSpace only accepts title and description per proto
    request.setTitle(input.title);
    request.setDescription(input.description);

    // Prepare metadata with auth token and owner id (internal UUID)
    const metadata = await createMetadataWithAuth();
    // Do not resolve or inject owner id on the client; backend will resolve from JWT.

    return new Promise((resolve) => {
      client.createSpace(request, metadata, (error: any, response: any) => {
        if (error) {
          console.error('CreateSpace error:', error);
          resolve({
            ok: false,
            error: { code: error.code || 'UNKNOWN', message: error.message }
          });
        } else {
          revalidatePath('/spaces');
          resolve({
            ok: true,
            data: protoSpaceToSpace(response)
          });
        }
      });
    });
  } catch (error) {
    console.error('CreateSpace exception:', error);
    return {
      ok: false,
      error: { code: 'INTERNAL', message: 'An unexpected error occurred' }
    };
  }
}

/**
 * Update an existing space
 */
export async function updateSpace(spaceId: string, input: UpdateSpaceInput): Promise<ActionResult<Space>> {
  try {
    const { userId } = await auth();
    if (!userId) {
      return {
        ok: false,
        error: { code: 'UNAUTHORIZED', message: 'User not authenticated' }
      };
    }

    // First verify the space belongs to the user
    const spaceResult = await getSpace(spaceId);
    if (!spaceResult.ok || !spaceResult.data) {
      return spaceResult;
    }

    const client = getKnowledgeServiceClient();
    const request = new knowledge.UpdateSpaceRequest();
    
    request.setId(spaceId);
    
    if (input.title !== undefined) {
      request.setTitle(input.title);
    }
    if (input.description !== undefined) {
      request.setDescription(input.description);
    }
    if (input.keywords !== undefined) {
      request.setKeywordsList(input.keywords);
    }
    if (input.icon !== undefined) {
      request.setIcon(input.icon);
    }

    const metadata = await createMetadataWithAuth();
    return new Promise((resolve) => {
      client.updateSpace(request, metadata, (error: any, response: any) => {
        if (error) {
          console.error('UpdateSpace error:', error);
          resolve({
            ok: false,
            error: { code: error.code || 'UNKNOWN', message: error.message }
          });
        } else {
          revalidatePath('/spaces');
          revalidatePath(`/spaces/${spaceId}`);
          resolve({
            ok: true,
            data: protoSpaceToSpace(response)
          });
        }
      });
    });
  } catch (error) {
    console.error('UpdateSpace exception:', error);
    return {
      ok: false,
      error: { code: 'INTERNAL', message: 'An unexpected error occurred' }
    };
  }
}

/**
 * Delete a space
 */
export async function deleteSpace(spaceId: string): Promise<ActionResult<void>> {
  try {
    const { userId } = await auth();
    if (!userId) {
      return {
        ok: false,
        error: { code: 'UNAUTHORIZED', message: 'User not authenticated' }
      };
    }

    // First verify the space belongs to the user
    const spaceResult = await getSpace(spaceId);
    if (!spaceResult.ok || !spaceResult.data) {
      return { ok: false, error: spaceResult.error };
    }

    const client = getKnowledgeServiceClient();
    const request = new knowledge.DeleteSpaceRequest();
    request.setId(spaceId);

    const metadata = await createMetadataWithAuth();
    return new Promise((resolve) => {
      client.deleteSpace(request, metadata, (error: any) => {
        if (error) {
          console.error('DeleteSpace error:', error);
          resolve({
            ok: false,
            error: { code: error.code || 'UNKNOWN', message: error.message }
          });
        } else {
          revalidatePath('/spaces');
          resolve({ ok: true });
        }
      });
    });
  } catch (error) {
    console.error('DeleteSpace exception:', error);
    return {
      ok: false,
      error: { code: 'INTERNAL', message: 'An unexpected error occurred' }
    };
  }
}

/**
 * List content sources for a space
 */
export async function listContentSources(spaceId: string): Promise<ActionResult<ContentSource[]>> {
  try {
    const { userId } = await auth();
    if (!userId) {
      return {
        ok: false,
        error: { code: 'UNAUTHORIZED', message: 'User not authenticated' }
      };
    }

    // First verify the space belongs to the user
    const spaceResult = await getSpace(spaceId);
    if (!spaceResult.ok || !spaceResult.data) {
      return { ok: false, error: spaceResult.error };
    }

    const client = getKnowledgeServiceClient();
    const request = new knowledge.ListContentSourcesRequest();
    request.setSpaceId(spaceId);

    const metadata = await createMetadataWithAuth();
    return new Promise((resolve) => {
      client.listContentSources(request, metadata, (error: any, response: any) => {
        if (error) {
          console.error('ListContentSources error:', error);
          resolve({
            ok: false,
            error: { code: error.code || 'UNKNOWN', message: error.message }
          });
        } else {
          const sources = response.getContentSourcesList().map(protoContentSourceToContentSource);
          resolve({
            ok: true,
            data: sources
          });
        }
      });
    });
  } catch (error) {
    console.error('ListContentSources exception:', error);
    return {
      ok: false,
      error: { code: 'INTERNAL', message: 'An unexpected error occurred' }
    };
  }
}

/**
 * Create upload URL for file upload
 */
export async function createUploadURL(input: CreateUploadURLRequest): Promise<ActionResult<CreateUploadURLResponse>> {
  try {
    const { userId } = await auth();
    if (!userId) {
      return {
        ok: false,
        error: { code: 'UNAUTHORIZED', message: 'User not authenticated' }
      };
    }

    // First verify the space belongs to the user
    const spaceResult = await getSpace(input.spaceId);
    if (!spaceResult.ok || !spaceResult.data) {
      return { ok: false, error: spaceResult.error };
    }

    const client = getKnowledgeServiceClient();
    const request = new knowledge.CreateUploadURLRequest();
    request.setSpaceId(input.spaceId);
    request.setFilename(input.filename);
    request.setMimeType(input.mimeType);
    request.setSizeBytes(input.sizeBytes);

    const metadata = await createMetadataWithAuth();
    return new Promise((resolve) => {
      client.createUploadURL(request, metadata, (error: any, response: any) => {
        if (error) {
          console.error('CreateUploadURL error:', error);
          resolve({
            ok: false,
            error: { code: error.code || 'UNKNOWN', message: error.message }
          });
        } else {
          resolve({
            ok: true,
            data: {
              uploadUrl: response.getUploadUrl(),
              contentSourceId: response.getContentSourceId(),
              expiresAt: timestampToISOString(response.getExpiresAt?.()),
            }
          });
        }
      });
    });
  } catch (error) {
    console.error('CreateUploadURL exception:', error);
    return {
      ok: false,
      error: { code: 'INTERNAL', message: 'An unexpected error occurred' }
    };
  }
}

/**
 * Confirm upload after file is uploaded
 */
export async function confirmUpload(contentSourceId: string): Promise<ActionResult<ContentSource>> {
  try {
    const { userId } = await auth();
    if (!userId) {
      return {
        ok: false,
        error: { code: 'UNAUTHORIZED', message: 'User not authenticated' }
      };
    }

    const client = getKnowledgeServiceClient();
    const request = new knowledge.ConfirmUploadRequest();
    request.setContentSourceId(contentSourceId);

    const metadata = await createMetadataWithAuth();
    return new Promise((resolve) => {
      client.confirmUpload(request, metadata, (error: any, response: any) => {
        if (error) {
          console.error('ConfirmUpload error:', error);
          resolve({
            ok: false,
            error: { code: error.code || 'UNKNOWN', message: error.message }
          });
        } else {
          resolve({
            ok: true,
            data: protoContentSourceToContentSource(response)
          });
        }
      });
    });
  } catch (error) {
    console.error('ConfirmUpload exception:', error);
    return {
      ok: false,
      error: { code: 'INTERNAL', message: 'An unexpected error occurred' }
    };
  }
}