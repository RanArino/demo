"use client";
import UploadModal from '../UploadModal';

interface UploadModalPageProps {
  params: {
    spaceId: string;
  };
  searchParams: { tab?: 'file' | 'google-drive' | 'link' | 'text' };
}

export default function UploadModalPage({ params, searchParams }: UploadModalPageProps) {
  const { spaceId } = params;

  return (
    <UploadModal
      spaceId={spaceId}
      isOpen={true}
      autoOpenTab={searchParams?.tab ?? 'file'}
      onUploadComplete={() => {}}
      onClose={() => {}}
    />
  );
}