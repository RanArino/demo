import type { Node, SpatialCoordinates, DisplayProps } from '@/api/generated/v1/canvas_pb';
import type { Struct } from '@bufbuild/protobuf';
import type {
  CanvasDisplayProps,
  CanvasRenderableNode,
  CanvasClusterNode,
  CanvasContentNode,
  CanvasChunkNode,
  Vec3,
} from './types';

const DEFAULT_POS: Vec3 = { x: 0, y: 0, z: 0 };

const DEFAULT_DISPLAY: Record<CanvasRenderableNode['kind'], CanvasDisplayProps> = {
  cluster: { size: 20, opacity: 0.8, shape: 'sphere', color: '#4A90E2' },
  content: { size: 15, opacity: 0.7, shape: 'icosahedron', color: '#FF6B6B' },
  chunk: { size: 4, opacity: 0.9, shape: 'sphere', color: '#F8E71C' },
};

function toVec3(coords?: SpatialCoordinates): Vec3 {
  if (!coords) {
    return DEFAULT_POS;
  }

  return {
    x: coords.x ?? 0,
    y: coords.y ?? 0,
    z: coords.z ?? 0,
  };
}

function toDisplay(kind: CanvasRenderableNode['kind'], props?: DisplayProps): CanvasDisplayProps {
  const defaults = DEFAULT_DISPLAY[kind];
  if (!props) {
    return defaults;
  }

  return {
    size: props.size || defaults.size,
    opacity: props.opacity ?? defaults.opacity,
    shape: (props.shape && props.shape.trim() !== '') ? props.shape : defaults.shape,
    color: (props.color && props.color.trim() !== '') ? props.color : defaults.color,
  };
}

export function convertProtoNode(node: Node): CanvasRenderableNode | null {
  if (!node?.node?.case) {
    return null;
  }

  if (node.node.case === 'cluster') {
    const raw = node.node.value;
    const base = raw.base;
    if (!base) {
      return null;
    }

    const chatContent = base.chatContent ?? undefined;

    const converted: CanvasClusterNode = {
      id: base.id,
      kind: 'cluster',
      abstractionLevel: 1 as const,
      position: toVec3(base.position3d),
      display: toDisplay('cluster', base.displayProps),
      keywords: base.keywords,
      title: raw.title ?? base.displayContent ?? raw.clusterScope,
      displayContent: base.displayContent ?? chatContent,
      chatContent,
      visibility: base.visibility ?? true,
      spaceId: base.spaceId,
      clusterScope: raw.clusterScope,
      memberCount: raw.memberCount,
      contextType: base.contextType,
    };

    return converted;
  }

  if (node.node.case === 'content') {
    const raw = node.node.value;
    const base = raw.base;
    if (!base) {
      return null;
    }

    const chatContent = base.chatContent ?? undefined;

    const converted: CanvasContentNode = {
      id: base.id,
      kind: 'content',
      abstractionLevel: 0 as const,
      position: toVec3(base.position3d),
      display: toDisplay('content', base.displayProps),
      keywords: base.keywords,
      title: raw.title ?? base.displayContent,
      displayContent: base.displayContent ?? chatContent,
      chatContent,
      visibility: base.visibility ?? true,
      spaceId: base.spaceId,
      contentSourceId: raw.contentSourceId,
      mediaType: raw.mediaType,
      tokenCount: raw.tokenCount,
      contextType: base.contextType,
      documentUrl: getStructStringField(raw.actionData, 'url'),
      documentTitle: raw.title ?? base.displayContent ?? undefined,
    };

    return converted;
  }

  if (node.node.case === 'chunk') {
    const raw = node.node.value;
    const base = raw.base;
    if (!base) {
      return null;
    }

    const chatContent = base.chatContent ?? undefined;

    const converted: CanvasChunkNode = {
      id: base.id,
      kind: 'chunk',
      abstractionLevel: base.abstractionLevel,
      position: toVec3(base.position3d),
      display: toDisplay('chunk', base.displayProps),
      keywords: base.keywords,
      title: base.displayContent ?? chatContent ?? raw.chunkType,
      displayContent: base.displayContent ?? chatContent,
      chatContent,
      visibility: base.visibility ?? true,
      contentSourceId: raw.contentSourceId,
      sequenceIndex: raw.sequenceIndex,
      parentNodeId: raw.contentSourceId || undefined,
      spaceId: base.spaceId,
      chunkType: raw.chunkType,
      contextType: base.contextType,
    };

    return converted;
  }

  return null;
}

export function convertProtoNodes(nodes: Node[] | undefined | null): CanvasRenderableNode[] {
  if (!nodes || nodes.length === 0) {
    return [];
  }

  const result: CanvasRenderableNode[] = [];

  nodes.forEach((node) => {
    const converted = convertProtoNode(node);
    if (converted) {
      result.push(converted);
    }
  });

  return result;
}

function getStructStringField(struct: Struct | undefined, key: string): string | undefined {
  if (!struct) {
    return undefined;
  }
  const fields: unknown = (struct as unknown as { fields?: unknown }).fields;
  if (!fields || typeof fields !== 'object') {
    return undefined;
  }
  const value = (fields as Record<string, unknown>)[key] as { stringValue?: unknown } | undefined;
  const str = value?.stringValue;
  return typeof str === 'string' && str.trim() !== '' ? str : undefined;
}
