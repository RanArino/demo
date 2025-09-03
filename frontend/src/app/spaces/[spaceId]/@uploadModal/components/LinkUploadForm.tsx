'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Link, Loader2 } from 'lucide-react';
import { ContentSource } from '@/app/spaces/types/content';

interface LinkUploadFormProps {
  spaceId: string;
  onUploadStart?: () => void;
  onUploadComplete?: (contentSources: ContentSource[]) => void;
  onUploadError?: (error: string) => void;
  disabled?: boolean;
}

export default function LinkUploadForm({
  onUploadStart,
  onUploadComplete,
  onUploadError,
  disabled = false
}: LinkUploadFormProps) {
  const [url, setUrl] = useState('');
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!url.trim()) {
      onUploadError?.('URL is required');
      return;
    }

    try {
      setIsLoading(true);
      onUploadStart?.();

      // TODO: Implement URL upload functionality
      // This would typically:
      // 1. Validate URL format
      // 2. Send to backend for content extraction
      // 3. Track processing status
      
      // Placeholder success response
      setTimeout(() => {
        setIsLoading(false);
        onUploadComplete?.([]);
        setUrl('');
        setTitle('');
        setDescription('');
      }, 2000);

    } catch (error) {
      setIsLoading(false);
      onUploadError?.(error instanceof Error ? error.message : 'Failed to upload URL');
    }
  };

  return (
    <div className="space-y-6">
      <div className="text-center">
        <div className="mx-auto h-12 w-12 flex items-center justify-center rounded-full bg-muted mb-4">
          <Link className="h-6 w-6 text-muted-foreground" />
        </div>
        <h3 className="text-lg font-medium">Add from URL</h3>
        <p className="text-sm text-muted-foreground mt-1">
          Import content from a web page or document URL
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="url">URL *</Label>
          <Input
            id="url"
            type="url"
            placeholder="https://example.com/document"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            disabled={disabled || isLoading}
            required
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="title">Title (optional)</Label>
          <Input
            id="title"
            type="text"
            placeholder="Custom title for this content"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            disabled={disabled || isLoading}
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="description">Description (optional)</Label>
          <Input
            id="description"
            type="text"
            placeholder="Brief description of the content"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            disabled={disabled || isLoading}
          />
        </div>

        <Button
          type="submit"
          className="w-full"
          disabled={disabled || isLoading || !url.trim()}
        >
          {isLoading ? (
            <>
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              Processing URL...
            </>
          ) : (
            <>
              <Link className="h-4 w-4 mr-2" />
              Import from URL
            </>
          )}
        </Button>
      </form>

      <div className="text-xs text-muted-foreground bg-muted/50 p-3 rounded-lg">
        <p className="font-medium mb-1">Supported URL types:</p>
        <ul className="space-y-1">
          <li>• Web pages and articles</li>
          <li>• Google Docs (public or with access)</li>
          <li>• PDF files hosted online</li>
          <li>• Other document formats</li>
        </ul>
      </div>
    </div>
  );
}