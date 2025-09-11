"use client";
import UploadModal from '../UploadModal';
import { useParams, useSearchParams } from 'next/navigation';

type AllowedTab = 'file' | 'google-drive' | 'link' | 'text';

export default function UploadModalPage() {
  const params = useParams();
  const searchParams = useSearchParams();

  const rawSpaceId = (params as Record<string, string | string[]>).spaceId;
  const spaceId = Array.isArray(rawSpaceId) ? rawSpaceId[0] : rawSpaceId;

  const tabParam = searchParams.get('tab');
  const autoOpenTab: AllowedTab =
    tabParam === 'google-drive' || tabParam === 'link' || tabParam === 'text' ? (tabParam as AllowedTab) : 'file';

  return (
    <UploadModal
      spaceId={spaceId}
      isOpen={true}
      autoOpenTab={autoOpenTab}
      onUploadComplete={() => {}}
      onClose={() => {}}
    />
  );
}