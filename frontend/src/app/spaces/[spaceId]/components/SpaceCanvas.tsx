'use client';

import { useEffect, useMemo, useRef, useState } from 'react';
import { useRouter } from 'next/navigation';
import type { Space } from '@/api/generated/v1/knowledge_pb';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import type { CanvasRenderableNode, CanvasNodeKind } from '@/lib/canvas';
import { CanvasScene } from '@/lib/canvas';
import { useCanvasData } from '../hooks/useCanvasData';
import { Grid3x3, Grid, LayoutGrid, Map as MapIcon, RefreshCw } from 'lucide-react';

interface SpaceCanvasProps {
  space: Space;
  className?: string;
}

export default function SpaceCanvas({ space, className }: SpaceCanvasProps) {
  const router = useRouter();
  const canvasContainerRef = useRef<HTMLDivElement | null>(null);
  const minimapContainerRef = useRef<HTMLDivElement | null>(null);
  const multiViewTopRef = useRef<HTMLDivElement | null>(null);
  const multiViewSideRef = useRef<HTMLDivElement | null>(null);
  const multiViewIsoRef = useRef<HTMLDivElement | null>(null);
  const multiViewFrontRef = useRef<HTMLDivElement | null>(null);
  const sceneRef = useRef<CanvasScene | null>(null);
  const resizeObserverRef = useRef<ResizeObserver | null>(null);

  const { nodes, isLoading, errorMessage, reload: reloadNodes, retryInMs } = useCanvasData(space.id);
  const [selectedNode, setSelectedNode] = useState<CanvasRenderableNode | null>(null);
  const [isGridVisible, setIsGridVisible] = useState(true);
  const [isMinimapVisible, setIsMinimapVisible] = useState(true);
  const [isMultiViewVisible, setIsMultiViewVisible] = useState(false);
  const [isMultiViewSwitching, setIsMultiViewSwitching] = useState(false);
  const [isLayeredView, setIsLayeredView] = useState(false);
  const [isLayeredSwitching, setIsLayeredSwitching] = useState(false);
  const [layerVisibility, setLayerVisibility] = useState<Record<CanvasNodeKind, boolean>>({
    cluster: true,
    content: true,
    chunk: true,
  });
  const [fps, setFps] = useState(0);
  const [edgePanSpeedPref, setEdgePanSpeedPref] = useState(120);
  const [edgePanThresholdPref, setEdgePanThresholdPref] = useState(0.12);
  const [zoomAggressiveness, setZoomAggressivenessPref] = useState(0.3);
  const [showSettings, setShowSettings] = useState(false);
  const [linkVisibility, setLinkVisibility] = useState<{ clusterEdges: boolean; semanticEdges: boolean }>({
    clusterEdges: false,
    semanticEdges: false,
  });

  const totalCounts = useMemo(() => {
    return nodes.reduce(
      (acc, node) => {
        acc.total += 1;
        acc[node.kind] += 1;
        return acc;
      },
      { total: 0, cluster: 0, content: 0, chunk: 0 }
    );
  }, [nodes]);

  useEffect(() => {
    if (typeof window === 'undefined') {
      return;
    }
    try {
      const stored = window.localStorage.getItem('canvas-ui-preferences');
      if (!stored) {
        return;
      }
      const prefs = JSON.parse(stored) as {
        grid?: boolean;
        minimap?: boolean;
        multiView?: boolean;
        layered?: boolean;
        visibility?: Record<CanvasNodeKind, boolean>;
        edgePanSpeed?: number;
        edgePanThreshold?: number;
        zoomAggressiveness?: number;
        links?: { clusterEdges: boolean; semanticEdges: boolean };
      };
      if (typeof prefs.grid === 'boolean') {
        setIsGridVisible(prefs.grid);
      }
      if (typeof prefs.minimap === 'boolean') {
        setIsMinimapVisible(prefs.minimap);
      }
      if (typeof prefs.multiView === 'boolean') {
        setIsMultiViewVisible(prefs.multiView);
      }
      if (typeof prefs.layered === 'boolean') {
        setIsLayeredView(prefs.layered);
      }
      if (prefs.visibility) {
        setLayerVisibility((prev) => ({ ...prev, ...prefs.visibility }));
      }
      if (typeof prefs.edgePanSpeed === 'number') {
        setEdgePanSpeedPref(prefs.edgePanSpeed);
      }
      if (typeof prefs.edgePanThreshold === 'number') {
        setEdgePanThresholdPref(prefs.edgePanThreshold);
      }
      if (typeof prefs.zoomAggressiveness === 'number') {
        setZoomAggressivenessPref(prefs.zoomAggressiveness);
      }
      if (prefs.links) {
        setLinkVisibility((prev) => ({ ...prev, ...prefs.links }));
      }
    } catch (error) {
      console.warn('[SpaceCanvas] Failed to load preferences', error);
    }
  }, []);

  useEffect(() => {
    if (typeof window === 'undefined') {
      return;
    }
    const prefs = {
      grid: isGridVisible,
      minimap: isMinimapVisible,
      multiView: isMultiViewVisible,
      layered: isLayeredView,
      visibility: layerVisibility,
      edgePanSpeed: edgePanSpeedPref,
      edgePanThreshold: edgePanThresholdPref,
      zoomAggressiveness,
      links: linkVisibility,
    };
    window.localStorage.setItem('canvas-ui-preferences', JSON.stringify(prefs));
  }, [isGridVisible, isMinimapVisible, isMultiViewVisible, isLayeredView, layerVisibility, edgePanSpeedPref, edgePanThresholdPref, zoomAggressiveness, linkVisibility]);

  useEffect(() => {
    const container = canvasContainerRef.current;
    if (!container || sceneRef.current) {
      return;
    }

    const scene = new CanvasScene({
      container,
      onNodeSelect: (node) => setSelectedNode(node),
      onFpsUpdate: (value) => setFps(value),
    });

    sceneRef.current = scene;
    scene.setMinimapContainer(isMinimapVisible ? minimapContainerRef.current : null);
    const multiViewConfigs = isMultiViewVisible
      ? [
          { id: 'top', type: 'top' as const, container: multiViewTopRef.current },
          { id: 'side', type: 'side' as const, container: multiViewSideRef.current },
          { id: 'iso', type: 'iso' as const, container: multiViewIsoRef.current },
          { id: 'front', type: 'front' as const, container: multiViewFrontRef.current },
        ]
      : [];

    scene.setMultiViewConfigs(multiViewConfigs);
    scene.setLayeredView(isLayeredView);
    (Object.entries(layerVisibility) as Array<[CanvasNodeKind, boolean]>).forEach(([kind, visible]) => {
      scene.setLayerVisibility(kind, visible);
    });
    scene.setEdgePanConfig({ speed: edgePanSpeedPref, threshold: edgePanThresholdPref });
    scene.setZoomAggressiveness(zoomAggressiveness);
    scene.setLinksConfig({ clusterEdges: linkVisibility.clusterEdges, semanticEdges: linkVisibility.semanticEdges });
    scene.updateSize();

    const resizeObserver = new ResizeObserver(() => {
      scene.updateSize();
    });
    resizeObserver.observe(container);
    resizeObserverRef.current = resizeObserver;

    return () => {
      resizeObserver.disconnect();
      resizeObserverRef.current = null;
      scene.dispose();
      sceneRef.current = null;
    };
  }, [isMinimapVisible, isMultiViewVisible, isLayeredView, layerVisibility, edgePanSpeedPref, edgePanThresholdPref, zoomAggressiveness, linkVisibility]);

  useEffect(() => {
    if (!sceneRef.current) {
      return;
    }

    const container = isMinimapVisible ? minimapContainerRef.current : null;
    sceneRef.current.setMinimapContainer(container);
    if (container) {
      sceneRef.current.updateSize();
    }
  }, [isMinimapVisible]);

  useEffect(() => {
    const scene = sceneRef.current;
    if (!scene) {
      return;
    }

    if (!isMultiViewVisible) {
      scene.setMultiViewConfigs([]);
      return;
    }

    if (!multiViewTopRef.current || !multiViewSideRef.current || !multiViewIsoRef.current || !multiViewFrontRef.current) {
      return;
    }

    scene.setMultiViewConfigs([
      { id: 'top', type: 'top', container: multiViewTopRef.current },
      { id: 'side', type: 'side', container: multiViewSideRef.current },
      { id: 'iso', type: 'iso', container: multiViewIsoRef.current },
      { id: 'front', type: 'front', container: multiViewFrontRef.current },
    ]);
  }, [isMultiViewVisible]);

  useEffect(() => {
    const scene = sceneRef.current;
    if (!scene) {
      return;
    }
    (Object.entries(layerVisibility) as Array<[CanvasNodeKind, boolean]>).forEach(([kind, visible]) => {
      scene.setLayerVisibility(kind, visible);
    });
  }, [layerVisibility]);

  useEffect(() => {
    sceneRef.current?.setLayeredView(isLayeredView);
  }, [isLayeredView]);

  useEffect(() => {
    sceneRef.current?.setEdgePanConfig({ speed: edgePanSpeedPref, threshold: edgePanThresholdPref });
  }, [edgePanSpeedPref, edgePanThresholdPref]);

  useEffect(() => {
    sceneRef.current?.setZoomAggressiveness(zoomAggressiveness);
  }, [zoomAggressiveness]);

  useEffect(() => {
    sceneRef.current?.setLinksConfig({ clusterEdges: linkVisibility.clusterEdges, semanticEdges: linkVisibility.semanticEdges });
  }, [linkVisibility]);

  useEffect(() => {
    if (sceneRef.current) {
      sceneRef.current.updateNodes(nodes);
    }
  }, [nodes]);

  useEffect(() => {
    if (sceneRef.current) {
      sceneRef.current.toggleGrid(isGridVisible);
    }
  }, [isGridVisible]);

  const handleToggleGrid = () => {
    setIsGridVisible((prev) => !prev);
  };

  const handleToggleMinimap = () => {
    setIsMinimapVisible((prev) => !prev);
  };

  const handleToggleMultiView = () => {
    setIsMultiViewSwitching(true);
    setIsMultiViewVisible((prev) => !prev);
    window.setTimeout(() => setIsMultiViewSwitching(false), 300);
  };

  const handleMultiViewFocus = (type: 'top' | 'side' | 'iso' | 'front') => () => {
    sceneRef.current?.jumpToMultiView(type);
  };

  const handleToggleLayeredView = () => {
    setIsLayeredSwitching(true);
    setIsLayeredView((prev) => !prev);
    window.setTimeout(() => setIsLayeredSwitching(false), 300);
  };

  const toggleLayerVisibility = (kind: CanvasNodeKind) => () => {
    setLayerVisibility((prev) => ({
      ...prev,
      [kind]: !prev[kind],
    }));
  };

  const toggleLinks = (key: 'clusterEdges' | 'semanticEdges') => () => {
    setLinkVisibility((prev) => ({ ...prev, [key]: !prev[key] }));
  };

  const multiViewPresets = [
    { key: 'top' as const, label: 'Top', ref: multiViewTopRef },
    { key: 'side' as const, label: 'Side', ref: multiViewSideRef },
    { key: 'iso' as const, label: 'Iso', ref: multiViewIsoRef },
    { key: 'front' as const, label: 'Front', ref: multiViewFrontRef },
  ];

  const selectionDetails = selectedNode
    ? (() => {
        const base = [
          { label: 'Node ID', value: selectedNode.id },
          { label: 'Type', value: selectedNode.kind },
          { label: 'Abstraction', value: selectedNode.abstractionLevel.toString() },
          { label: 'Position', value: `${selectedNode.position.x}, ${selectedNode.position.y}, ${selectedNode.position.z}` },
        ];
        const specifics: { label: string; value: string }[] = [];
        if (selectedNode.kind === 'content') {
          if (selectedNode.documentTitle) specifics.push({ label: 'Title', value: selectedNode.documentTitle });
          if (selectedNode.mediaType) specifics.push({ label: 'Media', value: selectedNode.mediaType });
          if (typeof selectedNode.tokenCount === 'number') specifics.push({ label: 'Tokens', value: String(selectedNode.tokenCount) });
          if (selectedNode.contentSourceId) specifics.push({ label: 'Source ID', value: selectedNode.contentSourceId });
          if (selectedNode.documentUrl) specifics.push({ label: 'URL', value: selectedNode.documentUrl });
        } else if (selectedNode.kind === 'chunk') {
          specifics.push({ label: 'Sequence', value: String(selectedNode.sequenceIndex) });
          if (selectedNode.chunkType) specifics.push({ label: 'Chunk Type', value: selectedNode.chunkType });
          if (selectedNode.contentSourceId) specifics.push({ label: 'Source ID', value: selectedNode.contentSourceId });
        } else if (selectedNode.kind === 'cluster') {
          specifics.push({ label: 'Scope', value: selectedNode.clusterScope });
          if (typeof selectedNode.memberCount === 'number') specifics.push({ label: 'Members', value: String(selectedNode.memberCount) });
        }
        if (selectedNode.title && !specifics.some(d => d.label === 'Title')) base.push({ label: 'Title', value: selectedNode.title });
        if (selectedNode.displayContent) specifics.push({ label: 'Display', value: selectedNode.displayContent });
        return [...base, ...specifics];
      })()
    : [];

  const handleKeyDown: React.KeyboardEventHandler<HTMLDivElement> = (e) => {
    if (!sceneRef.current) return;
    switch (e.key) {
      case 'ArrowLeft':
        e.preventDefault();
        sceneRef.current.panByDirection('left');
        break;
      case 'ArrowRight':
        e.preventDefault();
        sceneRef.current.panByDirection('right');
        break;
      case 'ArrowUp':
        e.preventDefault();
        sceneRef.current.panByDirection('up');
        break;
      case 'ArrowDown':
        e.preventDefault();
        sceneRef.current.panByDirection('down');
        break;
      case '+':
      case '=': // keyboards without shift
        e.preventDefault();
        sceneRef.current.zoomInCenter();
        break;
      case '-':
      case '_':
        e.preventDefault();
        sceneRef.current.zoomOutCenter();
        break;
      case 'Tab': {
        e.preventDefault();
        const delta = e.shiftKey ? -1 : 1;
        sceneRef.current.selectNextVisible(delta as 1 | -1);
        break;
      }
      case 'Enter':
        e.preventDefault();
        sceneRef.current.focusSelected();
        break;
      case 'Escape':
        e.preventDefault();
        sceneRef.current.clearSelection();
        break;
      default:
        break;
    }
  };

  return (
    <div className={cn('flex h-full flex-col bg-slate-950 text-slate-100', className)}>
      <header className="flex items-center justify-between border-b border-slate-800 px-5 py-3">
        <div>
          <h1 className="text-lg font-semibold text-slate-50">{space.title || 'Space Canvas'}</h1>
          {space.description && (
            <p className="text-sm text-slate-400 line-clamp-2">{space.description}</p>
          )}
        </div>
        <div className="flex items-center gap-2">
          <Badge variant="outline" className="border-slate-700 bg-slate-900 text-slate-200">
            {totalCounts.total} nodes
          </Badge>
          <div className="hidden items-center gap-1 sm:flex">
            <Badge variant="outline" className="border-slate-700 bg-slate-900 text-slate-300">
              {totalCounts.cluster} clusters
            </Badge>
            <Badge variant="outline" className="border-slate-700 bg-slate-900 text-slate-300">
              {totalCounts.content} content
            </Badge>
            <Badge variant="outline" className="border-slate-700 bg-slate-900 text-slate-300">
              {totalCounts.chunk} chunks
            </Badge>
          </div>
          <div className="hidden items-center gap-1 xl:flex">
            {(
              [
                { key: 'clusterEdges', label: 'Cluster Links' },
                { key: 'semanticEdges', label: 'Semantic Links' },
              ] as Array<{ key: 'clusterEdges' | 'semanticEdges'; label: string }>
            ).map(({ key, label }) => (
              <Button
                key={key}
                variant="ghost"
                size="sm"
                onClick={toggleLinks(key)}
                className={cn(
                  'h-8 px-3 text-xs font-medium uppercase tracking-wide text-slate-200 hover:bg-slate-800',
                  linkVisibility[key] ? 'bg-slate-800/50' : 'opacity-60'
                )}
                aria-pressed={linkVisibility[key]}
                aria-label={`${linkVisibility[key] ? 'Hide' : 'Show'} ${label.toLowerCase()}`}
              >
                {label}
              </Button>
            ))}
          </div>
          <div className="hidden items-center gap-1 lg:flex">
            {(
              [
                { key: 'cluster', label: 'Clusters' },
                { key: 'content', label: 'Content' },
                { key: 'chunk', label: 'Chunks' },
              ] as Array<{ key: CanvasNodeKind; label: string }>
            ).map(({ key, label }) => (
              <Button
                key={key}
                variant="ghost"
                size="sm"
                onClick={toggleLayerVisibility(key)}
                className={cn(
                  'h-8 px-3 text-xs font-medium uppercase tracking-wide text-slate-200 hover:bg-slate-800',
                  layerVisibility[key] ? 'bg-slate-800/50' : 'opacity-60'
                )}
              >
                {label}
              </Button>
            ))}
          </div>
          <Button
            variant="ghost"
            size="sm"
            className={cn('text-slate-200 hover:bg-slate-800', isMinimapVisible && 'bg-slate-800/50')}
            onClick={handleToggleMinimap}
            title={isMinimapVisible ? 'Hide minimap' : 'Show minimap'}
            aria-pressed={isMinimapVisible}
            aria-label={isMinimapVisible ? 'Hide minimap' : 'Show minimap'}
          >
            <MapIcon className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            className={cn('text-slate-200 hover:bg-slate-800', isMultiViewVisible && 'bg-slate-800/50')}
            onClick={handleToggleMultiView}
            title={isMultiViewVisible ? 'Hide multi-view' : 'Show multi-view'}
            aria-pressed={isMultiViewVisible}
            aria-label={isMultiViewVisible ? 'Hide multi-view' : 'Show multi-view'}
            disabled={isLoading || isMultiViewSwitching}
          >
            <LayoutGrid className={cn('h-4 w-4', isMultiViewSwitching && 'animate-pulse')} />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            className={cn('text-slate-200 hover:bg-slate-800', isLayeredView && 'bg-indigo-600/40 text-indigo-200 hover:bg-indigo-600/40')}
            onClick={handleToggleLayeredView}
            title={isLayeredView ? 'Disable side-layer view' : 'Enable side-layer view'}
            aria-pressed={isLayeredView}
            aria-label={isLayeredView ? 'Disable side-layer view' : 'Enable side-layer view'}
            disabled={isLoading || isLayeredSwitching}
          >
            <span className={cn(isLayeredSwitching && 'animate-pulse')}>Side View</span>
          </Button>
          <Button
            variant="ghost"
            size="sm"
            className="text-slate-200 hover:bg-slate-800"
            onClick={handleToggleGrid}
            title={isGridVisible ? 'Hide grid' : 'Show grid'}
            aria-pressed={isGridVisible}
            aria-label={isGridVisible ? 'Hide grid' : 'Show grid'}
          >
            {isGridVisible ? <Grid3x3 className="h-4 w-4" /> : <Grid className="h-4 w-4" />}
          </Button>
          <Button
            variant="ghost"
            size="sm"
            className={cn('text-slate-200 hover:bg-slate-800', showSettings && 'bg-slate-800/50')}
            onClick={() => setShowSettings((prev) => !prev)}
            title="Canvas settings"
            aria-pressed={showSettings}
            aria-label="Open canvas settings"
          >
            Settings
          </Button>
          <Button
            variant="ghost"
            size="sm"
            className="text-slate-200 hover:bg-slate-800"
            onClick={reloadNodes}
            disabled={isLoading}
            aria-busy={isLoading}
            aria-label="Reload nodes"
          >
            <RefreshCw className={cn('h-4 w-4', isLoading && 'animate-spin')} />
          </Button>
        </div>
      </header>

      <div className="relative flex-1 overflow-hidden">
        <div
          ref={canvasContainerRef}
          className="absolute inset-0"
          tabIndex={0}
          role="application"
          aria-label="3D canvas. Use arrow keys to pan, plus and minus to zoom, Tab to cycle selection, Enter to focus selected, Escape to clear."
          onKeyDown={handleKeyDown}
        />
        <div
          className={cn(
            'pointer-events-none absolute bottom-4 left-4 overflow-hidden rounded-lg border border-slate-700 bg-slate-900/80 backdrop-blur transition-opacity',
            'h-28 w-40 md:h-36 md:w-48',
            isMinimapVisible ? 'opacity-100' : 'opacity-0'
          )}
        >
          <div ref={minimapContainerRef} className="h-full w-full" />
          <div className="pointer-events-none absolute top-2 left-2 rounded bg-slate-900/80 px-2 py-1 text-[10px] font-semibold uppercase tracking-wide text-slate-200">
            {Math.round(fps)} FPS
          </div>
        </div>
        <div
          className={cn(
            'absolute top-4 right-4 z-20 transition-opacity',
            'w-44 sm:w-52',
            isMultiViewVisible ? 'pointer-events-auto opacity-100' : 'pointer-events-none opacity-0'
          )}
        >
          <div className="grid gap-2 sm:grid-cols-2">
            {multiViewPresets.map(({ key, label, ref }) => (
              <button
                key={key}
                type="button"
                onClick={handleMultiViewFocus(key)}
                className="group relative h-24 md:h-28 w-full overflow-hidden rounded-lg border border-slate-700 bg-slate-900/80 backdrop-blur transition ring-offset-2 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-400"
                aria-label={`Snap main view to ${label.toLowerCase()} preset`}
              >
                <div ref={ref} className="h-full w-full" />
                <span className="pointer-events-none absolute bottom-2 right-2 rounded bg-slate-900/80 px-2 py-1 text-xs font-medium uppercase text-slate-200 opacity-70 group-hover:opacity-100">
                  {label}
                </span>
              </button>
            ))}
          </div>
        </div>

        <div
          className={cn(
            'absolute top-20 right-4 z-30 w-64 rounded-lg border border-slate-700 bg-slate-900/95 p-4 text-xs text-slate-200 shadow-xl backdrop-blur transition-opacity',
            showSettings ? 'pointer-events-auto opacity-100' : 'pointer-events-none opacity-0'
          )}
        >
          <div className="mb-3 text-sm font-semibold text-slate-100">Navigation Settings</div>
          <div className="space-y-3">
            <label className="flex flex-col gap-1">
              <span className="uppercase tracking-wide text-[11px] text-slate-400">Edge Pan Speed ({edgePanSpeedPref.toFixed(0)})</span>
              <input
                type="range"
                min={40}
                max={240}
                step={10}
                value={edgePanSpeedPref}
                onChange={(event) => setEdgePanSpeedPref(Number(event.target.value))}
                className="accent-indigo-400"
              />
            </label>
            <label className="flex flex-col gap-1">
              <span className="uppercase tracking-wide text-[11px] text-slate-400">Edge Pan Threshold ({edgePanThresholdPref.toFixed(2)})</span>
              <input
                type="range"
                min={0.05}
                max={0.25}
                step={0.01}
                value={edgePanThresholdPref}
                onChange={(event) => setEdgePanThresholdPref(Number(event.target.value))}
                className="accent-indigo-400"
              />
            </label>
            <label className="flex flex-col gap-1">
              <span className="uppercase tracking-wide text-[11px] text-slate-400">Zoom Aggressiveness ({zoomAggressiveness.toFixed(2)})</span>
              <input
                type="range"
                min={0.1}
                max={0.6}
                step={0.05}
                value={zoomAggressiveness}
                onChange={(event) => setZoomAggressivenessPref(Number(event.target.value))}
                className="accent-indigo-400"
              />
            </label>
          </div>
        </div>

        {isLoading && (
          <div className="pointer-events-none absolute inset-0 flex items-center justify-center bg-slate-950/70 backdrop-blur-sm">
            <div className="rounded-md border border-slate-700 bg-slate-900 px-4 py-2 text-sm text-slate-200">
              Loading canvas data…
            </div>
          </div>
        )}

        {errorMessage && (
          <div className="absolute top-4 right-4 max-w-sm rounded-md border border-rose-700 bg-rose-950/80 px-4 py-3 text-sm text-rose-100">
            <p className="font-semibold">Unable to load canvas</p>
            <p className="mt-1 text-xs opacity-80">{errorMessage}</p>
            <div className="mt-2 flex items-center gap-2">
              <Button size="sm" variant="secondary" onClick={reloadNodes} aria-label="Retry now">
                Retry Now
              </Button>
              {retryInMs != null && (
                <span className="text-xs opacity-75">Retrying in {Math.ceil(retryInMs / 1000)}s…</span>
              )}
            </div>
          </div>
        )}

        {selectedNode && (
          <div className="absolute bottom-4 right-4 w-72 rounded-lg border border-slate-700 bg-slate-900/90 p-4 shadow-lg backdrop-blur">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-semibold text-slate-100">Selected node</h3>
              <Badge variant="outline" className="border-slate-700 text-xs text-slate-300">
                {selectedNode.kind}
              </Badge>
            </div>
            <dl className="mt-3 space-y-2 text-xs text-slate-300">
              {selectionDetails.map((detail) => (
                <div key={detail!.label}>
                  <dt className="font-medium text-slate-400">{detail!.label}</dt>
                  <dd className="mt-0.5 break-words text-slate-200">{detail!.value}</dd>
                </div>
              ))}
            </dl>
            {selectedNode.keywords.length > 0 && (
              <div className="mt-3 flex flex-wrap gap-1">
                {selectedNode.keywords.slice(0, 6).map((keyword) => (
                  <Badge
                    key={keyword}
                    variant="outline"
                    className="border-slate-700 bg-slate-800 text-[11px] text-slate-200"
                  >
                    {keyword}
                  </Badge>
                ))}
              </div>
            )}
            {selectedNode.kind === 'content' && (
              <div className="mt-3 flex gap-2">
                <Button
                  size="sm"
                  variant="secondary"
                  onClick={() => router.push(`/spaces/${space.id}/content/${selectedNode.contentSourceId}`)}
                >
                  Open Preview
                </Button>
                {selectedNode.documentUrl && (
                  <Button size="sm" variant="outline" onClick={() => window.open(selectedNode.documentUrl!, '_blank')}>
                    Open URL
                  </Button>
                )}
              </div>
            )}
          </div>
        )}

        {(linkVisibility.clusterEdges || linkVisibility.semanticEdges) && (
          <div className="pointer-events-none absolute top-4 left-4 z-10 rounded-md border border-slate-700 bg-slate-900/80 px-3 py-2 text-[11px] text-slate-200">
            <div className="font-semibold mb-1">Legend</div>
            {linkVisibility.clusterEdges && (
              <div className="flex items-center gap-2">
                <span className="inline-block h-0.5 w-4 bg-slate-400" /> Cluster Links
              </div>
            )}
            {linkVisibility.semanticEdges && (
              <div className="mt-1 flex items-center gap-2">
                <span className="inline-block h-0.5 w-4 bg-emerald-400" /> Semantic Links
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
