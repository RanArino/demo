'use client';

import { useState, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useToast } from '@/components/ui/use-toast';
import { 
  Upload, 
  Link as LinkIcon, 
  Type, 
  FileText,
  AlertCircle
} from 'lucide-react';
import FileUploadArea from './FileUploadArea';
import LinkUploadForm from './LinkUploadForm';
import TextUploadForm from './TextUploadForm';
import GoogleDriveUpload from './GoogleDriveUpload';
import { UploadModalProps, ContentSource } from '@/app/spaces/types/content';

type UploadTab = 'file' | 'google-drive' | 'link' | 'text';

interface TabConfig {
  id: UploadTab;
  label: string;
  icon: React.ReactNode;
}

const UPLOAD_TABS: TabConfig[] = [
  {
    id: 'file',
    label: 'Files',
    icon: <Upload className="h-4 w-4" />
  },
  {
    id: 'google-drive',
    label: 'Google Drive',
    icon: <FileText className="h-4 w-4" />
  },
  {
    id: 'link',
    label: 'Link',
    icon: <LinkIcon className="h-4 w-4" />
  },
  {
    id: 'text',
    label: 'Text',
    icon: <Type className="h-4 w-4" />
  }
];

export default function UploadModal({ 
  spaceId, 
  spaceName,
  isOpen = false, 
  onClose,
  onUploadComplete,
  autoOpenTab = 'file'
}: UploadModalProps & { spaceName?: string }) {
  const [activeTab, setActiveTab] = useState<UploadTab>(autoOpenTab);
  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const router = useRouter();
  const { toast } = useToast();

  const handleClose = useCallback(() => {
    if (isUploading) {
      // Don't close while uploading, show confirmation dialog
      const confirm = window.confirm(
        'Uploads are in progress. Are you sure you want to close?'
      );
      if (!confirm) return;
    }

    onClose?.();
    router.back();
  }, [isUploading, onClose, router]);

  const handleUploadStart = useCallback(() => {
    setIsUploading(true);
  }, []);

  const handleUploadComplete = useCallback((contentSources: ContentSource[]) => {
    setIsUploading(false);
    setError(null);
    onUploadComplete?.(contentSources);
    
    toast({
      title: 'Upload Complete',
      description: `Successfully uploaded ${contentSources.length} file${contentSources.length !== 1 ? 's' : ''}`,
    });
    
    // Close modal after successful upload
    setTimeout(() => {
      handleClose();
    }, 1000);
  }, [onUploadComplete, handleClose, toast]);

  const handleUploadError = useCallback((errorMessage: string) => {
    setIsUploading(false);
    setError(errorMessage);
    
    toast({
      title: 'Upload Failed',
      description: errorMessage,
      variant: 'destructive',
    });
  }, [toast]);

  const handleOpenChange = useCallback((open: boolean) => {
    if (!open) {
      if (activeTab !== 'file') {
        setActiveTab('file');
        return; // treat close as "back to Files" when on other tabs
      }
      handleClose();
    }
  }, [activeTab, handleClose]);

  return (
    <Dialog open={isOpen} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-[600px] max-h-[80vh] overflow-hidden flex flex-col pt-8 pr-8">
        {/* Accessible title for screen readers while keeping UI visually clean */}
        <DialogHeader>
          <DialogTitle className="sr-only">Upload to {spaceName ?? 'Space'}</DialogTitle>
        </DialogHeader>

        {/* Remove custom header and close button to avoid duplicate close icon and title */}

        <div className="flex-1 overflow-hidden">
          {/* Error Display */}
          {error && (
            <div className="mb-4 p-3 bg-destructive/10 border border-destructive/20 rounded-lg">
              <div className="flex items-start gap-2">
                <AlertCircle className="h-4 w-4 text-destructive mt-0.5 flex-shrink-0" />
                <div>
                  <h4 className="text-sm font-medium text-destructive">Upload Error</h4>
                  <p className="text-xs text-destructive/90 mt-1">{error}</p>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setError(null)}
                    className="mt-2 h-7 text-xs"
                  >
                    Dismiss
                  </Button>
                </div>
              </div>
            </div>
          )}

          <Tabs value={activeTab} onValueChange={(value) => setActiveTab(value as UploadTab)} className="h-full flex flex-col">

            <div className="flex-1 overflow-y-auto">
              <TabsContent value="file" className="mt-0 h-full">
                <FileUploadArea
                  spaceId={spaceId}
                  onUploadStart={handleUploadStart}
                  onUploadComplete={handleUploadComplete}
                  onUploadError={handleUploadError}
                  disabled={isUploading && activeTab !== 'file'}
                />
              </TabsContent>

              <TabsContent value="google-drive" className="mt-0 h-full">
                <GoogleDriveUpload
                  spaceId={spaceId}
                  onUploadStart={handleUploadStart}
                  onUploadComplete={handleUploadComplete}
                  onUploadError={handleUploadError}
                  disabled={isUploading && activeTab !== 'google-drive'}
                />
              </TabsContent>

              <TabsContent value="link" className="mt-0 h-full">
                <LinkUploadForm
                  spaceId={spaceId}
                  onUploadStart={handleUploadStart}
                  onUploadComplete={handleUploadComplete}
                  onUploadError={handleUploadError}
                  disabled={isUploading && activeTab !== 'link'}
                />
              </TabsContent>

              <TabsContent value="text" className="mt-0 h-full">
                <TextUploadForm
                  spaceId={spaceId}
                  onUploadStart={handleUploadStart}
                  onUploadComplete={handleUploadComplete}
                  onUploadError={handleUploadError}
                  disabled={isUploading && activeTab !== 'text'}
                />
              </TabsContent>
            </div>

            {/* Move file type buttons below the content area */}
            <TabsList className="grid w-full grid-cols-4 mt-4 bg-transparent p-0 gap-3 text-foreground">
              {UPLOAD_TABS.map((tab) => (
                <TabsTrigger
                  key={tab.id}
                  value={tab.id}
                  disabled={isUploading && activeTab !== tab.id}
                  className="flex items-center gap-2 border rounded-md bg-background shadow-sm data-[state=active]:shadow-md data-[state=active]:border-primary data-[state=inactive]:opacity-90"
                >
                  {tab.icon}
                  <span className="hidden sm:inline">{tab.label}</span>
                </TabsTrigger>
              ))}
            </TabsList>
          </Tabs>
        </div>
      </DialogContent>
    </Dialog>
  );
}