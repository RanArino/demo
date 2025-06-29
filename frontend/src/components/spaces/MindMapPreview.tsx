'use client';

import React, { useEffect, useRef } from 'react';

interface MindMapPreviewProps {
  keywords: string[];
  width: number;
  height: number;
}

const MindMapPreview: React.FC<MindMapPreviewProps> = ({ keywords, width, height }) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    ctx.clearRect(0, 0, width, height);

    const centerX = width / 2;
    const centerY = height / 2;
    const radius = Math.min(width, height) / 3;

    keywords.forEach((keyword, index) => {
      const angle = (index / keywords.length) * 2 * Math.PI;
      const x = centerX + radius * Math.cos(angle);
      const y = centerY + radius * Math.sin(angle);

      ctx.beginPath();
      ctx.arc(x, y, 5, 0, 2 * Math.PI);
      ctx.fillStyle = 'blue';
      ctx.fill();

      ctx.font = '12px Arial';
      ctx.textAlign = 'center';
      ctx.fillText(keyword, x, y + 15);
    });
  }, [keywords, width, height]);

  return <canvas ref={canvasRef} width={width} height={height} />;
};

export default MindMapPreview;