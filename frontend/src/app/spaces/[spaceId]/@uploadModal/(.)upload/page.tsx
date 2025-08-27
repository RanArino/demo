"use client";
import UploadModal from '@/app/spaces/components/UploadModal';

interface UploadModalPageProps {
  params: {
    spaceId: string;
  };
}

export default function UploadModalPage({ params }: UploadModalPageProps) {
  const { spaceId } = params;

  return (
    <UploadModal
      spaceId={spaceId}
      isOpen={true}
      onUploadComplete={() => {}}
      onClose={() => {}}
    />
  );
}