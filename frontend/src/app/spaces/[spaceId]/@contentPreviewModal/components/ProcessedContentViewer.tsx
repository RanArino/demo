'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import { generateDownloadURL, getContentSource } from '@/api/actions/contentActions';
import { Skeleton } from '@/components/ui/skeleton';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import remarkBreaks from 'remark-breaks';
import { Button } from '@/components/ui/button';
import { markdownToPlainText } from '@/app/spaces/utils/markdown';

interface ProcessedContentViewerProps {
  contentSourceId: string;
  onCopy?: (mode: 'markdown' | 'text', markdown?: string) => void;
}

export default function ProcessedContentViewer({ contentSourceId, onCopy }: ProcessedContentViewerProps) {
  const [markdown, setMarkdown] = useState<string>('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [viewMode, setViewMode] = useState<'rendered' | 'raw' | 'text'>('rendered');

  const fetchProcessed = useCallback(async () => {
    setLoading(true);
    setError(null);
    console.log('Fetching processed content for:', contentSourceId);
    const res = await generateDownloadURL(contentSourceId, 'processed');
    console.log('Download URL response:', res);
    if (!res.ok || !res.data) {
      setError(res.error?.message || 'Failed to get processed content URL');
      setLoading(false);
      return;
    }
    try {
      console.log('Fetching from URL:', res.data.url);
      const response = await fetch(res.data.url);
      console.log('Fetch response status:', response.status, response.headers.get('content-type'));
      const text = await response.text();
      console.log('Response text preview:', text.substring(0, 200));
      setMarkdown(text);
    } catch (error) {
      console.error('Fetch error:', error);
      setError('Failed to load processed content');
    } finally {
      setLoading(false);
    }
  }, [contentSourceId]);

  useEffect(() => {
    fetchProcessed();
  }, [fetchProcessed]);

  useEffect(() => {
    const listener = (e: Event) => {
      const detail = (e as CustomEvent).detail as { mode: 'markdown' | 'text' };
      onCopy?.(detail.mode, markdown);
    };
    document.addEventListener('content-preview-copy', listener as EventListener);
    return () => document.removeEventListener('content-preview-copy', listener as EventListener);
  }, [onCopy, markdown]);

  const content = useMemo(() => markdown, [markdown]);
  const plainText = useMemo(() => (markdown ? markdownToPlainText(markdown) : ''), [markdown]);

  if (loading) {
    return <Skeleton className="w-full h-[60vh]" />;
  }
  if (error) {
    return <div className="text-sm text-red-600 bg-red-50 p-3 rounded">{error}</div>;
  }

  return (
    <div className="border rounded h-[70vh] flex flex-col overflow-hidden">
      <div className="flex items-center gap-2 p-2 border-b bg-muted/30">
        <Button size="sm" variant={viewMode === 'rendered' ? 'secondary' : 'ghost'} onClick={() => setViewMode('rendered')}>Markdown</Button>
        <Button size="sm" variant={viewMode === 'raw' ? 'secondary' : 'ghost'} onClick={() => setViewMode('raw')}>Raw</Button>
        <Button size="sm" variant={viewMode === 'text' ? 'secondary' : 'ghost'} onClick={() => setViewMode('text')}>Text</Button>
      </div>
      <div className="flex-1 overflow-auto">
        {viewMode === 'rendered' && (
          <div className="prose max-w-none p-4 whitespace-pre-wrap break-words prose-pre:whitespace-pre-wrap">
            <ReactMarkdown remarkPlugins={[remarkGfm, remarkBreaks]}>{content}</ReactMarkdown>
          </div>
        )}
        {viewMode === 'raw' && (
          <pre className="p-4 text-sm whitespace-pre-wrap break-words">{content}</pre>
        )}
        {viewMode === 'text' && (
          <pre className="p-4 text-sm whitespace-pre-wrap break-words">{plainText}</pre>
        )}
      </div>
    </div>
  );
}
