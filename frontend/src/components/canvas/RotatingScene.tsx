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
  const [isDragging, setIsDragging] = useState(false);
  const [previousMouseX, setPreviousMouseX] = useState(0);
  const [targetRotation, setTargetRotation] = useState(0);
  const [currentRotation, setCurrentRotation] = useState(0);

  // Calculate node positions once and memoize
  const allNodesWithPositions = useMemo(() => {
    const positionize = (nodeList: MockNode[], layerKey: keyof typeof LAYER_CONFIG) => {
      const config = LAYER_CONFIG[layerKey];
      const radius = focusedLayer === layerKey ? config.radius * NODE_CONFIG.focusScaleMultiplier : config.radius;
      
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
  }, [nodes, focusedLayer]);

  const documentNodes = allNodesWithPositions.filter(n => n.content_entity_type === 'content_source');
  const clusterNodes = allNodesWithPositions.filter(n => n.content_entity_type === 'chunk_cluster');
  const chunkNodes = allNodesWithPositions.filter(n => n.content_entity_type === 'content_chunk');

  // Animate rotation towards target (disabled in focus mode)
  useFrame(() => {
    if (!isDragging && !focusedLayer && groupRef.current && Math.abs(targetRotation - currentRotation) > 0.001) {
      const newRotation = THREE.MathUtils.lerp(currentRotation, targetRotation, ROTATION_CONFIG.autoRotateSpeed);
      groupRef.current.rotation.y = newRotation;
      setCurrentRotation(newRotation);
    }
  });

  // Handle node click for auto-rotation (disabled in focus mode)
  const handleNodeClick = (nodeId: string) => {
    if (!focusedLayer) {
      const clickedNode = allNodesWithPositions.find(n => n.id === nodeId);
      if (clickedNode && groupRef.current) {
        const frontAngle = Math.atan2(camera.position.x, camera.position.z);
        const nodeAngle = Math.atan2(clickedNode.position_3d.x, clickedNode.position_3d.z);
        
        let finalTarget = frontAngle - nodeAngle;
        
        const twoPi = Math.PI * 2;
        let diff = (finalTarget - groupRef.current.rotation.y) % twoPi;
        if (diff < -Math.PI) diff += twoPi;
        if (diff > Math.PI) diff -= twoPi;

        setTargetRotation(groupRef.current.rotation.y + diff);
      }
    }
    onNodeClick(nodeId);
  };

  useEffect(() => {
    // Disable drag controls in focus mode
    if (focusedLayer) return;

    const handleMouseDown = (event: MouseEvent) => {
      setIsDragging(true);
      setPreviousMouseX(event.clientX);
    };

    const handleMouseMove = (event: MouseEvent) => {
      if (!isDragging || !groupRef.current) return;
      const deltaX = event.clientX - previousMouseX;
      const newRotation = currentRotation + deltaX * ROTATION_CONFIG.sensitivity;
      
      setCurrentRotation(newRotation);
      setTargetRotation(newRotation);
      groupRef.current.rotation.y = newRotation;
      
      setPreviousMouseX(event.clientX);
    };

    const handleMouseUp = () => {
      setIsDragging(false);
    };
    
    const domElement = gl.domElement;
    domElement.addEventListener('mousedown', handleMouseDown);
    window.addEventListener('mousemove', handleMouseMove);
    window.addEventListener('mouseup', handleMouseUp);

    return () => {
      domElement.removeEventListener('mousedown', handleMouseDown);
      window.removeEventListener('mousemove', handleMouseMove);
      window.removeEventListener('mouseup', handleMouseUp);
    };
  }, [gl.domElement, isDragging, previousMouseX, currentRotation, focusedLayer]);

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