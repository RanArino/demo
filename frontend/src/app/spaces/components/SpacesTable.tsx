import React, { useMemo } from 'react';
import { Space } from '@/api/generated/v1/knowledge_pb';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { 
  Pencil, 
  FileText, 
  Calendar
} from 'lucide-react';
import { EditSpaceForm } from './EditSpaceForm';
import { DocumentsDialog } from './DocumentsDialog';
import { useSpacesTable } from '../hooks/useSpacesTable';
import { SortIcon } from '@/components/common/SortIcon';
import { AccessIcon } from '@/components/common/AccessIcon';

interface SpacesTableProps {
  spaces: Space[];
  onSelect: (spaceId: string) => void;
  onEdit: (spaceId: string) => void;
  onDelete: (spaceId: string) => void;
  isDeleting: string | null;
}


const TABLE_HEADER_BG = 'bg-gray-50';

export function SpacesTable({
  spaces,
  onSelect,
  onEdit: _onEdit,
  onDelete: _onDelete,
  isDeleting: _isDeleting,
}: SpacesTableProps) {
  const {
    editingId,
    editedField,
    editedValue,
    sortConfig,
    handleEdit,
    handleSave,
    handleChange,
    requestSort,
    getSortIcon,
    getAccessColor,
    formatTimestamp,
  } = useSpacesTable();

  const sortedSpaces = useMemo(() => {
    const sortableSpaces = [...spaces];
    if (sortConfig !== null) {
      sortableSpaces.sort((a, b) => {
        let aValue: unknown;
        let bValue: unknown;

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
          const aDate = aValue && typeof aValue === 'object' && 'seconds' in aValue 
            ? new Date((aValue as { seconds: number }).seconds * 1000) 
            : new Date(0);
          const bDate = bValue && typeof bValue === 'object' && 'seconds' in bValue 
            ? new Date((bValue as { seconds: number }).seconds * 1000) 
            : new Date(0);
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

  return (
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
                      <SortIcon sortState={getSortIcon('title')} />
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
                      <SortIcon sortState={getSortIcon('stats.contentCount')} />
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
                      <SortIcon sortState={getSortIcon('createdAt')} />
                    </div>
            </TableHead>
            <TableHead
              onClick={() => requestSort('updatedAt')}
              className="cursor-pointer hover:bg-gray-100 transition-colors"
              aria-label="Sort by Last Updated Date"
            >
                    <div className="flex items-center gap-2">
                      Updated
                      <SortIcon sortState={getSortIcon('updatedAt')} />
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
                <div className="cursor-pointer hover:bg-blue-50 rounded px-2 py-1 -mx-2 -my-1 transition-colors">
                  <div className="flex flex-wrap gap-1">
                    {/* Keywords will be implemented when available in the data model */}
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
                <DocumentsDialog space={space} />
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
                  <AccessIcon level="" />
                </Badge>
              </TableCell>

              <TableCell>
                <Dialog>
                  <DialogTrigger asChild>
                    <Button 
                      variant="ghost" 
                      size="icon" 
                      className="h-8 w-8" 
                      aria-label={`Edit space ${space.title}`}
                    >
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
  );
}
