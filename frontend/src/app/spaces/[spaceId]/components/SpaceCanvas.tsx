'use client';

import { useState, useRef, useEffect } from 'react';
import { Space } from '@/app/spaces/types/spaces';
import { ContentSource } from '@/app/spaces/types/content';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ZoomIn, ZoomOut, RotateCcw, Plus, FileText, Link, Type, Upload } from 'lucide-react';
import { cn } from '@/lib/utils';

interface SpaceCanvasProps {
  space: Space;
  contentSources: ContentSource[];
  className?: string;
}

interface CanvasNode {
  id: string;
  x: number;
  y: number;
  width: number;
  height: number;
  type: 'document' | 'note' | 'link';
  title: string;
  content?: string;
  contentSourceId?: string;
}

export default function SpaceCanvas({ space, contentSources, className }: SpaceCanvasProps) {
  const canvasRef = useRef<HTMLDivElement>(null);
  const [zoom, setZoom] = useState(1);
  const [pan, setPan] = useState({ x: 0, y: 0 });
  const [isDragging, setIsDragging] = useState(false);
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 });
  const [nodes, setNodes] = useState<CanvasNode[]>([]);

  // Initialize nodes from content sources
  useEffect(() => {
    const initialNodes: CanvasNode[] = contentSources.map((source, index) => ({
      id: source.id,
      x: 100 + (index % 3) * 250,
      y: 100 + Math.floor(index / 3) * 200,
      width: 200,
      height: 150,
      type: 'document',
      title: source.title || source.filename || 'Untitled Document',
      contentSourceId: source.id,
    }));

    setNodes(initialNodes);
  }, [contentSources]);

  const handleZoomIn = () => {
    setZoom(prev => Math.min(prev * 1.2, 3));
  };

  const handleZoomOut = () => {
    setZoom(prev => Math.max(prev / 1.2, 0.3));
  };

  const handleResetView = () => {
    setZoom(1);
    setPan({ x: 0, y: 0 });
  };

  const handleMouseDown = (e: React.MouseEvent) => {
    if (e.target === canvasRef.current) {
      setIsDragging(true);
      setDragStart({ x: e.clientX - pan.x, y: e.clientY - pan.y });
    }
  };

  const handleMouseMove = (e: React.MouseEvent) => {
    if (isDragging) {
      setPan({
        x: e.clientX - dragStart.x,
        y: e.clientY - dragStart.y,
      });
    }
  };

  const handleMouseUp = () => {
    setIsDragging(false);
  };

  const getNodeIcon = (type: CanvasNode['type']) => {
    switch (type) {
      case 'document': return <FileText className="h-5 w-5" />;
      case 'link': return <Link className="h-5 w-5" />;
      case 'note': return <Type className="h-5 w-5" />;
      default: return <FileText className="h-5 w-5" />;
    }
  };

  const getNodeColor = (type: CanvasNode['type']) => {
    switch (type) {
      case 'document': return 'bg-blue-50 border-blue-200 text-blue-900';
      case 'link': return 'bg-green-50 border-green-200 text-green-900';
      case 'note': return 'bg-yellow-50 border-yellow-200 text-yellow-900';
      default: return 'bg-gray-50 border-gray-200 text-gray-900';
    }
  };

  const handleAddContent = () => {
    // TODO: Open upload modal
    console.log('Add content to space:', space.id);
  };

  return (
    <div className={cn("flex flex-col h-full bg-gray-50", className)}>
      {/* Canvas Toolbar */}
      <div className="flex items-center justify-between p-4 bg-white border-b border-gray-200">
        <div className="flex items-center gap-2">
          <h2 className="text-lg font-semibold text-gray-900">Canvas View</h2>
          <Badge variant="outline" className="text-xs">
            {nodes.length} items
          </Badge>
        </div>

        <div className="flex items-center gap-2">
          {/* Zoom Controls */}
          <div className="flex items-center gap-1 border border-gray-200 rounded-md">
            <Button
              variant="ghost"
              size="sm"
              onClick={handleZoomOut}
              className="h-8 w-8 p-0"
            >
              <ZoomOut className="h-4 w-4" />
            </Button>
            <span className="px-2 text-sm font-medium min-w-[3rem] text-center">
              {Math.round(zoom * 100)}%
            </span>
            <Button
              variant="ghost"
              size="sm"
              onClick={handleZoomIn}
              className="h-8 w-8 p-0"
            >
              <ZoomIn className="h-4 w-4" />
            </Button>
          </div>

          <Button
            variant="ghost"
            size="sm"
            onClick={handleResetView}
            className="flex items-center gap-2"
          >
            <RotateCcw className="h-4 w-4" />
            Reset
          </Button>

          <Button
            onClick={handleAddContent}
            size="sm"
            className="flex items-center gap-2"
          >
            <Plus className="h-4 w-4" />
            Add Content
          </Button>
        </div>
      </div>

      {/* Canvas Area */}
      <div className="flex-1 relative overflow-hidden">
        <div
          ref={canvasRef}
          className="w-full h-full cursor-grab active:cursor-grabbing"
          onMouseDown={handleMouseDown}
          onMouseMove={handleMouseMove}
          onMouseUp={handleMouseUp}
          onMouseLeave={handleMouseUp}
          style={{
            transform: `translate(${pan.x}px, ${pan.y}px) scale(${zoom})`,
            transformOrigin: '0 0',
          }}
        >
          {/* Grid Background */}
          <div 
            className="absolute inset-0 opacity-20"
            style={{
              backgroundImage: `
                linear-gradient(to right, #e5e7eb 1px, transparent 1px),
                linear-gradient(to bottom, #e5e7eb 1px, transparent 1px)
              `,
              backgroundSize: '20px 20px',
            }}
          />

          {/* Canvas Nodes */}
          {nodes.map((node) => (
            <div
              key={node.id}
              className={cn(
                "absolute border-2 rounded-lg shadow-sm cursor-pointer transition-all duration-200 hover:shadow-md hover:scale-105",
                getNodeColor(node.type)
              )}
              style={{
                left: node.x,
                top: node.y,
                width: node.width,
                height: node.height,
              }}
            >
              <div className="p-4 h-full flex flex-col">
                {/* Node Header */}
                <div className="flex items-center gap-2 mb-2">
                  {getNodeIcon(node.type)}
                  <span className="text-xs font-medium uppercase tracking-wide opacity-70">
                    {node.type}
                  </span>
                </div>

                {/* Node Title */}
                <h3 className="font-semibold text-sm line-clamp-2 mb-2">
                  {node.title}
                </h3>

                {/* Node Content Preview */}
                {node.content && (
                  <p className="text-xs opacity-70 line-clamp-3 flex-1">
                    {node.content}
                  </p>
                )}

                {/* Processing Status for Documents */}
                {node.type === 'document' && node.contentSourceId && (
                  <div className="mt-auto">
                    {(() => {
                      const source = contentSources.find(s => s.id === node.contentSourceId);
                      if (!source) return null;
                      
                      const statusColors = {
                        pending: 'bg-yellow-100 text-yellow-800',
                        processing: 'bg-blue-100 text-blue-800',
                        completed: 'bg-green-100 text-green-800',
                        failed: 'bg-red-100 text-red-800',
                      };

                      return (
                        <Badge 
                          variant="outline" 
                          className={`text-xs ${statusColors[source.processingStatus]}`}
                        >
                          {source.processingStatus}
                        </Badge>
                      );
                    })()}
                  </div>
                )}
              </div>
            </div>
          ))}

          {/* Empty State */}
          {nodes.length === 0 && (
            <div className="absolute inset-0 flex items-center justify-center">
              <div className="text-center max-w-md mx-auto p-8">
                <div className="w-16 h-16 mx-auto mb-4 bg-gray-100 rounded-full flex items-center justify-center">
                  <Upload className="h-8 w-8 text-gray-400" />
                </div>
                <h3 className="text-lg font-semibold text-gray-900 mb-2">
                  No content yet
                </h3>
                <p className="text-gray-600 mb-4">
                  Start building your knowledge space by adding documents, notes, or links.
                </p>
                <Button onClick={handleAddContent} className="flex items-center gap-2">
                  <Plus className="h-4 w-4" />
                  Add Your First Content
                </Button>
              </div>
            </div>
          )}
        </div>

        {/* Mini-map (placeholder) */}
        <div className="absolute bottom-4 right-4 w-32 h-24 bg-white border border-gray-200 rounded-lg shadow-sm opacity-80">
          <div className="p-2 text-xs text-gray-500 text-center">
            Mini-map
            <br />
            <span className="text-gray-400">(Coming soon)</span>
          </div>
        </div>
      </div>
    </div>
  );
}