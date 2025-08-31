'use client';

import { Button } from '@/components/ui/button';
import { Plus } from 'lucide-react';

interface ContentSourcesHeaderProps {
  title?: string;
  total: number;
  processing: number;
  failed: number;
  onAdd: () => void;
}

export default function ContentSourcesHeader({ title = 'Content Sources', total, processing, failed, onAdd }: ContentSourcesHeaderProps) {
  return (
    <div className="p-4 border-b border-gray-200">
      <div className="flex items-center justify-between mb-3">
        <h2 className="text-lg font-semibold text-gray-900">{title}</h2>
        <Button onClick={onAdd} size="sm" className="flex items-center gap-2">
          <Plus className="h-4 w-4" />
          Add
        </Button>
      </div>

      <div className="grid grid-cols-1 gap-2 text-xs">
        <div className="flex items-center justify-between">
          <span className="text-gray-600">Total:</span>
          <span className="font-medium">{total}</span>
        </div>
        {processing > 0 && (
          <div className="flex items-center justify-between">
            <span className="text-gray-600">Processing:</span>
            <span className="font-medium text-blue-600">{processing}</span>
          </div>
        )}
        {failed > 0 && (
          <div className="flex items-center justify-between">
            <span className="text-gray-600">Failed:</span>
            <span className="font-medium text-red-600">{failed}</span>
          </div>
        )}
      </div>
    </div>
  );
}
