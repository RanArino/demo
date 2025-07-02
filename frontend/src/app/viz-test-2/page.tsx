'use client';

import dynamic from 'next/dynamic';

const Canvas3D = dynamic(() => import('./components/Canvas3D'), { ssr: false });

export default function VizTest2Page() {
  return (
    <div className="relative h-screen w-screen">
      <Canvas3D />
    </div>
  );
}