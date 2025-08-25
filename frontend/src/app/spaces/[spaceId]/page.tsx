import { notFound } from 'next/navigation';
import { getSpace, listContentSources } from '@/api/actions/spaceActions';
import SpaceCanvas from './components/SpaceCanvas';
import DocumentsSection from './components/DocumentsSection';
import ChatSection from './components/ChatSection';
import LeftSidebar from './components/LeftSidebar';

interface SpaceDetailPageProps {
  params: {
    spaceId: string;
  };
}

export default async function SpaceDetailPage({ params }: SpaceDetailPageProps) {
  const { spaceId } = params;

  // Fetch space data and content sources in parallel
  const [spaceResult, contentSourcesResult] = await Promise.all([
    getSpace(spaceId),
    listContentSources(spaceId)
  ]);

  if (!spaceResult.ok || !spaceResult.data) {
    notFound();
  }

  const space = spaceResult.data;
  const contentSources = contentSourcesResult.ok ? contentSourcesResult.data || [] : [];

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Left Sidebar - Hover triggered with pin functionality */}
      <LeftSidebar space={space} />
      
      {/* Single-column layout with right panel */}
      <div className="flex h-screen">
        {/* Main Canvas Area - Full width */}
        <div className="flex-1">
          <SpaceCanvas 
            space={space} 
            contentSources={contentSources} 
            className="h-full" 
          />
        </div>

        {/* Right Panel - Documents and Chat */}
        <div className="w-80 xl:w-96 bg-white border-l border-gray-200 flex flex-col">
          {/* Documents Section */}
          <div className="flex-1 overflow-y-auto border-b border-gray-200">
            <DocumentsSection spaceId={spaceId} contentSources={contentSources} />
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