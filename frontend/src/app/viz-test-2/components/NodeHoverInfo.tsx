'use client';

import React from 'react';
import { MockNode } from '../mock-data';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';
import { 
  FileText,
  Layers,
  Hash,
  Tag,
  MapPin
} from 'lucide-react';

interface NodeHoverInfoProps {
  node: MockNode;
  position: { x: number; y: number };
}

export function NodeHoverInfo({ node, position }: NodeHoverInfoProps) {
  if (!node) return null;

  const getNodeTypeInfo = (contentEntityType: string) => {
    switch (contentEntityType) {
      case 'content_source':
        return { icon: FileText, label: 'Document', color: 'text-green-600', bgColor: 'bg-green-50' };
      case 'chunk_cluster':
        return { icon: Layers, label: 'Cluster', color: 'text-purple-600', bgColor: 'bg-purple-50' };
      case 'content_chunk':
        return { icon: Hash, label: 'Chunk', color: 'text-yellow-600', bgColor: 'bg-yellow-50' };
      default:
        return { icon: FileText, label: 'Unknown', color: 'text-gray-600', bgColor: 'bg-gray-50' };
    }
  };

  const typeInfo = getNodeTypeInfo(node.content_entity_type);
  const TypeIcon = typeInfo.icon;

  return (
    <div
      className="absolute bg-white border rounded-lg shadow-xl text-sm pointer-events-none z-50"
      style={{
        left: `${position.x + 20}px`,
        top: `${position.y + 20}px`,
        maxWidth: '320px',
        minWidth: '280px',
      }}
    >
      <Card className="border-0 shadow-none">
        <CardHeader className="pb-2">
          <div className="flex items-center gap-2 mb-2">
            <div className={`p-1.5 rounded-md ${typeInfo.bgColor}`}>
              <TypeIcon className={`w-4 h-4 ${typeInfo.color}`} />
            </div>
            <div className="flex-1">
              <CardTitle className="text-sm font-medium truncate">
                {node.action_data.textual_data.title}
              </CardTitle>
              <div className="flex items-center gap-2 text-xs text-gray-500">
                <span className={`${typeInfo.color} font-medium`}>{typeInfo.label}</span>
                <span>•</span>
                <span>Layer {node.depth_level}</span>
              </div>
            </div>
          </div>
        </CardHeader>
        
        <Separator />
        
        <CardContent className="pt-3 space-y-3">
          {/* Summary */}
          <div>
            <h4 className="text-xs font-semibold text-gray-700 mb-1">Summary</h4>
            <p className="text-xs text-gray-600 line-clamp-3">
              {node.action_data.textual_data.summary}
            </p>
          </div>

          {/* Keywords */}
          {node.action_data.textual_data.keywords?.length > 0 && (
            <div>
              <h4 className="text-xs font-semibold text-gray-700 mb-1 flex items-center gap-1">
                <Tag className="w-3 h-3" />
                Keywords
              </h4>
              <div className="flex flex-wrap gap-1">
                {node.action_data.textual_data.keywords.slice(0, 4).map((keyword, i) => (
                  <span key={i} className="px-2 py-0.5 bg-blue-100 text-blue-800 text-xs rounded-full">
                    {keyword}
                  </span>
                ))}
                {node.action_data.textual_data.keywords.length > 4 && (
                  <span className="text-xs text-gray-500">
                    +{node.action_data.textual_data.keywords.length - 4} more
                  </span>
                )}
              </div>
            </div>
          )}

          {/* Position (for testing) */}
          <div>
            <h4 className="text-xs font-semibold text-gray-700 mb-1 flex items-center gap-1">
              <MapPin className="w-3 h-3" />
              Position
            </h4>
            <div className="grid grid-cols-3 gap-2 text-xs">
              <div>
                <span className="text-gray-500">X:</span>
                <span className="ml-1 font-mono">{node.position_3d.x.toFixed(3)}</span>
              </div>
              <div>
                <span className="text-gray-500">Y:</span>
                <span className="ml-1 font-mono">{node.position_3d.y.toFixed(3)}</span>
              </div>
              <div>
                <span className="text-gray-500">Z:</span>
                <span className="ml-1 font-mono">{node.position_3d.z.toFixed(3)}</span>
              </div>
            </div>
          </div>

          {/* Metadata */}
          {node.action_data.general_metadata.file_type && (
            <div>
              <h4 className="text-xs font-semibold text-gray-700 mb-1">Type</h4>
              <span className="text-xs text-gray-600 uppercase">
                {node.action_data.general_metadata.file_type}
              </span>
              {node.action_data.general_metadata.word_count && (
                <span className="text-xs text-gray-500 ml-2">
                  • {node.action_data.general_metadata.word_count.toLocaleString()} words
                </span>
              )}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
} 