
import React from 'react';
import { Space } from '@/lib/types';
import { useRouter } from 'next/navigation';

interface MiniMapProps {
  spaces: Space[];
}

const MiniMap: React.FC<MiniMapProps> = ({ spaces }) => {
  const router = useRouter();

  return (
    <div className="p-4 bg-gray-100 rounded-lg shadow-md">
      <h3 className="text-lg font-semibold mb-2">Mini-Map</h3>
      <div className="grid grid-cols-3 gap-2">
        {spaces.map((space) => (
          <div
            key={space.id}
            className="w-16 h-16 bg-gray-300 rounded-md flex items-center justify-center text-xs cursor-pointer hover:bg-gray-400"
            onClick={() => router.push(`/spaces/minimap/${space.id}`)}
          >
            {space.icon}
          </div>
        ))}
      </div>
    </div>
  );
};

export default MiniMap;
