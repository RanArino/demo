'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Button } from '@/components/ui/button';
import { useToast } from '@/components/ui/use-toast';
import { generateDownloadURL, getContentSource } from '@/api/actions/contentActions';
import OriginalContentViewer from './OriginalContentViewer';
import ProcessedContentViewer from './ProcessedContentViewer';
import { ChevronDown, FileText } from 'lucide-react';
import { useRouter, usePathname } from 'next/navigation';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { ContentSource } from '@/app/spaces/types/content';

interface ContentPreviewModalProps {
  isOpen: boolean;
  onClose: () => void;
  contentSourceId?: string;
}

export default function ContentPreviewModal({ isOpen, onClose, contentSourceId }: ContentPreviewModalProps) {
  const [activeTab, setActiveTab] = useState<'original' | 'processed'>('processed');
  const [contentSource, setContentSource] = useState<ContentSource | null>(null);
  const { toast } = useToast();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (isOpen) {
      setActiveTab('processed');
      if (contentSourceId) {
        getContentSource(contentSourceId).then((result) => {
          if (result.ok && result.data) {
            setContentSource(result.data);
          }
        });
      }
    } else {
      setContentSource(null);
    }
  }, [isOpen, contentSourceId]);

  const title = useMemo(() => 'Preview', []);

  const originalFileFormat = useMemo(() => {
    if (!contentSource) return 'Original File';
    const mimeType = contentSource.mimeType;
    if (!mimeType) return 'Original File';
    const parts = mimeType.split('/');
    const fileType = parts.length > 1 ? parts[1] : parts[0];
    const friendlyNames: Record<string, string> = {
      pdf: 'PDF',
      'vnd.openxmlformats-officedocument.wordprocessingml.document': 'Word',
      plain: 'Text',
      markdown: 'Markdown',
    };
    return friendlyNames[fileType] || fileType.toUpperCase();
  }, [contentSource]);

  const handleDownload = useCallback(
    async (kind: 'original' | 'processed') => {
      if (!contentSourceId) return;
      const res = await generateDownloadURL(contentSourceId, kind);
      if (res.ok && res.data) {
        window.open(res.data.url, '_blank');
      } else {
        toast({
          title: 'Download failed',
          description: res.error?.message || 'Unable to get download URL',
          variant: 'destructive',
        });
      }
    },
    [contentSourceId, toast]
  );

  return (
    <Dialog
      open={isOpen}
      onOpenChange={(open) => {
        if (!open) {
          onClose();
          const match = (pathname ?? '').match(/^(\/spaces\/[^/]+)/);
          if (match) {
            router.push(match[1]);
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
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button size="sm" variant="secondary">
                    <FileText className="h-4 w-4 mr-2" />
                    Open
                    <ChevronDown className="h-4 w-4 ml-1" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem onClick={() => handleDownload('original')}>{originalFileFormat}</DropdownMenuItem>
                  <DropdownMenuItem onClick={() => handleDownload('processed')}>Text/Markdown</DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
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
              {contentSourceId && <OriginalContentViewer contentSourceId={contentSourceId} />}
            </TabsContent>
            <TabsContent value="processed">
              {contentSourceId && <ProcessedContentViewer contentSourceId={contentSourceId} />}
            </TabsContent>
          </div>
        </Tabs>
      </DialogContent>
    </Dialog>
  );
}