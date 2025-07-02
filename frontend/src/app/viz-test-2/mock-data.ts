// Mock data for the 3D document visualization
// Based on the data structure defined in data_struct.md

export interface MockNode {
  id: string;
  space_id: string;
  content_entity_id: string;
  content_entity_type: 'content_chunk' | 'chunk_cluster' | 'content_source' | 'source_cluster';
  parent_node_id?: string;
  depth_level: number;
  path_to_root: string;
  position_2d: { x: number; y: number };
  position_3d: { x: number; y: number; z: number };
  is_position_locked: boolean;
  visibility: boolean;
  display_props: {
    size: number;
    opacity: number;
    shape: string;
    color: string;
  };
  engagement_score: {
    canvas_score: number;
    chat_score: number;
    overall_score: number;
  };
  action_data: {
    textual_data: {
      title: string;
      summary: string;
      quote: string;
      keywords: string[];
      cluster_topics?: string[];
    };
    location_references: {
      document_id: string;
      page_number?: number;
      char_range?: { start: number; end: number };
      scroll_hint_percent?: number;
      context_heading?: string;
    };
    asset_references: {
      s3_key?: string;
      preview_url?: string;
      download_url?: string;
    };
    display_configuration: {
      highlight_style: {
        style: string;
        color: string;
        context_lines?: { before: number; after: number };
      };
    };
    general_metadata: {
      file_type: string;
      page_count?: number;
      word_count?: number;
      estimated_read_time_mins?: number;
    };
  };
  created_at: string;
  updated_at: string;
}

export interface MockEdge {
  id: string;
  start_node_id: string;
  end_node_id: string;
  description: string;
  style_metadata: {
    line_type: 'solid' | 'dashed' | 'dotted';
    line_weight: number;
    color: string;
    arrow_head_start: 'none' | 'filled_arrow' | 'open_arrow';
    arrow_head_end: 'none' | 'filled_arrow' | 'open_arrow';
  };
  created_at: string;
  updated_at: string;
  created_by: string;
  updated_by: string;
}

// Generate mock documents (Layer 2 - z=0.9)
const generateDocumentNodes = (): MockNode[] => {
  const documents = [
    { id: 'doc-1', title: 'AI Ethics Framework', topic: 'AI Ethics' },
    { id: 'doc-2', title: 'Machine Learning Principles', topic: 'ML Fundamentals' },
    { id: 'doc-3', title: 'Natural Language Processing', topic: 'NLP' },
    { id: 'doc-4', title: 'Computer Vision Advances', topic: 'Computer Vision' },
    { id: 'doc-5', title: 'Reinforcement Learning', topic: 'RL' },
  ];

  return documents.map((doc, index) => ({
    id: doc.id,
    space_id: 'space-1',
    content_entity_id: doc.id,
    content_entity_type: 'content_source' as const,
    depth_level: 2,
    path_to_root: `/${doc.id}`,
    position_2d: { x: Math.random() * 2 - 1, y: Math.random() * 2 - 1 },
    position_3d: { x: Math.random() * 2 - 1, y: Math.random() * 2 - 1, z: Math.random() * 2 - 1 },
    is_position_locked: false,
    visibility: true,
    display_props: {
      size: 15,
      opacity: 0.9,
      shape: 'sphere',
      color: '#90EE90',
    },
    engagement_score: {
      canvas_score: Math.random() * 0.5 + 0.5,
      chat_score: Math.random() * 0.5 + 0.5,
      overall_score: Math.random() * 0.5 + 0.5,
    },
    action_data: {
      textual_data: {
        title: doc.title,
        summary: `This document covers comprehensive aspects of ${doc.topic}, providing detailed insights and practical applications.`,
        quote: `Key insights from ${doc.title}...`,
        keywords: [doc.topic, 'research', 'technology', 'innovation'],
      },
      location_references: {
        document_id: doc.id,
        page_number: 1,
        scroll_hint_percent: 0,
        context_heading: 'Introduction',
      },
      asset_references: {
        s3_key: `documents/${doc.id}.pdf`,
        preview_url: `/assets/previews/${doc.id}.png`,
        download_url: `/api/v1/download/${doc.id}`,
      },
      display_configuration: {
        highlight_style: {
          style: 'document_boundary',
          color: '#e6ffe6',
        },
      },
      general_metadata: {
        file_type: 'pdf',
        page_count: Math.floor(Math.random() * 50) + 20,
        word_count: Math.floor(Math.random() * 5000) + 3000,
        estimated_read_time_mins: Math.floor(Math.random() * 30) + 15,
      },
    },
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  }));
};

