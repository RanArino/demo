'use client';

import React from 'react';
import { Line } from '@react-three/drei';
import * as THREE from 'three';
import { EDGE_CONFIG } from './config';
import { MockEdge, MockNode } from './types';
import { getNodeLayer } from './helpers';

export function Edge({ 
  edge, 
  nodes, 
  isHighlighted,
  hoveredNodeId,
  focusedLayer
}: { 
  edge: MockEdge; 
  nodes: MockNode[];
  isHighlighted?: boolean;
  hoveredNodeId: string | null;
  focusedLayer: string | null;
}) {
  const startNode = nodes.find(n => n.id === edge.start_node_id);
  const endNode = nodes.find(n => n.id === edge.end_node_id);
  
  if (!startNode || !endNode) return null;

  const startLayer = getNodeLayer(startNode);
  const endLayer = getNodeLayer(endNode);
  
  // Only show edges when hovering over a node
  const isRelatedToHover = hoveredNodeId && (edge.start_node_id === hoveredNodeId || edge.end_node_id === hoveredNodeId);
  
  if (!isRelatedToHover) {
    return null;
  }

  // In focus mode, only show edges within the focused layer
  if (focusedLayer) {
    if (startLayer !== focusedLayer || endLayer !== focusedLayer) {
      return null;
    }
  }

  const points = [
    new THREE.Vector3(startNode.position_3d.x, startNode.position_3d.y, startNode.position_3d.z),
    new THREE.Vector3(endNode.position_3d.x, endNode.position_3d.y, endNode.position_3d.z),
  ];

  const { color, lineWidth, opacity } = isHighlighted
    ? EDGE_CONFIG.highlight
    : EDGE_CONFIG.default;

  return (
    <Line
      points={points}
      color={color}
      lineWidth={lineWidth}
      transparent
      opacity={opacity}
    />
  );
} 