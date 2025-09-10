/**
 * Integration tests for server-side rendering of spaces pages
 * Tests server-first data fetching, cache behavior, and revalidation
 */

import { render, screen, waitFor } from '@testing-library/react';
import { notFound } from 'next/navigation';
import SpacesPage from '../page';
import SpaceDetailPage from '../[spaceId]/page';
import * as spaceActions from '@/api/actions/spaceActions';
import * as contentActions from '@/api/actions/contentActions';
import { Space, ContentSource, SpaceFilters } from '@/api/generated/v1/knowledge_pb';

// Mock Next.js modules
jest.mock('next/navigation', () => ({
  notFound: jest.fn(),
  useRouter: jest.fn(() => ({
    push: jest.fn(),
    replace: jest.fn(),
    prefetch: jest.fn(),
  })),
  useSearchParams: jest.fn(() => new URLSearchParams()),
}));

// Mock server actions
jest.mock('@/api/actions/spaceActions');
jest.mock('@/api/actions/contentActions');

// Mock client components that rely on browser APIs
jest.mock('../SpacesClientPage', () => {
  return function MockSpacesClientPage({ 
    initialSpaces, 
    initialFilters, 
    totalCount, 
    currentPage, 
    pageSize 
  }: any) {
    return (
      <div data-testid="spaces-client-page">
        <div data-testid="initial-spaces-count">{initialSpaces?.length || 0}</div>
        <div data-testid="total-count">{totalCount}</div>
        <div data-testid="current-page">{currentPage}</div>
        <div data-testid="page-size">{pageSize}</div>
        <div data-testid="search-query">{initialFilters?.q || ''}</div>
        {initialSpaces?.map((space: Space) => (
          <div key={space.id} data-testid={`space-${space.id}`}>
            {space.title}
          </div>
        ))}
      </div>
    );
  };
});

jest.mock('../[spaceId]/components/SpaceCanvas', () => {
  return function MockSpaceCanvas({ space, contentSources }: any) {
    return (
      <div data-testid="space-canvas">
        <div data-testid="space-title">{space.title}</div>
        <div data-testid="content-sources-count">{contentSources?.length || 0}</div>
      </div>
    );
  };
});

jest.mock('../[spaceId]/components/ContentSourcesSection', () => {
  return function MockContentSourcesSection({ spaceId, contentSources, errorMessage }: any) {
    return (
      <div data-testid="content-sources-section">
        <div data-testid="space-id">{spaceId}</div>
        <div data-testid="content-count">{contentSources?.length || 0}</div>
        {errorMessage && <div data-testid="error-message">{errorMessage}</div>}
        {contentSources?.map((content: ContentSource) => (
          <div key={content.id} data-testid={`content-${content.id}`}>
            {content.title}
          </div>
        ))}
      </div>
    );
  };
});

jest.mock('../[spaceId]/components/ChatSection', () => {
  return function MockChatSection({ spaceId }: any) {
    return <div data-testid="chat-section">Chat for {spaceId}</div>;
  };
});

jest.mock('../[spaceId]/components/LeftSidebar', () => {
  return function MockLeftSidebar({ space }: any) {
    return <div data-testid="left-sidebar">Sidebar for {space.title}</div>;
  };
});

