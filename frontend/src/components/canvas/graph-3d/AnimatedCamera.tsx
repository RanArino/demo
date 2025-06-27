'use client';

import React, { useState, useRef, useEffect } from 'react';
import { useFrame, useThree } from '@react-three/fiber';
import * as THREE from 'three';
import { CAMERA_CONFIG, LAYER_CONFIG, ANIMATION_CONFIG } from './config';
import { calculateDynamicZoom } from './helpers';
import { MockNode } from './types';

export function AnimatedCamera({ 
  focusedLayer,
  preFocusViewState,
  onAnimationComplete,
  controlsRef,
  nodes
}: {
  focusedLayer: string | null;
  preFocusViewState: { position: THREE.Vector3; target: THREE.Vector3; zoom: number } | null;
  onAnimationComplete?: () => void;
  controlsRef: React.RefObject<any>;
  nodes: MockNode[];
}) {
  const { camera } = useThree();
  const [isAnimating, setIsAnimating] = useState(false);
  const animationStartTime = useRef<number | null>(null);
  const startPosition = useRef<THREE.Vector3>(new THREE.Vector3());
  const startTarget = useRef<THREE.Vector3>(new THREE.Vector3());
  const startZoom = useRef<number>(1);
  const previousFocusState = useRef<string | null>(null);

  useFrame(() => {
    if (!isAnimating || !animationStartTime.current) return;

    const elapsed = Date.now() - animationStartTime.current;
    const progress = Math.min(elapsed / ANIMATION_CONFIG.cameraTransitionDuration, 1);
    
    // Easing function (ease-in-out)
    const easeProgress = progress < 0.5 
      ? 2 * progress * progress 
      : 1 - Math.pow(-2 * progress + 2, 3) / 2;

    let targetPosition: THREE.Vector3;
    let targetLookAt: THREE.Vector3;
    let targetZoom: number;

    if (focusedLayer) {
      const layerConfig = LAYER_CONFIG[focusedLayer as keyof typeof LAYER_CONFIG];
      targetPosition = new THREE.Vector3(...CAMERA_CONFIG.focusPosition);
      targetPosition.y = layerConfig.y + 200; // Position above the focused layer
      targetLookAt = new THREE.Vector3(0, layerConfig.y, 0);
      targetZoom = calculateDynamicZoom(nodes.length, layerConfig.radius);
    } else {
      // When returning from focus mode, use the saved state
      if (preFocusViewState) {
        targetPosition = preFocusViewState.position.clone();
        targetLookAt = preFocusViewState.target.clone();
        targetZoom = preFocusViewState.zoom;
      } else {
        // Fallback to default position
        targetPosition = new THREE.Vector3(...CAMERA_CONFIG.position);
        targetLookAt = new THREE.Vector3(...CAMERA_CONFIG.target);
        targetZoom = CAMERA_CONFIG.zoom;
      }
    }

    // Arc trajectory - create a curved path
    const midPoint = new THREE.Vector3().lerpVectors(startPosition.current, targetPosition, 0.5);
    midPoint.y += 100; // Raise the midpoint to create an arc
    
    let currentPosition: THREE.Vector3;
    if (progress < 0.5) {
      currentPosition = new THREE.Vector3().lerpVectors(startPosition.current, midPoint, easeProgress * 2);
    } else {
      currentPosition = new THREE.Vector3().lerpVectors(midPoint, targetPosition, (easeProgress - 0.5) * 2);
    }

    const currentTarget = new THREE.Vector3().lerpVectors(startTarget.current, targetLookAt, easeProgress);
    const currentZoom = THREE.MathUtils.lerp(startZoom.current, targetZoom, easeProgress);

    camera.position.copy(currentPosition);
    (camera as THREE.OrthographicCamera).zoom = currentZoom;
    camera.updateProjectionMatrix();

    // Update OrbitControls target during animation
    if (controlsRef.current) {
      controlsRef.current.target.copy(currentTarget);
      controlsRef.current.update();
    }

    if (progress >= 1) {
      setIsAnimating(false);
      animationStartTime.current = null;
      previousFocusState.current = focusedLayer;
      
      // Ensure final camera position is set correctly
      camera.position.copy(targetPosition);
      (camera as THREE.OrthographicCamera).zoom = targetZoom;
      camera.updateProjectionMatrix();
      
      if (controlsRef.current) {
        controlsRef.current.target.copy(targetLookAt);
        controlsRef.current.update();
      }
      
      onAnimationComplete?.();
    }
  });

  useEffect(() => {
    // Only start animation when focus state actually changes
    if (focusedLayer !== previousFocusState.current) {
      // Start animation
      const cam = camera as THREE.OrthographicCamera;
      startPosition.current.copy(camera.position);
      startTarget.current.copy(new THREE.Vector3(...CAMERA_CONFIG.target));
      startZoom.current = cam.zoom;
 
      animationStartTime.current = Date.now();
      setIsAnimating(true);
    }
  }, [focusedLayer, camera]);

  return null;
} 