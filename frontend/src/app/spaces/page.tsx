import { searchSpaces } from '@/api/actions/spaceActions';
import { SpaceFilters, ListSpacesRequest } from '@/api/generated/v1/knowledge_pb';
import SpacesClientPage from './SpacesClientPage';

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
  const filters = new SpaceFilters({
    q: searchParams.q || '',
    keywords: Array.isArray(searchParams.keywords) 
      ? searchParams.keywords 
      : searchParams.keywords?.split(',').filter(Boolean) || [],
    sortBy: searchParams.sortBy || 'created',
    sortOrder: searchParams.sortOrder || 'desc',
  });

  // Parse pagination parameters separately
  const page = searchParams.page ? parseInt(searchParams.page) : 1;
  const pageSize = searchParams.pageSize ? parseInt(searchParams.pageSize) : 20;

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

  const { spaces = [], totalCount = 0, page: resultPage = 1, pageSize: resultPageSize = 20 } = result.data || {};

  // Create initialFilters object for the client component
  const initialFilters: Partial<ListSpacesRequest> = {
    q: searchParams.q,
    keywords: Array.isArray(searchParams.keywords) 
      ? searchParams.keywords 
      : searchParams.keywords?.split(',').filter(Boolean),
  };

  return (
    <SpacesClientPage
      initialSpaces={spaces}
      initialFilters={initialFilters}
      totalCount={Number(totalCount)}
      currentPage={resultPage}
      pageSize={resultPageSize}
    />
  );
}
