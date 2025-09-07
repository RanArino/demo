import React, { useState } from 'react';
import { Space } from '@/api/generated/v1/knowledge_pb';
import { Timestamp } from '@bufbuild/protobuf';
import { safeTimestampToDate } from '@/lib/types';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { 
  Pencil, 
  ArrowUpDown, 
  ArrowUp, 
  ArrowDown, 
  Users, 
  FileText, 
  Calendar,
  Globe,
  Lock,
  Eye
} from 'lucide-react';
import { EditSpaceForm } from './EditSpaceForm';
import Link from 'next/link';

interface ListViewProps {
  spaces: Space[];
  onSelect: (spaceId: string) => void;
  onEdit: (spaceId: string) => void;
  onDelete: (spaceId: string) => void;
  isDeleting: string | null;
  loading?: boolean;
}

type SortKey = 'title' | 'stats.contentCount' | 'createdAt' | 'updatedAt';

export default function ListView({
  spaces: initialSpaces,
  onSelect,
  onEdit,
  onDelete,
  isDeleting,
  loading,
}: ListViewProps) {
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editedField, setEditedField] = useState<string | null>(null);
  const [editedValue, setEditedValue] = useState<string | string[] | null>(null);
  const [spaces, setSpaces] = useState<Space[]>(initialSpaces);
  const [sortConfig, setSortConfig] = useState<{ key: SortKey; direction: 'ascending' | 'descending' } | null>(null);

  const sortedSpaces = React.useMemo(() => {
    const sortableSpaces = [...spaces];
    if (sortConfig !== null) {
      sortableSpaces.sort((a, b) => {
        let aValue: any;
        let bValue: any;

        // Type-safe property access
        switch (sortConfig.key) {
          case 'title':
            aValue = a.title;
            bValue = b.title;
            break;
          case 'stats.contentCount':
            aValue = a.stats?.contentCount;
            bValue = b.stats?.contentCount;
            break;
          case 'createdAt':
            aValue = a.createdAt;
            bValue = b.createdAt;
            break;
          case 'updatedAt':
            aValue = a.updatedAt;
            bValue = b.updatedAt;
            break;
          default:
            return 0;
        }

        if (sortConfig.key === 'createdAt' || sortConfig.key === 'updatedAt') {
          // Handle Timestamp objects
          const aDate = safeTimestampToDate(aValue);
          const bDate = safeTimestampToDate(bValue);
          if (!aDate || !bDate) return 0;
          return sortConfig.direction === 'ascending'
            ? aDate.getTime() - bDate.getTime()
            : bDate.getTime() - aDate.getTime();
        }

        if (typeof aValue === 'string' && typeof bValue === 'string') {
          return sortConfig.direction === 'ascending' ? aValue.localeCompare(bValue) : bValue.localeCompare(aValue);
        }
        
        if (typeof aValue === 'number' && typeof bValue === 'number') {
          return sortConfig.direction === 'ascending' ? aValue - bValue : bValue - aValue;
        }

        return 0;
      });
    }
    return sortableSpaces;
  }, [spaces, sortConfig]);

  const requestSort = (key: SortKey) => {
    let direction: 'ascending' | 'descending' = 'ascending';
    if (sortConfig && sortConfig.key === key && sortConfig.direction === 'ascending') {
      direction = 'descending';
    }
    setSortConfig({ key, direction });
  };

  const handleEdit = (spaceId: string, field: string, value: string | string[]) => {
    setEditingId(spaceId);
    setEditedField(field);
    setEditedValue(value);
  };

  const handleSave = async (spaceId: string) => {
    if (editingId && editedField && editedValue !== null) {
      // Optimistically update the UI
      setSpaces(prevSpaces =>
        prevSpaces.map(space => {
          if (space.id === spaceId) {
            // Create a new Space instance with updated values
            const updatedSpace = new Space({
              ...space,
              [editedField]: editedValue,
              updatedAt: Timestamp.fromDate(new Date())
            });
            return updatedSpace;
          }
          return space;
        })
      );

      setEditingId(null);
      setEditedField(null);
      setEditedValue(null);
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    if (editedField === 'keywords') {
      setEditedValue(e.target.value.split(',').map(k => k.trim()));
    } else {
      setEditedValue(e.target.value);
    }
  };

  const getSortIcon = (key: SortKey) => {
    if (!sortConfig || sortConfig.key !== key) {
      return <ArrowUpDown className="h-4 w-4 text-gray-400" />;
    }
    return sortConfig.direction === 'ascending' ? 
      <ArrowUp className="h-4 w-4 text-blue-600" /> : 
      <ArrowDown className="h-4 w-4 text-blue-600" />;
  };

  const getAccessIcon = (level: string) => {
    // TODO: Implement access level logic based on new data model
    return <Globe className="h-4 w-4 text-green-600" />;
  };

  const getAccessColor = (level: string) => {
    // TODO: Implement access level logic based on new data model
    return 'bg-green-100 text-green-800 border-green-200';
  };

  const formatDate = (date: Date | string | undefined | null) => {
    if (!date) return '—';
    const d = typeof date === 'string' ? new Date(date) : date;
    if (!(d instanceof Date) || isNaN(d.getTime())) return '—';
    return d.toLocaleDateString();
  };

  const formatTimestamp = (timestamp: any) => {
    const date = safeTimestampToDate(timestamp);
    return date ? date.toLocaleDateString() : '—';
  };

  if (loading) {
    return (
      <div className="h-full flex flex-col">
        <div className="flex-shrink-0 flex items-center justify-between mb-4">
          <div className="h-6 w-32 bg-muted animate-pulse rounded" />
        </div>
        <div className="flex-1 min-h-0">
          <div className="bg-white rounded-lg border border-gray-200 overflow-hidden h-full flex flex-col">
            <div className="flex-shrink-0 h-12 bg-gray-50 animate-pulse" />
            <div className="flex-1 space-y-2 p-4">
              {Array.from({ length: 5 }).map((_, i) => (
                <div key={i} className="h-16 bg-muted animate-pulse rounded" />
              ))}
            </div>
          </div>
        </div>
      </div>
    );
  }

  const TABLE_HEADER_BG = 'bg-gray-50';

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
        {/* Empty state */}
        {spaces.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center">
            <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mb-4">
              <FileText className="h-8 w-8 text-gray-400" />
            </div>
            <h3 className="text-lg font-medium text-gray-900 mb-2">No spaces found</h3>
            <p className="text-gray-500 max-w-sm">
              Try adjusting your search criteria or create a new space to get started.
            </p>
          </div>
        ) : (
          /* Single Table with Sticky Header - keeps columns perfectly aligned */
          <div className="bg-white rounded-lg border border-gray-200 overflow-hidden h-full flex flex-col">
            <Table>
              <TableHeader className={`sticky top-0 z-10 ${TABLE_HEADER_BG}`}>
                <TableRow className={TABLE_HEADER_BG}>
                  <TableHead className="w-16">Icon</TableHead>
                  <TableHead
                    onClick={() => requestSort('title')}
                    className="cursor-pointer hover:bg-gray-100 transition-colors"
                    aria-label="Sort by Title"
                  >
                    <div className="flex items-center gap-2">
                      Title
                      {getSortIcon('title')}
                    </div>
                  </TableHead>
                  <TableHead>Keywords</TableHead>
                  <TableHead className="min-w-[200px]">Description</TableHead>
                  <TableHead
                    onClick={() => requestSort('stats.contentCount')}
                    className="cursor-pointer hover:bg-gray-100 transition-colors text-center"
                    aria-label="Sort by Document Count"
                  >
                    <div className="flex items-center justify-center gap-2">
                      <FileText className="h-4 w-4" />
                      Documents
                      {getSortIcon('stats.contentCount')}
                    </div>
                  </TableHead>
                  <TableHead
                    onClick={() => requestSort('createdAt')}
                    className="cursor-pointer hover:bg-gray-100 transition-colors"
                    aria-label="Sort by Creation Date"
                  >
                    <div className="flex items-center gap-2">
                      <Calendar className="h-4 w-4" />
                      Created
                      {getSortIcon('createdAt')}
                    </div>
                  </TableHead>
                  <TableHead
                    onClick={() => requestSort('updatedAt')}
                    className="cursor-pointer hover:bg-gray-100 transition-colors"
                    aria-label="Sort by Last Updated Date"
                  >
                    <div className="flex items-center gap-2">
                      Updated
                      {getSortIcon('updatedAt')}
                    </div>
                  </TableHead>
                  <TableHead>Access</TableHead>
                  <TableHead className="w-20">Actions</TableHead>
                </TableRow>
              </TableHeader>

              <TableBody>
                {sortedSpaces.map((space) => (
                  <TableRow key={space.id} className="hover:bg-gray-50">
                    <TableCell>
                      <span className="text-2xl">📚</span>
                    </TableCell>

                    <TableCell>
                      <div
                        onClick={() => handleEdit(space.id, 'title', space.title)}
                        className="cursor-pointer hover:bg-blue-50 rounded px-2 py-1 -mx-2 -my-1 transition-colors"
                      >
                        {editingId === space.id && editedField === 'title' ? (
                          <Input
                            value={editedValue as string}
                            onChange={handleChange}
                            onBlur={() => handleSave(space.id)}
                            className="h-8"
                            autoFocus
                          />
                        ) : (
                          <div>
                            <button
                              onClick={() => onSelect(space.id)}
                              className="font-medium text-gray-900 hover:text-blue-600 text-left"
                            >
                              {space.title}
                            </button>
                          </div>
                        )}
                      </div>
                    </TableCell>

                    <TableCell>
                      <div
                        className="cursor-pointer hover:bg-blue-50 rounded px-2 py-1 -mx-2 -my-1 transition-colors"
                      >
                        <div className="flex flex-wrap gap-1">
                        </div>
                      </div>
                    </TableCell>

                    <TableCell>
                      <div
                        onClick={() => handleEdit(space.id, 'description', space.description || '')}
                        className="cursor-pointer hover:bg-blue-50 rounded px-2 py-1 -mx-2 -my-1 transition-colors"
                      >
                        {editingId === space.id && editedField === 'description' ? (
                          <Input
                            value={editedValue as string}
                            onChange={handleChange}
                            onBlur={() => handleSave(space.id)}
                            className="h-8"
                            placeholder="Enter description"
                            autoFocus
                          />
                        ) : (
                          <span className="text-sm text-gray-600 line-clamp-2">
                            {space.description}
                          </span>
                        )}
                      </div>
                    </TableCell>

                    <TableCell className="text-center">
                      <Dialog>
                        <DialogTrigger asChild>
                          <Button variant="ghost" className="h-8 px-2 text-blue-600 hover:text-blue-800 hover:bg-blue-50" aria-label={`View documents for ${space.title}`}>
                            {space.stats?.contentCount.toString() || 0}
                          </Button>
                        </DialogTrigger>
                        <DialogContent aria-describedby={undefined}>
                          <DialogHeader>
                            <DialogTitle>Documents in {space.title}</DialogTitle>
                          </DialogHeader>
                          <div className="text-center py-8">
                            <FileText className="h-12 w-12 text-gray-400 mx-auto mb-4" />
                            <p className="text-gray-500">
                              Document list view will be implemented here
                            </p>
                          </div>
                        </DialogContent>
                      </Dialog>
                    </TableCell>

                    <TableCell>
                      <span className="text-sm text-gray-500">
                        {formatTimestamp(space.createdAt)}
                      </span>
                    </TableCell>

                    <TableCell>
                      <span className="text-sm text-gray-500">
                        {formatTimestamp(space.updatedAt)}
                      </span>
                    </TableCell>

                    <TableCell>
                      <Badge
                        variant="secondary"
                        className={`text-xs capitalize ${getAccessColor('')} flex items-center gap-1 w-fit`}
                      >
                        {getAccessIcon('')}
                        {/* {space.accessLevel} */}
                      </Badge>
                    </TableCell>

                    <TableCell>
                      <Dialog>
                        <DialogTrigger asChild>
                          <Button variant="ghost" size="icon" className="h-8 w-8" aria-label={`Edit space ${space.title}`}>
                            <Pencil className="h-4 w-4" />
                          </Button>
                        </DialogTrigger>
                        <DialogContent aria-describedby={undefined}>
                          <DialogHeader>
                            <DialogTitle>Edit Space</DialogTitle>
                          </DialogHeader>
                          <EditSpaceForm space={space} />
                        </DialogContent>
                      </Dialog>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </div>
    </div>
  );
}