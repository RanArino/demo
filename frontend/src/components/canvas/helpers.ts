import { MockNode } from './types';
import { LAYER_CONFIG } from './config';

// Helper function to get node layer
export const getNodeLayer = (node: MockNode): keyof typeof LAYER_CONFIG | null => {
  switch (node.content_entity_type) {
    case 'content_source':
      return 'documents';
    case 'chunk_cluster':
      return 'clusters';
    case 'content_chunk':
      return 'chunks';
    default:
      return null;
  }
};