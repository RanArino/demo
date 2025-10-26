import type { CanvasNodeKind, CanvasRenderableNode } from './types';

const KIND_PRIORITY: Record<CanvasNodeKind, number> = {
  cluster: 0,
  content: 1,
  chunk: 2,
};

interface MergeStats {
  added: number;
  updated: number;
}

/**
 * Lightweight in-memory store that keeps canvas nodes deduplicated while preserving
 * their arrival order per abstraction layer. This allows the UI to progressively
 * render nodes without re-sorting the entire array on every merge.
 */
export class GraphStore {
  private readonly nodes = new Map<string, CanvasRenderableNode>();
  private readonly insertionOrder = new Map<string, number>();
  private snapshot: CanvasRenderableNode[] | null = null;
  private sequence = 0;

  reset(): void {
    this.nodes.clear();
    this.insertionOrder.clear();
    this.snapshot = null;
    this.sequence = 0;
  }

  merge(batch: CanvasRenderableNode[]): MergeStats {
    let added = 0;
    let updated = 0;

    for (const node of batch) {
      if (!node || !node.id) {
        continue;
      }
      const existing = this.nodes.get(node.id);
      if (existing) {
        this.nodes.set(node.id, { ...existing, ...node });
        updated += 1;
      } else {
        this.nodes.set(node.id, node);
        this.insertionOrder.set(node.id, this.sequence += 1);
        added += 1;
      }
    }

    if (added > 0 || updated > 0) {
      this.snapshot = null;
    }

    return { added, updated };
  }

  getNodes(): CanvasRenderableNode[] {
    if (this.snapshot) {
      return [...this.snapshot];
    }

    const sorted = Array.from(this.nodes.values()).sort((a, b) => {
      const kindDiff = KIND_PRIORITY[a.kind] - KIND_PRIORITY[b.kind];
      if (kindDiff !== 0) {
        return kindDiff;
      }
      const aOrder = this.insertionOrder.get(a.id) ?? 0;
      const bOrder = this.insertionOrder.get(b.id) ?? 0;
      return aOrder - bOrder;
    });

    this.snapshot = sorted;
    return [...sorted];
  }

  counts(): Record<CanvasNodeKind, number> {
    const totals: Record<CanvasNodeKind, number> = {
      cluster: 0,
      content: 0,
      chunk: 0,
    };

    this.nodes.forEach((node) => {
      totals[node.kind] += 1;
    });

    return totals;
  }
}
