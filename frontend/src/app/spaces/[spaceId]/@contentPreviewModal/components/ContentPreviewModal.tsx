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
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { ContentSource } from '@/app/spaces/types/content';
import { getFriendlyNameFromMimeType } from '@/lib/mime-types';

interface ContentPreviewModalProps {
  isOpen: boolean;
  onClose: () => void;
  contentSourceId?: string;
}

const DIALOG_TITLE = 'Preview';

export default function ContentPreviewModal({ isOpen, onClose, contentSourceId }: ContentPreviewModalProps) {
  const [activeTab, setActiveTab] = useState<'original' | 'processed'>('processed');
  const [contentSource, setContentSource] = useState<ContentSource | null>(null);
  const { toast } = useToast();

  useEffect(() => {
    let isMounted = true;

    if (isOpen && contentSourceId) {
      setActiveTab('processed');
      getContentSource(contentSourceId).then((result) => {
        if (isMounted && result.ok && result.data) {
          setContentSource(result.data);
        }
      });
    } else {
      setContentSource(null);
    }

    return () => {
      isMounted = false;
    };
  }, [isOpen, contentSourceId]);

  const originalFileFormat = useMemo(() => {
    if (!contentSource?.mimeType) return 'Original File';
    return getFriendlyNameFromMimeType(contentSource.mimeType);
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

  const handleOpenChange = (open: boolean) => {
    if (!open) {
      onClose();
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={handleOpenChange}>
      <DialogContent className="max-w-5xl w-[90vw]">
        <DialogHeader>
          <div className="flex items-center justify-between">
            <DialogTitle className="truncate mr-3">{DIALOG_TITLE}</DialogTitle>
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

          <div className="mt-4 h-[60vh] overflow-y-auto">
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