'use client';

import React, { useRef } from 'react';
import { Text } from '@react-three/drei';
import * as THREE from 'three';
import { LAYER_CONFIG, NODE_CONFIG, ANIMATION_CONFIG } from './config';

export function LayerEllipse({ 
  layerKey,
  onLayerClick,
  onLayerHover,
  isHovered,
  isVisible = true,
  focusedLayer
}: { 
  layerKey: keyof typeof LAYER_CONFIG;
  onLayerClick: (layerKey: string) => void;
  onLayerHover: (layerKey: string | null) => void;
  isHovered: boolean;
  isVisible?: boolean;
  focusedLayer: string | null;
}) {
  const config = LAYER_CONFIG[layerKey];
  const isFocused = focusedLayer === layerKey;
  const scale = isFocused ? NODE_CONFIG.focusScaleMultiplier : 1;
  const ellipseRadius = config.radius * 1.5 * scale;
  const hoverTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  if (!isVisible) return null;

  const handlePointerEnter = () => {
    hoverTimeoutRef.current = setTimeout(() => {
      onLayerHover(layerKey);
    }, ANIMATION_CONFIG.hoverDelay);
  };

  const handlePointerLeave = () => {
    if (hoverTimeoutRef.current) {
      clearTimeout(hoverTimeoutRef.current);
      hoverTimeoutRef.current = null;
    }
    onLayerHover(null);
  };

  const handleClick = (e: any) => {
    e.stopPropagation();
    onLayerClick(layerKey);
  };

  return (
    <group position={[0, config.y, 0]}>
      {/* Background: A flat circle that appears as an ellipse */}
      <mesh 
        rotation={[-Math.PI / 2, 0, 0]}
        onClick={handleClick}
        onPointerEnter={handlePointerEnter}
        onPointerLeave={handlePointerLeave}
      >
        <circleGeometry args={[ellipseRadius, 64]} />
        <meshBasicMaterial 
          color={config.color}
          transparent
          opacity={isHovered ? 0.4 : 0.2}
          side={THREE.DoubleSide}
        />
      </mesh>
      
      {/* Border: A thin ring to outline the ellipse */}
      <mesh rotation={[-Math.PI / 2, 0, 0]}>
        <ringGeometry args={[ellipseRadius * 0.97, ellipseRadius, 64]} />
        <meshBasicMaterial 
          color={config.color}
          transparent
          opacity={isHovered ? 0.9 : 0.6}
          side={THREE.DoubleSide}
        />
      </mesh>
      
      <Text
        position={[0, 10, 0]}
        fontSize={16}
        color="#333333"
        anchorX="center"
        anchorY="middle"
      >
        {config.label}
      </Text>
    </group>
  );
} 