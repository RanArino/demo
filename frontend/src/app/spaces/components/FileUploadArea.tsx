'use client';

import { useState, useCallback, useRef } from 'react';
import { Button } from '@/components/ui/button';
import { 
  Upload, 
  FileText, 
  X, 
  AlertCircle,
  CheckCircle
} from 'lucide-react';
import { cn, formatSize } from '@/lib/utils';
import { 
  SUPPORTED_FILE_TYPES, 
  MAX_FILE_SIZE, 
  MAX_FILES_PER_UPLOAD,
  type SupportedMimeType,
  type ContentSource
} from '@/app/spaces/types/content';
import UploadProgressBar from './UploadProgressBar';
import { useUpload } from '@/app/spaces/hooks/useUpload';

interface FileUploadAreaProps {
  spaceId: string;
  onUploadStart?: () => void;
  onUploadComplete?: (contentSources: ContentSource[]) => void;
  onUploadError?: (error: string) => void;
  disabled?: boolean;
}

interface FileWithId extends File {
  id: string;
}

interface FileState {
  file: FileWithId;
  status: 'pending' | 'uploading' | 'processing' | 'completed' | 'failed';
  progress: number;
  error?: string;
  contentSourceId?: string;
}

function generateFileId(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return `file_${crypto.randomUUID()}`;
  }
  const randomPart = Math.random().toString(36).slice(2, 11);
  return `file_${Date.now()}_${randomPart}`;
}

function validateFile(file: File): string | null {
  // Check file size
  if (file.size > MAX_FILE_SIZE) {
    return `File size exceeds ${formatSize(MAX_FILE_SIZE)} limit`;
  }

  // Check file type
  const mimeType = file.type as SupportedMimeType;
  if (!SUPPORTED_FILE_TYPES[mimeType]) {
    const supportedTypes = Object.values(SUPPORTED_FILE_TYPES).join(', ');
    return `Unsupported file type. Supported: ${supportedTypes}`;
  }

  return null;
}

