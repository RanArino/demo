'use client';

import React, { useState, useEffect } from 'react';
import { Panel, PanelGroup, PanelResizeHandle } from 'react-resizable-panels';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { 
  Settings, 
  Info, 
  Maximize2, 
  RotateCcw,
  Eye,
  EyeOff,
  Layers
} from 'lucide-react';

// Import components
import Graph3DCanvas from '@/components/canvas';
import DocumentListView from './components/document-list-view';
import NodeDetailView from './components/node-detail-view';
import ChatInterface from './components/chat-interface';

// Import mock data
import { mockNodes, mockEdges, MockNode, MockEdge } from './mock-data';

export default function VizTestPage() {
  const [nodes, setNodes] = useState<MockNode[]>(mockNodes);
  const [edges, setEdges] = useState<MockEdge[]>(mockEdges);
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const [hoveredNodeId, setHoveredNodeId] = useState<string | null>(null);
  const [visibleDocuments, setVisibleDocuments] = useState<Set<string>>(
    new Set(mockNodes.filter(node => node.content_entity_type === 'content_source').map(node => node.id))
  );
  const [activeTab, setActiveTab] = useState('documents');
  const [layerVisibility, setLayerVisibility] = useState({
    documents: true,
    clusters: true,
    chunks: true,
  });

  // Helper function to find parent document
  const findParentDocument = (node: MockNode): MockNode | null => {
    if (node.content_entity_type === 'content_source') {
      return node;
    }
    
    // Simple logic: assume first part of node ID indicates document relation
    const docNodes = nodes.filter(n => n.content_entity_type === 'content_source');
    return docNodes.find(doc => 
      node.action_data.location_references.document_id === doc.id
    ) || null;
  };

  // Filter nodes based on visible documents and layer visibility
  const filteredNodes = nodes.filter(node => {
    // Check layer visibility
    const layerMap = {
      'content_source': 'documents',
      'chunk_cluster': 'clusters', 
      'content_chunk': 'chunks',
    };
    
    const layerKey = layerMap[node.content_entity_type as keyof typeof layerMap];
    if (!layerVisibility[layerKey as keyof typeof layerVisibility]) {
      return false;
    }

    // For document nodes, check if they're in visible documents
    if (node.content_entity_type === 'content_source') {
      return visibleDocuments.has(node.id);
    }
    
    // For clusters and chunks, check if their parent document is visible
    const parentDoc = findParentDocument(node);
    return parentDoc ? visibleDocuments.has(parentDoc.id) : true;
  });

  // Filter edges based on visible nodes
  const filteredEdges = edges.filter(edge => {
    const startNode = filteredNodes.find(n => n.id === edge.start_node_id);
    const endNode = filteredNodes.find(n => n.id === edge.end_node_id);
    return startNode && endNode;
  });

  const handleNodeClick = (nodeId: string) => {
    setSelectedNodeId(nodeId);
    setActiveTab('details');
  };

  const handleNodeHover = (nodeId: string | null) => {
    setHoveredNodeId(nodeId);
  };

  const handleDocumentToggle = (documentId: string) => {
    setVisibleDocuments(prev => {
      const newSet = new Set(prev);
      if (newSet.has(documentId)) {
        newSet.delete(documentId);
      } else {
        newSet.add(documentId);
      }
      return newSet;
    });
  };

  const handleDocumentAdd = (title: string, content: string) => {
    const newDocId = `doc-${Date.now()}`;
    const newDoc: MockNode = {
      id: newDocId,
      space_id: 'space-1',
      content_entity_id: newDocId,
      content_entity_type: 'content_source',
      depth_level: 2,
      path_to_root: `/${newDocId}`,
      position_2d: { x: Math.random() * 200 - 100, y: Math.random() * 200 - 100 },
      position_3d: { x: Math.random() * 200 - 100, y: Math.random() * 200 - 100, z: 200 },
      is_position_locked: false,
      visibility: true,
      display_props: {
        size: 15,
        opacity: 0.9,
        shape: 'sphere',
        color: '#90EE90',
      },
      engagement_score: {
        canvas_score: 0.5,
        chat_score: 0.5,
        overall_score: 0.5,
      },
      action_data: {
        textual_data: {
          title,
          summary: content,
          quote: `"${title}" - New document`,
          keywords: title.toLowerCase().split(' '),
        },
        location_references: {
          document_id: newDocId,
          page_number: 1,
          scroll_hint_percent: 0,
          context_heading: 'Introduction',
        },
        asset_references: {
          s3_key: `documents/${newDocId}.pdf`,
        },
        display_configuration: {
          highlight_style: {
            style: 'document_boundary',
            color: '#e6ffe6',
          },
        },
        general_metadata: {
          file_type: 'pdf',
          page_count: 1,
          word_count: content.split(' ').length,
          estimated_read_time_mins: Math.ceil(content.split(' ').length / 200),
        },
      },
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };

    setNodes(prev => [...prev, newDoc]);
    setVisibleDocuments(prev => new Set([...prev, newDocId]));
  };

  const handleDocumentDelete = (documentId: string) => {
    setNodes(prev => prev.filter(node => 
      node.id !== documentId && 
      node.action_data.location_references.document_id !== documentId
    ));
    setVisibleDocuments(prev => {
      const newSet = new Set(prev);
      newSet.delete(documentId);
      return newSet;
    });
    
    if (selectedNodeId === documentId) {
      setSelectedNodeId(null);
    }
  };

  const resetCamera = () => {
    // This would reset the camera in the 3D canvas
    console.log('Reset camera');
  };

  const toggleLayerVisibility = (layer: keyof typeof layerVisibility) => {
    setLayerVisibility(prev => ({
      ...prev,
      [layer]: !prev[layer]
    }));
  };

  return (
    <div className="h-screen w-full bg-gray-50">
      {/* Header */}
      <div className="h-16 bg-white border-b flex items-center justify-between px-6">
        <div className="flex items-center gap-4">
          <h1 className="text-xl font-bold">Document Vector Space Visualization</h1>
          <div className="flex items-center gap-2">
            <span className="text-sm text-gray-500">
              {filteredNodes.length} nodes • {filteredEdges.length} edges
            </span>
          </div>
        </div>
        
        <div className="flex items-center gap-2">
          {/* Layer visibility toggles */}
          <div className="flex items-center gap-1 mr-4">
            <Button
              variant={layerVisibility.documents ? "default" : "outline"}
              size="sm"
              onClick={() => toggleLayerVisibility('documents')}
              className="text-xs"
            >
              <Layers className="w-3 h-3 mr-1" />
              Docs
            </Button>
            <Button
              variant={layerVisibility.clusters ? "default" : "outline"}
              size="sm"
              onClick={() => toggleLayerVisibility('clusters')}
              className="text-xs"
            >
              <Layers className="w-3 h-3 mr-1" />
              Clusters
            </Button>
            <Button
              variant={layerVisibility.chunks ? "default" : "outline"}
              size="sm"
              onClick={() => toggleLayerVisibility('chunks')}
              className="text-xs"
            >
              <Layers className="w-3 h-3 mr-1" />
              Chunks
            </Button>
          </div>
          
          <Button variant="outline" size="sm" onClick={resetCamera}>
            <RotateCcw className="w-4 h-4" />
          </Button>
          <Button variant="outline" size="sm">
            <Settings className="w-4 h-4" />
          </Button>
          <Button variant="outline" size="sm">
            <Info className="w-4 h-4" />
          </Button>
        </div>
      </div>

      {/* Main content */}
      <div className="h-[calc(100vh-4rem)]">
        <PanelGroup direction="horizontal">
          {/* 3D Visualization Panel */}
          <Panel defaultSize={65} minSize={40}>
            <div className="h-full relative">
              <Graph3DCanvas
                nodes={filteredNodes}
                edges={filteredEdges}
                selectedNodeId={selectedNodeId || undefined}
                onNodeClick={handleNodeClick}
                onNodeHover={handleNodeHover}
              />
              
              {/* Overlay controls */}
              <div className="absolute top-4 left-4 bg-white/90 backdrop-blur-sm rounded-lg p-3 shadow-lg">
                <h3 className="text-sm font-medium mb-2">Controls</h3>
                <div className="text-xs text-gray-600 space-y-1">
                  <div>• Drag horizontally to rotate scene</div>
                  <div>• Scroll to zoom</div>
                  <div>• Click nodes for details</div>
                  <div>• Hover to highlight connections</div>
                </div>
              </div>

              {/* Layer legend */}
              <div className="absolute bottom-4 left-4 bg-white/90 backdrop-blur-sm rounded-lg p-3 shadow-lg">
                <h3 className="text-sm font-medium mb-2">Legend</h3>
                <div className="space-y-2 text-xs">
                  <div className="flex items-center gap-2">
                    <div className="w-3 h-3 rounded-full bg-green-400"></div>
                    <span>Documents (Layer 2)</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="w-3 h-3 rounded-full bg-purple-400"></div>
                    <span>Clusters (Layer 1)</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="w-3 h-3 rounded-full bg-yellow-400"></div>
                    <span>Chunks (Layer 0)</span>
                  </div>
                </div>
              </div>

              {/* Hovered node info */}
              {hoveredNodeId && (
                <div className="absolute top-4 right-4 bg-white/90 backdrop-blur-sm rounded-lg p-3 shadow-lg max-w-xs">
                  <div className="text-sm font-medium">
                    {nodes.find(n => n.id === hoveredNodeId)?.action_data.textual_data.title}
                  </div>
                  <div className="text-xs text-gray-600 mt-1">
                    {nodes.find(n => n.id === hoveredNodeId)?.content_entity_type.replace('_', ' ')}
                  </div>
                </div>
              )}
            </div>
          </Panel>

          <PanelResizeHandle className="w-2 bg-gray-200 hover:bg-gray-300 transition-colors" />

          {/* Side Panel */}
          <Panel defaultSize={35} minSize={25}>
            <div className="h-full bg-white">
              <Tabs value={activeTab} onValueChange={setActiveTab} className="h-full">
                <div className="border-b">
                  <TabsList className="grid w-full grid-cols-3">
                    <TabsTrigger value="documents">Documents</TabsTrigger>
                    <TabsTrigger value="details">Details</TabsTrigger>
                    <TabsTrigger value="chat">Chat</TabsTrigger>
                  </TabsList>
                </div>

                <TabsContent value="documents" className="h-[calc(100%-3rem)] m-0">
                  <DocumentListView
                    nodes={nodes}
                    onDocumentToggle={handleDocumentToggle}
                    onDocumentAdd={handleDocumentAdd}
                    onDocumentDelete={handleDocumentDelete}
                    visibleDocuments={visibleDocuments}
                  />
                </TabsContent>

                <TabsContent value="details" className="h-[calc(100%-3rem)] m-0">
                  <NodeDetailView
                    selectedNodeId={selectedNodeId}
                    onBack={() => setActiveTab('documents')}
                    onNodeClick={handleNodeClick}
                  />
                </TabsContent>

                <TabsContent value="chat" className="h-[calc(100%-3rem)] m-0">
                  <ChatInterface
                    nodes={filteredNodes}
                    selectedNodeId={selectedNodeId}
                  />
                </TabsContent>
              </Tabs>
            </div>
          </Panel>
        </PanelGroup>
      </div>
    </div>
  );
}
