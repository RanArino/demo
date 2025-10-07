'use server';

import { auth } from '@clerk/nextjs/server';
import { getKnowledgeServiceClient } from '../server-client';
import {
  ListContentSourcesRequest,
  CreateUploadURLRequest,
  CreateUploadURLResponse,
  ConfirmUploadRequest,
  GenerateDownloadURLRequest,
  DeleteContentSourceRequest,
  GetContentSourceRequest,
  ContentSource,
  ContentStatus,
  DownloadObjectKind,
} from '../generated/v1/knowledge_pb';
import { ActionResult } from '@/lib/types';
import { createAuthHeaders, sanitizeError, sanitizeProtobufForJson, isUnauthorizedError, logAuthFailure } from './utils';
import { cache } from 'react';
import { unstable_cache, revalidateTag, revalidatePath } from 'next/cache';

// Core fetcher for listContentSources
async function listContentSourcesCore(
  _userId: string,
  headers: Headers,
  spaceId: string,
  status?: 'uploading' | 'uploaded' | 'processing' | 'processed' | 'failed'
): Promise<ActionResult<ContentSource[]>> {
  try {
    const client = getKnowledgeServiceClient();

    // Properly construct the request object instead of using Partial
    const requestData: { spaceId: string; status?: ContentStatus } = { spaceId };
    if (status) {
      const statusMap = {
        uploading: ContentStatus.UPLOADING,
        uploaded: ContentStatus.UPLOADED,
        processing: ContentStatus.PROCESSING,
        processed: ContentStatus.PROCESSED,
        failed: ContentStatus.FAILED,
      };
      if (status in statusMap) {
        requestData.status = statusMap[status as keyof typeof statusMap];
      }
    }

    const request = new ListContentSourcesRequest(requestData);
    const response = await client.listContentSources(request, { headers });
    const { normalizeContentSourceForClient } = await import('./utils');
    const sanitizedItems = response.items
      .map(item => sanitizeProtobufForJson(item))
      .map(item => normalizeContentSourceForClient(item));
    return { ok: true, data: sanitizedItems };
  } catch (error) {
    if (isUnauthorizedError(error)) {
      logAuthFailure('listContentSourcesCore', error);
    } else {
      console.error('listContentSourcesCore error:', error);
    }
    return { ok: false, error: sanitizeError(error) };
  }
}

// Per-request memoization
const listContentSourcesMemoized = cache(async (_userId: string, headers: Headers, spaceId: string, status?: 'uploading' | 'uploaded' | 'processing' | 'processed' | 'failed') => {
  return listContentSourcesCore(_userId, headers, spaceId, status);
});

// Cross-request cache with tag and short TTL
const listContentSourcesCached = (_userId: string, headers: Headers, spaceId: string, status?: 'uploading' | 'uploaded' | 'processing' | 'processed' | 'failed') => {
  const statusKey = status ?? 'any';
  const key = [`listContentSources-${spaceId}-${statusKey}`];
  return unstable_cache(
    async () => listContentSourcesMemoized(_userId, headers, spaceId, status),
    key,
    {
      tags: [`content-sources-${spaceId}`],
      revalidate: 60,
    }
  )();
};

export async function listContentSources(
  spaceId: string,
  status?: 'uploading' | 'uploaded' | 'processing' | 'processed' | 'failed'
): Promise<ActionResult<ContentSource[]>> {
  try {
    const { userId } = await auth();
    if (!userId) return { ok: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };

    const headers = await createAuthHeaders();
    return await listContentSourcesCached(userId, headers, spaceId, status);
  } catch (error) {
    if (isUnauthorizedError(error)) {
      logAuthFailure('listContentSources', error);
    } else {
      console.error('listContentSources action error:', error);
    }
    return { ok: false, error: sanitizeError(error) };
  }
}

// Uncached variant for short-interval polling to avoid 60s cache staleness
export async function listContentSourcesUncached(
  spaceId: string,
  status?: 'uploading' | 'uploaded' | 'processing' | 'processed' | 'failed'
): Promise<ActionResult<ContentSource[]>> {
  try {
    const { userId } = await auth();
    if (!userId) return { ok: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };

    const headers = await createAuthHeaders();
    const client = getKnowledgeServiceClient();

    // Properly construct the request object instead of using Partial
    const requestData: { spaceId: string; status?: ContentStatus } = { spaceId };
    if (status) {
      const statusMap = {
        uploading: ContentStatus.UPLOADING,
        uploaded: ContentStatus.UPLOADED,
        processing: ContentStatus.PROCESSING,
        processed: ContentStatus.PROCESSED,
        failed: ContentStatus.FAILED,
      };
      if (status in statusMap) {
        requestData.status = statusMap[status as keyof typeof statusMap];
      }
    }

    const request = new ListContentSourcesRequest(requestData);
    const response = await client.listContentSources(request, { headers });
    const { normalizeContentSourceForClient } = await import('./utils');
    const sanitizedItems = response.items
      .map(item => sanitizeProtobufForJson(item))
      .map(item => normalizeContentSourceForClient(item));
    return { ok: true, data: sanitizedItems };
  } catch (error) {
    if (isUnauthorizedError(error)) {
      logAuthFailure('listContentSourcesUncached', error);
    } else {
      console.error('listContentSourcesUncached error:', error);
    }
    return { ok: false, error: sanitizeError(error) };
  }
}

