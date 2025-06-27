# 3D Document Vector Space Visualization - Consolidated Documentation

## 1. Overview

This document provides a comprehensive overview of the 3D Document Vector Space Visualization application. The application implements a 3D hierarchical document visualization system using React Three Fiber. It displays documents, clusters (sections), and text chunks in a three-layer vertical structure with interactive navigation and analysis features. The purpose of this application is to display three-layered hierarchical data consisting of documents, sections, and chunks in an interactive 3D space, allowing users to intuitively understand their relationships.

## 2. Functional Requirements

This section outlines the core functional requirements for the 3D visualization component.

### 2.1. Hierarchical Structure
Data is displayed in the following three hierarchical layers. Each layer is arranged vertically (Y-axis), and nodes are placed circularly on the XY plane of each layer. Note that "Sections" in the requirements correspond to "Clusters" in the implementation.

| Item               | Layer 2 (Documents) | Layer 1 (Clusters/Sections) | Layer 0 (Chunks) |
| :----------------- | :------------------ | :-------------------------- | :--------------- |
| **Y-coordinate**   | +120                | 0                           | -150             |
| **Circular Radius**| 70                  | 100                         | 120              |
| **Background Color**| Green (`#90EE90`)   | Purple (`#DDA0DD`)          | Blue (`#87CEEB`) |
| **Label**          | "Documents"         | "Sections"                  | "Chunks"         |

### 2.2. Scene Rotation
The rotation of the entire visualized 3D scene follows these specifications:
- **Rotation Axis**: The central Y-axis (vertical axis) of the scene.
- **Rotation Target**: All objects composing the scene (nodes, edges, background, etc.) rotate together as a single group.
- **Operation Method**: Horizontal mouse drag on the canvas.
- **Rotation Behavior**: During drag, rotation follows the amount of mouse movement.

### 2.3. Camera Control
- **Projection Method**: The camera must use an `OrthographicCamera` (parallel projection camera) to completely eliminate perspective distortion. This ensures all layers are always displayed in parallel.
- **Initial Viewpoint**: The camera is set to a position and zoom level from an oblique top-down view, where the entire scene is well-balanced.
- **Zoom**: Zoom in and out is possible by scrolling the mouse wheel.
- **Pan**: Panning the screen (parallel movement) is permitted to a limited extent.
- **Rotation**: User-controlled rotation of the camera itself (e.g., tilting) is not allowed.

### 2.4. Node Interaction
- **Click**:
    1. The scene automatically rotates around the Y-axis by the shortest distance so that the selected node is closest to the camera (front of the scene).
    2. Detailed information about the selected node is displayed in the side panel.
- **Hover**:
    1. Edges connected to the hovered node are highlighted.
    2. Brief information about the node (e.g., title) is displayed in a tooltip.

### 2.5. Edge Display
The display of edges indicating connections between nodes follows these rules:
- **Default State**:
    - **Within the same layer**: Edges are displayed as thin, semi-transparent grey (`#666666`) lines.
    - **Between different layers**: Edges are **hidden**.
- **Hover/Selected State**:
    - All edges connected to the target node (including inter-layer edges) are displayed in a highlight style: black (`#000000`), thick, and opaque.

### 2.6. UI/UX
- **Screen Layout**: The screen is divided into two parts:
    - **Left (65%)**: 3D visualization canvas.
    - **Right (35%)**: Side panel (with tabs for document list, node details, and chat function).
- **Operation Guide**: The following operation guide is always displayed on the screen:
    - Horizontal Drag: Scene Rotation
    - Scroll: Zoom
    - Node Click: Detail Display and Auto-Rotation
    - Node Hover: Connection Highlight

## 3. Architecture

### Core Technologies
- **React Three Fiber**: 3D rendering and scene management
- **Three.js**: Underlying 3D graphics engine
- **Shadcn/ui**: UI component library
- **React Resizable Panels**: Layout management
- **TypeScript**: Type safety and development experience

## 4. Project Structure

