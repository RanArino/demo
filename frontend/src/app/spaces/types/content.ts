// Content Sources types for the Knowledge Microservice (content_sources DB)
// This handles all content upload, processing, and management functionality
export interface ContentSource {
  id: string
  spaceId: string
  filename?: string
  title?: string
  mimeType: string
  sizeBytes: number
  uploadUrl?: string
  processingStatus: 'pending' | 'processing' | 'completed' | 'failed'
  createdAt: Date | string
  updatedAt: Date | string
  
  // Content-specific fields
  sourceType: 'file' | 'url' | 'text' | 'google_drive'
  sourceMetadata?: Record<string, any>
  extractedText?: string
  thumbnailUrl?: string
  
  // Processing details
  processingError?: string
  processingStartedAt?: Date | string
  processingCompletedAt?: Date | string
  
  // Content analysis results
  contentSummary?: string
  detectedLanguage?: string
  wordCount?: number
  pageCount?: number
}

export enum ContentSourceType {
  UNKNOWN = 'UNKNOWN',
  URL = 'URL',
  FILE = 'FILE',
  TEXT = 'TEXT'
}

export enum ContentSourceStatus {
  PENDING = 'PENDING',
  PROCESSING = 'PROCESSING',
  COMPLETED = 'COMPLETED',
  FAILED = 'FAILED'
}

// Upload management types (Knowledge microservice - content_sources DB)
export interface UploadSession {
  id: string
  spaceId: string
  files: UploadFile[]
  status: 'active' | 'completed' | 'cancelled'
  createdAt: Date | string
  completedAt?: Date | string
  totalFiles: number
  completedFiles: number
  failedFiles: number
}

export interface UploadFile {
  id: string
  sessionId: string
  filename: string
  sizeBytes: number
  mimeType: string
  uploadUrl?: string
  progress: number
  status: 'pending' | 'uploading' | 'processing' | 'completed' | 'failed'
  errorMessage?: string
  contentSourceId?: string
  startedAt?: Date | string
  completedAt?: Date | string
}

// Upload progress tracking
export interface UploadProgress {
  fileId: string
  filename: string
  progress: number
  status: 'uploading' | 'processing' | 'completed' | 'failed'
  bytesUploaded: number
  totalBytes: number
  uploadSpeed?: number
  estimatedTimeRemaining?: number
}

// Processing status for real-time updates
export interface ProcessingStatus {
  contentId: string
  status: 'pending' | 'processing' | 'completed' | 'failed'
  progress: number
  stage: 'upload' | 'extraction' | 'analysis' | 'indexing' | 'complete'
  message?: string
  error?: string
  startedAt?: Date | string
  completedAt?: Date | string
}

// Upload-related server action types
export interface CreateUploadSessionRequest {
  spaceId: string
  files: {
    filename: string
    mimeType: string
    sizeBytes: number
  }[]
}

export interface CreateUploadSessionResponse {
  sessionId: string
  uploadUrls: {
    fileId: string
    filename: string
    uploadUrl: string
    expiresAt: string
  }[]
}

export interface ConfirmUploadRequest {
  sessionId: string
  fileId: string
  success: boolean
  error?: string
}

export interface CreateContentSourceFromUrlRequest {
  spaceId: string
  url: string
  title?: string
  description?: string
}

export interface CreateContentSourceFromTextRequest {
  spaceId: string
  title: string
  content: string
  format?: 'plain' | 'markdown' | 'html'
}

export interface DeleteContentSourceRequest {
  contentSourceId: string
  spaceId: string
}

// Google Drive integration types
export interface GoogleDriveFile {
  id: string
  name: string
  mimeType: string
  size: number
  thumbnailLink?: string
  webViewLink: string
  downloadUrl?: string
  parents: string[]
  createdTime: string
  modifiedTime: string
}

export interface GoogleDriveAuthState {
  isAuthenticated: boolean
  accessToken?: string
  refreshToken?: string
  expiresAt?: Date
  userEmail?: string
}

export interface ImportGoogleDriveFileRequest {
  spaceId: string
  fileId: string
  filename: string
  mimeType: string
}

// Content-related component prop types
export interface ContentSourceCardProps {
  contentSource: ContentSource
  onView: (contentSource: ContentSource) => void
  onDownload: (contentSource: ContentSource) => void
  onDelete: (contentSource: ContentSource) => void
  isDeleting?: boolean
  showProcessingStatus?: boolean
}

export interface UploadModalProps {
  isOpen: boolean
  spaceId: string
  onClose: () => void
  onUploadComplete: (contentSources: ContentSource[]) => void
  autoOpenTab?: 'file' | 'google-drive' | 'link' | 'text'
}

export interface ProcessingStatusProps {
  status: ProcessingStatus
  showDetails?: boolean
  onRetry?: () => void
  onCancel?: () => void
}



// Upload modal UI state
export interface UploadModalState {
  isOpen: boolean
  spaceId: string | null
  activeTab: 'file' | 'google-drive' | 'link' | 'text'
  dragActive: boolean
  files: File[]
  uploadProgress: Record<string, UploadProgress>
  errors: UploadError[]
  isUploading: boolean
}

// Import error types from spaces.ts
export interface UploadError {
  type: 'validation' | 'network' | 'server' | 'processing' | 'authentication' | 'permission'
  message: string
  code: string
  retryable: boolean
  actions?: ErrorAction[]
  details?: Record<string, any>
  fileId?: string
  filename?: string
  uploadProgress?: number
}

export interface ErrorAction {
  label: string
  action: () => void | Promise<void>
  variant: 'primary' | 'secondary' | 'destructive'
  loading?: boolean
}

// File type validation constants
export const SUPPORTED_FILE_TYPES = {
  'application/pdf': '.pdf',
  'text/plain': '.txt',
  'text/markdown': '.md',
  'audio/mpeg': '.mp3',
  'audio/wav': '.wav',
  'audio/mp4': '.m4a',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document': '.docx',
  'application/msword': '.doc'
} as const

export type SupportedMimeType = keyof typeof SUPPORTED_FILE_TYPES

export const MAX_FILE_SIZE = 100 * 1024 * 1024 // 100MB
export const MAX_FILES_PER_UPLOAD = 10
export const UPLOAD_CHUNK_SIZE = 5 * 1024 * 1024 // 5MB chunks for large file uploads