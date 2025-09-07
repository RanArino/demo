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
import { ActionResult, safeTimestampToDate } from '@/lib/types';
import { Timestamp } from '@bufbuild/protobuf';
import { createAuthHeaders, sanitizeError } from './utils';



function protoTimestampToISOString(ts: Timestamp | undefined): string {
  if (!ts) return '';
  const date = safeTimestampToDate(ts);
  return date ? date.toISOString() : '';
}

export async function listContentSources(
  spaceId: string,
  status?: 'uploading' | 'uploaded' | 'processing' | 'processed' | 'failed'
): Promise<ActionResult<ContentSource[]>> {
  try {
    const { userId } = await auth();
    if (!userId) return { ok: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };

    const client = getKnowledgeServiceClient();
    const headers = await createAuthHeaders();

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
    return { ok: true, data: response.items };
  } catch (error) {
    console.error('listContentSources action error:', error);
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
    return { ok: true, data: response };
  } catch (error) {
    console.error('createUploadURL action error:', error);
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
    return { ok: true, data: response };
  } catch (error) {
    console.error('confirmUpload action error:', error);
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
    console.error('generateDownloadURL action error:', error);
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

    await client.deleteContentSource(request, { headers });
    return { ok: true };
  } catch (error) {
    console.error('deleteContentSource action error:', error);
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
    return { ok: true, data: response };
  } catch (error) {
    console.error('getContentSource action error:', error);
    return { ok: false, error: sanitizeError(error) };
  }
}