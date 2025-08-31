'use client';

import { useRouter } from 'next/navigation';
import { useState, useCallback, useMemo, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { FileText } from 'lucide-react';
import { cn } from '@/lib/utils';
import { ContentSource, ContentSourceStatus } from '@/app/spaces/types/content';
import { generateDownloadURL, deleteContentSource } from '@/api/actions/contentActions';
import { useToast } from '@/components/ui/use-toast';
import { useProcessingPoller } from '@/app/spaces/hooks/useProcessingPoller';
import ContentSourcesHeader from './ContentSourcesHeader';
import ContentSourceCard from '@/components/content/ContentSourceCard';

interface ContentSourcesSectionProps {
  spaceId: string;
  contentSources: ContentSource[];
  className?: string;
  errorMessage?: string;
}

export default function ContentSourcesSection({ spaceId, contentSources, className, errorMessage }: ContentSourcesSectionProps) {
  const router = useRouter();
  const { toast } = useToast();
  const [sources, setSources] = useState<ContentSource[]>(contentSources);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(() => new Set(contentSources.map((s) => s.id)));
  const [progressById, setProgressById] = useState<Record<string, number>>({});

  useEffect(() => {
    const onCreated = (e: CustomEvent<ContentSource>) => {
      const detail = e.detail;
      setSources((prev) => {
        if (prev.some((s) => s.id === detail.id)) return prev;
        return [detail, ...prev];
      });
    };
    const onUpdated = (e: CustomEvent<ContentSource>) => {
      const detail = e.detail;
      setSources((prev) => prev.map((s) => (s.id === detail.id ? { ...s, ...detail } : s)));
      if (detail.processingStatus === ContentSourceStatus.COMPLETED) {
        setSelectedIds((prev) => {
          const next = new Set(prev);
          next.add(detail.id);
          return next;
        });
        setProgressById((prev) => ({ ...prev, [detail.id]: 100 }));
        setTimeout(() => {
          setProgressById((prev) => {
            const { [detail.id]: _, ...rest } = prev;
            return rest;
          });
        }, 1500);
      }
    };
    const onProgress = (e: CustomEvent<{ contentSourceId: string; progress: number }>) => {
      const { contentSourceId, progress } = e.detail;
      setProgressById((prev) => ({ ...prev, [contentSourceId]: progress }));
    };
    const createdHandler = onCreated as unknown as EventListener;
    const updatedHandler = onUpdated as unknown as EventListener;
    const progressHandler = onProgress as unknown as EventListener;
    window.addEventListener('content:created', createdHandler);
    window.addEventListener('content:updated', updatedHandler);
    window.addEventListener('content:progress', progressHandler);
    return () => {
      window.removeEventListener('content:created', createdHandler);
      window.removeEventListener('content:updated', updatedHandler);
      window.removeEventListener('content:progress', progressHandler);
    };
  }, []);

  const handleAddContent = () => {
    router.push(`/spaces/${spaceId}/upload`);
  };

  const handleView = (source: ContentSource) => {
    router.push(`/spaces/${spaceId}/content/${source.id}`);
  };

  const handleSelectChange = (id: string, checked: boolean) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (checked) next.add(id); else next.delete(id);
      return next;
    });
  };

  const handleDownload = async (source: ContentSource, kind: 'original' | 'processed') => {
    const res = await generateDownloadURL(source.id, kind);
    if (res.ok && res.data) {
      window.open(res.data.url, '_blank');
    } else {
      toast({ title: 'Download failed', description: res.error?.message || 'Unable to get download URL', variant: 'destructive' });
    }
  };

  const handleDelete = async (source: ContentSource) => {
    if (!confirm('This will permanently delete the document. Continue?')) return;
    const prev = sources;
    setSources((s) => s.filter((x) => x.id !== source.id));
    const res = await deleteContentSource(source.id);
    if (res.ok) {
      toast({ title: 'Deleted', description: 'Document deleted successfully' });
    } else {
      toast({ title: 'Delete failed', description: res.error?.message || 'Unable to delete', variant: 'destructive' });
      setSources(prev);
    }
  };

  useProcessingPoller({
    spaceId,
    candidateIds: useMemo(() => sources.filter(s => s.processingStatus !== ContentSourceStatus.COMPLETED && s.processingStatus !== ContentSourceStatus.FAILED).map(s => s.id), [sources]),
    onUpdates: useCallback((updates) => {
      if (!updates || updates.length === 0) return;
      setSources((list) => list.map((s) => {
        const hit = updates.find((u) => u.id === s.id);
        return hit ? { ...s, processingStatus: hit.processingStatus } as ContentSource : s;
      }));

      updates.forEach((u) => {
        if (u.processingStatus === ContentSourceStatus.COMPLETED) {
          try {
            const completed = sources.find((s) => s.id === u.id);
            toast({ title: 'Processed Files', description: completed?.title || 'Your file is ready' });
          } catch {}
          setSelectedIds((prev) => {
            const next = new Set(prev);
            next.add(u.id);
            return next;
          });
          setProgressById((prev) => ({ ...prev, [u.id]: 100 }));
          setTimeout(() => {
            setProgressById((prev) => {
              const { [u.id]: _, ...rest } = prev;
              return rest;
            });
          }, 1500);
        }
        if (u.processingStatus === ContentSourceStatus.FAILED) {
          toast({ title: 'Processing failed', description: 'An error occurred while processing the file', variant: 'destructive' });
          setProgressById((prev) => {
            const { [u.id]: _, ...rest } = prev;
            return rest;
          });
        }
      });
    }, [sources, toast]),
  });

  const statusCounts = useMemo(() => ({
    total: sources.length,
    completed: sources.filter(s => s.processingStatus === ContentSourceStatus.COMPLETED).length,
    processing: sources.filter(s => s.processingStatus === ContentSourceStatus.PROCESSING).length,
    failed: sources.filter(s => s.processingStatus === ContentSourceStatus.FAILED).length,
  }), [sources]);

  return (
    <>
    <div className={cn("flex flex-col h-full", className)}>
      <ContentSourcesHeader
        total={statusCounts.total}
        processing={statusCounts.processing}
        failed={statusCounts.failed}
        onAdd={handleAddContent}
      />

      <div className="flex-1 overflow-y-auto p-4">
        {errorMessage && (
          <div className="mb-3 text-xs text-red-600 bg-red-50 border border-red-200 rounded p-2">
            {errorMessage}
          </div>
        )}
        {sources.length > 0 ? (
          <div className="space-y-3">
            {sources.map((source) => (
              <ContentSourceCard
                key={source.id}
                contentSource={source}
                onView={handleView}
                onDownload={handleDownload}
                onDelete={handleDelete}
                selected={selectedIds.has(source.id)}
                onSelectChange={handleSelectChange}
                progress={progressById[source.id]}
              />
            ))}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center h-full text-center">
            <div className="w-12 h-12 bg-gray-100 rounded-full flex items-center justify-center mb-3">
              <FileText className="h-6 w-6 text-gray-400" />
            </div>
            <h3 className="text-sm font-medium text-gray-900 mb-1">
              No content sources yet
            </h3>
            <p className="text-xs text-gray-600 mb-3">
              Add your first content source to get started
            </p>
            <Button
              onClick={handleAddContent}
              size="sm"
              variant="outline"
              className="flex items-center gap-2"
            >
              Add Content
            </Button>
          </div>
        )}
      </div>

    </div>

    {/* Preview modal handled via @contentPreviewModal route */}
  </>
  );
}
