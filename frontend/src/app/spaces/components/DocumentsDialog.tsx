import React, { useState, useCallback } from 'react';
import { Space, ContentSource } from '@/api/generated/v1/knowledge_pb';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { FileText, Loader2, ExternalLink } from 'lucide-react';
import { listContentSources, generateDownloadURL } from '@/api/actions/contentActions';

interface DocumentsDialogProps {
  space: Space;
  children?: React.ReactNode;
}

const TABLE_HEADER_BG = 'bg-gray-50';

export function DocumentsDialog({ space, children }: DocumentsDialogProps) {
  const [documents, setDocuments] = useState<ContentSource[]>([]);
  const [isLoadingDocs, setIsLoadingDocs] = useState(false);
  const [errorDocs, setErrorDocs] = useState<string | null>(null);
  const [loadingPreview, setLoadingPreview] = useState<string | null>(null);

  const formatFileSize = useCallback((bytes: bigint): string => {
    if (bytes === BigInt(0)) return '0 Bytes';
    const k = BigInt(1024);
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const bytesNumber = Number(bytes);
    const kNumber = Number(k);
    const i = Math.floor(Math.log(bytesNumber) / Math.log(kNumber));
    return parseFloat((bytesNumber / Math.pow(kNumber, i)).toFixed(2)) + ' ' + sizes[i];
  }, []);

  const handleOpenDocumentDialog = useCallback(async (spaceId: string) => {
    if (!spaceId) return;
    setIsLoadingDocs(true);
    setErrorDocs(null);
    setDocuments([]);
    try {
      const result = await listContentSources(spaceId);
      if (result.ok && result.data) {
        setDocuments(result.data);
      } else {
        setErrorDocs(result.error?.message || 'Failed to fetch documents.');
      }
    } catch (error: unknown) {
      setErrorDocs(error instanceof Error ? error.message : 'Failed to fetch documents.');
    } finally {
      setIsLoadingDocs(false);
    }
  }, []);

  const handleOpenPreview = useCallback(async (doc: ContentSource) => {
    if (!doc.id) return;
    
    setLoadingPreview(doc.id);
    try {
      const result = await generateDownloadURL(doc.id, 'original');
      if (result.ok && result.data?.url) {
        window.open(result.data.url, '_blank');
      } else {
        console.error('Failed to generate download URL:', result.error);
        // You could show a toast notification here
      }
    } catch (error) {
      console.error('Error generating download URL:', error);
      // You could show a toast notification here
    } finally {
      setLoadingPreview(null);
    }
  }, []);

  const triggerContent = children || (
    <Button 
      variant="ghost" 
      className="h-8 px-2 text-blue-600 hover:text-blue-800 hover:bg-blue-50" 
      aria-label={`View documents for ${space.title}`}
    >
      {space.stats?.contentCount.toString() || 0}
    </Button>
  );

  return (
    <Dialog onOpenChange={(open) => open && handleOpenDocumentDialog(space.id)}>
      <DialogTrigger asChild>
        {triggerContent}
      </DialogTrigger>
      <DialogContent 
        className="max-w-[80vw] max-h-[80vh] w-[80vw] h-[80vh] flex flex-col" 
        aria-describedby={undefined}
      >
        <DialogHeader>
          <DialogTitle>Documents in {space.title}</DialogTitle>
        </DialogHeader>
        <div className="flex-1 overflow-hidden">
          {isLoadingDocs ? (
            <div className="flex justify-center items-center h-full">
              <Loader2 className="h-8 w-8 animate-spin text-gray-500" />
            </div>
          ) : errorDocs ? (
            <div className="text-center py-8 text-red-600">
              <p>Error: {errorDocs}</p>
            </div>
          ) : documents.length > 0 ? (
            <div className="h-full overflow-auto">
              <Table>
                <TableHeader className={`sticky top-0 z-10 ${TABLE_HEADER_BG}`}>
                  <TableRow className={TABLE_HEADER_BG}>
                    <TableHead>Title</TableHead>
                    <TableHead>Keywords</TableHead>
                    <TableHead>Content Summary</TableHead>
                    <TableHead>File Type</TableHead>
                    <TableHead>Size</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {documents.map((doc) => (
                    <TableRow key={doc.id} className="hover:bg-gray-50" onClick={(e) => e.stopPropagation()}>
                      <TableCell className="font-medium">{doc.title}</TableCell>
                      <TableCell>{doc.keywords?.join(', ') || 'N/A'}</TableCell>
                      <TableCell className="max-w-xs truncate">{doc.contentSummary || 'N/A'}</TableCell>
                      <TableCell>{doc.mimeType || 'N/A'}</TableCell>
                      <TableCell>{formatFileSize(doc.sizeBytes)}</TableCell>
                      <TableCell>
                        <Button 
                          variant="outline" 
                          size="sm"
                          onClick={() => handleOpenPreview(doc)}
                          disabled={!doc.id || loadingPreview === doc.id}
                          aria-label={`Open preview for ${doc.title}`}
                        >
                          {loadingPreview === doc.id ? (
                            <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                          ) : (
                            <ExternalLink className="h-4 w-4 mr-2" />
                          )}
                          Open Preview
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          ) : (
            <div className="text-center py-8">
              <FileText className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <p className="text-gray-500">
                No documents found in this space.
              </p>
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
