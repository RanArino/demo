'use server';

import 'server-only';
import { Space } from '@/lib/types';
import { mockSpaces } from '@/app/spaces/mock-data';

interface GetSpacesParams {
  q?: string;
  keywords?: string | string[];
}

export async function getSpaces(params: GetSpacesParams = {}): Promise<Space[]> {
  const { q, keywords } = params;
  
  // For development, skip API call and use mock data directly
  // This prevents server-side fetch errors from breaking the page
  if (process.env.NODE_ENV === 'development') {
    console.log('Using mock data for development');
    return filterSpaces(mockSpaces, { q, keywords });
  }
  
  // Construct the URL with query parameters
  const url = new URL('http://localhost:8080/api/v1/spaces');
  
  if (q) {
    url.searchParams.append('q', q);
  }
  
  if (keywords) {
    if (Array.isArray(keywords)) {
      keywords.forEach(keyword => url.searchParams.append('keywords', keyword));
    } else {
      url.searchParams.append('keywords', keywords);
    }
  }

  try {
    // Replace with your actual API endpoint
    const res = await fetch(url.toString(), {
      // We can add caching options here if needed, e.g., next: { revalidate: 3600 }
      cache: 'no-store', // For now, disable caching to ensure fresh data
    });

    if (!res.ok) {
      // Log the error for debugging on the server
      console.error('Failed to fetch spaces:', res.status, res.statusText);
      // Fallback to mock data during development
      return filterSpaces(mockSpaces, { q, keywords });
    }

    const data = await res.json();
    return data.spaces || []; // Assuming the API returns { spaces: [...] }
  } catch (error) {
    console.error('Network or other error fetching spaces:', error);
    // In case of a network error, return mock data during development
    return filterSpaces(mockSpaces, { q, keywords });
  }
}

// Helper function to filter mock data based on search params
function filterSpaces(spaces: Space[], params: { q?: string; keywords?: string | string[] }): Space[] {
  const { q, keywords } = params;
  let filteredSpaces = [...spaces];

  // Filter by search query
  if (q && q.trim()) {
    const searchTerm = q.toLowerCase();
    filteredSpaces = filteredSpaces.filter(space =>
      space.title.toLowerCase().includes(searchTerm) ||
      space.description.toLowerCase().includes(searchTerm) ||
      space.keywords.some(keyword => keyword.toLowerCase().includes(searchTerm))
    );
  }

  // Filter by keywords
  if (keywords) {
    const keywordArray = Array.isArray(keywords) ? keywords : [keywords];
    if (keywordArray.length > 0) {
      filteredSpaces = filteredSpaces.filter(space =>
        keywordArray.some(keyword => 
          space.keywords.includes(keyword)
        )
      );
    }
  }

  return filteredSpaces;
}
