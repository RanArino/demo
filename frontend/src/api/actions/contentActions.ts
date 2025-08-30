'use server';

import { auth } from '@clerk/nextjs/server';
import * as grpc from '@grpc/grpc-js';
import { getKnowledgeServiceClient } from '../server-client';
const knowledge: any = require('../generated/v1/knowledge_pb');
import { ActionResult } from '@/app/spaces/types/shared';
import { ContentSource, ContentSourceStatus, ContentSourceType } from '@/app/spaces/types/content';

async function createMetadataWithAuth(): Promise<grpc.Metadata> {
  const { getToken } = await auth();
  const token = await getToken({ template: 'ms-user-auth' });
  const md = new grpc.Metadata();
  if (token) md.add('authorization', `Bearer ${token}`);
  return md;
}

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
  } catch {}
  return '';
}

function protoContentSourceToContentSource(protoSource: any): ContentSource {
  const statusMap: Record<number, ContentSourceStatus> = {
    0: ContentSourceStatus.PENDING,
    1: ContentSourceStatus.PROCESSING,
    2: ContentSourceStatus.COMPLETED,
    3: ContentSourceStatus.FAILED,
  };
  const status = protoSource.getStatus();
  const processingStatus = statusMap[status] || ContentSourceStatus.PENDING;
  return {
    id: protoSource.getId(),
    spaceId: protoSource.getSpaceId(),
    title: protoSource.getTitle() || undefined,
    mimeType: protoSource.getMimeType() || '',
    sizeBytes: protoSource.getSizeBytes() || 0,
    processingStatus,
    sourceType: ContentSourceType.FILE,
    createdAt: timestampToISOString(protoSource.getCreatedAt?.()),
    updatedAt: timestampToISOString(protoSource.getUpdatedAt?.()),
    contentSummary: protoSource.getContentSummary() || undefined,
  };
}

export interface CreateUploadURLRequest {
  spaceId: string;
  filename: string;
  mimeType: string;
  sizeBytes: number;
}
export interface CreateUploadURLResponse {
  uploadUrl: string;
  contentSourceId: string;
  expiresAt: string;
}

export async function listContentSources(spaceId: string): Promise<ActionResult<ContentSource[]>> {
  try {
    const { userId } = await auth();
    if (!userId) return { ok: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };
    const client = getKnowledgeServiceClient();
    const request = new knowledge.ListContentSourcesRequest();
    request.setSpaceId(spaceId);
    const metadata = await createMetadataWithAuth();
    return new Promise((resolve) => {
      client.listContentSources(request, metadata, (error: any, response: any) => {
        if (error) {
          resolve({ ok: false, error: { code: error.code || 'UNKNOWN', message: error.message } });
        } else {
          const sources = response.getItemsList().map(protoContentSourceToContentSource);
          resolve({ ok: true, data: sources });
        }
      });
    });
  } catch {
    return { ok: false, error: { code: 'INTERNAL', message: 'An unexpected error occurred' } };
  }
}

// Create upload URL for ORIGINAL object (first step)
export async function createUploadURL(input: CreateUploadURLRequest): Promise<ActionResult<CreateUploadURLResponse>> {
  try {
    const { userId } = await auth();
    if (!userId) return { ok: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };
    const client = getKnowledgeServiceClient();
    const request = new knowledge.CreateUploadURLRequest();
    request.setSpaceId(input.spaceId);
    request.setFilename(input.filename);
    request.setMimeType(input.mimeType);
    request.setSizeBytes(input.sizeBytes);
    request.setObjectKind(knowledge.DownloadObjectKind.DOWNLOAD_OBJECT_KIND_ORIGINAL);
    const metadata = await createMetadataWithAuth();
    return new Promise((resolve) => {
      client.createUploadURL(request, metadata, (error: any, response: any) => {
        if (error) {
          resolve({ ok: false, error: { code: error.code || 'UNKNOWN', message: error.message } });
        } else {
          resolve({
            ok: true,
            data: {
              uploadUrl: response.getUploadUrl(),
              contentSourceId: response.getContentSource()?.getId() || '',
              expiresAt: timestampToISOString(response.getExpiresAt?.()),
            },
          });
        }
      });
    });
  } catch {
    return { ok: false, error: { code: 'INTERNAL', message: 'An unexpected error occurred' } };
  }
}