export async function createUploadURL(
  request: CreateUploadURLRequest
): Promise<ActionResult<CreateUploadURLResponse>> {
  try {
    const { userId } = await auth();
    if (!userId) return { ok: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };

    const client = getKnowledgeServiceClient();
    const headers = await createAuthHeaders();

    // Use utility to rebuild/normalize request types (sizeBytes, enum fields)
    const { rebuildCreateUploadURLRequest } = await import('./utils');
    const rebuilt = rebuildCreateUploadURLRequest(request as CreateUploadURLRequest);
    const response = await client.createUploadURL(rebuilt, { headers });
    // Sanitize response (contains ContentSource with size_bytes BigInt field and Timestamp objects)
    const sanitizedResponse = sanitizeProtobufForJson(response);
    return { ok: true, data: sanitizedResponse };
  } catch (error) {
    if (isUnauthorizedError(error)) {
      logAuthFailure('createUploadURL', error);
    } else {
      console.error('createUploadURL action error:', error);
    }
    return { ok: false, error: sanitizeError(error) };
  }
}

export async function confirmUpload(
  contentSourceId: string,
  blobHash: string
): Promise<ActionResult<ContentSource>> {
  try {
    const { userId } = await auth();
    if (!userId) return { ok: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };

    const client = getKnowledgeServiceClient();
    const headers = await createAuthHeaders();
    const request = new ConfirmUploadRequest({ contentSourceId, blobHash });

    const response = await client.confirmUpload(request, { headers });
    // Sanitize ContentSource object (handles size_bytes BigInt field and Timestamp objects)
    const sanitizedResponse = sanitizeProtobufForJson(response);
    // Revalidate content lists for this space
    if (sanitizedResponse && (sanitizedResponse as any).spaceId) {
      const sid = (sanitizedResponse as any).spaceId as string;
      // Refresh content lists within the space (documents panel)
      revalidateTag(`content-sources-${sid}`);
      // Refresh the specific space cache (e.g., stats/counts used in detail views)
      revalidateTag(`space-${sid}`);
      // Refresh the spaces list (gallery/list page shows docs count in cards)
      if (userId) {
        revalidateTag(`spaces-list-${userId}`);
      }
      // Trigger page-level revalidation for both the detail and list pages
      revalidatePath(`/spaces/${sid}`);
      revalidatePath('/spaces');
    }
    return { ok: true, data: sanitizedResponse };
  } catch (error) {
    if (isUnauthorizedError(error)) {
      logAuthFailure('confirmUpload', error);
    } else {
      console.error('confirmUpload action error:', error);
    }
    return { ok: false, error: sanitizeError(error) };
  }
}

export async function generateDownloadURL(
  contentSourceId: string,
  kind: 'original' | 'processed' = 'original',
  expiresSeconds?: number
): Promise<ActionResult<{ url: string; expiresAt?: string; objectKey: string }>> {
  try {
    const { userId } = await auth();
    if (!userId) return { ok: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };

    const client = getKnowledgeServiceClient();
    const headers = await createAuthHeaders();
    const request = new GenerateDownloadURLRequest({
      contentSourceId,
      ...(expiresSeconds !== undefined && { expiresSeconds }),
      objectKind:
        kind === 'processed'
          ? DownloadObjectKind.PROCESSED
          : DownloadObjectKind.ORIGINAL,
    });

    const response = await client.generateDownloadURL(request, { headers });
    return {
      ok: true,
      data: {
        url: response.url,
        expiresAt: response.expiresAt ? sanitizeProtobufForJson(response.expiresAt) as unknown as string : undefined,
        objectKey: response.objectKey,
      },
    };
  } catch (error) {
    if (isUnauthorizedError(error)) {
      logAuthFailure('generateDownloadURL', error);
    } else {
      console.error('generateDownloadURL action error:', error);
    }
    return { ok: false, error: sanitizeError(error) };
  }
}

export async function deleteContentSource(contentSourceId: string): Promise<ActionResult<void>> {
  try {
    const { userId } = await auth();
    if (!userId) return { ok: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };

    const client = getKnowledgeServiceClient();
    const headers = await createAuthHeaders();
    const request = new DeleteContentSourceRequest({ id: contentSourceId });

    // Ideally we'd look up the spaceId before delete; assume client can pass or backend returns it
    // For now, trigger broad revalidation on all spaces pages
    await client.deleteContentSource(request, { headers });
    // We cannot infer spaceId directly; the UI calls this with context, so let the page refresh
    // If we had spaceId, we would call revalidateTag(`content-sources-${spaceId}`) and revalidatePath(`/spaces/${spaceId}`)
    return { ok: true };
  } catch (error) {
    if (isUnauthorizedError(error)) {
      logAuthFailure('deleteContentSource', error);
    } else {
      console.error('deleteContentSource action error:', error);
    }
    return { ok: false, error: sanitizeError(error) };
  }
}

export async function getContentSource(contentSourceId: string): Promise<ActionResult<ContentSource>> {
  try {
    const { userId } = await auth();
    if (!userId) return { ok: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };

    const client = getKnowledgeServiceClient();
    const headers = await createAuthHeaders();
    const request = new GetContentSourceRequest({ id: contentSourceId });

    const response = await client.getContentSource(request, { headers });
    // Sanitize ContentSource object (handles size_bytes BigInt field and Timestamp objects)
    const sanitizedResponse = sanitizeProtobufForJson(response);
    return { ok: true, data: sanitizedResponse };
  } catch (error) {
    if (isUnauthorizedError(error)) {
      logAuthFailure('getContentSource', error);
    } else {
      console.error('getContentSource action error:', error);
    }
    return { ok: false, error: sanitizeError(error) };
  }
}