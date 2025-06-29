'use client';

import { useState, useEffect, useMemo } from 'react';
import { Space } from '@/lib/types';
import { GalleryView } from '@/components/spaces/GalleryView';
import { ListView } from '@/components/spaces/ListView';
import { CanvasView } from '@/components/spaces/CanvasView';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Checkbox } from '@/components/ui/checkbox';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { CreateSpaceForm } from '@/components/spaces/CreateSpaceForm';

interface SpacesClientPageProps {
  initialSpaces: Space[];
}

export default function SpacesClientPage({ initialSpaces }: SpacesClientPageProps) {
  const [view, setView] = useState('gallery');
  const [spaces, setSpaces] = useState<Space[]>(initialSpaces);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedKeywords, setSelectedKeywords] = useState<string[]>([]);

  useEffect(() => {
    setSpaces(initialSpaces);
  }, [initialSpaces]);

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(e.target.value);
  };

  const handleKeywordChange = (keyword: string, checked: boolean) => {
    const newKeywords = checked
      ? [...selectedKeywords, keyword]
      : selectedKeywords.filter(k => k !== keyword);
    setSelectedKeywords(newKeywords);
  };

  const availableKeywords = useMemo(() => {
    const keywords = new Set<string>();
    initialSpaces.forEach(space => {
      space.keywords.forEach(keyword => keywords.add(keyword));
    });
    return Array.from(keywords);
  }, [initialSpaces]);

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="flex">
        {/* Sidebar */}
        <aside className="w-80 bg-white border-r border-gray-200 min-h-screen">
          <div className="p-6">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-lg font-semibold text-gray-900">Filters</h2>
              {(searchTerm || selectedKeywords.length > 0) && (
                <Button 
                  variant="ghost" 
                  size="sm"
                  onClick={() => {
                    setSearchTerm('');
                    setSelectedKeywords([]);
                    updateURLParams('', []);
                  }}
                  className="text-gray-500 hover:text-gray-700"
                >
                  Clear all
                </Button>
              )}
            </div>
            
            {/* Search Section */}
            <div className="mb-6">
              <h3 className="font-medium text-gray-700 mb-3">Search</h3>
              <Input
                type="text"
                placeholder="Search spaces, titles, descriptions..."
                value={searchTerm}
                onChange={handleSearchChange}
                className="w-full"
              />
            </div>

            {/* Keywords Section */}
            <div className="mb-6">
              <h3 className="font-medium text-gray-700 mb-3">Keywords ({availableKeywords.length})</h3>
              <div className="max-h-64 overflow-y-auto space-y-2">
                {availableKeywords.slice(0, 30).map(keyword => (
                  <div key={keyword} className="flex items-center space-x-2">
                    <Checkbox
                      id={keyword}
                      checked={selectedKeywords.includes(keyword)}
                      onCheckedChange={(checked) => handleKeywordChange(keyword, checked as boolean)}
                    />
                    <label 
                      htmlFor={keyword} 
                      className="text-sm text-gray-600 cursor-pointer flex-1 leading-none"
                    >
                      {keyword}
                    </label>
                    <span className="text-xs text-gray-400">
                      {initialSpaces.filter(s => s.keywords.includes(keyword)).length}
                    </span>
                  </div>
                ))}
                {availableKeywords.length > 30 && (
                  <p className="text-xs text-gray-400 pt-2">
                    +{availableKeywords.length - 30} more keywords
                  </p>
                )}
              </div>
            </div>
          </div>
        </aside>

        {/* Main Content */}
        <main className="flex-1">
          {/* Header */}
          <header className="bg-white border-b border-gray-200 px-6 py-4">
            <div className="flex items-center justify-between">
              <div>
                <h1 className="text-2xl font-bold text-gray-900">Knowledge Spaces</h1>
                <p className="text-gray-600 mt-1">
                  Organize and explore your documents in interactive spaces
                </p>
              </div>
              
              <div className="flex items-center gap-4">
                {/* View Switcher */}
                <div className="flex items-center gap-1 bg-gray-100 rounded-lg p-1">
                  <Button 
                    onClick={() => setView('gallery')} 
                    variant={view === 'gallery' ? 'default' : 'ghost'}
                    size="sm"
                    className="text-xs px-3"
                  >
                    Gallery
                  </Button>
                  <Button 
                    onClick={() => setView('list')} 
                    variant={view === 'list' ? 'default' : 'ghost'}
                    size="sm"
                    className="text-xs px-3"
                  >
                    List
                  </Button>
                  <Button 
                    onClick={() => setView('canvas')}
                    variant={view === 'canvas' ? 'default' : 'ghost'}
                    size="sm"
                    className="text-xs px-3"
                  >
                    Canvas
                  </Button>
                </div>

                {/* Create Button */}
                <Dialog>
                  <DialogTrigger asChild>
                    <Button className="bg-blue-600 hover:bg-blue-700">
                      Create New Space
                    </Button>
                  </DialogTrigger>
                  <DialogContent>
                    <DialogHeader>
                      <DialogTitle>Create New Space</DialogTitle>
                    </DialogHeader>
                    <CreateSpaceForm />
                  </DialogContent>
                </Dialog>
              </div>
            </div>
          </header>

          {/* Content Area */}
          <div className="p-6">
            {view === 'gallery' && <GalleryView spaces={spaces} />}
            {view === 'list' && <ListView spaces={spaces} />}
            {view === 'canvas' && <CanvasView spaces={spaces} />}
          </div>
        </main>
      </div>
    </div>
  );
}
