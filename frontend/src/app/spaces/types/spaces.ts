// Space-related types derived from knowledge.proto

export interface Space {
  id: string
  userId: string
  title: string
  description: string
  icon?: string
  keywords: string[]
  createdAt: string
  updatedAt: string
  contentCount?: number
  userCount?: number
}

export interface ContentSource {
  id: string
  spaceId: string
  name: string
  type: ContentSourceType
  url?: string
  size?: number
  mimeType?: string
  status: ContentSourceStatus
  createdAt: string
  updatedAt: string
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

// Server Action Result Types
export interface ActionResult<T> {
  ok: boolean
  data?: T
  error?: {
    code: string
    message: string
  }
}

export interface SpaceCardProps {
  space: SpaceWithStats
  view: ViewMode
  onSelect: (spaceId: string) => void
  onEdit: (spaceId: string) => void
  onDelete: (spaceId: string) => void
  isDeleting?: boolean
}