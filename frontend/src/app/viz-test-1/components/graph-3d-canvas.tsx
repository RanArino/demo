'use client';

import React, { useRef, useState, useEffect, useMemo } from 'react';
import { Canvas, useFrame, useThree } from '@react-three/fiber';
import { Text, Line, OrbitControls, OrthographicCamera } from '@react-three/drei';
import * as THREE from 'three';
import { MockNode, MockEdge } from '../mock-data';

interface Graph3DCanvasProps {
  nodes: MockNode[];
  edges: MockEdge[];
  selectedNodeId?: string;
  onNodeClick: (nodeId: string) => void;
  onNodeHover: (nodeId: string | null) => void;
}

// 設定可能パラメータ
const LAYER_CONFIG = {
  documents: { y: 120, radius: 70, color: '#90EE90', label: 'ドキュメント' },
  clusters: { y: 0, radius: 100, color: '#DDA0DD', label: 'セクション' },
  chunks: { y: -150, radius: 120, color: '#87CEEB', label: 'チャンク' }
};

const CAMERA_CONFIG = {
  position: [350, 250, 350] as [number, number, number],
  zoom: 1, // Default zoom level, smaller is wider.
  target: [0, -10, 0] as [number, number, number],
  // Focus view camera settings
  focusPosition: [0, 300, 0] as [number, number, number],
  baseFocusZoom: 1, // Base zoom level for focus mode
  focusTarget: [0, 0, 0] as [number, number, number]
};

const ROTATION_CONFIG = {
  sensitivity: 0.01,
  autoRotateSpeed: 0.1
};

const NODE_CONFIG = {
  defaultSize: 15,
  hoverScale: 1.1,
  selectedScale: 1.3,
  // Focus view scaling
  focusScaleMultiplier: 2.5
};

const EDGE_CONFIG = {
  default: { color: '#666666', lineWidth: 1, opacity: 0.4 },
  highlight: { color: '#000000', lineWidth: 2, opacity: 0.8 }
};

const ANIMATION_CONFIG = {
  cameraTransitionDuration: 1500, // ms
  hoverDelay: 500 // ms
};

// Node component
function Node({ 
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

// Edge component with visibility rules
function Edge({ 
  edge, 
  nodes, 
  isHighlighted,
  hoveredNodeId,
  focusedLayer
}: { 
  edge: MockEdge; 
  nodes: MockNode[];
  isHighlighted?: boolean;
  hoveredNodeId: string | null;
  focusedLayer: string | null;
}) {
  const startNode = nodes.find(n => n.id === edge.start_node_id);
  const endNode = nodes.find(n => n.id === edge.end_node_id);
  
  if (!startNode || !endNode) return null;

  const startLayer = getNodeLayer(startNode);
  const endLayer = getNodeLayer(endNode);
  const isCrossLayer = startLayer !== endLayer;
  
  // Only show edges when hovering over a node
  const isRelatedToHover = hoveredNodeId && (edge.start_node_id === hoveredNodeId || edge.end_node_id === hoveredNodeId);
  
  if (!isRelatedToHover) {
    return null;
  }

  // In focus mode, only show edges within the focused layer
  if (focusedLayer) {
    if (startLayer !== focusedLayer || endLayer !== focusedLayer) {
      return null;
    }
  }

  const points = [
    new THREE.Vector3(startNode.position_3d.x, startNode.position_3d.y, startNode.position_3d.z),
    new THREE.Vector3(endNode.position_3d.x, endNode.position_3d.y, endNode.position_3d.z),
  ];

  const { color, lineWidth, opacity } = isHighlighted
    ? EDGE_CONFIG.highlight
    : EDGE_CONFIG.default;

  return (
    <Line
      points={points}
      color={color}
      lineWidth={lineWidth}
      transparent
      opacity={opacity}
    />
  );
}

// Helper function to get node layer
function getNodeLayer(node: MockNode): string {
  switch (node.content_entity_type) {
    case 'content_source': return 'documents';
    case 'chunk_cluster': return 'clusters';
    case 'content_chunk': return 'chunks';
    default: return 'unknown';
  }
}

// Layer ellipse background with interaction
function LayerEllipse({ 
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

// "Dumb" component for rendering pre-positioned nodes
function CircularLayer({ 
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

// Helper function to calculate dynamic zoom based on node count
function calculateDynamicZoom(nodeCount: number, layerRadius: number): number {
  // Base calculations for optimal zoom
  const baseZoom = CAMERA_CONFIG.baseFocusZoom;
  
  // Calculate zoom based on node count and layer radius
  // More nodes = need to zoom out more (smaller zoom value)
  // Fewer nodes = can zoom in more (larger zoom value)
  
  if (nodeCount <= 5) {
    // Few nodes (documents) - zoom in more
    return baseZoom * 1.5;
  } else if (nodeCount <= 10) {
    // Medium nodes (clusters) - standard zoom
    return baseZoom;
  } else {
    // Many nodes (chunks) - zoom out more
    const scaleFactor = Math.max(0.3, 1 - (nodeCount - 10) * 0.05);
    return baseZoom * scaleFactor;
  }
}

// Camera animation component
function AnimatedCamera({ 
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

// Rotating scene group with auto-rotation on click
function RotatingScene({ 
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

// Back button component
function BackButton({ onBack, isVisible }: { onBack: () => void; isVisible: boolean }) {
  if (!isVisible) return null;

  return (
    <div className="absolute top-4 left-4 z-10">
      <button
        onClick={onBack}
        className="px-4 py-2 bg-blue-500 text-white rounded-lg shadow-lg hover:bg-blue-600 transition-colors"
      >
        ← 戻る
      </button>
    </div>
  );
}

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