// Generate mock clusters (Layer 1 - z=0)
const generateClusterNodes = (): MockNode[] => {
  const clusters = [
    { id: 'cluster-1', title: 'Ethics & Principles', docs: ['doc-1', 'doc-2'] },
    { id: 'cluster-2', title: 'Applied AI Technologies', docs: ['doc-3', 'doc-4'] },
    { id: 'cluster-3', title: 'Learning Systems', docs: ['doc-2', 'doc-5'] },
    { id: 'cluster-4', title: 'Data Processing', docs: ['doc-3', 'doc-4'] },
    { id: 'cluster-5', title: 'Advanced Methods', docs: ['doc-4', 'doc-5'] },
    { id: 'cluster-6', title: 'Theoretical Foundations', docs: ['doc-1', 'doc-5'] },
  ];

  return clusters.map((cluster, index) => ({
    id: cluster.id,
    space_id: 'space-1',
    content_entity_id: cluster.id,
    content_entity_type: 'chunk_cluster' as const,
    depth_level: 1,
    path_to_root: `/${cluster.id}`,
    position_2d: { x: Math.random() * 2 - 1, y: Math.random() * 2 - 1 },
    position_3d: { x: Math.random() * 2 - 1, y: Math.random() * 2 - 1, z: Math.random() * 2 - 1 },
    is_position_locked: false,
    visibility: true,
    display_props: {
      size: 12,
      opacity: 0.8,
      shape: 'sphere',
      color: '#DDA0DD',
    },
    engagement_score: {
      canvas_score: Math.random() * 0.4 + 0.3,
      chat_score: Math.random() * 0.4 + 0.3,
      overall_score: Math.random() * 0.4 + 0.3,
    },
    action_data: {
      textual_data: {
        title: cluster.title,
        summary: `Cluster containing related content about ${cluster.title}`,
        quote: `Key topics in ${cluster.title}...`,
        keywords: ['cluster', 'grouping', 'related'],
        cluster_topics: [cluster.title, 'Related Concepts'],
      },
      location_references: {
        document_id: cluster.docs[0],
        context_heading: cluster.title,
      },
      asset_references: {},
      display_configuration: {
        highlight_style: {
          style: 'cluster_boundary',
          color: '#f0e6ff',
        },
      },
      general_metadata: {
        file_type: 'cluster',
      },
    },
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  }));
};

// Generate mock text chunks (Layer 0 - z=-0.9)
const generateChunkNodes = (): MockNode[] => {
  const chunks: MockNode[] = [];
  const chunkTopics = [
    'Introduction to AI Ethics',
    'Bias in Machine Learning',
    'Fairness Principles',
    'Neural Network Architectures',
    'Deep Learning Fundamentals',
    'Supervised Learning Methods',
    'Text Processing Algorithms',
    'Language Model Training',
    'Semantic Analysis',
    'Image Recognition Techniques',
    'Convolutional Networks',
    'Object Detection',
    'Policy Gradient Methods',
    'Q-Learning Algorithms',
    'Reward Function Design',
  ];

  chunkTopics.forEach((topic, index) => {    
    chunks.push({
      id: `chunk-${index + 1}`,
      space_id: 'space-1',
      content_entity_id: `chunk-${index + 1}`,
      content_entity_type: 'content_chunk' as const,
      parent_node_id: `cluster-${Math.floor(index / 3) + 1}`,
      depth_level: 0,
      path_to_root: `/cluster-${Math.floor(index / 3) + 1}/chunk-${index + 1}`,
      position_2d: { x: Math.random() * 2 - 1, y: Math.random() * 2 - 1 },
      position_3d: { x: Math.random() * 2 - 1, y: Math.random() * 2 - 1, z: Math.random() * 2 - 1 },
      is_position_locked: false,
      visibility: true,
      display_props: {
        size: 8,
        opacity: 0.7,
        shape: 'sphere',
        color: '#FFFFE0',
      },
      engagement_score: {
        canvas_score: Math.random() * 0.3 + 0.2,
        chat_score: Math.random() * 0.3 + 0.2,
        overall_score: Math.random() * 0.3 + 0.2,
      },
      action_data: {
        textual_data: {
          title: topic,
          summary: `Detailed content about ${topic}`,
          quote: `"${topic} represents a fundamental concept in this domain..."`,
          keywords: topic.toLowerCase().split(' '),
        },
        location_references: {
          document_id: `doc-${Math.floor(index / 3) + 1}`,
          page_number: Math.floor(Math.random() * 10) + 1,
          char_range: { start: index * 100, end: (index + 1) * 100 },
          scroll_hint_percent: Math.random(),
          context_heading: topic,
        },
        asset_references: {},
        display_configuration: {
          highlight_style: {
            style: 'chunk_boundary',
            color: '#ffffcc',
            context_lines: { before: 2, after: 2 },
          },
        },
        general_metadata: {
          file_type: 'text',
          word_count: Math.floor(Math.random() * 200) + 100,
          estimated_read_time_mins: Math.floor(Math.random() * 5) + 2,
        },
      },
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    });
  });

  return chunks;
};

