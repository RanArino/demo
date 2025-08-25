import { create } from 'zustand';
import { SpacesUIState, ViewMode } from '@/app/spaces/types/spaces';

interface SpacesStore extends SpacesUIState {
  // Actions
  setView: (view: ViewMode) => void;
  setSearchTerm: (term: string) => void;
  setSelectedKeywords: (keywords: string[]) => void;
  setSortBy: (sortBy: SpacesUIState['sortBy']) => void;
  setSortOrder: (order: SpacesUIState['sortOrder']) => void;
  setPage: (page: number) => void;
  setPageSize: (size: number) => void;
  setIsCreating: (isCreating: boolean) => void;
  setIsDeleting: (spaceId: string | null) => void;
  setLastError: (error: string | null) => void;
  setSelectedSpaceId: (spaceId: string | null) => void;
  clearFilters: () => void;
  reset: () => void;
}

const initialState: SpacesUIState = {
  view: 'gallery',
  searchTerm: '',
  selectedKeywords: [],
  sortBy: 'created',
  sortOrder: 'desc',
  page: 1,
  pageSize: 20,
  isCreating: false,
  isDeleting: null,
  lastError: null,
  selectedSpaceId: null,
};

export const useSpacesStore = create<SpacesStore>((set) => ({
  ...initialState,

  // View actions
  setView: (view) => set({ view }),
  
  // Filter actions
  setSearchTerm: (searchTerm) => set({ searchTerm, page: 1 }), // Reset page on search
  setSelectedKeywords: (selectedKeywords) => set({ selectedKeywords, page: 1 }), // Reset page on filter
  setSortBy: (sortBy) => set({ sortBy }),
  setSortOrder: (sortOrder) => set({ sortOrder }),
  
  // Pagination actions
  setPage: (page) => set({ page }),
  setPageSize: (pageSize) => set({ pageSize, page: 1 }), // Reset page on size change
  
  // UI state actions
  setIsCreating: (isCreating) => set({ isCreating }),
  setIsDeleting: (isDeleting) => set({ isDeleting }),
  setLastError: (lastError) => set({ lastError }),
  setSelectedSpaceId: (selectedSpaceId) => set({ selectedSpaceId }),
  
  // Utility actions
  clearFilters: () => set({
    searchTerm: '',
    selectedKeywords: [],
    sortBy: 'created',
    sortOrder: 'desc',
    page: 1,
  }),
  
  reset: () => set(initialState),
}));

// Selector hooks for commonly used combinations
export const useSpacesView = () => useSpacesStore((state) => state.view);
export const useSpacesFilters = () => useSpacesStore((state) => ({
  searchTerm: state.searchTerm,
  selectedKeywords: state.selectedKeywords,
  sortBy: state.sortBy,
  sortOrder: state.sortOrder,
  page: state.page,
  pageSize: state.pageSize,
}));
export const useSpacesUIState = () => useSpacesStore((state) => ({
  isCreating: state.isCreating,
  isDeleting: state.isDeleting,
  lastError: state.lastError,
  selectedSpaceId: state.selectedSpaceId,
}));