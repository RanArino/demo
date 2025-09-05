'use client';

import { useState, useEffect, useCallback } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useSpacesStore } from '@/stores/spacesStore';
import { Space, SpaceFilters } from './types/spaces';
import SpaceFiltersComponent from './components/SpaceFilters';
import GalleryView from './components/GalleryView';
import ListView from './components/ListView';
import CanvasView from './components/CanvasView';
import CreateSpaceDialog from './components/CreateSpaceDialog';
import { Button } from '@/components/ui/button';
import { searchSpaces, deleteSpace } from '@/api/actions/spaceActions';
import { useToast } from '@/components/ui/use-toast';

interface SpacesClientPageProps {
  initialSpaces: Space[];
  initialFilters: SpaceFilters;
  totalCount: number;
  currentPage: number;
  pageSize: number;
}

export default function SpacesClientPage({
  initialSpaces,
  initialFilters,
  totalCount,
  currentPage,
  pageSize,
}: SpacesClientPageProps) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { toast } = useToast();
  
  const [spaces, setSpaces] = useState(initialSpaces);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(totalCount);
  const [showCreateDialog, setShowCreateDialog] = useState(false);

  // Zustand store
  const {
    view,
    searchTerm,
    selectedKeywords,
    sortBy,
    sortOrder,
    page,
    isDeleting,
    setView,
    setSearchTerm,
    setSelectedKeywords,
    setSortBy,
    setSortOrder,
    setPage,
    setIsDeleting,
    clearFilters,
  } = useSpacesStore();

  // Initialize store from URL params
  useEffect(() => {
    if (initialFilters.q) setSearchTerm(initialFilters.q);
    if (initialFilters.keywords) setSelectedKeywords(initialFilters.keywords);
    if (initialFilters.sortBy) setSortBy(initialFilters.sortBy);
    if (initialFilters.sortOrder) setSortOrder(initialFilters.sortOrder);
    if (initialFilters.page) setPage(initialFilters.page);
  }, []);

  // Update URL when filters change
  const updateURL = useCallback((filters: SpaceFilters) => {
    const params = new URLSearchParams();
    
    if (filters.q) params.set('q', filters.q);
    if (filters.keywords && filters.keywords.length > 0) {
      params.set('keywords', filters.keywords.join(','));
    }
    if (filters.page && filters.page > 1) params.set('page', filters.page.toString());
    if (filters.pageSize && filters.pageSize !== 20) params.set('pageSize', filters.pageSize.toString());
    if (filters.sortBy) params.set('sortBy', filters.sortBy);
    if (filters.sortOrder) params.set('sortOrder', filters.sortOrder);
    
    router.push(`/spaces?${params.toString()}`);
  }, [router]);

  // Fetch spaces with filters
  const fetchSpaces = useCallback(async (filters: SpaceFilters) => {
    setLoading(true);
    try {
      const result = await searchSpaces(filters);
      if (result.ok && result.data) {
        setSpaces(result.data.spaces);
        setTotal(result.data.totalCount);
      } else {
        toast({
          title: 'Error',
          description: result.error?.message || 'Failed to fetch spaces',
          variant: 'destructive',
        });
      }
    } catch (error) {
      toast({
        title: 'Error',
        description: 'An unexpected error occurred',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  }, [toast]);

  // Debounced search
  useEffect(() => {
    const timer = setTimeout(() => {
      if (searchTerm !== initialFilters.q) {
        const filters = {
          q: searchTerm,
          keywords: selectedKeywords,
          sortBy,
          sortOrder,
          page,
          pageSize,
        };
        updateURL(filters);
        fetchSpaces(filters);
      }
    }, 500);

    return () => clearTimeout(timer);
  }, [searchTerm]);

  // Handle filter changes (non-search)
  useEffect(() => {
    const filters = {
      q: searchTerm,
      keywords: selectedKeywords,
      sortBy,
      sortOrder,
      page,
      pageSize,
    };
    updateURL(filters);
    fetchSpaces(filters);
  }, [selectedKeywords, sortBy, sortOrder, page]);

  // Handle space selection
  const handleSpaceSelect = (spaceId: string) => {
    router.push(`/spaces/${spaceId}`);
  };

  // Handle space edit
  const handleSpaceEdit = (spaceId: string) => {
    // Open edit modal or navigate to edit page
    router.push(`/spaces/${spaceId}/edit`);
  };

  // Handle space delete
  const handleSpaceDelete = async (spaceId: string) => {
    if (!confirm('Are you sure you want to delete this space?')) return;
    
    setIsDeleting(spaceId);
    try {
      const result = await deleteSpace(spaceId);
      if (result.ok) {
        toast({
          title: 'Success',
          description: 'Space deleted successfully',
        });
        // Remove from local state
        setSpaces(spaces.filter(s => s.id !== spaceId));
      } else {
        toast({
          title: 'Error',
          description: result.error?.message || 'Failed to delete space',
          variant: 'destructive',
        });
      }
    } catch (error) {
      toast({
        title: 'Error',
        description: 'An unexpected error occurred',
        variant: 'destructive',
      });
    } finally {
      setIsDeleting(null);
    }
  };

  // Handle clear filters
  const handleClearFilters = () => {
    clearFilters();
    router.push('/spaces');
    fetchSpaces({});
  };

  // View components based on view mode
  const renderView = () => {
    const props = {
      spaces,
      onSelect: handleSpaceSelect,
      onEdit: handleSpaceEdit,
      onDelete: handleSpaceDelete,
      isDeleting,
      loading,
    };

    switch (view) {
      case 'list':
        return <ListView {...props} />;
      case 'canvas':
        return <CanvasView {...props} />;
      case 'gallery':
      default:
        return <GalleryView {...props} />;
    }
  };

  return (
    <div className="flex h-screen bg-background">
      {/* Sidebar */}
      <SpaceFiltersComponent
        searchTerm={searchTerm}
        selectedKeywords={selectedKeywords}
        onSearchChange={setSearchTerm}
        onKeywordSelect={setSelectedKeywords}
        onClearAll={handleClearFilters}
        availableKeywords={[]} // TODO: Fetch available keywords
      />

      {/* Main Content */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Header */}
        <div className="border-b border-border p-6">
          <div className="flex items-center justify-between">
            <h1 className="text-3xl font-bold">Knowledge Spaces</h1>
            
            <div className="flex items-center gap-4">
              {/* View Switcher */}
              <div className="flex items-center bg-muted rounded-lg p-1">
                <Button
                  variant={view === 'gallery' ? 'default' : 'ghost'}
                  size="sm"
                  onClick={() => setView('gallery')}
                >
                  Gallery
                </Button>
                <Button
                  variant={view === 'list' ? 'default' : 'ghost'}
                  size="sm"
                  onClick={() => setView('list')}
                >
                  List
                </Button>
                <Button
                  variant={view === 'canvas' ? 'default' : 'ghost'}
                  size="sm"
                  onClick={() => setView('canvas')}
                >
                  Canvas
                </Button>
              </div>

              {/* Create Button */}
              <Button onClick={() => setShowCreateDialog(true)}>
                Create New Space
              </Button>
            </div>
          </div>

          {/* Results count */}
          <p className="text-sm text-muted-foreground mt-2">
            {total} spaces found
          </p>
        </div>

        {/* Content Area */}
        <div className="flex-1 overflow-auto p-6">
          {loading ? (
            <div className="flex items-center justify-center h-full">
              <div className="text-muted-foreground">Loading spaces...</div>
            </div>
          ) : spaces.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-full">
              <h2 className="text-xl font-semibold mb-2">No spaces found</h2>
              <p className="text-muted-foreground mb-4">
                {searchTerm || selectedKeywords.length > 0
                  ? 'Try adjusting your filters'
                  : 'Create your first space to get started'}
              </p>
              {!searchTerm && selectedKeywords.length === 0 && (
                <Button onClick={() => setShowCreateDialog(true)}>
                  Create Your First Space
                </Button>
              )}
            </div>
          ) : (
            renderView()
          )}
        </div>

        {/* Pagination */}
        {total > pageSize && (
          <div className="border-t border-border p-4">
            <div className="flex items-center justify-between">
              <Button
                variant="outline"
                onClick={() => setPage(page - 1)}
                disabled={page === 1}
              >
                Previous
              </Button>
              
              <span className="text-sm text-muted-foreground">
                Page {page} of {Math.ceil(total / pageSize)}
              </span>
              
              <Button
                variant="outline"
                onClick={() => setPage(page + 1)}
                disabled={page >= Math.ceil(total / pageSize)}
              >
                Next
              </Button>
            </div>
          </div>
        )}
      </div>

      {/* Create Space Dialog */}
      {showCreateDialog && (
        <CreateSpaceDialog
          open={showCreateDialog}
          onClose={() => setShowCreateDialog(false)}
          onSuccess={(newSpace) => {
            setShowCreateDialog(false);
            // Refresh the spaces list
            fetchSpaces({
              q: searchTerm,
              keywords: selectedKeywords,
              sortBy,
              sortOrder,
              page,
              pageSize,
            });
          }}
        />
      )}
    </div>
  );
}