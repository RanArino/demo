// Canvas-related types for the Canvas Microservice
// These types define the minimal necessary structure for future Canvas service implementation

// Core canvas data structure
export interface CanvasData {
  id: string
  spaceId: string
  nodes: CanvasNode[]
  connections: CanvasConnection[]
  layout: CanvasLayout
  metadata?: CanvasMetadata
  createdAt: Date | string
  updatedAt: Date | string
}

// Canvas node representing content or concepts
export interface CanvasNode {
  id: string
  canvasId: string
  type: 'document' | 'concept' | 'note' | 'cluster'
  title: string
  content?: string
  position: CanvasPosition
  size: CanvasSize
  style?: CanvasNodeStyle
  metadata?: Record<string, any>
  
  // Relationships to content sources (from Knowledge microservice)
  contentSourceId?: string
  
  // Node state
  locked?: boolean
  visible?: boolean
  selected?: boolean
  
  createdAt: Date | string
  updatedAt: Date | string
}

// Canvas connection between nodes
export interface CanvasConnection {
  id: string
  canvasId: string
  sourceNodeId: string
  targetNodeId: string
  type: 'reference' | 'similarity' | 'custom' | 'hierarchy'
  strength: number
  label?: string
  style?: CanvasConnectionStyle
  metadata?: Record<string, any>
  
  createdAt: Date | string
  updatedAt: Date | string
}

// Canvas layout configuration
export interface CanvasLayout {
  type: 'force' | 'hierarchical' | 'circular' | 'custom' | 'manual'
  algorithmSettings: Record<string, any>
  viewport: CanvasViewport
  gridSettings?: CanvasGridSettings
}

// Supporting types for canvas elements
export interface CanvasPosition {
  x: number
  y: number
  z?: number // For 3D layouts in the future
}

export interface CanvasSize {
  width: number
  height: number
}

export interface CanvasViewport {
  center: CanvasPosition
  zoom: number
  rotation?: number
}

export interface CanvasNodeStyle {
  backgroundColor?: string
  borderColor?: string
  borderWidth?: number
  borderRadius?: number
  fontSize?: number
  fontColor?: string
  opacity?: number
  shadow?: boolean
}

export interface CanvasConnectionStyle {
  color?: string
  width?: number
  style?: 'solid' | 'dashed' | 'dotted'
  arrowType?: 'none' | 'arrow' | 'circle'
  opacity?: number
}

export interface CanvasGridSettings {
  enabled: boolean
  size: number
  color?: string
  opacity?: number
  snapToGrid?: boolean
}

export interface CanvasMetadata {
  version: string
  last_layout_algorithm?: string
  auto_layout_enabled?: boolean
  collaboration_enabled?: boolean
  view_mode?: 'edit' | 'view' | 'present'
}

// Canvas operations and interactions
export interface CanvasOperation {
  id: string
  canvasId: string
  type: 'create_node' | 'update_node' | 'delete_node' | 'create_connection' | 'delete_connection' | 'layout_change'
  data: Record<string, any>
  user_id?: string
  timestamp: Date | string
}

// Canvas collaboration (for real-time editing)
export interface CanvasCollaborator {
  userId: string
  cursorPosition?: CanvasPosition
  selectedNodes?: string[]
  active: boolean
  lastSeen: Date | string
}

// Canvas API request/response types
export interface CreateCanvasRequest {
  spaceId: string
  layoutType?: CanvasLayout['type']
  initialNodes?: Omit<CanvasNode, 'id' | 'canvasId' | 'createdAt' | 'updatedAt'>[]
}

export interface UpdateCanvasLayoutRequest {
  canvasId: string
  layout: CanvasLayout
}

export interface CreateNodeRequest {
  canvasId: string
  node: Omit<CanvasNode, 'id' | 'canvasId' | 'createdAt' | 'updatedAt'>
}

export interface UpdateNodeRequest {
  nodeId: string
  updates: Partial<Pick<CanvasNode, 'title' | 'content' | 'position' | 'size' | 'style' | 'metadata'>>
}

export interface CreateConnectionRequest {
  canvasId: string
  connection: Omit<CanvasConnection, 'id' | 'canvasId' | 'createdAt' | 'updatedAt'>
}

// Canvas UI state (for frontend components)
export interface CanvasUIState {
  canvasId: string | null
  viewport: CanvasViewport
  selectedNodes: string[]
  selectedConnections: string[]
  toolMode: 'select' | 'pan' | 'create_node' | 'create_connection'
  isLoading: boolean
  isSaving: boolean
  showGrid: boolean
  showMinimap: boolean
  collaborationActive: boolean
  collaborators: CanvasCollaborator[]
}

// Canvas component props (for React components)
export interface CanvasViewerProps {
  spaceId: string
  canvasData?: CanvasData
  readonly?: boolean
  onNodeSelect?: (nodeIds: string[]) => void
  onNodeUpdate?: (nodeId: string, updates: Partial<CanvasNode>) => void
  onConnectionCreate?: (connection: Omit<CanvasConnection, 'id' | 'canvasId' | 'createdAt' | 'updatedAt'>) => void
  onLayoutChange?: (layout: CanvasLayout) => void
}

export interface CanvasToolbarProps {
  canvasState: CanvasUIState
  onToolChange: (tool: CanvasUIState['toolMode']) => void
  onLayoutChange: (layoutType: CanvasLayout['type']) => void
  onZoomChange: (zoom: number) => void
  onToggleGrid: () => void
  onToggleMinimap: () => void
}

// Canvas integration with Knowledge microservice
export interface CanvasContentSourceLink {
  canvasNodeId: string
  contentSourceId: string
  linkType: 'represents' | 'references' | 'derived_from'
  createdAt: Date | string
}

// Error types specific to Canvas operations
export interface CanvasError {
  type: 'canvas_not_found' | 'node_not_found' | 'connection_invalid' | 'layout_failed' | 'permission_denied'
  message: string
  code: string
  canvasId?: string
  nodeId?: string
  connectionId?: string
}

// Constants for Canvas service
export const CANVAS_CONSTANTS = {
  MAX_NODES_PER_CANVAS: 1000,
  MAX_CONNECTIONS_PER_CANVAS: 5000,
  MIN_ZOOM: 0.1,
  MAX_ZOOM: 5.0,
  DEFAULT_NODE_SIZE: { width: 200, height: 100 },
  DEFAULT_VIEWPORT: { center: { x: 0, y: 0 }, zoom: 1.0 }
} as const

export type CanvasNodeType = CanvasNode['type']
export type CanvasConnectionType = CanvasConnection['type']
export type CanvasLayoutType = CanvasLayout['type']
export type CanvasToolMode = CanvasUIState['toolMode']