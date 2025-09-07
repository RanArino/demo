'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Type, Loader2 } from 'lucide-react';
import { ContentSource } from '@/api/generated/v1/knowledge_pb';

interface TextUploadFormProps {
  spaceId: string;
  onUploadStart?: () => void;
  onUploadComplete?: (contentSources: ContentSource[]) => void;
  onUploadError?: (error: string) => void;
  disabled?: boolean;
}

export default function TextUploadForm({
  onUploadStart,
  onUploadComplete,
  onUploadError,
  disabled = false
}: TextUploadFormProps) {
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [format, setFormat] = useState<'plain' | 'markdown' | 'html'>('plain');
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!title.trim() || !content.trim()) {
      onUploadError?.('Title and content are required');
      return;
    }

    try {
      setIsLoading(true);
      onUploadStart?.();

      // TODO: Implement text upload functionality
      // This would typically:
      // 1. Validate input
      // 2. Send to backend for content processing
      // 3. Track processing status
      
      // Placeholder success response
      setTimeout(() => {
        setIsLoading(false);
        onUploadComplete?.([]);
        setTitle('');
        setContent('');
        setFormat('plain');
      }, 1500);

    } catch (error) {
      setIsLoading(false);
      onUploadError?.(error instanceof Error ? error.message : 'Failed to save text content');
    }
  };

  const characterCount = content.length;
  const wordCount = content.trim() ? content.trim().split(/\s+/).length : 0;

  return (
    <div className="space-y-6">
      <div className="text-center">
        <div className="mx-auto h-12 w-12 flex items-center justify-center rounded-full bg-muted mb-4">
          <Type className="h-6 w-6 text-muted-foreground" />
        </div>
        <h3 className="text-lg font-medium">Add Text Content</h3>
        <p className="text-sm text-muted-foreground mt-1">
          Paste or type content directly into your space
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="title">Title *</Label>
          <Input
            id="title"
            type="text"
            placeholder="Enter a title for this content"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            disabled={disabled || isLoading}
            required
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="format">Format</Label>
          <select
            id="format"
            value={format}
            onChange={(e) => setFormat(e.target.value as 'plain' | 'markdown' | 'html')}
            disabled={disabled || isLoading}
            className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
          >
            <option value="plain">Plain Text</option>
            <option value="markdown">Markdown</option>
            <option value="html">HTML</option>
          </select>
        </div>

        <div className="space-y-2">
          <Label htmlFor="content">Content *</Label>
          <textarea
            id="content"
            placeholder="Paste or type your content here..."
            value={content}
            onChange={(e) => setContent(e.target.value)}
            disabled={disabled || isLoading}
            required
            className="flex min-h-[200px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 resize-y"
          />
          
          {/* Character and word count */}
          <div className="flex justify-between text-xs text-muted-foreground">
            <span>{characterCount} characters</span>
            <span>{wordCount} words</span>
          </div>
        </div>

        <Button
          type="submit"
          className="w-full"
          disabled={disabled || isLoading || !title.trim() || !content.trim()}
        >
          {isLoading ? (
            <>
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              Saving Content...
            </>
          ) : (
            <>
              <Type className="h-4 w-4 mr-2" />
              Save Content
            </>
          )}
        </Button>
      </form>

      <div className="text-xs text-muted-foreground bg-muted/50 p-3 rounded-lg">
        <p className="font-medium mb-1">Supported formats:</p>
        <ul className="space-y-1">
          <li>• <strong>Plain Text:</strong> Regular text content</li>
          <li>• <strong>Markdown:</strong> Formatted text with markdown syntax</li>
          <li>• <strong>HTML:</strong> Rich content with HTML markup</li>
        </ul>
      </div>
    </div>
  );
}