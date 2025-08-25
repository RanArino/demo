// Shared types, enums, and interfaces used across the spaces domain
// (spaces, content, canvas) to avoid duplication and maintain consistency.

// =========================================================================
// ENUMS
// =========================================================================

export enum ContentSourceType {
  FILE = 'file',
  URL = 'url', 
  TEXT = 'text',
  GOOGLE_DRIVE = 'google_drive'
}

export enum ContentSourceStatus {
  PENDING = 'pending',
  PROCESSING = 'processing', 
  COMPLETED = 'completed',
  FAILED = 'failed'
}

export enum ProcessingStage {
  UPLOAD = 'upload',
  EXTRACTION = 'extraction',
  ANALYSIS = 'analysis', 
  INDEXING = 'indexing',
  COMPLETE = 'complete'
}

export enum AccessLevel {
  PRIVATE = 'private',
  SHARED = 'shared',
  PUBLIC = 'public'
}

// =========================================================================
// BASE INTERFACES
// =========================================================================

export interface BaseEntity {
  id: string
  createdAt: Date | string
  updatedAt: Date | string
}

export interface BaseEntityWithUser extends BaseEntity {
  userId: string
}

export interface Timestamps {
  createdAt: Date | string
  updatedAt: Date | string
}

// =========================================================================
// ERROR HANDLING
// =========================================================================

export interface ErrorState {
  type: 'validation' | 'network' | 'server' | 'processing' | 'authentication' | 'permission'
  message: string
  code: string
  retryable: boolean
  actions?: ErrorAction[]
  details?: Record<string, any>
}

export interface ErrorAction {
  label: string
  action: () => void | Promise<void>
  variant: 'primary' | 'secondary' | 'destructive'
  loading?: boolean
}

export interface ValidationError {
  field: string
  message: string
  code: string
}

export interface FormErrors {
  [field: string]: ValidationError[]
}

export interface ApiError {
  code: string
  message: string
  details?: Record<string, any>
  timestamp: string
  requestId?: string
}

// =========================================================================
// API RESPONSE TYPES
// =========================================================================

export interface ActionResult<T> {
  ok: boolean
  data?: T
  error?: {
    code: string
    message: string
    details?: Record<string, any>
  }
}

export interface ServerActionResult<T> extends ActionResult<T> {
  timestamp: string
  requestId?: string
}

// =========================================================================
// UTILITY TYPES
// =========================================================================

export type DeepPartial<T> = {
  [P in keyof T]?: T[P] extends object ? DeepPartial<T[P]> : T[P]
}

export type RequiredFields<T, K extends keyof T> = T & Required<Pick<T, K>>

export type OptionalFields<T, K extends keyof T> = Omit<T, K> & Partial<Pick<T, K>>

// =========================================================================
// COMMON UI STATE
// =========================================================================

export interface LoadingStates {
  [key: string]: boolean
}

export interface FilterState {
  searchTerm: string
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
  page: number
  pageSize: number
}

// =========================================================================
// METADATA
// =========================================================================

export interface ContentMetadata {
  mimeType: string
  sizeBytes: number
  filename?: string
  title?: string
  extractedText?: string
  thumbnailUrl?: string
  contentSummary?: string
  detectedLanguage?: string
  wordCount?: number
  pageCount?: number
}

export interface ProcessingMetadata {
  processingStartedAt?: Date | string
  processingCompletedAt?: Date | string
  processingError?: string
  stage: ProcessingStage
  progress: number
}
