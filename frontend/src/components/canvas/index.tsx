'use client';

import React, { useRef, useState } from 'react';
import { Canvas } from '@react-three/fiber';
import { OrbitControls, PerspectiveCamera } from '@react-three/drei';
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
  const [preFocusViewState, setPreFocusViewState] = useState<{ position: THREE.Vector3; target: THREE.Vector3; fov: number } | null>(null);
  const controlsRef = useRef<any>(null);

  const handleLayerClick = (layerKey: string) => {
    if (focusedLayer === layerKey) return;

    if (!focusedLayer && controlsRef.current) {
      const camera = controlsRef.current.object as THREE.PerspectiveCamera;
      setPreFocusViewState({
        position: camera.position.clone(),
        target: controlsRef.current.target.clone(),
        fov: camera.fov,
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
        style={{ background: 'linear-gradient(to bottom, #111111, #222222)' }}
      >
        {/* Use Perspective Camera for a cinematic, 3D feel */}
        <PerspectiveCamera 
          makeDefault
          position={CAMERA_CONFIG.position}
          fov={50} // Field of view, defines the extent of the scene that is seen
          near={0.1}
          far={5000}
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

        <ambientLight intensity={0.5} />
        <directionalLight position={[10, 10, 5]} intensity={1} />
        <pointLight position={[0, -200, 0]} intensity={2} color="#5555ff" />
        <spotLight position={[300, 300, 300]} angle={0.3} penumbra={1} intensity={2} castShadow />
        
        <OrbitControls 
          ref={controlsRef}
          makeDefault
          enableRotate={!focusedLayer} // Allow rotation only in overview
          enablePan={!focusedLayer}
          enableZoom={true}
          zoomSpeed={0.8}
          target={CAMERA_CONFIG.target}
          minDistance={100} // Set min zoom distance
          maxDistance={1200} // Set max zoom distance
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
      </Canvas>
    </div>
  );
} 