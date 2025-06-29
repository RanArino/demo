'use client';

import React, { useRef } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';
import { NODE_CONFIG } from './config';
import { MockNode } from './types';

export function Node({ 
  node, 
  isSelected, 
  isHovered, 
  onClick, 
  onPointerOver, 
  onPointerOut,
  focusedLayer
}: {
  node: MockNode;
  isSelected: boolean;
  isHovered: boolean;
  onClick: () => void;
  onPointerOver: () => void;
  onPointerOut: () => void;
  focusedLayer: string | null;
}) {
  const meshRef = useRef<THREE.Mesh>(null!);
  
  useFrame(() => {
    if (meshRef.current) {
      const baseScale = focusedLayer ? NODE_CONFIG.focusScaleMultiplier : 1;
      if (isSelected) {
        meshRef.current.scale.setScalar(baseScale * NODE_CONFIG.selectedScale);
      } else if (isHovered) {
        meshRef.current.scale.setScalar(baseScale * NODE_CONFIG.hoverScale);
      } else {
        meshRef.current.scale.setScalar(baseScale);
      }
    }
  });

  return (
    <mesh
      ref={meshRef}
      position={[node.position_3d.x, node.position_3d.y, node.position_3d.z]}
      onClick={(e) => { e.stopPropagation(); onClick(); }}
      onPointerOver={(e) => { e.stopPropagation(); onPointerOver(); }}
      onPointerOut={onPointerOut}
    >
      <sphereGeometry args={[NODE_CONFIG.defaultSize, 16, 16]} />
      <meshLambertMaterial 
        color={isSelected ? '#ff4444' : isHovered ? '#44ff44' : '#ffffff'}
        transparent
        opacity={0.9}
      />
    </mesh>
  );
} 