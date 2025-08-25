// Spaces types for the Knowledge Microservice (spaces DB)
// This handles space management, collaboration, and metadata
// Related types: ./content.ts (content_sources DB), ./canvas.ts (Canvas microservice)

export interface Space {
  id: string
  userId: string
  title: string
  description: string
  icon?: string
  coverImage?: string
  keywords: string[]
  accessLevel: 'private' | 'shared' | 'public'
  documentCount: number
  totalSizeBytes: number
  createdAt: Date | string
  lastUpdatedAt: Date | string
  contentCount?: number
  userCount?: number

  // Enhanced fields for new functionality
  collaboration_settings?: CollaborationSettings
  processing_stats?: ProcessingStats
}

// Collaboration settings for spaces
export interface CollaborationSettings {
  allowComments: boolean
  allowEditing: boolean
  shareSettings: {
    publicLink?: string
    expiresAt?: Date
    permissions: 'view' | 'comment' | 'edit'
  }
  invitedUsers: CollaborationUser[]
}

export interface CollaborationUser {
  userId: string
  email: string
  role: 'viewer' | 'editor' | 'admin'
  invitedAt: Date
  lastActive?: Date
}

// Processing statistics for spaces
export interface ProcessingStats {
  totalDocuments: number
  processedDocuments: number
  failedDocuments: number
  totalProcessingTime: number
  lastProcessedAt?: Date
  averageProcessingTime: number
}



// Error handling types
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



// Validation error types
export interface ValidationError {
  field: string
  message: string
  code: string
}

export interface FormErrors {
  [field: string]: ValidationError[]
}

// API error response type
export interface ApiError {
  code: string
  message: string
  details?: Record<string, any>
  timestamp: string
  requestId?: string
}

export interface SpaceFilters {
  q?: string              // Search query
  keywords?: string[]     // Filter by keywords
  sortBy?: 'name' | 'created' | 'updated' | 'documents'
  sortOrder?: 'asc' | 'desc'
  page?: number
  pageSize?: number
}

export interface SpaceWithStats extends Space {
  stats?: {
    contentCount: number
    userCount: number
    lastActivity?: string
  }
}

export interface CreateSpaceInput {
  title: string
  description: string
  keywords?: string[]
  icon?: string
}

export interface UpdateSpaceInput {
  title?: string
  description?: string
  keywords?: string[]
  icon?: string
}

export interface CreateUploadURLRequest {
  spaceId: string
  filename: string
  mimeType: string
  sizeBytes: number
}

export interface CreateUploadURLResponse {
  uploadUrl: string
  contentSourceId: string
  expiresAt: string
}

export interface SearchSpacesRequest {
  q?: string
  keywords?: string[]
  ownerId?: string
  page?: number
  pageSize?: number
  sortKey?: string
  sortDirection?: 'asc' | 'desc'
}

export interface SearchSpacesResponse {
  spaces: Space[]
  totalCount: number
  page: number
  pageSize: number
}

// UI-specific types
export type ViewMode = 'gallery' | 'list' | 'canvas'

export interface SpacesUIState {
  view: ViewMode
  searchTerm: string
  selectedKeywords: string[]
  sortBy: SpaceFilters['sortBy']
  sortOrder: SpaceFilters['sortOrder']
  page: number
  pageSize: number
  isCreating: boolean
  isDeleting: string | null // spaceId being deleted
  lastError: string | null
  selectedSpaceId: string | null
}

// Enhanced state management for spaces and content
// Note: Content-related types are imported from ./content.ts
export interface EnhancedSpacesState extends SpacesUIState {
  // Content management (types from content.ts)
  contentSources: Record<string, any[]>  // spaceId -> content sources
  uploadProgress: Record<string, any>   // fileId -> progress  
  processingStatus: Record<string, any> // contentId -> status

  // Upload sessions (types from content.ts)
  activeSessions: Record<string, any>    // sessionId -> session

  // Error states
  errors: Record<string, ErrorState>              // errorId -> error

  // Loading states
  loadingStates: {
    spaces: boolean
    contentSources: Record<string, boolean>       // spaceId -> loading
    uploads: Record<string, boolean>              // sessionId -> loading
  }
}



// Space detail page UI state
export interface SpaceDetailState {
  spaceId: string
  activeSection: 'canvas' | 'documents' | 'chat'
  canvasZoom: number
  canvasCenter: { x: number; y: number }
  selectedNodes: string[]
  sidebarCollapsed: boolean
  documentsPanelCollapsed: boolean
}

// Grid size options for gallery view
export type GridSize = 'small' | 'medium' | 'large'

// Sort options for different views
export type SortOption = {
  key: string
  label: string
  direction: 'asc' | 'desc'
}

// Filter options
export interface FilterOptions {
  accessLevels: ('private' | 'shared' | 'public')[]
  dateRange?: {
    start: Date
    end: Date
  }
  sizeRange?: {
    min: number
    max: number
  }
  hasDocuments?: boolean
}

// Server Action Result Types
export interface ActionResult<T> {
  ok: boolean
  data?: T
  error?: {
    code: string
    message: string
    details?: Record<string, any>
  }
}

// Enhanced server action types
export interface ServerActionResult<T> extends ActionResult<T> {
  timestamp: string
  requestId?: string
}



export interface SpaceCardProps {
  space: SpaceWithStats
  view: ViewMode
  onSelect: (spaceId: string) => void
  onEdit: (spaceId: string) => void
  onDelete: (spaceId: string) => void
  isDeleting?: boolean
}

// Space-specific component prop types
export interface SpaceHeaderProps {
  space: Space
  onEdit: () => void
  onDelete: () => void
  onShare: () => void
  isEditing?: boolean
}

// Utility types
export type DeepPartial<T> = {
  [P in keyof T]?: T[P] extends object ? DeepPartial<T[P]> : T[P]
}

export type RequiredFields<T, K extends keyof T> = T & Required<Pick<T, K>>

export type OptionalFields<T, K extends keyof T> = Omit<T, K> & Partial<Pick<T, K>>

// Constants for UI
export const GRID_SIZES = {
  small: { columns: 4, cardHeight: 200 },
  medium: { columns: 3, cardHeight: 250 },
  large: { columns: 2, cardHeight: 300 }
} as const