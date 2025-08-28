"use client";

import { useRouter, useSearchParams } from 'next/navigation';
import UploadModal from '@/app/spaces/components/UploadModal';

interface PageProps {
  params: { spaceId: string };
}

export default function UploadFallbackPage({ params }: PageProps) {
  const { spaceId } = params;
  const router = useRouter();
  const searchParams = useSearchParams();

  const tabParam = (searchParams.get('tab') as 'file' | 'google-drive' | 'link' | 'text' | null) || 'file';

  return (
    <UploadModal
      spaceId={spaceId}
      isOpen={true}
      autoOpenTab={tabParam}
      onClose={() => router.push(`/spaces/${spaceId}`)}
    />
  );
}


