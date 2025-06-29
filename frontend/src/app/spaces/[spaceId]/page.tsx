import { notFound } from 'next/navigation';
import { mockSpaces } from '@/app/spaces/mock-data';

interface SpacePageProps {
  params: Promise<{
    spaceId: string;
  }>;
}

export default async function SpacePage({ params }: SpacePageProps) {
  const { spaceId } = await params;
  const space = mockSpaces.find(s => s.id === spaceId);

  if (!space) {
    notFound();
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-white border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="py-6">
            <div className="flex items-center space-x-4">
              <span className="text-4xl">{space.icon}</span>
              <div>
                <h1 className="text-3xl font-bold text-gray-900">{space.title}</h1>
                <p className="text-gray-600">{space.description}</p>
              </div>
            </div>
            <div className="mt-4 flex flex-wrap gap-2">
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
        </div>
      </div>
      
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Left Panel - Canvas (placeholder) */}
          <div className="lg:col-span-2">
            <div className="bg-white rounded-lg shadow p-6">
              <h2 className="text-xl font-semibold mb-4">Canvas View</h2>
              <div className="h-96 bg-gray-100 rounded-lg flex items-center justify-center">
                <p className="text-gray-500">3D/2D visualization will be implemented here</p>
              </div>
            </div>
          </div>
          
          {/* Right Panel - Document Preview and Chat */}
          <div className="space-y-6">
            {/* Document Preview */}
            <div className="bg-white rounded-lg shadow p-6">
              <h3 className="text-lg font-semibold mb-4">Documents ({space.document_count})</h3>
              <div className="space-y-2">
                <p className="text-gray-500">Document list will be implemented here</p>
                <div className="h-32 bg-gray-50 rounded border-2 border-dashed border-gray-300 flex items-center justify-center">
                  <p className="text-gray-400">Drag & drop documents here</p>
                </div>
              </div>
            </div>
            
            {/* Chat Session */}
            <div className="bg-white rounded-lg shadow p-6">
              <h3 className="text-lg font-semibold mb-4">Chat Session</h3>
              <div className="h-64 bg-gray-50 rounded p-4">
                <p className="text-gray-500">AI chat interface will be implemented here</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}