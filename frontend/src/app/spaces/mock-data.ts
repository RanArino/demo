
import { Space, Node, Edge, ContentSource } from '@/lib/types';

export const mockSpaces: Space[] = [
  {
    id: 'spc_1',
    title: 'Project Gemini',
    description: 'A space for the Gemini project, containing all related documents and resources.',
    icon: '🚀',
    cover_image: 'https://images.unsplash.com/photo-1518186285589-2f7649de83e0?w=400&h=240&fit=crop',
    keywords: ['gemini', 'ai', 'project'],
    owner_id: 'user_1',
    created_at: new Date('2023-01-15T09:30:00Z'),
    last_updated_at: new Date('2023-05-20T14:00:00Z'),
    document_count: 10,
    total_size_bytes: 1024 * 1024 * 50, // 50MB
    access_level: 'private',
  },
  {
    id: 'spc_2',
    title: 'Marketing Q2',
    description: 'All marketing materials for the second quarter.',
    icon: '📈',
    cover_image: 'https://images.unsplash.com/photo-1542744173-8e7e53415bb0?w=400&h=240&fit=crop',
    keywords: ['marketing', 'q2', 'campaign'],
    owner_id: 'user_2',
    created_at: new Date('2023-04-01T11:00:00Z'),
    last_updated_at: new Date('2023-06-01T18:45:00Z'),
    document_count: 25,
    total_size_bytes: 1024 * 1024 * 120, // 120MB
    access_level: 'shared',
  },
  {
    id: 'spc_3',
    title: 'Research Archive',
    description: 'Collection of research papers and academic resources for ongoing studies.',
    icon: '🔬',
    cover_image: 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=400&h=240&fit=crop',
    keywords: ['research', 'academic', 'papers', 'studies'],
    owner_id: 'user_1',
    created_at: new Date('2023-02-10T14:15:00Z'),
    last_updated_at: new Date('2023-06-15T10:30:00Z'),
    document_count: 45,
    total_size_bytes: 1024 * 1024 * 200, // 200MB
    access_level: 'public',
  },
  {
    id: 'spc_4',
    title: 'Design System',
    description: 'UI components, guidelines, and design assets for the product team.',
    icon: '🎨',
    cover_image: 'https://images.unsplash.com/photo-1541462608143-67571c6738dd?w=400&h=240&fit=crop',
    keywords: ['design', 'ui', 'components', 'guidelines'],
    owner_id: 'user_3',
    created_at: new Date('2023-03-05T16:20:00Z'),
    last_updated_at: new Date('2023-06-10T12:45:00Z'),
    document_count: 18,
    total_size_bytes: 1024 * 1024 * 75, // 75MB
    access_level: 'shared',
  },
  {
    id: 'spc_5',
    title: 'Client Presentations',
    description: 'Presentation materials and client-facing documentation.',
    icon: '💼',
    cover_image: 'https://images.unsplash.com/photo-1497366216548-37526070297c?w=400&h=240&fit=crop',
    keywords: ['presentations', 'client', 'meetings', 'docs'],
    owner_id: 'user_2',
    created_at: new Date('2023-03-20T11:10:00Z'),
    last_updated_at: new Date('2023-06-05T15:20:00Z'),
    document_count: 32,
    total_size_bytes: 1024 * 1024 * 95, // 95MB
    access_level: 'private',
  },
  {
    id: 'spc_6',
    title: 'Data Science Lab',
    description: 'Jupyter notebooks, datasets, and ML model experiments.',
    icon: '📊',
    cover_image: 'https://images.unsplash.com/photo-1551288049-bebda4e38f71?w=400&h=240&fit=crop',
    keywords: ['data', 'science', 'ml', 'jupyter', 'datasets'],
    owner_id: 'user_4',
    created_at: new Date('2023-04-12T08:30:00Z'),
    last_updated_at: new Date('2023-06-20T09:15:00Z'),
    document_count: 67,
    total_size_bytes: 1024 * 1024 * 350, // 350MB
    access_level: 'shared',
  },
];

export const mockContentSources: ContentSource[] = [
    {
        id: 'cs_1',
        owner_id: 'user_1',
        title: 'Gemini Project Proposal',
        media_type: 'document',
        source: 'upload',
        status: 'processed',
        num_chunks: 15,
        total_size_bytes: 1024 * 1024 * 5, // 5MB
        created_at: new Date('2023-01-15T10:00:00Z'),
        updated_at: new Date('2023-01-15T10:05:00Z'),
    },
    {
        id: 'cs_2',
        owner_id: 'user_1',
        title: 'Market Research Q1',
        media_type: 'document',
        source: 'gdrive',
        status: 'processed',
        num_chunks: 50,
        total_size_bytes: 1024 * 1024 * 10, // 10MB
        created_at: new Date('2023-01-20T14:30:00Z'),
        updated_at: new Date('2023-01-20T14:35:00Z'),
    }
];

export const mockNodes: Node[] = [
  {
    id: 'node_1',
    space_id: 'spc_1',
    content_source_id: 'cs_1',
    parent_node_id: null,
    content_entity_type: 'content_source',
    media_type: 'document',
    depth_level: 0,
    position_2d: { x: 100, y: 100 },
    position_3d: { x: 100, y: 100, z: 0 },
    is_position_locked: false,
    visibility: true,
    display_props: { color: '#FF6347', size: 50, shape: 'circle' },
    engagement_score: { overall_score: 0.8 },
    action_data: { title: 'Gemini Project Proposal' },
    created_at: new Date('2023-01-15T10:05:00Z'),
    updated_at: new Date('2023-01-15T10:05:00Z'),
  },
  {
    id: 'node_2',
    space_id: 'spc_1',
    content_source_id: 'cs_1',
    parent_node_id: 'node_1',
    content_entity_type: 'content_chunk',
    media_type: 'text',
    depth_level: 1,
    position_2d: { x: 200, y: 150 },
    position_3d: { x: 200, y: 150, z: 20 },
    is_position_locked: false,
    visibility: true,
    display_props: { color: '#4682B4', size: 20, shape: 'square' },
    engagement_score: { overall_score: 0.9 },
    action_data: { title: 'Introduction' },
    created_at: new Date('2023-01-15T10:06:00Z'),
    updated_at: new Date('2023-01-15T10:06:00Z'),
  },
];

export const mockEdges: Edge[] = [
  {
    id: 'edge_1',
    start_node_id: 'node_1',
    end_node_id: 'node_2',
    description: 'contains',
    style_metadata: {
      line_type: 'solid',
      line_weight: 1,
      color: '#888888',
      arrow_head_start: 'none',
      arrow_head_end: 'filled_arrow',
    },
    created_by: 'system',
    updated_by: 'system',
    created_at: new Date('2023-01-15T10:06:00Z'),
    updated_at: new Date('2023-01-15T10:06:00Z'),
  },
];
