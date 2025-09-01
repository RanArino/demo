'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Button } from '@/components/ui/button';
import { useToast } from '@/components/ui/use-toast';
import { generateDownloadURL } from '@/api/actions/contentActions';
import { copyMarkdownToClipboard, markdownToPlainText } from '@/app/spaces/utils/markdown';
import OriginalContentViewer from './OriginalContentViewer';
import ProcessedContentViewer from './ProcessedContentViewer';
import { Clipboard, ClipboardCheck, Download, ChevronDown } from 'lucide-react';
import { useRouter, usePathname } from 'next/navigation';

interface ContentPreviewModalProps {
  isOpen: boolean;
  onClose: () => void;
  // Optional: allow callers to pass an id directly via route
  contentSourceId?: string;
}

export default function ContentPreviewModal({ isOpen, onClose, contentSourceId }: ContentPreviewModalProps) {
  const [activeTab, setActiveTab] = useState<'original' | 'processed'>('processed');
  const [showCopyMenu, setShowCopyMenu] = useState(false);
  const [isCopying, setIsCopying] = useState(false);
  const { toast } = useToast();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (isOpen) {
      setActiveTab('processed');
      setShowCopyMenu(false);
    }
  }, [isOpen]);

  const title = useMemo(() => 'Preview', []);

  const handleDownload = useCallback(async (kind: 'original' | 'processed') => {
    if (!contentSourceId) return;
    const res = await generateDownloadURL(contentSourceId, kind);
    if (res.ok && res.data) {
      window.open(res.data.url, '_blank');
    } else {
      toast({ title: 'Download failed', description: res.error?.message || 'Unable to get download URL', variant: 'destructive' });
    }
  }, [contentSourceId, toast]);

  const handleCopy = useCallback(async (mode: 'markdown' | 'text', markdown?: string) => {
    if (!markdown) return;
    setIsCopying(true);
    try {
      if (mode === 'markdown') {
        await copyMarkdownToClipboard(markdown);
      } else {
        await navigator.clipboard.writeText(markdownToPlainText(markdown));
      }
      toast({ title: 'Copied', description: mode === 'markdown' ? 'Markdown copied to clipboard' : 'Plain text copied to clipboard' });
    } catch {
      toast({ title: 'Copy failed', description: 'Unable to write to clipboard', variant: 'destructive' });
    } finally {
      setIsCopying(false);
      setShowCopyMenu(false);
    }
  }, [toast]);

  return (
    <Dialog
      open={isOpen}
      onOpenChange={(open) => {
        if (!open) {
          onClose();
          const segments = (pathname ?? '').split('/').filter(Boolean);
          if (segments.length >= 2) {
            const target = '/' + segments.slice(0, 2).join('/');
            router.push(target);
          } else {
            router.push('/spaces');
          }
        }
      }}
    >
      <DialogContent className="max-w-5xl w-[90vw]" aria-describedby={undefined}>
        <DialogHeader>
          <div className="flex items-center justify-between">
            <DialogTitle className="truncate mr-3">{title}</DialogTitle>
            <div className="relative flex items-center gap-2">
              <div className="relative">
                <Button size="sm" variant="secondary" onClick={() => setShowCopyMenu((v) => !v)}>
                  {isCopying ? <ClipboardCheck className="h-4 w-4 mr-2" /> : <Clipboard className="h-4 w-4 mr-2" />} Copy <ChevronDown className="h-4 w-4 ml-1" />
                </Button>
                {showCopyMenu && (
                  <div className="absolute right-0 mt-2 w-44 bg-white border rounded-lg shadow-lg z-10">
                    <button className="w-full text-left px-3 py-2 hover:bg-gray-50 text-sm" onClick={() => document.dispatchEvent(new CustomEvent('content-preview-copy', { detail: { mode: 'markdown' } }))}>Copy as Markdown</button>
                    <button className="w-full text-left px-3 py-2 hover:bg-gray-50 text-sm" onClick={() => document.dispatchEvent(new CustomEvent('content-preview-copy', { detail: { mode: 'text' } }))}>Copy as Text</button>
                  </div>
                )}
              </div>
              <div className="relative">
                <Button size="sm" onClick={() => handleDownload('original')}>
                  <Download className="h-4 w-4 mr-2" /> Download
                </Button>
                <div className="absolute right-0" />
              </div>
              <div className="relative">
                <Button size="sm" variant="outline" onClick={() => handleDownload('processed')}>Processed (.md)</Button>
              </div>
            </div>
          </div>
        </DialogHeader>

        <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as 'original' | 'processed')}>
          <TabsList>
            <TabsTrigger value="original">Original</TabsTrigger>
            <TabsTrigger value="processed">Processed</TabsTrigger>
          </TabsList>

          <div className="mt-4">
            <TabsContent value="original">
              {contentSourceId && (
                <OriginalContentViewer contentSourceId={contentSourceId} />
              )}
            </TabsContent>
            <TabsContent value="processed">
              {contentSourceId && (
                <ProcessedContentViewer contentSourceId={contentSourceId} onCopy={handleCopy} />
              )}
            </TabsContent>
          </div>
        </Tabs>
      </DialogContent>
    </Dialog>
  );
}
