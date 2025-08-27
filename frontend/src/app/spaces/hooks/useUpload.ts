'use client';

import { useState, useCallback } from 'react';
import { createUploadURL, confirmUpload } from '@/api/actions/spaceActions';
import { ContentSource } from '@/app/spaces/types/content';

interface UploadOptions {
  spaceId: string;
  onUploadComplete?: (contentSources: ContentSource[]) => void;
  onUploadError?: (error: string) => void;
}


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
      // Step 1: Get upload URL from backend
      const uploadUrlResult = await createUploadURL({
        spaceId,
        filename: file.name,
        mimeType: file.type,
        sizeBytes: file.size
      });

      if (!uploadUrlResult.ok || !uploadUrlResult.data) {
        throw new Error(uploadUrlResult.error?.message || 'Failed to create upload URL');
      }

      const { uploadUrl, contentSourceId } = uploadUrlResult.data;

      // Step 2: Upload file directly to cloud storage with progress tracking
      const xhr = new XMLHttpRequest();
      
      return new Promise<void>((resolve, reject) => {
        // Track upload progress
        xhr.upload.addEventListener('progress', (event) => {
          if (event.lengthComputable) {
            const progress = Math.round((event.loaded / event.total) * 100);
            setUploadState(prev => ({ ...prev, progress }));
            onProgress?.(progress);
          }
        });

        // Handle completion
        xhr.addEventListener('load', async () => {
          if (xhr.status >= 200 && xhr.status < 300) {
            try {
              // Step 3: Confirm upload with backend
              onProcessing?.(contentSourceId);
              
              const confirmResult = await confirmUpload(contentSourceId);
              
              if (!confirmResult.ok || !confirmResult.data) {
                throw new Error(confirmResult.error?.message || 'Failed to confirm upload');
              }

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
        xhr.setRequestHeader('Content-Type', file.type);
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