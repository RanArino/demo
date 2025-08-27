'use client';

import { useRouter } from 'next/navigation';
import { ContentSource } from '@/app/spaces/types/content';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { 
  Plus, 
  FileText, 
  Link, 
  Type, 
  Download, 
  Trash2, 
  Eye, 
  MoreVertical,
  Clock,
  CheckCircle,
  XCircle,
  Loader2
} from 'lucide-react';
import { cn, formatSize } from '@/lib/utils';

interface DocumentsSectionProps {
  spaceId: string;
  contentSources: ContentSource[];
  className?: string;
}

interface ContentSourceCardProps {
  contentSource: ContentSource;
  onView: (source: ContentSource) => void;
  onDownload: (source: ContentSource) => void;
  onDelete: (source: ContentSource) => void;
}

function ContentSourceCard({ contentSource, onView, onDownload, onDelete }: ContentSourceCardProps) {
  const getSourceIcon = () => {
    switch (contentSource.sourceType) {
      case 'file': return <FileText className="h-4 w-4" />;
      case 'url': return <Link className="h-4 w-4" />;
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

  const getStatusColor = () => {
    switch (contentSource.processingStatus) {
      case 'pending': return 'bg-yellow-100 text-yellow-800 border-yellow-200';
      case 'processing': return 'bg-blue-100 text-blue-800 border-blue-200';
      case 'completed': return 'bg-green-100 text-green-800 border-green-200';
      case 'failed': return 'bg-red-100 text-red-800 border-red-200';
      default: return 'bg-gray-100 text-gray-800 border-gray-200';
    }
  };

  

  const formatDate = (date: Date | string) => {
    const d = typeof date === 'string' ? new Date(date) : date;
    return d.toLocaleDateString();
  };

  return (
    <div className="bg-white border border-gray-200 rounded-lg p-4 hover:shadow-sm transition-shadow">
      {/* Header */}
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center gap-2 min-w-0 flex-1">
          {getSourceIcon()}
          <h3 className="font-medium text-gray-900 truncate">
            {contentSource.title || contentSource.filename || 'Untitled'}
          </h3>
        </div>
        
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onView(contentSource)}
            className="h-8 w-8 p-0"
          >
            <Eye className="h-4 w-4" />
          </Button>
          
          {contentSource.processingStatus === 'completed' && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onDownload(contentSource)}
              className="h-8 w-8 p-0"
            >
              <Download className="h-4 w-4" />
            </Button>
          )}
          
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onDelete(contentSource)}
            className="h-8 w-8 p-0 text-red-600 hover:text-red-700 hover:bg-red-50"
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* Status and Metadata */}
      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <Badge 
            variant="outline" 
            className={`text-xs flex items-center gap-1 ${getStatusColor()}`}
          >
            {getStatusIcon()}
            {contentSource.processingStatus}
          </Badge>
          
          <span className="text-xs text-gray-500">
            {formatSize(contentSource.sizeBytes)}
          </span>
        </div>

        <div className="text-xs text-gray-500">
          Added {formatDate(contentSource.createdAt)}
        </div>

        {/* Processing Error */}
        {contentSource.processingStatus === 'failed' && contentSource.processingError && (
          <div className="text-xs text-red-600 bg-red-50 p-2 rounded">
            {contentSource.processingError}
          </div>
        )}

        {/* Content Preview */}
        {contentSource.extractedText && (
          <div className="text-xs text-gray-600 line-clamp-2 bg-gray-50 p-2 rounded">
            {contentSource.extractedText.substring(0, 100)}...
          </div>
        )}
      </div>
    </div>
  );
}

export default function DocumentsSection({ spaceId, contentSources, className }: DocumentsSectionProps) {
  const router = useRouter();

  const handleAddContent = () => {
    // Navigate to upload modal using the parallel route
    router.push(`/spaces/${spaceId}/upload`);
  };

  const handleViewContent = (source: ContentSource) => {
    // TODO: Implement content viewer
    console.log('View content:', source.id);
  };

  const handleDownloadContent = (source: ContentSource) => {
    // TODO: Implement download functionality
    console.log('Download content:', source.id);
  };

  const handleDeleteContent = (source: ContentSource) => {
    // TODO: Implement delete functionality with confirmation
    console.log('Delete content:', source.id);
  };

  const getStatusCounts = () => {
    const counts = {
      total: contentSources.length,
      completed: contentSources.filter(s => s.processingStatus === 'completed').length,
      processing: contentSources.filter(s => s.processingStatus === 'processing').length,
      failed: contentSources.filter(s => s.processingStatus === 'failed').length,
    };
    return counts;
  };

  const statusCounts = getStatusCounts();

  return (
    <div className={cn("flex flex-col h-full", className)}>
      {/* Header */}
      <div className="p-4 border-b border-gray-200">
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-lg font-semibold text-gray-900">Documents</h2>
          <Button
            onClick={handleAddContent}
            size="sm"
            className="flex items-center gap-2"
          >
            <Plus className="h-4 w-4" />
            Add
          </Button>
        </div>

        {/* Status Summary */}
        <div className="grid grid-cols-2 gap-2 text-xs">
          <div className="flex items-center justify-between">
            <span className="text-gray-600">Total:</span>
            <span className="font-medium">{statusCounts.total}</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-gray-600">Ready:</span>
            <span className="font-medium text-green-600">{statusCounts.completed}</span>
          </div>
          {statusCounts.processing > 0 && (
            <div className="flex items-center justify-between">
              <span className="text-gray-600">Processing:</span>
              <span className="font-medium text-blue-600">{statusCounts.processing}</span>
            </div>
          )}
          {statusCounts.failed > 0 && (
            <div className="flex items-center justify-between">
              <span className="text-gray-600">Failed:</span>
              <span className="font-medium text-red-600">{statusCounts.failed}</span>
            </div>
          )}
        </div>
      </div>

      {/* Content List */}
      <div className="flex-1 overflow-y-auto p-4">
        {contentSources.length > 0 ? (
          <div className="space-y-3">
            {contentSources.map((source) => (
              <ContentSourceCard
                key={source.id}
                contentSource={source}
                onView={handleViewContent}
                onDownload={handleDownloadContent}
                onDelete={handleDeleteContent}
              />
            ))}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center h-full text-center">
            <div className="w-12 h-12 bg-gray-100 rounded-full flex items-center justify-center mb-3">
              <FileText className="h-6 w-6 text-gray-400" />
            </div>
            <h3 className="text-sm font-medium text-gray-900 mb-1">
              No documents yet
            </h3>
            <p className="text-xs text-gray-600 mb-3">
              Add your first document to get started
            </p>
            <Button
              onClick={handleAddContent}
              size="sm"
              variant="outline"
              className="flex items-center gap-2"
            >
              <Plus className="h-4 w-4" />
              Add Document
            </Button>
          </div>
        )}
      </div>

    </div>
  );
}