import { MockNode } from './types';
import { CAMERA_CONFIG } from './config';

// Helper function to get node layer
export function getNodeLayer(node: MockNode): string {
  switch (node.content_entity_type) {
    case 'content_source': return 'documents';
    case 'chunk_cluster': return 'clusters';
    case 'content_chunk': return 'chunks';
    default: return 'unknown';
  }
}

// Helper function to calculate dynamic zoom based on node count
export function calculateDynamicZoom(nodeCount: number, layerRadius: number): number {
  // Base calculations for optimal zoom
  const baseZoom = CAMERA_CONFIG.baseFocusZoom;
  
  // Calculate zoom based on node count and layer radius
  // More nodes = need to zoom out more (smaller zoom value)
  // Fewer nodes = can zoom in more (larger zoom value)
  
  if (nodeCount <= 5) {
    // Few nodes (documents) - zoom in more
    return baseZoom * 1.5;
  } else if (nodeCount <= 10) {
    // Medium nodes (clusters) - standard zoom
    return baseZoom;
  } else {
    // Many nodes (chunks) - zoom out more
    const scaleFactor = Math.max(0.3, 1 - (nodeCount - 10) * 0.05);
    return baseZoom * scaleFactor;
  }
} 