import React from 'react';
import { ArrowUpDown, ArrowUp, ArrowDown } from 'lucide-react';

interface SortIconProps {
  sortState: 'sortable' | 'asc' | 'desc';
}

export function SortIcon({ sortState }: SortIconProps) {
  switch (sortState) {
    case 'asc':
      return <ArrowUp className="h-4 w-4 text-blue-600" />;
    case 'desc':
      return <ArrowDown className="h-4 w-4 text-blue-600" />;
    default:
      return <ArrowUpDown className="h-4 w-4 text-gray-400" />;
  }
}
