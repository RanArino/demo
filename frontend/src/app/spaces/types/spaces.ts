import type { 
  BaseEntityWithUser, 
  AccessLevel, 
  ErrorState, 
  ErrorAction, 
  ValidationError, 
  FormErrors, 
  ApiError, 
  ActionResult, 
  ServerActionResult,
  DeepPartial,
  RequiredFields,
  OptionalFields,
  FilterState
} from './shared'

// =========================================================================
// SPACE INTERFACES
// =========================================================================

export interface Space extends BaseEntityWithUser {
  title: string
  description: string
  icon?: string
  coverImage?: string
  keywords: string[]
  accessLevel: AccessLevel
  documentCount: number
  totalSizeBytes: number
  lastUpdatedAt: Date | string
  contentCount?: number
  userCount?: number
  collaborationSettings?: CollaborationSettings
  processingStats?: ProcessingStats
}

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

export interface ProcessingStats {
  totalDocuments: number
  processedDocuments: number
  failedDocuments: number
  totalProcessingTime: number
  lastProcessedAt?: Date
  averageProcessingTime: number
}

// =========================================================================
// SPACE OPERATIONS
// =========================================================================

export interface SpaceFilters extends FilterState {
  q?: string
  keywords?: string[]
  sortBy?: 'name' | 'created' | 'updated' | 'documents'
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

// =========================================================================
// SEARCH & PAGINATION
// =========================================================================

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

// =========================================================================
// UI STATE MANAGEMENT
// =========================================================================

export type ViewMode = 'gallery' | 'list' | 'canvas'
export type GridSize = 'small' | 'medium' | 'large'

export interface SpacesUIState {
  view: ViewMode
  searchTerm: string
  selectedKeywords: string[]
  sortBy: SpaceFilters['sortBy']
  sortOrder: SpaceFilters['sortOrder']
  page: number
  pageSize: number
  isCreating: boolean
  isDeleting: string | null
  lastError: string | null
  selectedSpaceId: string | null
}

export interface EnhancedSpacesState extends SpacesUIState {
  contentSources: Record<string, any[]>
  uploadProgress: Record<string, any>
  processingStatus: Record<string, any>
  activeSessions: Record<string, any>
  errors: Record<string, ErrorState>
  loadingStates: {
    spaces: boolean
    contentSources: Record<string, boolean>
    uploads: Record<string, boolean>
  }
}

export interface SpaceDetailState {
  spaceId: string
  activeSection: 'canvas' | 'documents' | 'chat'
  canvasZoom: number
  canvasCenter: { x: number; y: number }
  selectedNodes: string[]
  sidebarCollapsed: boolean
  documentsPanelCollapsed: boolean
}

// =========================================================================
// COMPONENT PROPS
// =========================================================================

export interface SpaceCardProps {
  space: SpaceWithStats
  view: ViewMode
  onSelect: (spaceId: string) => void
  onEdit: (spaceId: string) => void
  onDelete: (spaceId: string) => void
  isDeleting?: boolean
}

export interface SpaceHeaderProps {
  space: Space
  onEdit: () => void
  onDelete: () => void
  onShare: () => void
  isEditing?: boolean
}

// =========================================================================
// FILTERING & SORTING
// =========================================================================

export type SortOption = {
  key: string
  label: string
  direction: 'asc' | 'desc'
}

export interface FilterOptions {
  accessLevels: AccessLevel[]
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

// =========================================================================
// CONSTANTS
// =========================================================================

export const GRID_SIZES = {
  small: { columns: 4, cardHeight: 200 },
  medium: { columns: 3, cardHeight: 250 },
  large: { columns: 2, cardHeight: 300 }
} as const

export const SORT_OPTIONS: SortOption[] = [
  { key: 'lastUpdatedAt', label: 'Last Updated', direction: 'desc' },
  { key: 'createdAt', label: 'Date Created', direction: 'desc' },
  { key: 'title', label: 'Title', direction: 'asc' },
  { key: 'documentCount', label: 'Documents', direction: 'desc' }
] as const

export const DEFAULT_PAGE_SIZE = 12
export const MAX_KEYWORDS = 10
export const MAX_DESCRIPTION_LENGTH = 500
export const MAX_TITLE_LENGTH = 100

// =========================================================================
// RE-EXPORT SHARED TYPES
// =========================================================================

export type { 
  ErrorState, 
  ErrorAction, 
  ValidationError, 
  FormErrors, 
  ApiError, 
  ActionResult, 
  ServerActionResult,
  DeepPartial,
  RequiredFields,
  OptionalFields
}