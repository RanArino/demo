import React from 'react';

export interface LoadingStateProps {
  /** Layout type - 'table' for table-like loading, 'grid' for grid-like loading */
  layout?: 'table' | 'grid';
  /** Number of skeleton items to show */
  itemCount?: number;
  /** Custom height for skeleton items */
  itemHeight?: string;
  /** Custom width for header skeleton */
  headerWidth?: string;
  /** Custom height for header skeleton */
  headerHeight?: string;
  /** Grid columns for grid layout */
  gridColumns?: string;
  /** Gap between grid items */
  gridGap?: string;
  /** Custom className for the container */
  className?: string;
}

export function LoadingState({
  layout = 'table',
  itemCount = 5,
  itemHeight = 'h-16',
  headerWidth = 'w-32',
  headerHeight = 'h-6',
  gridColumns = 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4',
  gridGap = 'gap-4',
  className = '',
}: LoadingStateProps) {
  if (layout === 'grid') {
    return (
      <div className={`space-y-6 ${className}`}>
        {/* Header skeleton */}
        <div className="flex items-center justify-between">
          <div className={`${headerHeight} ${headerWidth} bg-muted animate-pulse rounded`} />
          <div className="h-8 w-24 bg-muted animate-pulse rounded-lg" />
        </div>
        
        {/* Grid skeleton */}
        <div className={`grid ${gridColumns} ${gridGap}`}>
          {Array.from({ length: itemCount }).map((_, i) => (
            <div key={i} className={`h-[380px] bg-muted animate-pulse rounded-xl`} />
          ))}
        </div>
      </div>
    );
  }

  // Default table layout
  return (
    <div className={`h-full flex flex-col ${className}`}>
      <div className="flex-shrink-0 flex items-center justify-between mb-4">
        <div className={`${headerHeight} ${headerWidth} bg-muted animate-pulse rounded`} />
      </div>
      <div className="flex-1 min-h-0">
        <div className="bg-white rounded-lg border border-gray-200 overflow-hidden h-full flex flex-col">
          <div className="flex-shrink-0 h-12 bg-gray-50 animate-pulse" />
          <div className="flex-1 space-y-2 p-4">
            {Array.from({ length: itemCount }).map((_, i) => (
              <div key={i} className={`${itemHeight} bg-muted animate-pulse rounded`} />
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
