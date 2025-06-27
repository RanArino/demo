'use client';

import React from 'react';
import { Node } from './Node';
import { MockNode } from './types';

export function CircularLayer({ 
  nodes, 
  selectedNodeId, 
  hoveredNodeId,
  onNodeClick, 
  onNodeHover,
  focusedLayer,
  isVisible = true
}: {
  nodes: MockNode[];
  selectedNodeId?: string;
  hoveredNodeId: string | null;
  onNodeClick: (nodeId: string) => void;
  onNodeHover: (nodeId: string | null) => void;
  focusedLayer: string | null;
  isVisible?: boolean;
}) {
  if (!isVisible) return null;

  return (
    <>
      {nodes.map((node) => (
        <Node
          key={node.id}
          node={node}
          isSelected={selectedNodeId === node.id}
          isHovered={hoveredNodeId === node.id}
          onClick={() => onNodeClick(node.id)}
          onPointerOver={() => onNodeHover(node.id)}
          onPointerOut={() => onNodeHover(null)}
          focusedLayer={focusedLayer}
        />
      ))}
    </>
  );
} 