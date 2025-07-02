'use client';

import React, { useState } from 'react';
import { MockNode } from '../mock-data';
import { ThreeEvent } from '@react-three/fiber';

// CONFIGURATION
const USE_SPHERICAL_TRANSFORMATION = true; // Set to false for cylindrical projection

// Z-baseline positions for each layer
const LAYER_Z_BASELINES = {
  document: 0,    // Layer 2: content_source (documents)
  cluster: -2,    // Layer 1: chunk_cluster (sections)
  chunk: -4       // Layer 0: content_chunk (chunks)
} as const;

// Transform square coordinates to cylinder coordinates
const transformSquareToCylinder = (x: number, y: number, z: number) => {
  const u = x * Math.sqrt(1 - (y * y) / 2);
  const v = y * Math.sqrt(1 - (x * x) / 2);
  
  return { x: u, y: v, z: z };
};

// Transform cube coordinates to spherical coordinates (volumetric)
const transformCubeToSphere = (x: number, y: number, z: number) => {
  const x2 = x * x;
  const y2 = y * y;
  const z2 = z * z;

  const u = x * Math.sqrt(1 - y2 / 2 - z2 / 2 + (y2 * z2) / 3);
  const v = y * Math.sqrt(1 - z2 / 2 - x2 / 2 + (z2 * x2) / 3);
  const w = z * Math.sqrt(1 - x2 / 2 - y2 / 2 + (x2 * y2) / 3);

  return { x: u, y: v, z: w };
};

// Get layer baseline z position based on content type
const getLayerBaseline = (contentEntityType: string): number => {
  switch (contentEntityType) {
    case 'content_source': return LAYER_Z_BASELINES.document;
    case 'chunk_cluster': return LAYER_Z_BASELINES.cluster;
    case 'content_chunk': return LAYER_Z_BASELINES.chunk;
    default: return 0;
  }
};

interface NodeProps {
  node: MockNode;
  onHover: (node: MockNode | null, position: { x: number; y: number } | null) => void;
}

export function Node({ node, onHover }: NodeProps) {
  const [isHovered, setIsHovered] = useState(false);

  const layerBaseline = getLayerBaseline(node.content_entity_type);
  let transformedPosition;

  if (USE_SPHERICAL_TRANSFORMATION) {
    // For spherical, transform with normalized [-1, 1] coords FIRST, then apply layer offset
    const sphericalPosition = transformCubeToSphere(
      node.position_3d.x,
      node.position_3d.y,
      node.position_3d.z
    );
    transformedPosition = {
      x: sphericalPosition.x,
      y: sphericalPosition.y,
      z: sphericalPosition.z + layerBaseline,
    };
  } else {
    // For cylindrical, the z-coordinate is independent
    const cylindricalPosition = transformSquareToCylinder(
      node.position_3d.x,
      node.position_3d.y,
      node.position_3d.z
    );
    transformedPosition = {
      x: cylindricalPosition.x,
      y: cylindricalPosition.y,
      z: cylindricalPosition.z + layerBaseline,
    };
  }

  const handlePointerOver = (e: ThreeEvent<PointerEvent>) => {
    e.stopPropagation();
    setIsHovered(true);
    document.body.style.cursor = 'pointer';
    onHover(node, { x: e.clientX, y: e.clientY });
  };

  const handlePointerOut = () => {
    setIsHovered(false);
    document.body.style.cursor = 'auto';
    onHover(null, null);
  };

  // Color based on content type
  const getNodeColor = () => {
    switch (node.content_entity_type) {
      case 'content_source': return '#90EE90'; // Light green for documents
      case 'chunk_cluster': return '#DDA0DD'; // Plum for clusters
      case 'content_chunk': return '#FFFFE0'; // Light yellow for chunks
      default: return '#FFFFFF';
    }
  };

  return (
    <mesh 
      position={[transformedPosition.x, transformedPosition.y, transformedPosition.z]}
      onPointerOver={handlePointerOver}
      onPointerOut={handlePointerOut}
    >
      <sphereGeometry args={[isHovered ? 0.07 : 0.05, 16, 16]} />
      <meshLambertMaterial 
        color={getNodeColor()}
        transparent
        opacity={isHovered ? 1.0 : 0.8}
      />
    </mesh>
  );
}