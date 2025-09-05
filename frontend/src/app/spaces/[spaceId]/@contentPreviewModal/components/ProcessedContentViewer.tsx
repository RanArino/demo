'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import { generateDownloadURL } from '@/api/actions/contentActions';
import { Skeleton } from '@/components/ui/skeleton';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import remarkBreaks from 'remark-breaks';
import { Button } from '@/components/ui/button';
import { markdownToPlainText, copyMarkdownToClipboard } from '@/app/spaces/utils/markdown';
import { useToast } from '@/components/ui/use-toast';
import { Clipboard, ClipboardCheck } from 'lucide-react';

interface ProcessedContentViewerProps {
  contentSourceId: string;
}

export default function ProcessedContentViewer({ contentSourceId }: ProcessedContentViewerProps) {
  const [markdown, setMarkdown] = useState<string>('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [viewMode, setViewMode] = useState<'markdown' | 'text'>('markdown');
  const [isCopying, setIsCopying] = useState(false);
  const { toast } = useToast();

  const fetchProcessed = useCallback(async () => {
    setLoading(true);
    setError(null);
    const res = await generateDownloadURL(contentSourceId, 'processed');
    if (!res.ok || !res.data) {
      setError(res.error?.message || 'Failed to get processed content URL');
      setLoading(false);
      return;
    }
    try {
      const response = await fetch(res.data.url);
      const text = await response.text();
      setMarkdown(text);
    } catch (error) {
      setError('Failed to load processed content');
    } finally {
      setLoading(false);
    }
  }, [contentSourceId]);

  useEffect(() => {
    fetchProcessed();
  }, [fetchProcessed]);

  const plainText = useMemo(() => (markdown ? markdownToPlainText(markdown) : ''), [markdown]);

  const handleCopy = useCallback(async () => {
    if (!markdown) return;
    setIsCopying(true);
    try {
      if (viewMode === 'markdown') {
        await copyMarkdownToClipboard(markdown);
        toast({
          title: 'Copied',
          description: 'Markdown copied to clipboard',
        });
      } else {
        await navigator.clipboard.writeText(plainText);
        toast({
          title: 'Copied',
          description: 'Plain text copied to clipboard',
        });
      }
    } catch {
      toast({ title: 'Copy failed', description: 'Unable to write to clipboard', variant: 'destructive' });
    } finally {
      setTimeout(() => setIsCopying(false), 1000);
    }
  }, [markdown, plainText, viewMode, toast]);

  if (loading) {
    return <Skeleton className="w-full h-[60vh]" />;
  }
  if (error) {
    return <div className="text-sm text-red-600 bg-red-50 p-3 rounded">{error}</div>;
  }

  return (
    <div className="border rounded h-[70vh] flex flex-col overflow-hidden">
      <div className="flex items-center justify-between p-2 border-b bg-muted/30">
        <div className="flex items-center gap-2">
          <Button size="sm" variant={viewMode === 'markdown' ? 'secondary' : 'ghost'} onClick={() => setViewMode('markdown')}>
            Markdown
          </Button>
          <Button size="sm" variant={viewMode === 'text' ? 'secondary' : 'ghost'} onClick={() => setViewMode('text')}>
            Text
          </Button>
        </div>
        <Button size="sm" variant="secondary" onClick={handleCopy} disabled={isCopying}>
          {isCopying ? <ClipboardCheck className="h-4 w-4 mr-2" /> : <Clipboard className="h-4 w-4 mr-2" />}
          Copy
        </Button>
      </div>
      <div className="flex-1 overflow-auto">
        {viewMode === 'markdown' && (
          <div className="prose max-w-none p-4 whitespace-pre-wrap break-words prose-pre:whitespace-pre-wrap">
            <ReactMarkdown remarkPlugins={[remarkGfm, remarkBreaks]}>{markdown}</ReactMarkdown>
          </div>
        )}
        {viewMode === 'text' && <pre className="p-4 text-sm whitespace-pre-wrap break-words">{plainText}</pre>}
      </div>
    </div>
  );
}