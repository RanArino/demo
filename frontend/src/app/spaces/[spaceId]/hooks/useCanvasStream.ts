'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import type { CanvasRenderableNode, CanvasNodeKind } from '@/lib/canvas';
import { GraphStore } from '@/lib/canvas';

type StreamMetaEvent = {
  event: 'meta';
  message?: string;
  limits?: Record<string, number>;
};

type StreamNodesEvent = {
  event: 'nodes';
  level: CanvasNodeKind;
  parentId?: string;
  nodes: CanvasRenderableNode[];
};

type StreamErrorEvent = {
  event: 'error';
  message?: string;
};

type StreamEndEvent = {
  event: 'end';
  counts?: Partial<Record<CanvasNodeKind, number>>;
};

type StreamEvent = StreamMetaEvent | StreamNodesEvent | StreamErrorEvent | StreamEndEvent;

interface CanvasStreamState {
  nodes: CanvasRenderableNode[];
  isLoading: boolean;
  errorMessage: string | null;
  retryInMs: number | null;
  reload: () => void;
  progress: Record<CanvasNodeKind, number>;
  mergeNodes: (batch: CanvasRenderableNode[]) => void;
}

const STREAM_ENDPOINT = (spaceId: string) => `/api/spaces/${spaceId}/canvas/stream`;

