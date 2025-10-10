export type CanvasNodeKind = 'cluster' | 'content' | 'chunk';

export interface Vec3 {
  x: number;
  y: number;
  z: number;
}

export interface CanvasDisplayProps {
  size: number;
  opacity?: number;
  shape?: string;
  color?: string;
}

export interface CanvasNodeCommon {
  id: string;
  kind: CanvasNodeKind;
  abstractionLevel: number;
  position: Vec3;
  display: CanvasDisplayProps;
  keywords: string[];
  title?: string;
  displayContent?: string;
  visibility: boolean;
  contextType?: string;
}

export interface CanvasClusterNode extends CanvasNodeCommon {
  kind: 'cluster';
  abstractionLevel: 1;
  spaceId: string;
  clusterScope: string;
  memberCount?: number;
}

export interface CanvasContentNode extends CanvasNodeCommon {
  kind: 'content';
  abstractionLevel: 0;
  contentSourceId: string;
  spaceId: string;
  mediaType?: string;
  tokenCount?: number;
  documentUrl?: string;
  documentTitle?: string;
}

export interface CanvasChunkNode extends CanvasNodeCommon {
  kind: 'chunk';
  abstractionLevel: number;
  contentSourceId: string;
  sequenceIndex: number;
  parentNodeId?: string;
  spaceId: string;
  chunkType?: string;
}

export type CanvasRenderableNode = CanvasClusterNode | CanvasContentNode | CanvasChunkNode;

export interface CanvasGraphData {
  nodes: CanvasRenderableNode[];
}
