'use client';

import { Progress } from '@/components/ui/progress';
import { Loader2 } from 'lucide-react';
import { cn } from '@/lib/utils';

interface UploadProgressBarProps {
  progress: number;
  status: 'uploading' | 'processing' | 'completed' | 'failed';
  uploadSpeed?: number;
  estimatedTimeRemaining?: number;
  className?: string;
}

function formatSpeed(bytesPerSecond: number): string {
  if (bytesPerSecond < 1024) {
    return `${bytesPerSecond.toFixed(0)} B/s`;
  } else if (bytesPerSecond < 1024 * 1024) {
    return `${(bytesPerSecond / 1024).toFixed(1)} KB/s`;
  } else {
    return `${(bytesPerSecond / (1024 * 1024)).toFixed(1)} MB/s`;
  }
}

function formatTime(seconds: number): string {
  if (seconds < 60) {
    return `${Math.round(seconds)}s`;
  } else if (seconds < 3600) {
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = Math.round(seconds % 60);
    return remainingSeconds > 0 ? `${minutes}m ${remainingSeconds}s` : `${minutes}m`;
  } else {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return minutes > 0 ? `${hours}h ${minutes}m` : `${hours}h`;
  }
}

function getStatusColor(status: UploadProgressBarProps['status']): string {
  switch (status) {
    case 'uploading':
      return 'bg-blue-500';
    case 'processing':
      return 'bg-amber-500';
    case 'completed':
      return 'bg-green-500';
    case 'failed':
      return 'bg-red-500';
    default:
      return 'bg-primary';
  }
}

function getStatusText(status: UploadProgressBarProps['status']): string {
  switch (status) {
    case 'uploading':
      return 'Uploading';
    case 'processing':
      return 'Processing';
    case 'completed':
      return 'Complete';
    case 'failed':
      return 'Failed';
    default:
      return 'Unknown';
  }
}

export default function UploadProgressBar({
  progress,
  status,
  uploadSpeed,
  estimatedTimeRemaining,
  className
}: UploadProgressBarProps) {
  const progressValue = Math.max(0, Math.min(100, progress));
  const isProcessing = status === 'processing';
  const isCompleted = status === 'completed';
  const isFailed = status === 'failed';

  return (
    <div className={cn('space-y-2', className)}>
      {/* Progress Bar */}
      <div className="relative">
        <Progress 
          value={isProcessing ? undefined : progressValue} 
          className="h-2" 
        />
        {/* Custom colored fill for different statuses */}
        <div
          className={cn(
            'absolute top-0 left-0 h-full rounded-full transition-all duration-300',
            getStatusColor(status)
          )}
          style={{
            width: isProcessing ? '100%' : `${progressValue}%`,
            animation: isProcessing ? 'pulse 2s infinite' : undefined
          }}
        />
      </div>

      {/* Status and Details */}
      <div className="flex items-center justify-between text-xs text-muted-foreground">
        <div className="flex items-center gap-1">
          {isProcessing && (
            <Loader2 className="h-3 w-3 animate-spin" />
          )}
          <span className={cn(
            'font-medium',
            isCompleted && 'text-green-600',
            isFailed && 'text-destructive'
          )}>
            {getStatusText(status)}
          </span>
          {!isProcessing && !isCompleted && !isFailed && (
            <span>({progressValue.toFixed(0)}%)</span>
          )}
        </div>

        {/* Upload details (only show during upload) */}
        {status === 'uploading' && (uploadSpeed || estimatedTimeRemaining) && (
          <div className="flex items-center gap-3">
            {uploadSpeed && uploadSpeed > 0 && (
              <span>{formatSpeed(uploadSpeed)}</span>
            )}
            {estimatedTimeRemaining && estimatedTimeRemaining > 0 && (
              <span>{formatTime(estimatedTimeRemaining)} remaining</span>
            )}
          </div>
        )}
      </div>
    </div>
  );
}