export function useCanvasStream(spaceId: string): CanvasStreamState {
  const [nodes, setNodes] = useState<CanvasRenderableNode[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [retryInMs, setRetryInMs] = useState<number | null>(null);
  const [progress, setProgress] = useState<Record<CanvasNodeKind, number>>({
    cluster: 0,
    content: 0,
    chunk: 0,
  });

  const graphStoreRef = useRef(new GraphStore());
  const controllerRef = useRef<AbortController | null>(null);
  const retryTimeoutRef = useRef<number | null>(null);
  const retryTickerRef = useRef<number | null>(null);
  const flushTimeoutRef = useRef<number | null>(null);
  const attemptsRef = useRef(0);
  const isFetchingRef = useRef(false);
  const isActiveRef = useRef(true);
  const pendingErrorRef = useRef<string | null>(null);
  const fetchRef = useRef<() => void>();
  const hasReceivedFirstRef = useRef(false);

  // Utilities -------------------------------------------------------------

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
    const baseDelay = 2000;
    const maxDelay = 60000;
    const delay = Math.min(maxDelay, Math.floor(baseDelay * Math.pow(2, Math.max(0, attempt - 1))));
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
      fetchRef.current?.();
    }, delay);
  }, []);

  const clearFlush = useCallback(() => {
    if (flushTimeoutRef.current !== null) {
      window.clearTimeout(flushTimeoutRef.current);
      flushTimeoutRef.current = null;
    }
  }, []);

  const flushNodes = useCallback(() => {
    if (!isActiveRef.current) {
      return;
    }
    setNodes(graphStoreRef.current.getNodes());
    setProgress(graphStoreRef.current.counts());
  }, []);

  const scheduleFlush = useCallback(() => {
    if (flushTimeoutRef.current !== null) {
      return;
    }
    flushTimeoutRef.current = window.setTimeout(() => {
      flushTimeoutRef.current = null;
      flushNodes();
    }, 32);
  }, [flushNodes]);

  const mergeNodes = useCallback((batch: CanvasRenderableNode[]) => {
    if (!batch || batch.length === 0) {
      return;
    }
    const stats = graphStoreRef.current.merge(batch);
    if (stats.added > 0 || stats.updated > 0) {
      scheduleFlush();
    }
  }, [scheduleFlush]);

  // Streaming -------------------------------------------------------------

  const startStream = useCallback(async () => {
    if (isFetchingRef.current) {
      return;
    }

    isFetchingRef.current = true;
    setIsLoading(true);
    pendingErrorRef.current = null;
    hasReceivedFirstRef.current = false;
    controllerRef.current?.abort();
    const controller = new AbortController();
    controllerRef.current = controller;

    graphStoreRef.current.reset();
    setNodes([]);
    setProgress({ cluster: 0, content: 0, chunk: 0 });

    try {
      const response = await fetch(STREAM_ENDPOINT(spaceId), {
        method: 'GET',
        cache: 'no-store',
        signal: controller.signal,
      });

      if (!response.ok) {
        throw new Error(await response.text());
      }

      const reader = response.body?.getReader();
      if (!reader) {
        throw new Error('Streaming not supported by the browser');
      }

      const decoder = new TextDecoder();
      let buffer = '';
      while (true) {
        const { value, done } = await reader.read();
        if (done) {
          break;
        }
        buffer += decoder.decode(value, { stream: true });
        let newlineIndex = buffer.indexOf('\n');
        while (newlineIndex >= 0) {
          const raw = buffer.slice(0, newlineIndex).trim();
          buffer = buffer.slice(newlineIndex + 1);
          newlineIndex = buffer.indexOf('\n');
          if (!raw) {
            continue;
          }
          try {
            const evt = JSON.parse(raw) as StreamEvent;
            if (evt.event === 'nodes' && Array.isArray(evt.nodes) && evt.nodes.length > 0) {
              const stats = graphStoreRef.current.merge(evt.nodes);
              if (stats.added > 0 || stats.updated > 0) {
                scheduleFlush();
              }
              if (!hasReceivedFirstRef.current) {
                hasReceivedFirstRef.current = true;
                setIsLoading(false);
              }
            } else if (evt.event === 'error') {
              pendingErrorRef.current = evt.message ?? 'Stream error';
              break;
            } else if (evt.event === 'end') {
              if (evt.counts) {
                setProgress((prev) => ({ ...prev, ...evt.counts }));
              }
            }
          } catch (error) {
            globalThis.console?.warn?.('[useCanvasStream] Failed to parse stream chunk', error, raw);
          }
        }

        if (pendingErrorRef.current) {
          break;
        }
      }

      const trailing = buffer.trim();
      if (!pendingErrorRef.current && trailing) {
        try {
          const evt = JSON.parse(trailing) as StreamEvent;
          if (evt.event === 'nodes' && Array.isArray(evt.nodes) && evt.nodes.length > 0) {
            const stats = graphStoreRef.current.merge(evt.nodes);
            if (stats.added > 0 || stats.updated > 0) {
              scheduleFlush();
            }
            if (!hasReceivedFirstRef.current) {
              hasReceivedFirstRef.current = true;
              setIsLoading(false);
            }
          } else if (evt.event === 'error') {
            pendingErrorRef.current = evt.message ?? 'Stream error';
          } else if (evt.event === 'end' && evt.counts) {
            setProgress((prev) => ({ ...prev, ...evt.counts }));
          }
        } catch (error) {
          globalThis.console?.warn?.('[useCanvasStream] Failed to parse trailing chunk', error, trailing);
        }
      }

      if (pendingErrorRef.current) {
        throw new Error(pendingErrorRef.current);
      }

      attemptsRef.current = 0;
      clearRetry();
      scheduleFlush();
      setErrorMessage(null);
    } catch (error) {
      if (!isActiveRef.current) {
        return;
      }
      if (error instanceof DOMException && error.name === 'AbortError') {
        return;
      }
      const message = error instanceof Error ? error.message : 'Failed to stream canvas data';
      setErrorMessage(message);
      attemptsRef.current += 1;
      scheduleRetry();
    } finally {
      if (isActiveRef.current) {
        setIsLoading(false);
      }
      clearFlush();
      isFetchingRef.current = false;
    }
  }, [spaceId, clearFlush, clearRetry, scheduleFlush, scheduleRetry]);

  fetchRef.current = startStream;

  // Lifecycle -------------------------------------------------------------

  useEffect(() => {
    return () => {
      isActiveRef.current = false;
      controllerRef.current?.abort();
      clearRetry();
      clearFlush();
    };
  }, [clearRetry, clearFlush]);

  useEffect(() => {
    isActiveRef.current = true;
    attemptsRef.current = 0;
    clearRetry();
    startStream();

    return () => {
      controllerRef.current?.abort();
    };
  }, [spaceId, clearRetry, startStream]);

  const reload = useCallback(() => {
    attemptsRef.current = 0;
    clearRetry();
    startStream();
  }, [clearRetry, startStream]);

  return {
    nodes,
    isLoading,
    errorMessage,
    retryInMs,
    reload,
    progress,
    mergeNodes,
  };
}
