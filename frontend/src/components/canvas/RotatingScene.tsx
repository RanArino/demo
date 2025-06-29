'use client';

import React, { useRef, useState, useMemo, useEffect } from 'react';
import { useFrame, useThree } from '@react-three/fiber';
import * as THREE from 'three';
import { LayerEllipse } from './LayerEllipse';
import { CircularLayer } from './CircularLayer';
import { Edge } from './Edge';
import { MockNode, MockEdge } from './types';
import { LAYER_CONFIG, NODE_CONFIG, ROTATION_CONFIG } from './config';

export function RotatingScene({ 
  nodes, 
  edges, 
  selectedNodeId, 
  hoveredNodeId,
  onNodeClick, 
  onNodeHover,
  focusedLayer,
  hoveredLayer,
  onLayerClick,
  onLayerHover
}: {
  nodes: MockNode[];
  edges: MockEdge[];
  selectedNodeId?: string;
  hoveredNodeId: string | null;
  onNodeClick: (nodeId: string) => void;
  onNodeHover: (nodeId: string | null) => void;
  focusedLayer: string | null;
  hoveredLayer: string | null;
  onLayerClick: (layerKey: string) => void;
  onLayerHover: (layerKey: string | null) => void;
}) {
  const groupRef = useRef<THREE.Group>(null);
  const { gl, camera } = useThree();
  const [targetRotation, setTargetRotation] = useState(0);
  const currentRotation = useRef(0);

  // Calculate node positions once and memoize
  const allNodesWithPositions = useMemo(() => {
    const positionize = (nodeList: MockNode[], layerKey: keyof typeof LAYER_CONFIG) => {
      const config = LAYER_CONFIG[layerKey];
      const radius = config.radius;
      
      return nodeList.map((node, index) => {
        const angle = (index / nodeList.length) * Math.PI * 2;
        const x = Math.cos(angle) * radius;
        const z = Math.sin(angle) * radius;
        return { ...node, position_3d: { x, y: config.y, z } };
      });
    };
    
    return [
      ...positionize(nodes.filter(n => n.content_entity_type === 'content_source'), 'documents'),
      ...positionize(nodes.filter(n => n.content_entity_type === 'chunk_cluster'), 'clusters'),
      ...positionize(nodes.filter(n => n.content_entity_type === 'content_chunk'), 'chunks')
    ];
  }, [nodes]);

  const documentNodes = allNodesWithPositions.filter(n => n.content_entity_type === 'content_source');
  const clusterNodes = allNodesWithPositions.filter(n => n.content_entity_type === 'chunk_cluster');
  const chunkNodes = allNodesWithPositions.filter(n => n.content_entity_type === 'content_chunk');

  // Animate rotation towards target
  useFrame(() => {
    if (groupRef.current) {
      const currentY = groupRef.current.rotation.y;
      if (Math.abs(targetRotation - currentY) > 0.001) {
        const newRotation = THREE.MathUtils.lerp(currentY, targetRotation, ROTATION_CONFIG.autoRotateSpeed);
        groupRef.current.rotation.y = newRotation;
      }
    }
  });

  // Handle node click to center it
  const handleNodeClick = (nodeId: string) => {
    if (!focusedLayer) {
      const clickedNode = allNodesWithPositions.find(n => n.id === nodeId);
      if (clickedNode && groupRef.current) {
        // Calculate the angle of the camera in the XZ plane
        const cameraAngle = Math.atan2(camera.position.x, camera.position.z);
        
        // The node's current world angle is its local angle plus the group's rotation
        const nodeLocalAngle = Math.atan2(clickedNode.position_3d.x, clickedNode.position_3d.z);
        const currentGroupRotation = groupRef.current.rotation.y;
        
        // We want the node's final world angle to be the same as the camera's angle
        // targetGroupRotation + nodeLocalAngle = cameraAngle
        const targetGroupRotation = cameraAngle - nodeLocalAngle;
        
        // Find the shortest rotation path
        const twoPi = Math.PI * 2;
        let diff = (targetGroupRotation - currentGroupRotation) % twoPi;
        if (diff < -Math.PI) diff += twoPi;
        if (diff > Math.PI) diff -= twoPi;
        
        setTargetRotation(currentGroupRotation + diff);
      }
    }
    onNodeClick(nodeId);
  };

  return (
    <group ref={groupRef}>

      <LayerEllipse 
        layerKey="documents" 
        onLayerClick={onLayerClick}
        onLayerHover={onLayerHover}
        isHovered={hoveredLayer === 'documents'}
        isVisible={!focusedLayer || focusedLayer === 'documents'}
        focusedLayer={focusedLayer}
      />
      <LayerEllipse 
        layerKey="clusters" 
        onLayerClick={onLayerClick}
        onLayerHover={onLayerHover}
        isHovered={hoveredLayer === 'clusters'}
        isVisible={!focusedLayer || focusedLayer === 'clusters'}
        focusedLayer={focusedLayer}
      />
      <LayerEllipse 
        layerKey="chunks" 
        onLayerClick={onLayerClick}
        onLayerHover={onLayerHover}
        isHovered={hoveredLayer === 'chunks'}
        isVisible={!focusedLayer || focusedLayer === 'chunks'}
        focusedLayer={focusedLayer}
      />
      
      <CircularLayer
        nodes={documentNodes}
        selectedNodeId={selectedNodeId}
        hoveredNodeId={hoveredNodeId}
        onNodeClick={handleNodeClick}
        onNodeHover={onNodeHover}
        focusedLayer={focusedLayer}
        isVisible={!focusedLayer || focusedLayer === 'documents'}
      />
      <CircularLayer
        nodes={clusterNodes}
        selectedNodeId={selectedNodeId}
        hoveredNodeId={hoveredNodeId}
        onNodeClick={handleNodeClick}
        onNodeHover={onNodeHover}
        focusedLayer={focusedLayer}
        isVisible={!focusedLayer || focusedLayer === 'clusters'}
      />
      <CircularLayer
        nodes={chunkNodes}
        selectedNodeId={selectedNodeId}
        hoveredNodeId={hoveredNodeId}
        onNodeClick={handleNodeClick}
        onNodeHover={onNodeHover}
        focusedLayer={focusedLayer}
        isVisible={!focusedLayer || focusedLayer === 'chunks'}
      />
      
      {edges.map(edge => {
        const isHighlighted = hoveredNodeId ? 
          (edge.start_node_id === hoveredNodeId || edge.end_node_id === hoveredNodeId) : false;
        const isSelected = selectedNodeId ? 
          (edge.start_node_id === selectedNodeId || edge.end_node_id === selectedNodeId) : false;
        
        return (
          <Edge 
            key={edge.id} 
            edge={edge} 
            nodes={allNodesWithPositions}
            isHighlighted={isHighlighted || isSelected}
            hoveredNodeId={hoveredNodeId}
            focusedLayer={focusedLayer}
          />
        );
      })}
    </group>
  );
} 