```
frontend/src/app/viz-test-1/
├── page.tsx                 # Main application page with layout
├── mock-data.ts            # Mock data generation and utilities
├── data_struct.md          # Database schema specification
├── requirement.md          # Feature requirements and specifications
└── components/
    ├── graph-3d/               # 3D visualization components
    │   ├── index.tsx           # Main canvas component
    │   ├── Node.tsx            # Node component
    │   ├── Edge.tsx            # Edge component
    │   └── ... (other components)
    ├── document-list-view.tsx  # Component to list and manage documents
    ├── node-detail-view.tsx    # Component to show details of a selected node
    └── chat-interface.tsx      # Component for user interaction via chat
```

## 5. Data Model

### Node Structure (`MockNode`)
Based on `data_struct.md`, nodes represent content entities with:

- **Identification**: `id`, `space_id`, `content_entity_id`
- **Hierarchy**: `parent_node_id`, `depth_level`, `path_to_root`
- **Positioning**: `position_2d`, `position_3d`, `is_position_locked`
- **Visualization**: `display_props` (size, opacity, shape, color)
- **Engagement**: `engagement_score` (canvas_score, chat_score, overall_score)
- **Content Data**: `action_data` with textual_data, location_references, asset_references

### Node Types
1. **Documents** (`content_source`): Layer 2 - Green spheres
2. **Clusters** (`chunk_cluster`): Layer 1 - Purple spheres  
3. **Chunks** (`content_chunk`): Layer 0 - Blue spheres

### Edge Structure (`MockEdge`)
Connections between nodes with:
- **Endpoints**: `start_node_id`, `end_node_id`
- **Styling**: `style_metadata` (line_type, weight, color, arrows)
- **Metadata**: description, creation/update tracking

## 6. Components

This section describes the key components of the application.

### 6.1. Main Page (`page.tsx`)
The root component orchestrating the entire application.

#### Key Features:
- **Resizable Panel Layout**: Three-panel design with 3D canvas, side panel, and details
- **State Management**: Manages nodes, edges, selections, and visibility
- **Layer Control**: Toggle visibility for documents/clusters/chunks
- **Document Filtering**: Show/hide documents and their related content

### 6.2. 3D Graph Canvas (`/components/canvas/graph-3d/`)
The core 3D visualization component implementing the hierarchical structure. It contains sub-components for Nodes, Edges, Layer Ellipses, and the Rotating Scene logic.

### 6.3. Document List View (`document-list-view.tsx`)
Document management interface providing CRUD operations.

### 6.4. Node Detail View (`node-detail-view.tsx`)
Comprehensive information display for selected nodes.

### 6.5. Chat Interface (`chat-interface.tsx`)
Simulated RAG (Retrieval-Augmented Generation) chat system.

## 7. Configuration
The application's key visual and interaction parameters are centrally managed in configuration objects for maintainability, as specified in the requirements.

```typescript
// Layer and layout configuration
const LAYER_CONFIG = {
  documents: { y: 120, radius: 70, color: '#90EE90', label: 'Documents' },
  clusters:  { y: 0,   radius: 100, color: '#DDA0DD', label: 'Sections' },
  chunks:    { y: -150,radius: 120, color: '#87CEEB', label: 'Chunks' }
};

// Camera initial setup using Orthographic projection
const CAMERA_CONFIG = {
  position: [350, 250, 350],
  zoom: 1.0, // Initial zoom level. Smaller value = wider view.
  target: [0, -10, 0]
};

// Interaction and animation settings
const ROTATION_CONFIG = {
  sensitivity: 0.01,       // Drag rotation sensitivity
  autoRotateSpeed: 0.1     // Click-to-focus animation speed
};

const NODE_CONFIG = {
  defaultSize: 15,
  hoverScale: 1.1,
  selectedScale: 1.3
};

const EDGE_CONFIG = {
  default:   { color: '#666666', lineWidth: 1, opacity: 0.4 },
  highlight: { color: '#000000', lineWidth: 2, opacity: 0.8 }
};
```

## 8. Out of Scope
The following items are outside the scope of the current requirements:
- User-adjustable camera tilt functionality.
- Use of a `PerspectiveCamera`.
- Persistent display of inter-layer edges in the default state. 