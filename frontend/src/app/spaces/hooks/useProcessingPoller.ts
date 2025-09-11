'use client';

import { useCallback, useEffect, useRef } from 'react';
import { ContentSource, ContentStatus } from '@/api/generated/v1/knowledge_pb';

export interface ProcessingPollerUpdate {
  id: string;
  status: ContentStatus;
}

interface UseProcessingPollerOptions {
  spaceId: string;
  /** Called with any updates detected during a poll cycle. */
  onUpdates: (updates: ProcessingPollerUpdate[]) => void;
  /** Only emit updates for these IDs if provided (prevents initial spam). */
  candidateIds?: string[];
  /** Base interval in ms. Will backoff if no changes. */
  intervalMs?: number;
}

/**
 * Polls the backend for content sources that are still processing and
 * notifies the caller when any of them transition to a terminal state.
 * One-way, no WebSocket required.
 */
export function useProcessingPoller({ spaceId, onUpdates, candidateIds, intervalMs = 3000 }: UseProcessingPollerOptions) {
  const timerRef = useRef<number | null>(null);
  const backoffRef = useRef<number>(intervalMs);
  const runningRef = useRef<boolean>(false);
  const lastEmittedRef = useRef<Map<string, ContentStatus>>(new Map());

  const clearTimer = () => {
    if (timerRef.current) {
      window.clearTimeout(timerRef.current);
      timerRef.current = null;
    }
  };

  const scheduleNext = useCallback((nextMs: number) => {
    clearTimer();
    timerRef.current = window.setTimeout(() => {
      void tick();
    }, nextMs) as unknown as number;
  }, []);

  const tick = useCallback(async () => {
    if (runningRef.current) return;
    runningRef.current = true;
    try {
      // Pause if tab not visible
      if (typeof document !== 'undefined' && document.visibilityState !== 'visible') {
        runningRef.current = false;
        scheduleNext(Math.max(backoffRef.current, intervalMs));
        return;
      }

      const { listContentSourcesUncached } = await import('@/api/actions/contentActions');
      const result = await listContentSourcesUncached(spaceId, 'processing');
      if (!result.ok || !result.data) {
        // Retry later with backoff
        backoffRef.current = Math.min(backoffRef.current * 1.5, 15000);
        runningRef.current = false;
        scheduleNext(backoffRef.current);
        return;
      }

      const processing = result.data as ContentSource[];

      // Even if none appear in processing now, candidates may have just transitioned
      // to processed/failed. Do not return early; continue to check terminal sets.

      // Also fetch any recently completed in case we missed the transition
      const completedRes = await listContentSourcesUncached(spaceId, 'processed');
      const failedRes = await listContentSourcesUncached(spaceId, 'failed');
      let completed = completedRes.ok && completedRes.data ? completedRes.data : [];
      let failed = failedRes.ok && failedRes.data ? failedRes.data : [];

      // If candidateIds provided, ensure we directly check those IDs too, to avoid
      // missing transitions due to list filters or timing.
      if (candidateIds && candidateIds.length > 0) {
        const { getContentSource } = await import('@/api/actions/contentActions');
        await Promise.allSettled(candidateIds.map(async (id) => {
          const res = await getContentSource(id);
          if (res.ok && res.data) {
            if (res.data.status === ContentStatus.PROCESSED && !completed.find(s => s.id === id)) {
              completed = completed.concat([res.data]);
            } else if (res.data.status === ContentStatus.FAILED && !failed.find(s => s.id === id)) {
              failed = failed.concat([res.data]);
            }
          }
        }));
      }

      const updates: ProcessingPollerUpdate[] = [];
      const candidateSet = candidateIds ? new Set(candidateIds) : undefined;

      const maybeEmit = (id: string, status: ContentStatus) => {
        if (candidateSet && !candidateSet.has(id)) return;
        const prev = lastEmittedRef.current.get(id);
        if (prev === status) return;
        lastEmittedRef.current.set(id, status);
        updates.push({ id, status: status });
      };

      for (const s of completed) maybeEmit(s.id, ContentStatus.PROCESSED);
      for (const s of failed) maybeEmit(s.id, ContentStatus.FAILED);

      if (updates.length > 0) {
        onUpdates(updates);
        // When changes detected, reset backoff to be responsive
        backoffRef.current = intervalMs;
      } else {
        backoffRef.current = Math.min(backoffRef.current * 1.3, 10000);
      }

      runningRef.current = false;
      scheduleNext(backoffRef.current);
    } catch {
      runningRef.current = false;
      backoffRef.current = Math.min(backoffRef.current * 1.5, 15000);
      scheduleNext(backoffRef.current);
    }
  }, [intervalMs, onUpdates, scheduleNext, spaceId]);

  useEffect(() => {
    backoffRef.current = intervalMs;
    runningRef.current = false;
    clearTimer();
    void tick();
    return () => {
      clearTimer();
    };
  }, [intervalMs, spaceId, tick]);

  useEffect(() => {
    const onVisibility = () => {
      if (document.visibilityState === 'visible') {
        backoffRef.current = intervalMs;
        void tick();
      }
    };
    document.addEventListener('visibilitychange', onVisibility);
    return () => document.removeEventListener('visibilitychange', onVisibility);
  }, [intervalMs, tick]);
}


