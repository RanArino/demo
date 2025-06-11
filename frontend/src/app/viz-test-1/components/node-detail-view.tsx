'use client';

import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { 
  ArrowLeft,
  FileText,
  Layers,
  Hash,
  MapPin,
  Clock,
  Eye,
  Download,
  Share,
  BookOpen,
  Tag,
  Link as LinkIcon
} from 'lucide-react';
import { MockNode, getNodeById, getChildNodes, getEdgesForNode } from '../mock-data';

interface NodeDetailViewProps {
  selectedNodeId?: string | null;
  onBack: () => void;
  onNodeClick: (nodeId: string) => void;
}

export default function NodeDetailView({ 
  selectedNodeId, 
  onBack, 
  onNodeClick 
}: NodeDetailViewProps) {
  if (!selectedNodeId) {
    return (
      <div className="h-full flex flex-col items-center justify-center text-gray-500">
        <FileText className="w-16 h-16 mb-4 opacity-30" />
        <p className="text-lg font-medium">No node selected</p>
        <p className="text-sm">Click on a node in the visualization to see details</p>
      </div>
    );
  }

  const node = getNodeById(selectedNodeId);
  if (!node) {
    return (
      <div className="h-full flex flex-col items-center justify-center text-red-500">
        <FileText className="w-16 h-16 mb-4 opacity-30" />
        <p className="text-lg font-medium">Node not found</p>
        <p className="text-sm">The selected node could not be found</p>
      </div>
    );
  }

  const childNodes = getChildNodes(node.id);
  const edges = getEdgesForNode(node.id);
  
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
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="p-4 border-b">
        <div className="flex items-center gap-3 mb-4">
          <Button variant="ghost" size="sm" onClick={onBack}>
            <ArrowLeft className="w-4 h-4" />
          </Button>
          <div className={`p-2 rounded-lg ${typeInfo.bgColor}`}>
            <TypeIcon className={`w-5 h-5 ${typeInfo.color}`} />
          </div>
          <div className="flex-1">
            <h2 className="text-lg font-semibold truncate">
              {node.action_data.textual_data.title}
            </h2>
            <div className="flex items-center gap-2 text-sm text-gray-500">
              <span className={`${typeInfo.color} font-medium`}>{typeInfo.label}</span>
              <span>•</span>
              <span>Layer {node.depth_level}</span>
            </div>
          </div>
        </div>

        {/* Action buttons */}
        <div className="flex gap-2">
          <Button variant="outline" size="sm" className="flex-1">
            <Eye className="w-4 h-4 mr-2" />
            View
          </Button>
          <Button variant="outline" size="sm" className="flex-1">
            <Download className="w-4 h-4 mr-2" />
            Download
          </Button>
          <Button variant="outline" size="sm" className="flex-1">
            <Share className="w-4 h-4 mr-2" />
            Share
          </Button>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-4 space-y-6">
        {/* Summary */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <BookOpen className="w-4 h-4" />
              Summary
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-gray-700 leading-relaxed">
              {node.action_data.textual_data.summary}
            </p>
            {node.action_data.textual_data.quote && (
              <blockquote className="mt-4 pl-4 border-l-4 border-blue-200 italic text-sm text-gray-600">
                {node.action_data.textual_data.quote}
              </blockquote>
            )}
          </CardContent>
        </Card>

        {/* Keywords & Topics */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <Tag className="w-4 h-4" />
              Keywords & Topics
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <div>
                <h4 className="text-sm font-medium text-gray-600 mb-2">Keywords</h4>
                <div className="flex flex-wrap gap-2">
                  {node.action_data.textual_data.keywords.map((keyword, index) => (
                    <span
                      key={index}
                      className="px-3 py-1 bg-blue-100 text-blue-800 text-xs rounded-full"
                    >
                      {keyword}
                    </span>
                  ))}
                </div>
              </div>
              
              {node.action_data.textual_data.cluster_topics && (
                <div>
                  <h4 className="text-sm font-medium text-gray-600 mb-2">Cluster Topics</h4>
                  <div className="flex flex-wrap gap-2">
                    {node.action_data.textual_data.cluster_topics.map((topic, index) => (
                      <span
                        key={index}
                        className="px-3 py-1 bg-purple-100 text-purple-800 text-xs rounded-full"
                      >
                        {topic}
                      </span>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Location & Context */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <MapPin className="w-4 h-4" />
              Location & Context
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <span className="text-xs font-medium text-gray-500">Document ID</span>
                  <p className="text-sm">{node.action_data.location_references.document_id}</p>
                </div>
                {node.action_data.location_references.page_number && (
                  <div>
                    <span className="text-xs font-medium text-gray-500">Page</span>
                    <p className="text-sm">{node.action_data.location_references.page_number}</p>
                  </div>
                )}
              </div>
              
              {node.action_data.location_references.context_heading && (
                <div>
                  <span className="text-xs font-medium text-gray-500">Context</span>
                  <p className="text-sm">{node.action_data.location_references.context_heading}</p>
                </div>
              )}
              
              {node.action_data.location_references.char_range && (
                <div>
                  <span className="text-xs font-medium text-gray-500">Character Range</span>
                  <p className="text-sm">
                    {node.action_data.location_references.char_range.start} - {node.action_data.location_references.char_range.end}
                  </p>
                </div>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Metadata */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <Clock className="w-4 h-4" />
              Metadata
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <span className="text-xs font-medium text-gray-500">File Type</span>
                <p>{node.action_data.general_metadata.file_type?.toUpperCase()}</p>
              </div>
              {node.action_data.general_metadata.word_count && (
                <div>
                  <span className="text-xs font-medium text-gray-500">Word Count</span>
                  <p>{node.action_data.general_metadata.word_count?.toLocaleString()}</p>
                </div>
              )}
              {node.action_data.general_metadata.page_count && (
                <div>
                  <span className="text-xs font-medium text-gray-500">Page Count</span>
                  <p>{node.action_data.general_metadata.page_count}</p>
                </div>
              )}
              {node.action_data.general_metadata.estimated_read_time_mins && (
                <div>
                  <span className="text-xs font-medium text-gray-500">Read Time</span>
                  <p>{node.action_data.general_metadata.estimated_read_time_mins} min</p>
                </div>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Engagement Score */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Engagement Metrics</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div>
                <div className="flex justify-between text-sm mb-1">
                  <span>Canvas Score</span>
                  <span>{Math.round(node.engagement_score.canvas_score * 100)}%</span>
                </div>
                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div 
                    className="bg-blue-600 h-2 rounded-full"
                    style={{ width: `${node.engagement_score.canvas_score * 100}%` }}
                  />
                </div>
              </div>
              
              <div>
                <div className="flex justify-between text-sm mb-1">
                  <span>Chat Score</span>
                  <span>{Math.round(node.engagement_score.chat_score * 100)}%</span>
                </div>
                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div 
                    className="bg-green-600 h-2 rounded-full"
                    style={{ width: `${node.engagement_score.chat_score * 100}%` }}
                  />
                </div>
              </div>
              
              <div>
                <div className="flex justify-between text-sm mb-1">
                  <span>Overall Score</span>
                  <span className="font-medium">{Math.round(node.engagement_score.overall_score * 100)}%</span>
                </div>
                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div 
                    className="bg-purple-600 h-2 rounded-full"
                    style={{ width: `${node.engagement_score.overall_score * 100}%` }}
                  />
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Connections */}
        {(childNodes.length > 0 || edges.length > 0) && (
          <Card>
            <CardHeader>
              <CardTitle className="text-base flex items-center gap-2">
                <LinkIcon className="w-4 h-4" />
                Connections
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {childNodes.length > 0 && (
                  <div>
                    <h4 className="text-sm font-medium text-gray-600 mb-2">
                      Child Nodes ({childNodes.length})
                    </h4>
                    <div className="space-y-2">
                      {childNodes.slice(0, 5).map((child) => (
                        <div 
                          key={child.id}
                          className="flex items-center gap-2 p-2 bg-gray-50 rounded-md cursor-pointer hover:bg-gray-100"
                          onClick={() => onNodeClick(child.id)}
                        >
                          <div className={`w-2 h-2 rounded-full ${
                            child.content_entity_type === 'content_chunk' ? 'bg-yellow-400' : 'bg-purple-400'
                          }`} />
                          <span className="text-sm flex-1">{child.action_data.textual_data.title}</span>
                        </div>
                      ))}
                      {childNodes.length > 5 && (
                        <p className="text-xs text-gray-500">
                          +{childNodes.length - 5} more children
                        </p>
                      )}
                    </div>
                  </div>
                )}
                
                {edges.length > 0 && (
                  <div>
                    <h4 className="text-sm font-medium text-gray-600 mb-2">
                      Connected Edges ({edges.length})
                    </h4>
                    <div className="space-y-2">
                      {edges.slice(0, 5).map((edge) => (
                        <div key={edge.id} className="text-xs text-gray-600 p-2 bg-gray-50 rounded">
                          <span className="font-medium">{edge.description}</span>
                          <span className="ml-2">({edge.style_metadata.line_type})</span>
                        </div>
                      ))}
                      {edges.length > 5 && (
                        <p className="text-xs text-gray-500">
                          +{edges.length - 5} more connections
                        </p>
                      )}
                    </div>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
} 