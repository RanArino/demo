'use client';

import { useState } from 'react';
import { Space } from '@/api/generated/v1/knowledge_pb';
import SpaceCard from './SpaceCard';
import { Button } from '@/components/ui/button';
import { Grid3X3, Grid2X2, LayoutGrid } from 'lucide-react';
import { LoadingState } from '@/components/common/LoadingState';

interface GalleryViewProps {
  spaces: Space[];
  onSelect: (spaceId: string) => void;
  onEdit: (spaceId: string) => void;
  onDelete: (spaceId: string) => void;
  isDeleting: string | null;
  loading?: boolean;
}

export default function GalleryView({
  spaces,
  onSelect,
  onEdit,
  onDelete,
  isDeleting,
  loading,
}: GalleryViewProps) {
  const [cardSize, setCardSize] = useState<'small' | 'medium' | 'large'>('medium');

  const sizeClasses = {
    small: 'grid-cols-2 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6',
    medium: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4',
    large: 'grid-cols-1 md:grid-cols-1 lg:grid-cols-2 xl:grid-cols-3',
  };

  const gapClasses = {
    small: 'gap-3',
    medium: 'gap-4',
    large: 'gap-6',
  };

  if (loading) {
    return (
      <LoadingState
        layout="grid"
        itemCount={8}
        gridColumns={sizeClasses[cardSize]}
        gridGap={gapClasses[cardSize]}
      />
    );
  }

  return (
    <div className="space-y-6">
      {/* Header with controls */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <h2 className="text-lg font-medium text-gray-700">Gallery View</h2>
          <span className="text-sm text-gray-500">({spaces.length} spaces)</span>
        </div>
        
        <div className="flex items-center gap-1 bg-gray-100 rounded-lg p-1">
          <Button 
            onClick={() => setCardSize('small')} 
            variant={cardSize === 'small' ? 'default' : 'ghost'}
            size="sm"
            className="h-8 w-8 p-0"
            aria-label="Small grid"
          >
            <Grid3X3 className="h-4 w-4" />
          </Button>
          <Button 
            onClick={() => setCardSize('medium')} 
            variant={cardSize === 'medium' ? 'default' : 'ghost'}
            size="sm"
            className="h-8 w-8 p-0"
            aria-label="Medium grid"
          >
            <Grid2X2 className="h-4 w-4" />
          </Button>
          <Button 
            onClick={() => setCardSize('large')}
            variant={cardSize === 'large' ? 'default' : 'ghost'}
            size="sm"
            className="h-8 w-8 p-0"
            aria-label="Large grid"
          >
            <LayoutGrid className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* Empty state */}
      {spaces.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mb-4">
            <LayoutGrid className="h-8 w-8 text-gray-400" />
          </div>
          <h3 className="text-lg font-medium text-gray-900 mb-2">No spaces found</h3>
          <p className="text-gray-500 max-w-sm">
            Try adjusting your search criteria or create a new space to get started.
          </p>
        </div>
      ) : (
        /* Grid layout */
        <div className={`grid ${sizeClasses[cardSize]} ${gapClasses[cardSize]}`}>
          {spaces.map((space) => (
            <SpaceCard
              key={space.id}
              space={space}
              onSelect={onSelect}
              onEdit={onEdit}
              onDelete={onDelete}
              isDeleting={isDeleting === space.id}
            />
          ))}
        </div>
      )}
    </div>
  );
}