export async function confirmUpload(
  contentSourceId: string,
  kind: 'original' | 'processed',
  blobHash: string,
): Promise<ActionResult<ContentSource>> {
  try {
    const { userId } = await auth();
    if (!userId) return { ok: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };
    const client = getKnowledgeServiceClient();
    const request = new knowledge.ConfirmUploadRequest();
    request.setContentSourceId(contentSourceId);
    request.setObjectKind(
      kind === 'processed'
        ? knowledge.DownloadObjectKind.DOWNLOAD_OBJECT_KIND_PROCESSED
        : knowledge.DownloadObjectKind.DOWNLOAD_OBJECT_KIND_ORIGINAL,
    );
    request.setBlobHash(blobHash);
    const metadata = await createMetadataWithAuth();
    return new Promise((resolve) => {
      client.confirmUpload(request, metadata, (error: any, response: any) => {
        if (error) {
          resolve({ ok: false, error: { code: error.code || 'UNKNOWN', message: error.message } });
        } else {
          resolve({ ok: true, data: protoContentSourceToContentSource(response) });
        }
      });
    });
  } catch {
    return { ok: false, error: { code: 'INTERNAL', message: 'An unexpected error occurred' } };
  }
}

export async function generateDownloadURL(
  contentSourceId: string,
  kind: 'original' | 'processed' = 'original',
  expiresSeconds?: number,
): Promise<ActionResult<{ url: string; expiresAt: string }>> {
  try {
    const { userId } = await auth();
    if (!userId) return { ok: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };
    const client = getKnowledgeServiceClient();
    const request = new knowledge.GenerateDownloadURLRequest();
    request.setContentSourceId(contentSourceId);
    if (expiresSeconds && expiresSeconds > 0) request.setExpiresSeconds(expiresSeconds);
    request.setObjectKind(
      kind === 'processed'
        ? knowledge.DownloadObjectKind.DOWNLOAD_OBJECT_KIND_PROCESSED
        : knowledge.DownloadObjectKind.DOWNLOAD_OBJECT_KIND_ORIGINAL,
    );
    const metadata = await createMetadataWithAuth();
    return new Promise((resolve) => {
      client.generateDownloadURL(request, metadata, (error: any, response: any) => {
        if (error) {
          resolve({ ok: false, error: { code: error.code || 'UNKNOWN', message: error.message } });
        } else {
          resolve({
            ok: true,
            data: {
              url: response.getUrl(),
              expiresAt: timestampToISOString(response.getExpiresAt?.()),
            },
          });
        }
      });
    });
  } catch {
    return { ok: false, error: { code: 'INTERNAL', message: 'An unexpected error occurred' } };
  }
}

export async function deleteContentSource(contentSourceId: string): Promise<ActionResult<void>> {
  try {
    const { userId } = await auth();
    if (!userId) return { ok: false, error: { code: 'UNAUTHORIZED', message: 'User not authenticated' } };
    const client = getKnowledgeServiceClient();
    const request = new knowledge.DeleteContentSourceRequest();
    request.setId(contentSourceId);
    const metadata = await createMetadataWithAuth();
    return new Promise((resolve) => {
      client.deleteContentSource(request, metadata, (error: any) => {
        if (error) {
          resolve({ ok: false, error: { code: error.code || 'UNKNOWN', message: error.message } });
        } else {
          resolve({ ok: true });
        }
      });
    });
  } catch {
    return { ok: false, error: { code: 'INTERNAL', message: 'An unexpected error occurred' } };
  }
}


