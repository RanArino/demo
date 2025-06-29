import React, { useState } from 'react';
import { Space } from '@/lib/types';
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
import { updateSpace } from '@/app/spaces/actions';
import Link from 'next/link';

interface ListViewProps {
  spaces: Space[];
}

type SortKey = 'title' | 'document_count' | 'created_at' | 'last_updated_at';

export function ListView({ spaces: initialSpaces }: ListViewProps) {
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

        if (typeof aValue === 'string' && typeof bValue === 'string') {
          return sortConfig.direction === 'ascending' ? aValue.localeCompare(bValue) : bValue.localeCompare(aValue);
        } else if (typeof aValue === 'number' && typeof bValue === 'number') {
          return sortConfig.direction === 'ascending' ? aValue - bValue : bValue - aValue;
        } else if (aValue instanceof Date && bValue instanceof Date) {
          return sortConfig.direction === 'ascending' ? aValue.getTime() - bValue.getTime() : bValue.getTime() - aValue.getTime();
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
      const formData = new FormData();
      formData.append(editedField, Array.isArray(editedValue) ? editedValue.join(', ') : editedValue);

      await updateSpace(spaceId, {}, formData);

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
    setEditedValue(e.target.value);
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
          /* Table Container - Scrollable */
          <div className="bg-white rounded-lg border border-gray-200 overflow-hidden h-full flex flex-col">
            {/* Table Header - Fixed */}
            <div className="flex-shrink-0">
              <Table>
                <TableHeader>
                  <TableRow className="bg-gray-50">
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
              </Table>
            </div>
            
            {/* Table Body - Scrollable */}
            <div className="flex-1 overflow-auto">
              <Table>
                <TableBody>
                  {sortedSpaces.map((space) => (
                    <TableRow key={space.id} className="hover:bg-gray-50">
                      <TableCell>
                        <span className="text-2xl">{space.icon}</span>
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
                              <Link href={`/spaces/${space.id}`} className="font-medium text-gray-900 hover:text-blue-600">
                                {space.title}
                              </Link>
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
                              value={(editedValue as string[]).join(', ')} 
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
                              {space.document_count}
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
                          {new Date(space.created_at).toLocaleDateString()}
                        </span>
                      </TableCell>

                      <TableCell>
                        <span className="text-sm text-gray-500">
                          {new Date(space.last_updated_at).toLocaleDateString()}
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
          </div>
        )}
      </div>
    </div>
  );
}