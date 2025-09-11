'use client';

import { useState, useCallback } from 'react';
// Server actions are dynamically imported at call time to avoid stale action IDs during HMR
import { ContentSource, ContentStatus, DownloadObjectKind } from '@/api/generated/v1/knowledge_pb';
import { Timestamp, protoInt64 } from '@bufbuild/protobuf';

interface UploadOptions {
  spaceId: string;
  onUploadComplete?: (contentSources: ContentSource[]) => void;
  onUploadError?: (error: string) => void;
}

/**
 * Represents the state of a file upload.
 *
 * - isUploading: Indicates whether an upload is currently in progress.
 * - progress: Upload progress percentage in the range 0-100.
 * - error: Error message if a failure occurred; null when no error.
 */
interface UploadState {
  isUploading: boolean;
  progress: number;
  error: string | null;
}

export function useUpload({ spaceId, onUploadComplete, onUploadError }: UploadOptions) {
  const [uploadState, setUploadState] = useState<UploadState>({
    isUploading: false,
    progress: 0,
    error: null
  });

  const uploadFile = useCallback(async (
    file: File,
    onProgress?: (progress: number) => void,
    onProcessing?: (contentSourceId: string) => void,
    onComplete?: () => void,
    onError?: (error: string) => void
  ) => {
    setUploadState(prev => ({ ...prev, isUploading: true, error: null }));

    try {
      // Step 1: Get upload URL from server action (dynamic import to avoid stale action id)
      const { createUploadURL } = await import('@/api/actions/contentActions');
      
      const uploadUrlResult = await createUploadURL({
        spaceId,
        filename: file.name,
        mimeType: file.type,
        sizeBytes: file.size,
        objectKind: DownloadObjectKind.ORIGINAL,
        title: file.name,
      } as any);

      if (!uploadUrlResult.ok || !uploadUrlResult.data) {
        throw new Error(uploadUrlResult.error?.message || 'Failed to create upload URL');
      }

      const { uploadUrl, contentSource } = uploadUrlResult.data;
      const contentSourceId = contentSource?.id;
      
      if (!contentSourceId) {
        throw new Error('Failed to get content source ID from upload response');
      }

      // Broadcast creation so the Documents list can render a placeholder card
      try {
        const placeholder = new ContentSource({
          id: contentSourceId,
          spaceId,
          title: file.name,
          mimeType: file.type || 'application/octet-stream',
          sizeBytes: protoInt64.parse(file.size),
          status: ContentStatus.UPLOADING,
          createdAt: Timestamp.fromDate(new Date()),
          updatedAt: Timestamp.fromDate(new Date()),
        });
        window.dispatchEvent(new CustomEvent('content:created', { detail: placeholder }));
      } catch {}

      // Step 2: Upload file directly to cloud storage with progress tracking
      const xhr = new XMLHttpRequest();
      
      return new Promise<void>((resolve, reject) => {
        // Track upload progress
        xhr.upload.addEventListener('progress', (event) => {
          if (event.lengthComputable) {
            const progress = Math.round((event.loaded / event.total) * 100);
            setUploadState(prev => ({ ...prev, progress }));
            onProgress?.(progress);
            // Broadcast progress so DocumentsSection can reflect per-card progress
            try {
              window.dispatchEvent(new CustomEvent('content:progress', { detail: { contentSourceId, progress } }));
            } catch {}
          }
        });

        // Handle completion
        xhr.addEventListener('load', async () => {
          if (xhr.status >= 200 && xhr.status < 300) {
            try {
              // Step 3: Confirm upload with backend
              onProcessing?.(contentSourceId);
              
              // Confirm the ORIGINAL upload (first step). We don't yet know blob_hash here,
              // so compute it on client for now to satisfy API; backend accepts and stores it.
              const buffer = await file.arrayBuffer();
              const hashBuffer = await crypto.subtle.digest('SHA-256', buffer);
              const hashArray = Array.from(new Uint8Array(hashBuffer));
              const blobHash = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
              const { confirmUpload } = await import('@/api/actions/contentActions');
              const confirmResult = await confirmUpload(contentSourceId, blobHash);
              
              if (!confirmResult.ok || !confirmResult.data) {
                throw new Error(confirmResult.error?.message || 'Failed to confirm upload');
              }

              // Broadcast updated server copy (may include status/summary)
              try {
                window.dispatchEvent(new CustomEvent('content:updated', { detail: confirmResult.data }));
              } catch {}

              onComplete?.();
              onUploadComplete?.([confirmResult.data]);
              
              setUploadState(prev => ({ ...prev, isUploading: false }));
              resolve();
            } catch (error) {
              const errorMessage = error instanceof Error ? error.message : 'Upload confirmation failed';
              onError?.(errorMessage);
              onUploadError?.(errorMessage);
              setUploadState(prev => ({ ...prev, error: errorMessage, isUploading: false }));
              reject(error);
            }
          } else {
            const errorMessage = `Upload failed with status ${xhr.status}`;
            onError?.(errorMessage);
            onUploadError?.(errorMessage);
            setUploadState(prev => ({ ...prev, error: errorMessage, isUploading: false }));
            reject(new Error(errorMessage));
          }
        });

        // Handle errors
        xhr.addEventListener('error', () => {
          const errorMessage = 'Network error during upload';
          onError?.(errorMessage);
          onUploadError?.(errorMessage);
          setUploadState(prev => ({ ...prev, error: errorMessage, isUploading: false }));
          reject(new Error(errorMessage));
        });

        // Handle abort
        xhr.addEventListener('abort', () => {
          const errorMessage = 'Upload was cancelled';
          onError?.(errorMessage);
          setUploadState(prev => ({ ...prev, error: errorMessage, isUploading: false }));
          reject(new Error(errorMessage));
        });

        // Start the upload
        xhr.open('PUT', uploadUrl);
        const contentType = file.type && file.type.trim() !== '' ? file.type : 'application/octet-stream';
        xhr.setRequestHeader('Content-Type', contentType);
        xhr.send(file);
      });

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Upload failed';
      onError?.(errorMessage);
      onUploadError?.(errorMessage);
      setUploadState(prev => ({ ...prev, error: errorMessage, isUploading: false }));
      throw error;
    }
  }, [spaceId, onUploadComplete, onUploadError]);

  const reset = useCallback(() => {
    setUploadState({
      isUploading: false,
      progress: 0,
      error: null
    });
  }, []);

  return {
    uploadFile,
    isUploading: uploadState.isUploading,
    progress: uploadState.progress,
    error: uploadState.error,
    reset
  };
}