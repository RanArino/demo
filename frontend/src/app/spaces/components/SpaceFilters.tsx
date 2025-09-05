import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { cn } from '@/lib/utils';

interface SpaceFiltersProps {
  searchTerm: string;
  selectedKeywords: string[];
  onSearchChange: (term: string) => void;
  onKeywordSelect: (keywords: string[]) => void;
  onClearAll: () => void;
  availableKeywords: string[];
  className?: string;
}

export default function SpaceFilters({
  searchTerm,
  selectedKeywords,
  onSearchChange,
  onKeywordSelect,
  onClearAll,
  availableKeywords,
  className,
}: SpaceFiltersProps) {
  const [localSearchTerm, setLocalSearchTerm] = useState(searchTerm);

  useEffect(() => {
    setLocalSearchTerm(searchTerm);
  }, [searchTerm]);

  const handleSearchChange = (value: string) => {
    setLocalSearchTerm(value);
    onSearchChange(value);
  };

  const toggleKeyword = (keyword: string) => {
    if (selectedKeywords.includes(keyword)) {
      onKeywordSelect(selectedKeywords.filter(k => k !== keyword));
    } else {
      onKeywordSelect([...selectedKeywords, keyword]);
    }
  };

  const hasActiveFilters = searchTerm || selectedKeywords.length > 0;

  // TODO: remove this mock keywords later
  // Mock keywords for now - in production, these would come from the backend
  const mockKeywords = [
    'AI', 'Machine Learning', 'Research', 'Documentation',
    'Projects', 'Notes', 'Ideas', 'References',
    'Articles', 'Books', 'Videos', 'Tutorials'
  ];

  const keywords = availableKeywords.length > 0 ? availableKeywords : mockKeywords;

  return (
    <aside className={cn(
      "w-[280px] border-r border-border p-6 bg-background flex flex-col h-full",
      className
    )}>
      {/* Header */}
      <div className="mb-6">
        <h2 className="text-lg font-semibold mb-4">Filters</h2>
        
        {/* Clear All Button */}
        {hasActiveFilters && (
          <Button
            variant="outline"
            size="sm"
            className="w-full mb-4"
            onClick={onClearAll}
          >
            Clear all filters
          </Button>
        )}
      </div>

      {/* Search Input */}
      <div className="mb-6">
        <Label htmlFor="search" className="mb-2 block">
          Search
        </Label>
        <Input
          id="search"
          type="text"
          placeholder="Search spaces..."
          value={localSearchTerm}
          onChange={(e) => handleSearchChange(e.target.value)}
          className="w-full"
        />
      </div>

      {/* Keywords Section */}
      <div className="flex-1 overflow-auto">
        <Label className="mb-3 block">Keywords</Label>
        <div className="space-y-2">
          {keywords.map((keyword) => (
            <button
              key={keyword}
              onClick={() => toggleKeyword(keyword)}
              className={cn(
                "w-full text-left px-3 py-2 rounded-md text-sm transition-colors",
                selectedKeywords.includes(keyword)
                  ? "bg-primary text-primary-foreground"
                  : "hover:bg-accent hover:text-accent-foreground"
              )}
            >
              <span className="flex items-center justify-between">
                {keyword}
                {selectedKeywords.includes(keyword) && (
                  <span className="text-xs"></span>
                )}
              </span>
            </button>
          ))}
        </div>
      </div>

      {/* Sort Options */}
      <div className="mt-6 pt-6 border-t border-border">
        <Label className="mb-3 block">Sort by</Label>
        <select className="w-full px-3 py-2 rounded-md border border-input bg-background text-sm">
          <option value="created-desc">Newest first</option>
          <option value="created-asc">Oldest first</option>
          <option value="name-asc">Name (A-Z)</option>
          <option value="name-desc">Name (Z-A)</option>
          <option value="documents-desc">Most documents</option>
          <option value="documents-asc">Least documents</option>
        </select>
      </div>
    </aside>
  );
}