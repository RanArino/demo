
'use client';

import { useRouter } from 'next/navigation';

export default function MiniMapModal() {
  const router = useRouter();

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
      <div className="bg-white p-8 rounded-lg">
        <h2 className="text-xl font-bold">Expanded Mini-Map</h2>
        <p>This is a placeholder for the expanded mini-map view.</p>
        <button onClick={() => router.back()} className="mt-4 px-4 py-2 bg-blue-500 text-white rounded">
          Close
        </button>
      </div>
    </div>
  );
}
