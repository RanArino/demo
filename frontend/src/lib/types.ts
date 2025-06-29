
export interface Space {
  id: string;
  title: string;
  description: string;
  icon: string;
  cover_image: string;
  keywords: string[];
  owner_id: string;
  created_at: Date;
  last_updated_at: Date;
  document_count: number;
  total_size_bytes: number;
  access_level: 'private' | 'shared' | 'public';
}

export interface ContentSource {
    id: string;
    owner_id: string;
    title: string;
    media_type: string;
    source: string;
    status: string;
    num_chunks: number;
    total_size_bytes: number;
    created_at: Date;
    updated_at: Date;
}

export interface Node {
  id: string;
  space_id: string;
  content_source_id: string | null;
  parent_node_id: string | null;
  content_entity_type: string;
  media_type: string;
  depth_level: number;
  position_2d: { x: number; y: number };
  position_3d: { x: number; y: number; z: number };
  is_position_locked: boolean;
  visibility: boolean;
  display_props: Record<string, unknown>;
  engagement_score: Record<string, unknown>;
  action_data: Record<string, unknown>;
  created_at: Date;
  updated_at: Date;
}

export interface Edge {
  id: string;
  start_node_id: string;
  end_node_id: string;
  description: string;
  style_metadata: Record<string, unknown>;
  created_by: string;
  updated_by: string;
  created_at: Date;
  updated_at: Date;
}
