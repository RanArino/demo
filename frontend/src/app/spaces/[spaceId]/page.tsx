import { notFound } from 'next/navigation';
import { use } from 'react';
import { getSpace } from '@/api/actions/spaceActions';
import { listContentSources } from '@/api/actions/contentActions';
import SpaceCanvas from './components/SpaceCanvas';
import ContentSourcesSection from './components/ContentSourcesSection';
import ChatSection from './components/ChatSection';
import LeftSidebar from './components/LeftSidebar';

interface SpaceDetailPageProps {
  params: Promise<{ spaceId: string }>;
}

export default async function SpaceDetailPage({ params }: SpaceDetailPageProps) {
  const { spaceId } = await params;

  // Fetch space data and content sources in parallel
  const [spaceResult, contentSourcesResult] = await Promise.all([
    getSpace(spaceId),
    listContentSources(spaceId, 'processed')
  ]);

  if (!spaceResult.ok || !spaceResult.data) {
    notFound();
  }

  const space = spaceResult.data;
  const contentSources = contentSourcesResult.ok ? contentSourcesResult.data || [] : [];
  const contentError = contentSourcesResult.ok ? undefined : (contentSourcesResult.error?.message || 'Failed to load documents');

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Left Sidebar - Hover triggered with pin functionality */}
      <LeftSidebar space={space} />
      
      {/* Single-column layout with right panel */}
      <div
        className="flex h-screen transition-all duration-300 ease-out"
        style={{ marginLeft: 'var(--sidebar-offset, 0px)' }}
      >
        {/* Main Canvas Area - Full width */}
        <div className="flex-1">
          <SpaceCanvas 
            space={space} 
            className="h-full" 
          />
        </div>

        {/* Right Panel - Documents and Chat */}
        <div className="w-80 xl:w-96 bg-white border-l border-gray-200 flex flex-col">
          {/* Documents Section */}
          <div className="flex-1 overflow-y-auto border-b border-gray-200">
            <ContentSourcesSection spaceId={spaceId} contentSources={contentSources} errorMessage={contentError} />
          </div>
          
          {/* Chat Section */}
          <div className="h-96">
            <ChatSection spaceId={spaceId} />
          </div>
        </div>
      </div>
    </div>
  );
}
