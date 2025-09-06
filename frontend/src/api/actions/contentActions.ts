'use server';

import { auth } from '@clerk/nextjs/server';
import { getKnowledgeServiceClient } from '../server-client';
import {
  ListContentSourcesRequest,
  CreateUploadURLRequest,
  ConfirmUploadRequest,
  GenerateDownloadURLRequest,
  DeleteContentSourceRequest,
  GetContentSourceRequest,
  ContentSource as ProtoContentSource,
  ContentStatus,
  DownloadObjectKind,
} from '../generated/v1/knowledge_pb';
import { ActionResult } from '@/app/spaces/types/shared';
import {
  ContentSource,
  ContentSourceStatus,
  ContentSourceType,
  CreateUploadURLRequest as AppCreateUploadURLRequest,
  CreateUploadURLResponse,
} from '@/app/spaces/types/content';
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

function protoContentSourceToContentSource(protoSource: ProtoContentSource): ContentSource {
  const statusMap: Record<ContentStatus, ContentSourceStatus> = {
    [ContentStatus.CONTENT_STATUS_UNSPECIFIED]: ContentSourceStatus.PENDING,
    [ContentStatus.UPLOADING]: ContentSourceStatus.PROCESSING,
    [ContentStatus.UPLOADED]: ContentSourceStatus.PROCESSING,
    [ContentStatus.PROCESSING]: ContentSourceStatus.PROCESSING,
    [ContentStatus.PROCESSED]: ContentSourceStatus.COMPLETED,
    [ContentStatus.FAILED]: ContentSourceStatus.FAILED,
  };
  const processingStatus = statusMap[protoSource.status] || ContentSourceStatus.PENDING;

  return {
    id: protoSource.id,
    spaceId: protoSource.spaceId,
    title: protoSource.title || undefined,
    mimeType: protoSource.mimeType,
    sizeBytes: Number(protoSource.sizeBytes),
    processingStatus,
    sourceType: ContentSourceType.FILE, // This might need to be mapped from the proto if available
    createdAt: protoTimestampToISOString(protoSource.createdAt),
    updatedAt: protoTimestampToISOString(protoSource.updatedAt),
    contentSummary: protoSource.contentSummary || undefined,
  };
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
    const sources = response.items.map(protoContentSourceToContentSource);
    return { ok: true, data: sources };
  } catch (error) {
    console.error('listContentSources action error:', error);
    return { ok: false, error: sanitizeError(error) };
  }
}

export async function createUploadURL(
  input: AppCreateUploadURLRequest
): Promise<ActionResult<CreateUploadURLResponse>> {
  try {
    const { userId } = await auth();
    if (!userId) return { ok: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };

    const client = getKnowledgeServiceClient();
    const headers = await createAuthHeaders();

    const request = new CreateUploadURLRequest({
      spaceId: input.spaceId,
      filename: input.filename,
      mimeType: input.mimeType,
      sizeBytes: BigInt(input.sizeBytes),
      objectKind:
        input.objectKind === 'processed'
          ? DownloadObjectKind.PROCESSED
          : DownloadObjectKind.ORIGINAL,
    });

    const response = await client.createUploadURL(request, { headers });
    return {
      ok: true,
      data: {
        uploadUrl: response.uploadUrl,
        contentSourceId: response.contentSource?.id || '',
        expiresAt: protoTimestampToISOString(response.expiresAt),
        objectKey: response.objectKey,
      },
    };
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
    return { ok: true, data: protoContentSourceToContentSource(response) };
  } catch (error) {
    console.error('confirmUpload action error:', error);
    return { ok: false, error: sanitizeError(error) };
  }
}

export async function generateDownloadURL(
  contentSourceId: string,
  kind: 'original' | 'processed' = 'original',
  expiresSeconds?: number
): Promise<ActionResult<{ url: string; expiresAt: string }>> {
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
        expiresAt: protoTimestampToISOString(response.expiresAt),
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
    return { ok: true, data: protoContentSourceToContentSource(response) };
  } catch (error) {
    console.error('getContentSource action error:', error);
    return { ok: false, error: sanitizeError(error) };
  }
}