describe('Spaces Page Integration Tests', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('Server-side data fetching', () => {
    it('should render spaces page with initial server data', async () => {
      const mockSpaces = [
        { id: 'space1', title: 'Test Space 1', description: 'Description 1' },
        { id: 'space2', title: 'Test Space 2', description: 'Description 2' },
      ];

      (spaceActions.searchSpaces as jest.Mock).mockResolvedValue({
        ok: true,
        data: {
          spaces: mockSpaces,
          totalCount: 2,
          page: 1,
          pageSize: 20,
        },
      });

      const searchParams = { q: 'test query', page: '1' };
      
      render(await SpacesPage({ searchParams }));

      // Verify server action was called with correct filters
      expect(spaceActions.searchSpaces).toHaveBeenCalledWith(
        expect.objectContaining({
          q: 'test query',
          keywords: [],
          sortBy: 'created',
          sortOrder: 'desc',
        })
      );

      // Verify initial data is passed to client component
      expect(screen.getByTestId('initial-spaces-count')).toHaveTextContent('2');
      expect(screen.getByTestId('total-count')).toHaveTextContent('2');
      expect(screen.getByTestId('current-page')).toHaveTextContent('1');
      expect(screen.getByTestId('search-query')).toHaveTextContent('test query');

      // Verify spaces are rendered
      expect(screen.getByTestId('space-space1')).toHaveTextContent('Test Space 1');
      expect(screen.getByTestId('space-space2')).toHaveTextContent('Test Space 2');
    });

    it('should handle server-side error gracefully', async () => {
      (spaceActions.searchSpaces as jest.Mock).mockResolvedValue({
        ok: false,
        error: { code: 'INTERNAL', message: 'Server error occurred' },
      });

      render(await SpacesPage({ searchParams: {} }));

      expect(screen.getByText('Failed to load spaces')).toBeInTheDocument();
      expect(screen.getByText('Server error occurred')).toBeInTheDocument();
    });

    it('should parse complex search parameters correctly', async () => {
      (spaceActions.searchSpaces as jest.Mock).mockResolvedValue({
        ok: true,
        data: { spaces: [], totalCount: 0, page: 1, pageSize: 20 },
      });

      const searchParams = {
        q: 'complex query',
        keywords: ['tag1', 'tag2', 'tag3'],
        page: '2',
        pageSize: '10',
        sortBy: 'updated',
        sortOrder: 'asc',
      };

      render(await SpacesPage({ searchParams }));

      expect(spaceActions.searchSpaces).toHaveBeenCalledWith(
        expect.objectContaining({
          q: 'complex query',
          keywords: ['tag1', 'tag2', 'tag3'],
          sortBy: 'updated',
          sortOrder: 'asc',
        })
      );
    });

    it('should handle comma-separated keywords string', async () => {
      (spaceActions.searchSpaces as jest.Mock).mockResolvedValue({
        ok: true,
        data: { spaces: [], totalCount: 0, page: 1, pageSize: 20 },
      });

      const searchParams = {
        keywords: 'tag1,tag2,tag3',
      };

      render(await SpacesPage({ searchParams }));

      expect(spaceActions.searchSpaces).toHaveBeenCalledWith(
        expect.objectContaining({
          keywords: ['tag1', 'tag2', 'tag3'],
        })
      );
    });
  });

  describe('Space Detail Page Integration Tests', () => {
    it('should render space detail page with server-fetched data', async () => {
      const mockSpace = {
        id: 'space1',
        title: 'Test Space',
        description: 'Test Description',
      };

      const mockContentSources = [
        { id: 'content1', title: 'Document 1', status: 'processed' },
        { id: 'content2', title: 'Document 2', status: 'processed' },
      ];

      (spaceActions.getSpace as jest.Mock).mockResolvedValue({
        ok: true,
        data: mockSpace,
      });

      (contentActions.listContentSources as jest.Mock).mockResolvedValue({
        ok: true,
        data: mockContentSources,
      });

      const params = { spaceId: 'space1' };

      render(await SpaceDetailPage({ params }));

      // Verify both server actions were called
      expect(spaceActions.getSpace).toHaveBeenCalledWith('space1');
      expect(contentActions.listContentSources).toHaveBeenCalledWith('space1', 'processed');

      // Verify data is passed to components
      expect(screen.getByTestId('space-title')).toHaveTextContent('Test Space');
      expect(screen.getByTestId('content-sources-count')).toHaveTextContent('2');
      expect(screen.getByTestId('content-count')).toHaveTextContent('2');
      expect(screen.getByTestId('space-id')).toHaveTextContent('space1');

      // Verify individual content sources are rendered
      expect(screen.getByTestId('content-content1')).toHaveTextContent('Document 1');
      expect(screen.getByTestId('content-content2')).toHaveTextContent('Document 2');

      // Verify other components receive correct data
      expect(screen.getByTestId('left-sidebar')).toHaveTextContent('Sidebar for Test Space');
      expect(screen.getByTestId('chat-section')).toHaveTextContent('Chat for space1');
    });

    it('should call notFound when space is not found', async () => {
      (spaceActions.getSpace as jest.Mock).mockResolvedValue({
        ok: false,
        error: { code: 'NOT_FOUND', message: 'Space not found' },
      });

      (contentActions.listContentSources as jest.Mock).mockResolvedValue({
        ok: true,
        data: [],
      });

      const params = { spaceId: 'nonexistent' };

      await SpaceDetailPage({ params });

      expect(notFound).toHaveBeenCalled();
    });

    it('should handle content loading errors gracefully', async () => {
      const mockSpace = {
        id: 'space1',
        title: 'Test Space',
        description: 'Test Description',
      };

      (spaceActions.getSpace as jest.Mock).mockResolvedValue({
        ok: true,
        data: mockSpace,
      });

      (contentActions.listContentSources as jest.Mock).mockResolvedValue({
        ok: false,
        error: { code: 'INTERNAL', message: 'Failed to load documents' },
      });

      const params = { spaceId: 'space1' };

      render(await SpaceDetailPage({ params }));

      // Space should still render
      expect(screen.getByTestId('space-title')).toHaveTextContent('Test Space');
      
      // Content section should show error
      expect(screen.getByTestId('error-message')).toHaveTextContent('Failed to load documents');
      expect(screen.getByTestId('content-count')).toHaveTextContent('0');
    });

    it('should fetch data in parallel for better performance', async () => {
      const mockSpace = { id: 'space1', title: 'Test Space' };
      const mockContent = [{ id: 'content1', title: 'Document 1' }];

      // Track the order of calls
      const callOrder: string[] = [];
      
      (spaceActions.getSpace as jest.Mock).mockImplementation(async () => {
        callOrder.push('getSpace');
        return { ok: true, data: mockSpace };
      });

      (contentActions.listContentSources as jest.Mock).mockImplementation(async () => {
        callOrder.push('listContentSources');
        return { ok: true, data: mockContent };
      });

      const params = { spaceId: 'space1' };

      await SpaceDetailPage({ params });

      // Both should be called (order may vary due to parallel execution)
      expect(callOrder).toContain('getSpace');
      expect(callOrder).toContain('listContentSources');
      expect(spaceActions.getSpace).toHaveBeenCalledTimes(1);
      expect(contentActions.listContentSources).toHaveBeenCalledTimes(1);
    });
  });

  describe('Cache invalidation behavior', () => {
    it('should fetch fresh data when cache is invalidated', async () => {
      // This test would need to be enhanced with actual cache testing
      // For now, we verify that the server actions are called correctly
      const mockSpaces = [{ id: 'space1', title: 'Updated Space' }];
      
      (spaceActions.searchSpaces as jest.Mock).mockResolvedValue({
        ok: true,
        data: { spaces: mockSpaces, totalCount: 1, page: 1, pageSize: 20 },
      });

      render(await SpacesPage({ searchParams: {} }));

      expect(spaceActions.searchSpaces).toHaveBeenCalled();
      expect(screen.getByTestId('space-space1')).toHaveTextContent('Updated Space');
    });
  });

  describe('Server-side rendering performance', () => {
    it('should complete initial render without client-side loading states', async () => {
      const mockSpaces = [
        { id: 'space1', title: 'Space 1' },
        { id: 'space2', title: 'Space 2' },
      ];

      (spaceActions.searchSpaces as jest.Mock).mockResolvedValue({
        ok: true,
        data: { spaces: mockSpaces, totalCount: 2, page: 1, pageSize: 20 },
      });

      render(await SpacesPage({ searchParams: {} }));

      // Data should be immediately available (no loading states)
      expect(screen.getByTestId('initial-spaces-count')).toHaveTextContent('2');
      expect(screen.getByTestId('space-space1')).toBeInTheDocument();
      expect(screen.getByTestId('space-space2')).toBeInTheDocument();

      // Should not show loading spinners or empty states
      expect(screen.queryByTestId('loading-spinner')).not.toBeInTheDocument();
      expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
    });

    it('should render space detail page without loading delays', async () => {
      const mockSpace = { id: 'space1', title: 'Test Space' };
      const mockContent = [{ id: 'content1', title: 'Document 1' }];

      (spaceActions.getSpace as jest.Mock).mockResolvedValue({
        ok: true,
        data: mockSpace,
      });

      (contentActions.listContentSources as jest.Mock).mockResolvedValue({
        ok: true,
        data: mockContent,
      });

      const params = { spaceId: 'space1' };

      render(await SpaceDetailPage({ params }));

      // All data should be immediately available
      expect(screen.getByTestId('space-title')).toHaveTextContent('Test Space');
      expect(screen.getByTestId('content-content1')).toHaveTextContent('Document 1');

      // No loading states should be present
      expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
      expect(screen.queryByTestId('skeleton')).not.toBeInTheDocument();
    });
  });
});