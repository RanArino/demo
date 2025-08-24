import { Suspense } from 'react';
import { searchSpaces } from '@/api/actions/spaceActions';
import SpacesClientPage from './SpacesClientPage';
import SpacesLoading from './loading';

interface SearchParams {
  q?: string;
  keywords?: string | string[];
  page?: string;
  pageSize?: string;
  sortBy?: string;
  sortOrder?: string;
}

export default async function SpacesPage({
  searchParams,
}: {
  searchParams: SearchParams;
}) {
  // Parse search parameters
  const filters = {
    q: searchParams.q,
    keywords: Array.isArray(searchParams.keywords) 
      ? searchParams.keywords 
      : searchParams.keywords?.split(',').filter(Boolean),
    page: searchParams.page ? parseInt(searchParams.page) : 1,
    pageSize: searchParams.pageSize ? parseInt(searchParams.pageSize) : 20,
    sortBy: searchParams.sortBy as any,
    sortOrder: searchParams.sortOrder as any,
  };

  // Fetch initial spaces data
  const result = await searchSpaces(filters);
  
  // Handle error state
  if (!result.ok) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-semibold text-destructive mb-2">
            Failed to load spaces
          </h2>
          <p className="text-muted-foreground">
            {result.error?.message || 'An unexpected error occurred'}
          </p>
        </div>
      </div>
    );
  }

  const { spaces = [], totalCount = 0, page = 1, pageSize = 20 } = result.data || {};

  return (
    <Suspense fallback={<SpacesLoading />}>
      <SpacesClientPage
        initialSpaces={spaces}
        initialFilters={filters}
        totalCount={totalCount}
        currentPage={page}
        pageSize={pageSize}
      />
    </Suspense>
  );
}