export default function FileUploadArea({
  spaceId,
  onUploadStart,
  onUploadComplete,
  onUploadError,
  disabled = false
}: FileUploadAreaProps) {
  const [isDragActive, setIsDragActive] = useState(false);
  const [files, setFiles] = useState<FileState[]>([]);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const dragCountRef = useRef(0);

  const { uploadFile, isUploading } = useUpload({
    spaceId,
    onUploadComplete: (contentSources) => {
      onUploadComplete?.(contentSources);
    },
    onUploadError: (error) => {
      onUploadError?.(error);
    }
  });

  // Handle file selection
  const handleFiles = useCallback((selectedFiles: FileList | File[]) => {
    const fileArray = Array.from(selectedFiles);
    
    // Check total file count
    if (files.length + fileArray.length > MAX_FILES_PER_UPLOAD) {
      onUploadError?.(`Maximum ${MAX_FILES_PER_UPLOAD} files allowed per upload`);
      return;
    }

    const newFiles: FileState[] = [];
    const validationErrors: string[] = [];

    fileArray.forEach((file) => {
      const error = validateFile(file);
      if (error) {
        validationErrors.push(`${file.name}: ${error}`);
      } else {
        const fileWithId = Object.assign(file, { id: generateFileId() });
        newFiles.push({
          file: fileWithId,
          status: 'pending',
          progress: 0
        });
      }
    });

    if (validationErrors.length > 0) {
      onUploadError?.(validationErrors.join(', '));
    }

    if (newFiles.length > 0) {
      setFiles(prev => [...prev, ...newFiles]);
    }
  }, [files.length, onUploadError]);

  // File input change handler
  const handleFileInputChange = useCallback((event: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFiles = event.target.files;
    if (selectedFiles && selectedFiles.length > 0) {
      handleFiles(selectedFiles);
    }
    // Reset input to allow selecting the same files again
    event.target.value = '';
  }, [handleFiles]);

  // Drag and drop handlers
  const handleDragEnter = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    dragCountRef.current++;
    if (dragCountRef.current === 1) {
      setIsDragActive(true);
    }
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    dragCountRef.current--;
    if (dragCountRef.current === 0) {
      setIsDragActive(false);
    }
  }, []);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  }, []);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    
    dragCountRef.current = 0;
    setIsDragActive(false);

    const droppedFiles = e.dataTransfer.files;
    if (droppedFiles && droppedFiles.length > 0) {
      handleFiles(droppedFiles);
    }
  }, [handleFiles]);

  // Remove file from list
  const removeFile = useCallback((fileId: string) => {
    setFiles(prev => prev.filter(f => f.file.id !== fileId));
  }, []);

  // Start upload process
  const startUpload = useCallback(async () => {
    const pendingFiles = files.filter(f => f.status === 'pending');
    if (pendingFiles.length === 0) return;

    onUploadStart?.();

    // Update all pending files to uploading
    setFiles(prev => prev.map(f => 
      f.status === 'pending' 
        ? { ...f, status: 'uploading' as const }
        : f
    ));

    try {
      for (const fileState of pendingFiles) {
        await uploadFile(
          fileState.file,
          (progress) => {
            setFiles(prev => prev.map(f =>
              f.file.id === fileState.file.id
                ? { ...f, progress }
                : f
            ));
          },
          (contentSourceId) => {
            setFiles(prev => prev.map(f =>
              f.file.id === fileState.file.id
                ? { ...f, status: 'processing', contentSourceId }
                : f
            ));
          },
          () => {
            setFiles(prev => prev.map(f =>
              f.file.id === fileState.file.id
                ? { ...f, status: 'completed' }
                : f
            ));
          },
          (error) => {
            setFiles(prev => prev.map(f =>
              f.file.id === fileState.file.id
                ? { ...f, status: 'failed', error }
                : f
            ));
          }
        );
      }
    } catch (error) {
      console.error('Upload error:', error);
      onUploadError?.(error instanceof Error ? error.message : 'Upload failed');
    }
  }, [files, uploadFile, onUploadStart, onUploadError]);

  // Clear completed files
  const clearCompleted = useCallback(() => {
    setFiles(prev => prev.filter(f => f.status !== 'completed'));
  }, []);

  // Reset all files
  const resetFiles = useCallback(() => {
    setFiles([]);
  }, []);

  const hasFiles = files.length > 0;
  const hasPendingFiles = files.some(f => f.status === 'pending');
  const hasCompletedFiles = files.some(f => f.status === 'completed');
  const allCompleted = files.length > 0 && files.every(f => f.status === 'completed');

  return (
    <div className="space-y-4">
      {/* Drop Zone */}
      <div
        className={cn(
          'relative border-2 border-dashed rounded-lg p-8 transition-colors',
          'hover:bg-muted/50 focus-within:bg-muted/50',
          isDragActive && 'border-primary bg-primary/5',
          !isDragActive && 'border-muted-foreground/25',
          disabled && 'opacity-50 pointer-events-none'
        )}
        onDragEnter={handleDragEnter}
        onDragLeave={handleDragLeave}
        onDragOver={handleDragOver}
        onDrop={handleDrop}
      >
        <input
          ref={fileInputRef}
          type="file"
          multiple
          accept={Object.keys(SUPPORTED_FILE_TYPES).join(',')}
          onChange={handleFileInputChange}
          className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
          disabled={disabled}
        />

        <div className="text-center space-y-4">
          <div className="mx-auto h-12 w-12 flex items-center justify-center rounded-full bg-muted">
            <Upload className="h-6 w-6 text-muted-foreground" />
          </div>
          
          <div>
            <p className="text-lg font-medium">
              {isDragActive ? 'Drop files here' : 'Drag & drop files here'}
            </p>
            <p className="text-sm text-muted-foreground mt-1">
              or{' '}
              <button
                type="button"
                className="text-primary hover:underline"
                onClick={() => fileInputRef.current?.click()}
                disabled={disabled}
              >
                browse to choose files
              </button>
            </p>
          </div>

          <div className="text-xs text-muted-foreground space-y-1">
            <p>Supported formats: PDF, TXT, Markdown, Audio files</p>
            <p>Maximum file size: {formatSize(MAX_FILE_SIZE)} • Maximum {MAX_FILES_PER_UPLOAD} files</p>
          </div>
        </div>
      </div>

      {/* File List */}
      {hasFiles && (
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <h4 className="text-sm font-medium">Files ({files.length})</h4>
            <div className="flex items-center gap-2">
              {hasCompletedFiles && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={clearCompleted}
                  disabled={disabled}
                >
                  Clear Completed
                </Button>
              )}
              <Button
                variant="outline"
                size="sm"
                onClick={resetFiles}
                disabled={disabled || isUploading}
              >
                Clear All
              </Button>
            </div>
          </div>

          <div className="space-y-2 max-h-60 overflow-y-auto">
            {files.map((fileState) => (
              <div
                key={fileState.file.id}
                className="flex items-center gap-3 p-3 border rounded-lg bg-background"
              >
                <FileText className="h-5 w-5 text-muted-foreground flex-shrink-0" />
                
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between">
                    <p className="text-sm font-medium truncate">
                      {fileState.file.name}
                    </p>
                    <div className="flex items-center gap-2">
                      <span className="text-xs text-muted-foreground">
                        {formatSize(fileState.file.size)}
                      </span>
                      {fileState.status === 'pending' && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => removeFile(fileState.file.id)}
                          className="h-6 w-6 p-0 hover:bg-destructive/10 hover:text-destructive"
                        >
                          <X className="h-3 w-3" />
                        </Button>
                      )}
                    </div>
                  </div>
                  
                  {(fileState.status === 'uploading' || fileState.status === 'processing') && (
                    <UploadProgressBar
                      progress={fileState.progress}
                      status={fileState.status}
                      className="mt-2"
                    />
                  )}
                  
                  {fileState.status === 'completed' && (
                    <div className="flex items-center gap-1 mt-1">
                      <CheckCircle className="h-3 w-3 text-green-600" />
                      <span className="text-xs text-green-600">Upload complete</span>
                    </div>
                  )}
                  
                  {fileState.status === 'failed' && fileState.error && (
                    <div className="flex items-center gap-1 mt-1">
                      <AlertCircle className="h-3 w-3 text-destructive" />
                      <span className="text-xs text-destructive">{fileState.error}</span>
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>

          {/* Upload Button */}
          {hasPendingFiles && (
            <div className="flex justify-end">
              <Button
                onClick={startUpload}
                disabled={disabled || isUploading || !hasPendingFiles}
                className="min-w-[120px]"
              >
                {isUploading ? (
                  <>Uploading...</>
                ) : (
                  <>
                    <Upload className="h-4 w-4 mr-2" />
                    Upload Files
                  </>
                )}
              </Button>
            </div>
          )}

          {/* Success Message */}
          {allCompleted && (
            <div className="flex items-center justify-center gap-2 text-green-600 bg-green-50 p-3 rounded-lg">
              <CheckCircle className="h-5 w-5" />
              <span className="text-sm font-medium">All files uploaded successfully!</span>
            </div>
          )}
        </div>
      )}
    </div>
  );
}