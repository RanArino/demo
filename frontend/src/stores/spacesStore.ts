import { create } from 'zustand';
import { PlainMessage } from '@bufbuild/protobuf';
import { SpacesUIState, ViewMode } from '@/api/generated/v1/knowledge_pb';

// Extract plain data type from protobuf class (without protobuf methods)
type SpacesState = PlainMessage<SpacesUIState>;

interface SpacesStore extends SpacesState {
  // Actions
  setView: (view: ViewMode) => void;
  setSearchTerm: (term: string) => void;
  setSelectedKeywords: (keywords: string[]) => void;
  setSortBy: (sortBy: string) => void;
  setSortOrder: (order: string) => void;
  setPage: (page: number) => void;
  setPageSize: (size: number) => void;
  setIsCreating: (isCreating: boolean) => void;
  setIsDeleting: (spaceId: string | undefined) => void;
  setLastError: (error: string | undefined) => void;
  setSelectedSpaceId: (spaceId: string | undefined) => void;
  clearFilters: () => void;
  reset: () => void;
}

const initialState: SpacesState = {
  view: ViewMode.GALLERY,
  searchTerm: '',
  selectedKeywords: [],
  sortBy: 'created',
  sortOrder: 'desc',
  page: 1,
  pageSize: 20,
  isCreating: false,
  isDeleting: undefined,
  lastError: undefined,
  selectedSpaceId: undefined,
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
    isDeleting: undefined,
    lastError: undefined,
    selectedSpaceId: undefined,
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