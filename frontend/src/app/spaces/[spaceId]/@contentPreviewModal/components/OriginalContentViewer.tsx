'use client';

import { useEffect, useState } from 'react';
import { generateDownloadURL } from '@/api/actions/contentActions';
import { Skeleton } from '@/components/ui/skeleton';

interface OriginalContentViewerProps {
  contentSourceId: string;
}

export default function OriginalContentViewer({ contentSourceId }: OriginalContentViewerProps) {
  const [url, setUrl] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let mounted = true;
    setLoading(true);
    setError(null);
    (async () => {
      const res = await generateDownloadURL(contentSourceId, 'original');
      if (!mounted) return;
      if (res.ok && res.data) {
        setUrl(res.data.url);
      } else {
        setError(res.error?.message || 'Failed to load original content');
      }
      setLoading(false);
    })();
    return () => { mounted = false; };
  }, [contentSourceId]);

  if (loading) {
    return <Skeleton className="w-full h-[60vh]" />;
  }
  if (error) {
    return <div className="text-sm text-red-600 bg-red-50 p-3 rounded">{error}</div>;
  }
  if (!url) return null;

  return (
    <div className="w-full h-[70vh] border rounded">
      <iframe src={url} className="w-full h-full" />
    </div>
  );
}
