'use client';

import React, { useRef, useState } from 'react';
import { Canvas } from '@react-three/fiber';
import { OrbitControls, OrthographicCamera } from '@react-three/drei';
import * as THREE from 'three';
import { Graph3DCanvasProps } from './types';
import { CAMERA_CONFIG } from './config';
import { BackButton } from './BackButton';
import { AnimatedCamera } from './AnimatedCamera';
import { RotatingScene } from './RotatingScene';
import { getNodeLayer } from './helpers';

// Main 3D Graph Canvas component
export default function Graph3DCanvas({ 
  nodes, 
  edges, 
  selectedNodeId, 
  onNodeClick, 
  onNodeHover 
}: Graph3DCanvasProps) {
  const [hoveredNodeId, setHoveredNodeId] = useState<string | null>(null);
  const [focusedLayer, setFocusedLayer] = useState<string | null>(null);
  const [hoveredLayer, setHoveredLayer] = useState<string | null>(null);
  const [preFocusViewState, setPreFocusViewState] = useState<{ position: THREE.Vector3; target: THREE.Vector3; zoom: number } | null>(null);
  const controlsRef = useRef<any>(null);

  const handleLayerClick = (layerKey: string) => {
    if (focusedLayer === layerKey) return;

    if (!focusedLayer && controlsRef.current) {
      const camera = controlsRef.current.object as THREE.OrthographicCamera;
      setPreFocusViewState({
        position: camera.position.clone(),
        target: controlsRef.current.target.clone(),
        zoom: camera.zoom,
      });
    }
    setFocusedLayer(layerKey);
    setHoveredLayer(null);
  };

  const handleBackClick = () => {
    setFocusedLayer(null);
    setHoveredLayer(null);
  };

  return (
    <div className="w-full h-full cursor-grab active:cursor-grabbing relative">
      <BackButton onBack={handleBackClick} isVisible={!!focusedLayer} />
      
      <Canvas
        style={{ background: 'linear-gradient(to bottom, #f0f0f0, #e0e0e0)' }}
      >
        {/* Use Orthographic Camera for a parallel projection, eliminating perspective distortion */}
        <OrthographicCamera 
          makeDefault
          position={CAMERA_CONFIG.position}
          zoom={CAMERA_CONFIG.zoom}
          near={1}
          far={2000}
        />

        <AnimatedCamera 
          focusedLayer={focusedLayer}
          preFocusViewState={preFocusViewState}
          controlsRef={controlsRef}
          nodes={focusedLayer ? nodes.filter(n => getNodeLayer(n) === focusedLayer) : []}
          onAnimationComplete={() => {
            // After animation finishes, if we are back to multi-layer view, clear the state
            if (!focusedLayer) {
              setPreFocusViewState(null);
            }
          }}
        />

        <ambientLight intensity={0.8} />
        <directionalLight position={[100, 200, 100]} intensity={0.6} />
        <pointLight position={[0, 150, 200]} intensity={0.4} />
        
        <OrbitControls 
          ref={controlsRef}
          makeDefault
          enableRotate={false}
          enablePan={!focusedLayer}
          enableZoom={true}
          zoomSpeed={0.5}
          target={CAMERA_CONFIG.target}
        />
        
        <RotatingScene
          nodes={nodes}
          edges={edges}
          selectedNodeId={selectedNodeId}
          hoveredNodeId={hoveredNodeId}
          onNodeClick={onNodeClick}
          onNodeHover={setHoveredNodeId}
          focusedLayer={focusedLayer}
          hoveredLayer={hoveredLayer}
          onLayerClick={handleLayerClick}
          onLayerHover={setHoveredLayer}
        />
        
        {!focusedLayer && (
          <gridHelper 
            args={[400, 20, '#cccccc', '#dddddd']} 
            position={[0, -200, 0]} 
          />
        )}
      </Canvas>
    </div>
  );
} 