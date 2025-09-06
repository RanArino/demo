'use client';

import { useState, useCallback, useRef } from 'react';
import { Button } from '@/components/ui/button';
import { 
  Upload, 
  FileText, 
  X
} from 'lucide-react';
import { cn, formatSize } from '@/lib/utils';
import { 
  SUPPORTED_FILE_TYPES, 
  MAX_FILE_SIZE, 
  MAX_FILES_PER_UPLOAD,
  type SupportedMimeType
} from '@/lib/files';
import { ContentSource } from '@/api/generated/v1/knowledge_pb';
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
}

interface DropZoneProps {
  isDragActive: boolean;
  disabled: boolean;
  onDragEnter: (e: React.DragEvent) => void;
  onDragLeave: (e: React.DragEvent) => void;
  onDragOver: (e: React.DragEvent) => void;
  onDrop: (e: React.DragEvent) => void;
  fileInputRef: React.RefObject<HTMLInputElement | null>;
  onFileInputChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
}

function DropZone({
  isDragActive,
  disabled,
  onDragEnter,
  onDragLeave,
  onDragOver,
  onDrop,
  fileInputRef,
  onFileInputChange
}: DropZoneProps) {
  return (
    <div
      className={cn(
        'relative border-2 border-dashed rounded-lg p-8 transition-colors',
        'hover:bg-muted/50 focus-within:bg-muted/50',
        isDragActive && 'border-primary bg-primary/5',
        !isDragActive && 'border-muted-foreground/25',
        disabled && 'opacity-50 pointer-events-none'
      )}
      onDragEnter={onDragEnter}
      onDragLeave={onDragLeave}
      onDragOver={onDragOver}
      onDrop={onDrop}
    >
      <input
        ref={fileInputRef}
        type="file"
        multiple
        accept={Object.keys(SUPPORTED_FILE_TYPES).join(',')}
        onChange={onFileInputChange}
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
  );
}

interface FilesHeaderProps {
  count: number;
  onClearAll: () => void;
  disabled: boolean;
}

function FilesHeader({ count, onClearAll, disabled }: FilesHeaderProps) {
  return (
    <div className="flex items-center justify-between">
      <h4 className="text-sm font-medium">Files ({count})</h4>
      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={onClearAll}
          disabled={disabled}
        >
          Clear All
        </Button>
      </div>
    </div>
  );
}

interface FileItemProps {
  file: FileWithId;
  onRemove: (id: string) => void;
}

function FileItem({ file, onRemove }: FileItemProps) {
  return (
    <div className="flex items-center gap-3 p-3 border rounded-lg bg-background">
      <FileText className="h-5 w-5 text-muted-foreground flex-shrink-0" />
      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between">
          <p className="text-sm font-medium truncate">{file.name}</p>
          <div className="flex items-center gap-2">
            <span className="text-xs text-muted-foreground">{formatSize(file.size)}</span>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onRemove(file.id)}
              className="h-6 w-6 p-0 hover:bg-destructive/10 hover:text-destructive"
            >
              <X className="h-3 w-3" />
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}

interface FileListProps {
  files: FileState[];
  onRemove: (id: string) => void;
}

function FileList({ files, onRemove }: FileListProps) {
  return (
    <div className="space-y-2 max-h-60 overflow-y-auto">
      {files.map((fs) => (
        <FileItem key={fs.file.id} file={fs.file} onRemove={onRemove} />
      ))}
    </div>
  );
}

interface UploadActionsProps {
  hasFiles: boolean;
  disabled: boolean;
  isUploading: boolean;
  onClick: () => void;
}

function UploadActions({ hasFiles, disabled, isUploading, onClick }: UploadActionsProps) {
  if (!hasFiles) return null;
  return (
    <div className="flex justify-end">
      <Button
        onClick={onClick}
        disabled={disabled || isUploading || !hasFiles}
        className="min-w-[120px]"
      >
        <>
          <Upload className="h-4 w-4 mr-2" />
          Upload Files
        </>
      </Button>
    </div>
  );
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
        newFiles.push({ file: fileWithId });
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


  // Reset all files
  const resetFiles = useCallback(() => {
    setFiles([]);
  }, []);

  const hasFiles = files.length > 0;

  return (
    <div className="space-y-4">
      <DropZone
        isDragActive={isDragActive}
        disabled={disabled}
        onDragEnter={handleDragEnter}
        onDragLeave={handleDragLeave}
        onDragOver={handleDragOver}
        onDrop={handleDrop}
        fileInputRef={fileInputRef}
        onFileInputChange={handleFileInputChange}
      />

      {/* File List */}
      {hasFiles && (
        <div className="space-y-3">
          <FilesHeader
            count={files.length}
            onClearAll={resetFiles}
            disabled={disabled || isUploading}
          />

          <FileList files={files} onRemove={removeFile} />

          <UploadActions
            hasFiles={hasFiles}
            disabled={disabled}
            isUploading={isUploading}
            onClick={() => {
              if (!hasFiles) return;
              onUploadStart?.();
              const currentFiles = [...files];
              currentFiles.forEach((fileState) => {
                void uploadFile(
                  fileState.file,
                  undefined,
                  undefined,
                  undefined,
                  (error) => {
                    onUploadError?.(error);
                  }
                );
              });
            }}
          />
        </div>
      )}
    </div>
  );
}