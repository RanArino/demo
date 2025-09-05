import type { 
  BaseEntity,
  ContentSourceType,
  ContentSourceStatus,
  ErrorState,
  ContentMetadata,
  ProcessingMetadata
} from './shared'

// =========================================================================
// CONTENT SOURCE INTERFACES
// =========================================================================

export interface ContentSource extends BaseEntity, ContentMetadata {
  spaceId: string
  sourceType: ContentSourceType
  processingStatus: ContentSourceStatus
  sourceMetadata?: Record<string, any>
  processingError?: string
  processingStartedAt?: Date | string
  processingCompletedAt?: Date | string
  uploadUrl?: string
}

// =========================================================================
// UPLOAD MANAGEMENT
// =========================================================================

export interface UploadSession extends BaseEntity {
  spaceId: string
  files: UploadFile[]
  status: 'active' | 'completed' | 'cancelled'
  completedAt?: Date | string
  totalFiles: number
  completedFiles: number
  failedFiles: number
}

export interface UploadFile extends BaseEntity {
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

export interface ProcessingStatus extends ProcessingMetadata {
  contentId: string
  status: ContentSourceStatus
  message?: string
  error?: string
  startedAt?: Date | string
  completedAt?: Date | string
}

// =========================================================================
// API REQUEST/RESPONSE TYPES
// =========================================================================

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

export interface CreateUploadURLRequest {
  spaceId: string;
  filename: string;
  mimeType: string;
  sizeBytes: number;
  objectKind?: 'original' | 'processed';
}
export interface CreateUploadURLResponse {
  uploadUrl: string;
  contentSourceId: string;
  expiresAt: string;
  objectKey: string;
}

// =========================================================================\n// GOOGLE DRIVE INTEGRATION\n// =========================================================================

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

// =========================================================================\n// COMPONENT PROPS\n// =========================================================================

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
  onClose?: () => void
  onUploadComplete?: (contentSources: ContentSource[]) => void
  autoOpenTab?: 'file' | 'google-drive' | 'link' | 'text'
}

export interface ProcessingStatusProps {
  status: ProcessingStatus
  showDetails?: boolean
  onRetry?: () => void
  onCancel?: () => void
}

// =========================================================================\n// UI STATE\n// =========================================================================

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

export interface UploadError extends ErrorState {
  fileId?: string
  filename?: string
  uploadProgress?: number
}

// =========================================================================\n// CONSTANTS\n// =========================================================================

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
export const UPLOAD_CHUNK_SIZE = 5 * 1024 * 1024 // 5MB chunks

// =========================================================================\n// RE-EXPORT SHARED TYPES\n// =========================================================================

export { ContentSourceType, ContentSourceStatus } from './shared'