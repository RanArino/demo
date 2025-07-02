'use client';

import React, { useState } from 'react';
import { Canvas } from '@react-three/fiber';
import { OrbitControls } from '@react-three/drei';
import { mockNodes, MockNode } from '../mock-data';
import { Node } from './Node';
import { NodeHoverInfo } from './NodeHoverInfo';

export default function Canvas3D() {
  const [hoveredNode, setHoveredNode] = useState<{ node: MockNode; position: { x: number, y: number } } | null>(null);

  const handleNodeHover = (node: MockNode | null, position: { x: number; y: number } | null) => {
    if (node && position) {
      setHoveredNode({ node, position });
    } else {
      setHoveredNode(null);
    }
  };

  return (
    <>
      <Canvas
        camera={{
          position: [0, 0, 5],
          fov: 60
        }}
      >
        <ambientLight intensity={0.5} />
        <pointLight position={[10, 10, 10]} />
        
        {mockNodes.map((node) => (
          <Node key={node.id} node={node} onHover={handleNodeHover} />
        ))}
        
        <OrbitControls 
          enablePan={true}
          enableZoom={true}
          enableRotate={true}
        />
      </Canvas>
      {hoveredNode && <NodeHoverInfo node={hoveredNode.node} position={hoveredNode.position} />}
    </>
  );
}