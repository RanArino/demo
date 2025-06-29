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

export interface Graph3DCanvasProps {
  nodes: MockNode[];
  edges: MockEdge[];
  selectedNodeId?: string;
  onNodeClick: (nodeId: string) => void;
  onNodeHover: (nodeId: string | null) => void;
} 