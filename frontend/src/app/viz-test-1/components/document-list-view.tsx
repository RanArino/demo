'use client';

import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Separator } from '@/components/ui/separator';
import { 
  FileText, 
  Plus, 
  Eye, 
  EyeOff, 
  Trash2, 
  Search,
  Download,
  Calendar,
  Clock
} from 'lucide-react';
import { MockNode, getNodesByLayer } from '../mock-data';

interface DocumentListViewProps {
  nodes: MockNode[];
  onDocumentToggle: (documentId: string) => void;
  onDocumentAdd: (title: string, content: string) => void;
  onDocumentDelete: (documentId: string) => void;
  visibleDocuments: Set<string>;
}

export default function DocumentListView({
  nodes,
  onDocumentToggle,
  onDocumentAdd,
  onDocumentDelete,
  visibleDocuments,
}: DocumentListViewProps) {
  const [searchTerm, setSearchTerm] = useState('');
  const [newDocTitle, setNewDocTitle] = useState('');
  const [newDocContent, setNewDocContent] = useState('');
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);

  // Get document nodes (Layer 2)
  const documentNodes = getNodesByLayer(2);

  // Filter documents based on search term
  const filteredDocuments = documentNodes.filter(doc =>
    doc.action_data.textual_data.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
    doc.action_data.textual_data.keywords.some(keyword =>
      keyword.toLowerCase().includes(searchTerm.toLowerCase())
    )
  );

  const handleAddDocument = () => {
    if (newDocTitle.trim() && newDocContent.trim()) {
      onDocumentAdd(newDocTitle, newDocContent);
      setNewDocTitle('');
      setNewDocContent('');
      setIsAddDialogOpen(false);
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString();
  };

  const formatReadTime = (minutes: number) => {
    return `${minutes} min read`;
  };

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="p-4 border-b">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold">Documents</h2>
          <Dialog open={isAddDialogOpen} onOpenChange={setIsAddDialogOpen}>
            <DialogTrigger asChild>
              <Button size="sm" className="flex items-center gap-2">
                <Plus className="w-4 h-4" />
                Add Document
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Add New Document</DialogTitle>
              </DialogHeader>
              <div className="space-y-4">
                <Input
                  placeholder="Document title"
                  value={newDocTitle}
                  onChange={(e) => setNewDocTitle(e.target.value)}
                />
                <textarea
                  className="w-full p-2 border rounded-md min-h-[120px] resize-none"
                  placeholder="Document content or summary"
                  value={newDocContent}
                  onChange={(e) => setNewDocContent(e.target.value)}
                />
                <div className="flex justify-end gap-2">
                  <Button variant="outline" onClick={() => setIsAddDialogOpen(false)}>
                    Cancel
                  </Button>
                  <Button onClick={handleAddDocument}>
                    Add Document
                  </Button>
                </div>
              </div>
            </DialogContent>
          </Dialog>
        </div>

        {/* Search */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
          <Input
            placeholder="Search documents..."
            className="pl-10"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
      </div>

      {/* Document List */}
      <div className="flex-1 overflow-y-auto p-4 space-y-3">
        {filteredDocuments.length === 0 ? (
          <div className="text-center text-gray-500 py-8">
            <FileText className="w-12 h-12 mx-auto mb-4 opacity-50" />
            <p>No documents found</p>
            {searchTerm && (
              <p className="text-sm">Try adjusting your search terms</p>
            )}
          </div>
        ) : (
          filteredDocuments.map((doc) => {
            const isVisible = visibleDocuments.has(doc.id);
            const metadata = doc.action_data.general_metadata;
            
            return (
              <Card 
                key={doc.id} 
                className={`transition-all duration-200 hover:shadow-md ${
                  isVisible ? 'border-blue-200 bg-blue-50/50' : 'border-gray-200'
                }`}
              >
                <CardHeader className="pb-2">
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <CardTitle className="text-sm font-medium mb-1">
                        {doc.action_data.textual_data.title}
                      </CardTitle>
                      <p className="text-xs text-gray-600 line-clamp-2">
                        {doc.action_data.textual_data.summary}
                      </p>
                    </div>
                    <div className="flex items-center gap-1 ml-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => onDocumentToggle(doc.id)}
                        className="h-8 w-8 p-0"
                      >
                        {isVisible ? (
                          <Eye className="w-4 h-4 text-blue-600" />
                        ) : (
                          <EyeOff className="w-4 h-4 text-gray-400" />
                        )}
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => onDocumentDelete(doc.id)}
                        className="h-8 w-8 p-0 text-red-500 hover:text-red-700"
                      >
                        <Trash2 className="w-4 h-4" />
                      </Button>
                    </div>
                  </div>
                </CardHeader>
                
                <CardContent className="pt-0">
                  {/* Keywords */}
                  <div className="flex flex-wrap gap-1 mb-3">
                    {doc.action_data.textual_data.keywords.slice(0, 3).map((keyword, index) => (
                      <span
                        key={index}
                        className="px-2 py-1 bg-gray-100 text-gray-700 text-xs rounded-full"
                      >
                        {keyword}
                      </span>
                    ))}
                    {doc.action_data.textual_data.keywords.length > 3 && (
                      <span className="px-2 py-1 bg-gray-100 text-gray-500 text-xs rounded-full">
                        +{doc.action_data.textual_data.keywords.length - 3} more
                      </span>
                    )}
                  </div>

                  <Separator className="my-2" />

                  {/* Metadata */}
                  <div className="grid grid-cols-2 gap-2 text-xs text-gray-500">
                    <div className="flex items-center gap-1">
                      <FileText className="w-3 h-3" />
                      <span>{metadata.file_type?.toUpperCase()}</span>
                    </div>
                    <div className="flex items-center gap-1">
                      <Calendar className="w-3 h-3" />
                      <span>{formatDate(doc.created_at)}</span>
                    </div>
                    {metadata.page_count && (
                      <div className="flex items-center gap-1">
                        <span>{metadata.page_count} pages</span>
                      </div>
                    )}
                    {metadata.estimated_read_time_mins && (
                      <div className="flex items-center gap-1">
                        <Clock className="w-3 h-3" />
                        <span>{formatReadTime(metadata.estimated_read_time_mins)}</span>
                      </div>
                    )}
                  </div>

                  {/* Engagement Score */}
                  <div className="mt-2">
                    <div className="flex items-center justify-between text-xs">
                      <span className="text-gray-500">Engagement</span>
                      <span className="font-medium">
                        {Math.round(doc.engagement_score.overall_score * 100)}%
                      </span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-1.5 mt-1">
                      <div 
                        className="bg-blue-600 h-1.5 rounded-full transition-all duration-300"
                        style={{ width: `${doc.engagement_score.overall_score * 100}%` }}
                      />
                    </div>
                  </div>

                  {/* Actions */}
                  <div className="flex items-center gap-2 mt-3">
                    <Button variant="outline" size="sm" className="flex-1 text-xs">
                      <Download className="w-3 h-3 mr-1" />
                      Download
                    </Button>
                    <Button variant="outline" size="sm" className="flex-1 text-xs">
                      View Details
                    </Button>
                  </div>
                </CardContent>
              </Card>
            );
          })
        )}
      </div>

      {/* Footer Stats */}
      <div className="p-4 border-t bg-gray-50">
        <div className="grid grid-cols-3 gap-4 text-center">
          <div>
            <div className="text-lg font-semibold text-blue-600">{documentNodes.length}</div>
            <div className="text-xs text-gray-500">Total</div>
          </div>
          <div>
            <div className="text-lg font-semibold text-green-600">{visibleDocuments.size}</div>
            <div className="text-xs text-gray-500">Visible</div>
          </div>
          <div>
            <div className="text-lg font-semibold text-gray-600">{filteredDocuments.length}</div>
            <div className="text-xs text-gray-500">Filtered</div>
          </div>
        </div>
      </div>
    </div>
  );
} 