'use client';

import React, { useState, useRef, useEffect } from 'react';
import { useFrame, useThree } from '@react-three/fiber';
import * as THREE from 'three';
import { CAMERA_CONFIG, LAYER_CONFIG, ANIMATION_CONFIG } from './config';
import { MockNode } from './types';

export function AnimatedCamera({ 
  focusedLayer,
  preFocusViewState,
  onAnimationComplete,
  controlsRef,
  nodes
}: {
  focusedLayer: string | null;
  preFocusViewState: { position: THREE.Vector3; target: THREE.Vector3; fov: number } | null;
  onAnimationComplete?: () => void;
  controlsRef: React.RefObject<any>;
  nodes: MockNode[];
}) {
  const { camera } = useThree();
  const [isAnimating, setIsAnimating] = useState(false);
  const animationStartTime = useRef<number | null>(null);
  const startState = useRef<{ position: THREE.Vector3; target: THREE.Vector3; fov: number; }>({
    position: new THREE.Vector3(),
    target: new THREE.Vector3(),
    fov: 50
  });
  const previousFocusState = useRef<string | null>(null);

  useFrame(() => {
    if (!isAnimating || !animationStartTime.current) return;

    const elapsed = Date.now() - animationStartTime.current;
    const progress = Math.min(elapsed / ANIMATION_CONFIG.cameraTransitionDuration, 1);
    
    // Easing function (ease-in-out)
    const easeProgress = progress < 0.5 
      ? 4 * progress * progress * progress
      : 1 - Math.pow(-2 * progress + 2, 3) / 2;

    let targetPosition: THREE.Vector3;
    let targetLookAt: THREE.Vector3;
    let targetFov: number;

    if (focusedLayer) {
      const layerConfig = LAYER_CONFIG[focusedLayer as keyof typeof LAYER_CONFIG];
      const distance = layerConfig.radius * 2.5; // Adjust distance based on layer size
      targetPosition = new THREE.Vector3(0, layerConfig.y, distance);
      targetLookAt = new THREE.Vector3(0, layerConfig.y, 0);
      targetFov = 60;
    } else {
      if (preFocusViewState) {
        targetPosition = preFocusViewState.position.clone();
        targetLookAt = preFocusViewState.target.clone();
        targetFov = preFocusViewState.fov;
      } else {
        targetPosition = new THREE.Vector3(...CAMERA_CONFIG.position);
        targetLookAt = new THREE.Vector3(...CAMERA_CONFIG.target);
        targetFov = 50;
      }
    }

    // Interpolate camera properties
    const currentPosition = new THREE.Vector3().lerpVectors(startState.current.position, targetPosition, easeProgress);
    const currentTarget = new THREE.Vector3().lerpVectors(startState.current.target, targetLookAt, easeProgress);
    const currentFov = THREE.MathUtils.lerp(startState.current.fov, targetFov, easeProgress);

    camera.position.copy(currentPosition);
    (camera as THREE.PerspectiveCamera).fov = currentFov;
    camera.updateProjectionMatrix();

    if (controlsRef.current) {
      controlsRef.current.target.copy(currentTarget);
      controlsRef.current.update();
    }

    if (progress >= 1) {
      setIsAnimating(false);
      animationStartTime.current = null;
      previousFocusState.current = focusedLayer;
      
      // Ensure final state is set
      camera.position.copy(targetPosition);
      (camera as THREE.PerspectiveCamera).fov = targetFov;
      camera.updateProjectionMatrix();
      
      if (controlsRef.current) {
        controlsRef.current.target.copy(targetLookAt);
        controlsRef.current.update();
      }
      
      onAnimationComplete?.();
    }
  });

  useEffect(() => {
    if (focusedLayer !== previousFocusState.current) {
      const cam = camera as THREE.PerspectiveCamera;
      startState.current = {
        position: cam.position.clone(),
        target: controlsRef.current ? controlsRef.current.target.clone() : new THREE.Vector3(),
        fov: cam.fov,
      };
 
      animationStartTime.current = Date.now();
      setIsAnimating(true);
    }
  }, [focusedLayer, camera, controlsRef]);

  return null;
} 