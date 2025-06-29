
import { Suspense } from 'react';
import { getSpaces } from '@/lib/data/spaces';
import SpacesClientPage from '@/app/spaces/SpacesClientPage';
import { SpacesSkelton } from '@/components/spaces/Skelton';

interface SpacesPageProps {
  searchParams: Promise<{
    q?: string;
    keywords?: string | string[];
  }>;
}

export default async function SpacesPage({ searchParams }: SpacesPageProps) {
  const resolvedSearchParams = await searchParams;
  const spaces = await getSpaces(resolvedSearchParams);

  return (
    <Suspense fallback={<SpacesSkelton />}>
      <SpacesClientPage initialSpaces={spaces} />
    </Suspense>
  );
}
