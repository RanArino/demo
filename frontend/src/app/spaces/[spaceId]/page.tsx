import { notFound } from 'next/navigation';
import { getSpace, listContentSources } from '@/api/actions/spaceActions';
import SpaceHeader from './components/SpaceHeader';
import SpaceCanvas from './components/SpaceCanvas';
import DocumentsSection from './components/DocumentsSection';
import ChatSection from './components/ChatSection';

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
      {/* Three-column responsive layout */}
      <div className="flex flex-col lg:flex-row h-screen">
        {/* Left Sidebar - Space Header and Navigation */}
        <div className="lg:w-80 xl:w-96 bg-white border-r border-gray-200 flex flex-col">
          <SpaceHeader space={space} />
          
          {/* Navigation/Metadata Section */}
          <div className="flex-1 p-6 overflow-y-auto">
            <div className="space-y-6">
              {/* Space Stats */}
              <div className="grid grid-cols-2 gap-4">
                <div className="bg-gray-50 rounded-lg p-4">
                  <div className="text-2xl font-bold text-gray-900">
                    {space.documentCount || 0}
                  </div>
                  <div className="text-sm text-gray-600">Documents</div>
                </div>
                <div className="bg-gray-50 rounded-lg p-4">
                  <div className="text-2xl font-bold text-gray-900">
                    {((space.totalSizeBytes || 0) / (1024 * 1024)).toFixed(1)}
                  </div>
                  <div className="text-sm text-gray-600">MB Used</div>
                </div>
              </div>

              {/* Keywords */}
              {space.keywords && space.keywords.length > 0 && (
                <div>
                  <h3 className="text-sm font-medium text-gray-900 mb-2">Keywords</h3>
                  <div className="flex flex-wrap gap-2">
                    {space.keywords.map((keyword) => (
                      <span
                        key={keyword}
                        className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800"
                      >
                        {keyword}
                      </span>
                    ))}
                  </div>
                </div>
              )}

              {/* Access Level */}
              <div>
                <h3 className="text-sm font-medium text-gray-900 mb-2">Access Level</h3>
                <div className="flex items-center gap-2">
                  <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800 capitalize">
                    {space.accessLevel}
                  </span>
                </div>
              </div>

              {/* Timestamps */}
              <div className="space-y-2 text-sm text-gray-600">
                <div>
                  <span className="font-medium">Created:</span>{' '}
                  {new Date(space.createdAt).toLocaleDateString()}
                </div>
                <div>
                  <span className="font-medium">Updated:</span>{' '}
                  {new Date(space.lastUpdatedAt).toLocaleDateString()}
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Main Content Area - Canvas */}
        <div className="flex-1 flex flex-col">
          <SpaceCanvas space={space} contentSources={contentSources} />
        </div>

        {/* Right Sidebar - Documents and Chat */}
        <div className="lg:w-80 xl:w-96 bg-white border-l border-gray-200 flex flex-col">
          {/* Documents Section */}
          <div className="flex-1 border-b border-gray-200">
            <DocumentsSection spaceId={spaceId} contentSources={contentSources} />
          </div>
          
          {/* Chat Section */}
          <div className="h-80">
            <ChatSection spaceId={spaceId} />
          </div>
        </div>
      </div>
    </div>
  );
}