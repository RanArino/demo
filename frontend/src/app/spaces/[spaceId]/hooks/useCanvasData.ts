import { useCallback, useEffect, useRef, useState } from 'react';
import type { CanvasRenderableNode } from '@/lib/canvas';

interface CanvasApiResponse {
  nodes: CanvasRenderableNode[];
}

export function useCanvasData(spaceId: string) {
  const [nodes, setNodes] = useState<CanvasRenderableNode[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const controllerRef = useRef<AbortController | null>(null);

  const fetchNodes = useCallback(async () => {
    controllerRef.current?.abort();
    const controller = new AbortController();
    controllerRef.current = controller;

    try {
      const response = await fetch(`/api/spaces/${spaceId}/canvas/nodes`, {
        method: 'GET',
        cache: 'no-store',
        signal: controller.signal,
      });

      if (!response.ok) {
        throw new Error(await response.text());
      }

      const payload = (await response.json()) as CanvasApiResponse;
      setNodes(payload.nodes ?? []);
      setErrorMessage(null);
    } catch (error) {
      if (error instanceof DOMException && error.name === 'AbortError') {
        return;
      }
      const message = error instanceof Error ? error.message : 'Failed to load canvas data';
      setErrorMessage(message);
    } finally {
      setIsLoading(false);
    }
  }, [spaceId]);

  useEffect(() => {
    setIsLoading(true);
    fetchNodes();
    return () => {
      controllerRef.current?.abort();
    };
  }, [fetchNodes]);

  useEffect(() => {
    const intervalId = window.setInterval(() => {
      fetchNodes();
    }, 10_000);

    return () => window.clearInterval(intervalId);
  }, [fetchNodes]);

  return {
    nodes,
    isLoading,
    errorMessage,
    reload: fetchNodes,
  };
}
