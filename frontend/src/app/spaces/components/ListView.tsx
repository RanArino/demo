import React from 'react';
import { Space } from '@/api/generated/v1/knowledge_pb';
import { SpacesTable } from './SpacesTable';
import { EmptyState } from '@/components/common/EmptyState';
import { LoadingState } from '@/components/common/LoadingState';

interface ListViewProps {
  spaces: Space[];
  onEdit: (spaceId: string) => void;
  onDelete: (spaceId: string) => void;
  isDeleting: string | null;
  loading?: boolean;
}

export default function ListView({
  spaces,
  onEdit,
  onDelete,
  isDeleting,
  loading,
}: ListViewProps) {
  if (loading) {
    return <LoadingState />;
  }

  return (
    <div className="h-full flex flex-col">
      {/* Header - Fixed */}
      <div className="flex-shrink-0 flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <h2 className="text-lg font-medium text-gray-700">List View</h2>
          <span className="text-sm text-gray-500">({spaces.length} spaces)</span>
        </div>
      </div>

      {/* Content Area - Scrollable */}
      <div className="flex-1 min-h-0">
        {spaces.length === 0 ? (
          <EmptyState />
        ) : (
          <SpacesTable
            spaces={spaces}
            onEdit={onEdit}
            onDelete={onDelete}
            isDeleting={isDeleting}
          />
        )}
      </div>
    </div>
  );
}