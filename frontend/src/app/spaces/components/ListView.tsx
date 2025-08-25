import React, { useState } from 'react';
import { Space } from '../types/spaces';
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

type SortKey = 'title' | 'document_count' | 'created_at' | 'last_updated_at';

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
        const aValue = a[sortConfig.key];
        const bValue = b[sortConfig.key];

        if (sortConfig.key === 'created_at' || sortConfig.key === 'last_updated_at') {
          // Ensure values are parsed as dates for correct comparison
          const aDate = new Date(aValue as string | Date);
          const bDate = new Date(bValue as string | Date);
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
        prevSpaces.map(space =>
          space.id === spaceId
            ? { ...space, [editedField]: editedValue, last_updated_at: new Date() }
            : space
        )
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
    switch (level) {
      case 'public': return <Globe className="h-4 w-4 text-green-600" />;
      case 'shared': return <Users className="h-4 w-4 text-blue-600" />;
      case 'private': return <Lock className="h-4 w-4 text-gray-600" />;
      default: return <Eye className="h-4 w-4 text-gray-600" />;
    }
  };

  const getAccessColor = (level: string) => {
    switch (level) {
      case 'public': return 'bg-green-100 text-green-800 border-green-200';
      case 'shared': return 'bg-blue-100 text-blue-800 border-blue-200';  
      case 'private': return 'bg-gray-100 text-gray-800 border-gray-200';
      default: return 'bg-gray-100 text-gray-800 border-gray-200';
    }
  };

  const formatDate = (date: Date | string | undefined | null) => {
    if (!date) return '—';
    const d = typeof date === 'string' ? new Date(date) : date;
    if (!(d instanceof Date) || isNaN(d.getTime())) return '—';
    return d.toLocaleDateString();
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
                  >
                    <div className="flex items-center gap-2">
                      Title
                      {getSortIcon('title')}
                    </div>
                  </TableHead>
                  <TableHead>Keywords</TableHead>
                  <TableHead className="min-w-[200px]">Description</TableHead>
                  <TableHead
                    onClick={() => requestSort('document_count')}
                    className="cursor-pointer hover:bg-gray-100 transition-colors text-center"
                  >
                    <div className="flex items-center justify-center gap-2">
                      <FileText className="h-4 w-4" />
                      Documents
                      {getSortIcon('document_count')}
                    </div>
                  </TableHead>
                  <TableHead
                    onClick={() => requestSort('created_at')}
                    className="cursor-pointer hover:bg-gray-100 transition-colors"
                  >
                    <div className="flex items-center gap-2">
                      <Calendar className="h-4 w-4" />
                      Created
                      {getSortIcon('created_at')}
                    </div>
                  </TableHead>
                  <TableHead
                    onClick={() => requestSort('last_updated_at')}
                    className="cursor-pointer hover:bg-gray-100 transition-colors"
                  >
                    <div className="flex items-center gap-2">
                      Updated
                      {getSortIcon('last_updated_at')}
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
                      <span className="text-2xl">{space.icon || '📚'}</span>
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
                        onClick={() => handleEdit(space.id, 'keywords', space.keywords)}
                        className="cursor-pointer hover:bg-blue-50 rounded px-2 py-1 -mx-2 -my-1 transition-colors"
                      >
                        {editingId === space.id && editedField === 'keywords' ? (
                          <Input
                            value={Array.isArray(editedValue) ? editedValue.join(', ') : ''}
                            onChange={handleChange}
                            onBlur={() => handleSave(space.id)}
                            className="h-8"
                            placeholder="Enter keywords, separated by commas"
                            autoFocus
                          />
                        ) : (
                          <div className="flex flex-wrap gap-1">
                            {space.keywords.slice(0, 3).map((keyword) => (
                              <Badge key={keyword} variant="outline" className="text-xs">
                                {keyword}
                              </Badge>
                            ))}
                            {space.keywords.length > 3 && (
                              <Badge variant="outline" className="text-xs text-gray-400">
                                +{space.keywords.length - 3}
                              </Badge>
                            )}
                          </div>
                        )}
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
                          <Button variant="ghost" className="h-8 px-2 text-blue-600 hover:text-blue-800 hover:bg-blue-50">
                            {space.document_count || 0}
                          </Button>
                        </DialogTrigger>
                        <DialogContent>
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
                        {formatDate(space.created_at)}
                      </span>
                    </TableCell>

                    <TableCell>
                      <span className="text-sm text-gray-500">
                        {formatDate(space.last_updated_at)}
                      </span>
                    </TableCell>

                    <TableCell>
                      <Badge
                        variant="secondary"
                        className={`text-xs capitalize ${getAccessColor(space.access_level)} flex items-center gap-1 w-fit`}
                      >
                        {getAccessIcon(space.access_level)}
                        {space.access_level}
                      </Badge>
                    </TableCell>

                    <TableCell>
                      <Dialog>
                        <DialogTrigger asChild>
                          <Button variant="ghost" size="icon" className="h-8 w-8">
                            <Pencil className="h-4 w-4" />
                          </Button>
                        </DialogTrigger>
                        <DialogContent>
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