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
import { Timestamp } from '@bufbuild/protobuf';
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

    const request: Partial<ListContentSourcesRequest> = { spaceId };
    if (status) {
      const statusMap = {
        uploading: ContentStatus.UPLOADING,
        uploaded: ContentStatus.UPLOADED,
        processing: ContentStatus.PROCESSING,
        processed: ContentStatus.PROCESSED,
        failed: ContentStatus.FAILED,
      };
      if (status in statusMap) {
        request.status = statusMap[status as keyof typeof statusMap];
      }
    }

    const response = await client.listContentSources(request, { headers });
    const sanitizedItems = response.items.map(item => sanitizeProtobufForJson(item));
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

export async function createUploadURL(
  request: CreateUploadURLRequest
): Promise<ActionResult<CreateUploadURLResponse>> {
  try {
    const { userId } = await auth();
    if (!userId) return { ok: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };

    const client = getKnowledgeServiceClient();
    const headers = await createAuthHeaders();

    const response = await client.createUploadURL(request, { headers });
    // Sanitize response (contains ContentSource with size_bytes BigInt field)
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
    // Sanitize ContentSource object (handles size_bytes BigInt field)
    const sanitizedResponse = sanitizeProtobufForJson(response);
    // Revalidate content lists for this space
    if (sanitizedResponse && (sanitizedResponse as any).spaceId) {
      const sid = (sanitizedResponse as any).spaceId as string;
      revalidateTag(`content-sources-${sid}`);
      revalidatePath(`/spaces/${sid}`);
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
): Promise<ActionResult<{ url: string; expiresAt: Timestamp | undefined; objectKey: string }>> {
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
        expiresAt: response.expiresAt,
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
    // Sanitize ContentSource object (handles size_bytes BigInt field)
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