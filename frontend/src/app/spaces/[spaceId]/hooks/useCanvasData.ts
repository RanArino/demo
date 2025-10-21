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
  const retryTimeoutRef = useRef<number | null>(null);
  const retryTickerRef = useRef<number | null>(null);
  const attemptsRef = useRef(0);
  const [retryInMs, setRetryInMs] = useState<number | null>(null);
  const isActiveRef = useRef(true);
  const isFetchingRef = useRef(false);
  const fetchNodesRef = useRef<() => Promise<void>>();
  const lastVisibilityStateRef = useRef<DocumentVisibilityState | null>(null);

  useEffect(() => {
    return () => {
      isActiveRef.current = false;
    };
  }, []);

  const clearRetry = useCallback(() => {
    if (retryTimeoutRef.current) {
      window.clearTimeout(retryTimeoutRef.current);
      retryTimeoutRef.current = null;
    }
    if (retryTickerRef.current) {
      window.clearInterval(retryTickerRef.current);
      retryTickerRef.current = null;
    }
    setRetryInMs(null);
  }, []);

  const scheduleRetry = useCallback(() => {
    const attempt = attemptsRef.current;
    const base = 2000; // 2s
    const maxDelay = 60000; // 60s
    const delay = Math.min(maxDelay, Math.floor(base * Math.pow(2, Math.max(0, attempt - 1))));
    const start = performance.now();
    setRetryInMs(delay);
    if (retryTickerRef.current) {
      window.clearInterval(retryTickerRef.current);
    }
    retryTickerRef.current = window.setInterval(() => {
      const elapsed = performance.now() - start;
      const remaining = Math.max(0, delay - elapsed);
      setRetryInMs(remaining);
    }, 250);
    if (retryTimeoutRef.current) {
      window.clearTimeout(retryTimeoutRef.current);
    }
    retryTimeoutRef.current = window.setTimeout(() => {
      retryTimeoutRef.current = null;
      setRetryInMs(null);
      fetchNodesRef.current?.();
    }, delay);
  }, []);

  const fetchNodes = useCallback(async () => {
    if (isFetchingRef.current) {
      return;
    }
    isFetchingRef.current = true;
    setIsLoading(true);
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
      if (!isActiveRef.current) {
        return;
      }
      setNodes(payload.nodes ?? []);
      setErrorMessage(null);
      attemptsRef.current = 0;
      clearRetry();
    } catch (error) {
      if (error instanceof DOMException && error.name === 'AbortError') {
        return;
      }
      if (!isActiveRef.current) {
        return;
      }
      const message = error instanceof Error ? error.message : 'Failed to load canvas data';
      setErrorMessage(message);
      attemptsRef.current += 1;
      scheduleRetry();
    } finally {
      if (isActiveRef.current) {
        setIsLoading(false);
      }
      isFetchingRef.current = false;
    }
  }, [spaceId, clearRetry, scheduleRetry]);

  fetchNodesRef.current = fetchNodes;

  useEffect(() => {
    setIsLoading(true);
    fetchNodes();
    return () => {
      controllerRef.current?.abort();
      clearRetry();
    };
  }, [fetchNodes, clearRetry]);

  useEffect(() => {
    lastVisibilityStateRef.current = document.visibilityState;
    const handleVisibility = () => {
      const state = document.visibilityState;
      if (state !== 'visible') {
        lastVisibilityStateRef.current = state;
        return;
      }
      if (lastVisibilityStateRef.current === 'visible') {
        return;
      }
      lastVisibilityStateRef.current = state;
      clearRetry();
      fetchNodesRef.current?.();
    };
    const handleFocus = () => {
      if (document.visibilityState !== 'visible') {
        return;
      }
      handleVisibility();
    };

    document.addEventListener('visibilitychange', handleVisibility);
    window.addEventListener('focus', handleFocus);

    return () => {
      document.removeEventListener('visibilitychange', handleVisibility);
      window.removeEventListener('focus', handleFocus);
    };
  }, [fetchNodes, clearRetry]);

  return {
    nodes,
    isLoading,
    errorMessage,
    reload: () => {
      attemptsRef.current = 0;
      clearRetry();
      setIsLoading(true);
      fetchNodes();
    },
    retryInMs,
  };
}
