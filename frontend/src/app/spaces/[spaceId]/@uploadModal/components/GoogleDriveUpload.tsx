'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { FileText, Loader2, ExternalLink } from 'lucide-react';
import { ContentSource } from '@/api/generated/v1/knowledge_pb';

interface GoogleDriveUploadProps {
  spaceId: string;
  onUploadStart?: () => void;
  onUploadComplete?: (contentSources: ContentSource[]) => void;
  onUploadError?: (error: string) => void;
  disabled?: boolean;
}

export default function GoogleDriveUpload({
  onUploadStart,
  onUploadComplete,
  onUploadError,
  disabled = false
}: GoogleDriveUploadProps) {
  const [isConnecting, setIsConnecting] = useState(false);
  const [isConnected, setIsConnected] = useState(false);

  const handleConnect = async () => {
    try {
      setIsConnecting(true);
      
      // TODO: Implement Google Drive OAuth flow
      // This would typically:
      // 1. Redirect to Google OAuth
      // 2. Handle callback and store tokens
      // 3. Show file picker interface
      
      // Placeholder - simulate connection
      setTimeout(() => {
        setIsConnecting(false);
        setIsConnected(true);
      }, 2000);

    } catch (error) {
      setIsConnecting(false);
      onUploadError?.(error instanceof Error ? error.message : 'Failed to connect to Google Drive');
    }
  };

  const handleSelectFiles = async () => {
    try {
      onUploadStart?.();
      
      // TODO: Implement Google Drive file picker
      // This would typically:
      // 1. Show Google Drive file picker
      // 2. Handle file selection
      // 3. Import selected files
      
      // Placeholder success response
      setTimeout(() => {
        onUploadComplete?.([]);
      }, 2000);

    } catch (error) {
      onUploadError?.(error instanceof Error ? error.message : 'Failed to import files from Google Drive');
    }
  };

  if (!isConnected) {
    return (
      <div className="space-y-6">
        <div className="text-center">
          <div className="mx-auto h-12 w-12 flex items-center justify-center rounded-full bg-muted mb-4">
            <FileText className="h-6 w-6 text-muted-foreground" />
          </div>
          <h3 className="text-lg font-medium">Connect Google Drive</h3>
          <p className="text-sm text-muted-foreground mt-1">
            Import documents directly from your Google Drive
          </p>
        </div>

        <div className="bg-muted/50 p-4 rounded-lg">
          <h4 className="font-medium text-sm mb-2">What you can import:</h4>
          <ul className="text-sm text-muted-foreground space-y-1">
            <li>• Google Docs, Sheets, and Slides</li>
            <li>• PDFs and other documents</li>
            <li>• Images and media files</li>
            <li>• Any files stored in your Drive</li>
          </ul>
        </div>

        <Button
          onClick={handleConnect}
          disabled={disabled || isConnecting}
          className="w-full"
        >
          {isConnecting ? (
            <>
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              Connecting...
            </>
          ) : (
            <>
              <ExternalLink className="h-4 w-4 mr-2" />
              Connect Google Drive
            </>
          )}
        </Button>

        <div className="text-xs text-muted-foreground text-center">
          <p>
            By connecting, you agree to allow access to your Google Drive files.
            <br />
            You can disconnect at any time in your account settings.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="text-center">
        <div className="mx-auto h-12 w-12 flex items-center justify-center rounded-full bg-green-100 mb-4">
          <FileText className="h-6 w-6 text-green-600" />
        </div>
        <h3 className="text-lg font-medium">Google Drive Connected</h3>
        <p className="text-sm text-muted-foreground mt-1">
          Select files to import into your space
        </p>
      </div>

      <div className="space-y-4">
        <Button
          onClick={handleSelectFiles}
          disabled={disabled}
          className="w-full"
        >
          <FileText className="h-4 w-4 mr-2" />
          Select Files from Drive
        </Button>

        <Button
          variant="outline"
          onClick={() => setIsConnected(false)}
          disabled={disabled}
          className="w-full"
        >
          Disconnect Google Drive
        </Button>
      </div>

      <div className="text-xs text-muted-foreground bg-blue-50 p-3 rounded-lg">
        <p className="font-medium mb-1">Tips for importing:</p>
        <ul className="space-y-1">
          <li>• Select multiple files at once to bulk import</li>
          <li>• Large files may take longer to process</li>
          <li>• Google Docs will be converted to text format</li>
        </ul>
      </div>
    </div>
  );
}