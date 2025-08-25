// Re-export all types for easier imports

// Shared types and utilities
export * as Shared from './shared'

// Knowledge Microservice types
export * as Spaces from './spaces'      // spaces DB
export * as Content from './content'    // content_sources DB

// Canvas Microservice types  
export * as Canvas from './canvas'      // Canvas service (separate microservice)

// Convenience re-exports for commonly used types
export type { 
  ErrorState, 
  ErrorAction, 
  ValidationError,
  ApiError,
  ActionResult,
  ServerActionResult
} from './shared'

export type { Space, SpaceWithStats, ViewMode } from './spaces'
export type { ContentSource, UploadSession } from './content'
export type { CanvasData, CanvasNode, CanvasConnection } from './canvas'