// Generate mock edges
const generateEdges = (nodes: MockNode[]): MockEdge[] => {
  const edges: MockEdge[] = [];
  const documents = nodes.filter(n => n.content_entity_type === 'content_source');
  const clusters = nodes.filter(n => n.content_entity_type === 'chunk_cluster');
  const chunks = nodes.filter(n => n.content_entity_type === 'content_chunk');

  // Document to Cluster edges
  documents.forEach(doc => {
    const relatedClusters = clusters.filter(() => Math.random() > 0.5);
    relatedClusters.forEach(cluster => {
      edges.push({
        id: `edge-${doc.id}-${cluster.id}`,
        start_node_id: doc.id,
        end_node_id: cluster.id,
        description: 'contains',
        style_metadata: {
          line_type: 'solid',
          line_weight: 2,
          color: '#666666',
          arrow_head_start: 'none',
          arrow_head_end: 'filled_arrow',
        },
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        created_by: 'system',
        updated_by: 'system',
      });
    });
  });

  // Cluster to Chunk edges
  chunks.forEach(chunk => {
    if (chunk.parent_node_id) {
      edges.push({
        id: `edge-${chunk.parent_node_id}-${chunk.id}`,
        start_node_id: chunk.parent_node_id,
        end_node_id: chunk.id,
        description: 'includes',
        style_metadata: {
          line_type: 'solid',
          line_weight: 1,
          color: '#999999',
          arrow_head_start: 'none',
          arrow_head_end: 'filled_arrow',
        },
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        created_by: 'system',
        updated_by: 'system',
      });
    }
  });

  // Cross-layer relationships
  chunks.forEach(chunk => {
    const relatedChunks = chunks.filter(c => c.id !== chunk.id && Math.random() > 0.8);
    relatedChunks.forEach(relatedChunk => {
      edges.push({
        id: `edge-${chunk.id}-${relatedChunk.id}`,
        start_node_id: chunk.id,
        end_node_id: relatedChunk.id,
        description: 'related to',
        style_metadata: {
          line_type: 'dashed',
          line_weight: 1,
          color: '#CCCCCC',
          arrow_head_start: 'none',
          arrow_head_end: 'open_arrow',
        },
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        created_by: 'system',
        updated_by: 'system',
      });
    });
  });

  return edges;
};

// Export mock data
export const mockNodes: MockNode[] = [
  ...generateDocumentNodes(),
  ...generateClusterNodes(),
  ...generateChunkNodes(),
];

export const mockEdges: MockEdge[] = generateEdges(mockNodes);

// Helper functions for data access
export const getNodesByLayer = (layer: number): MockNode[] => {
  const layerMap = {
    2: 'content_source',
    1: 'chunk_cluster',
    0: 'content_chunk',
  };
  return mockNodes.filter(node => node.content_entity_type === layerMap[layer as keyof typeof layerMap]);
};

export const getNodeById = (id: string): MockNode | undefined => {
  return mockNodes.find(node => node.id === id);
};

export const getEdgesForNode = (nodeId: string): MockEdge[] => {
  return mockEdges.filter(edge => edge.start_node_id === nodeId || edge.end_node_id === nodeId);
};

export const getChildNodes = (parentId: string): MockNode[] => {
  return mockNodes.filter(node => node.parent_node_id === parentId);
}; 