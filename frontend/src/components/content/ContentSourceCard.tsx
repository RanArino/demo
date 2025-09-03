'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { 
  FileText, 
  Link as LinkIcon, 
  Type, 
  Download, 
  Trash2, 
  Clock,
  CheckCircle,
  XCircle,
  Loader2
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { ContentSource, ContentSourceStatus } from '@/app/spaces/types/content';
import { formatDate, fileTypeLabel } from '@/lib/contentSource';

export interface ContentSourceCardProps {
  contentSource: ContentSource;
  onView: (contentSource: ContentSource) => void;
  onDownload: (contentSource: ContentSource, kind: 'original' | 'processed') => void;
  onDelete: (contentSource: ContentSource) => void;
  selected: boolean;
  onSelectChange: (id: string, checked: boolean) => void;
  progress?: number;
}

export default function ContentSourceCard({ contentSource, onView, onDownload, onDelete, selected, onSelectChange, progress }: ContentSourceCardProps) {
  const [showDownloadMenu, setShowDownloadMenu] = useState(false);

  const getSourceIcon = () => {
    switch (contentSource.sourceType) {
      case 'file': return <FileText className="h-4 w-4" />;
      case 'url': return <LinkIcon className="h-4 w-4" />;
      case 'text': return <Type className="h-4 w-4" />;
      case 'google_drive': return <FileText className="h-4 w-4" />;
      default: return <FileText className="h-4 w-4" />;
    }
  };

  const getStatusIcon = () => {
    switch (contentSource.processingStatus) {
      case 'pending': return <Clock className="h-3 w-3" />;
      case 'processing': return <Loader2 className="h-3 w-3 animate-spin" />;
      case 'completed': return <CheckCircle className="h-3 w-3" />;
      case 'failed': return <XCircle className="h-3 w-3" />;
      default: return <Clock className="h-3 w-3" />;
    }
  };

  const isReady = contentSource.processingStatus === ContentSourceStatus.COMPLETED;

  return (
    <div
      className={cn(
      "relative bg-white border border-gray-200 rounded-lg p-4 hover:shadow-sm transition-shadow",
      !isReady && "opacity-70"
    )}
      onClick={() => onView(contentSource)}
      role="button"
      tabIndex={0}
    >
      {!isReady && (
        <div className="absolute inset-0 rounded-lg bg-transparent pointer-events-none" />
      )}
      {/* Row header */}
      <div className="flex items-center gap-3 mb-2">
        {isReady ? (
          <input
            type="checkbox"
            checked={selected}
            onChange={(e) => onSelectChange(contentSource.id, e.currentTarget.checked)}
            onClick={(e) => e.stopPropagation()}
            className="h-4 w-4 border-gray-400 rounded"
          />
        ) : (
          <div className="flex items-center gap-2 text-blue-600">
            <Loader2 className="h-4 w-4 animate-spin" />
            <span className="text-xs">
              {contentSource.processingStatus === 'pending' && 'Initializing...'}
              {contentSource.processingStatus === 'processing' && 'Processing...'}
              {contentSource.processingStatus === 'failed' && 'Failed'}
            </span>
          </div>
        )}
        <h3 className="font-medium text-gray-900 truncate flex-1">
          {contentSource.title || contentSource.filename || 'Untitled'}
        </h3>
      </div>

      {/* Secondary row: left (file type + added date), right (status + actions) */}
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-3">
          <span className="text-xs px-3 py-1 border rounded-md">{fileTypeLabel(contentSource.mimeType, contentSource.sourceType)}</span>
          <div className="text-xs text-gray-500">Added {formatDate(contentSource.createdAt)}</div>
        </div>
        <div className="flex items-center gap-3" onClick={(e) => e.stopPropagation()}>
          {contentSource.processingStatus === 'completed' && (
            <div className="relative">
              <Button
                variant="ghost"
                size="icon"
                onClick={() => setShowDownloadMenu((v) => !v)}
                className="h-8 w-8 p-0 flex items-center justify-center"
                aria-haspopup="menu"
                aria-expanded={showDownloadMenu}
              >
                <Download className="h-4 w-4" />
              </Button>
              {showDownloadMenu && (
                <div className="absolute right-0 mt-1 w-40 bg-white border rounded-md shadow-md z-10">
                  <button
                    className="w-full text-left px-3 py-2 text-sm hover:bg-gray-50"
                    onClick={() => { setShowDownloadMenu(false); onDownload(contentSource, 'original'); }}
                  >
                    Original
                  </button>
                  <button
                    className="w-full text-left px-3 py-2 text-sm hover:bg-gray-50"
                    onClick={() => { setShowDownloadMenu(false); onDownload(contentSource, 'processed'); }}
                  >
                    Processed (Markdown)
                  </button>
                </div>
              )}
            </div>
          )}
          <Button
            variant="ghost"
            size="icon"
            onClick={() => onDelete(contentSource)}
            className="h-8 w-8 p-0 flex items-center justify-center text-red-600 hover:text-red-700 hover:bg-red-50"
            aria-label="Delete"
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* Progress bar for in-flight uploads */}
      {contentSource.processingStatus !== 'completed' && typeof progress === 'number' && (
        <div className="mt-2">
          <div className="h-1.5 bg-gray-200 rounded">
            <div className="h-1.5 bg-blue-500 rounded" style={{ width: `${progress}%` }} />
          </div>
          <div className="text-[10px] text-gray-500 mt-1">{progress}%</div>
        </div>
      )}

      {/* Error message if failed */}
      {contentSource.processingStatus === 'failed' && contentSource.processingError && (
        <div className="text-xs text-red-600 bg-red-50 p-2 rounded mb-3">
          {contentSource.processingError}
        </div>
      )}

      {/* No overlay button; container handles click so checkbox/icons work correctly */}
    </div>
  );
}
