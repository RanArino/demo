"use client";
import { useRouter } from 'next/navigation';
import { use } from 'react';
import ContentPreviewModal from '@/app/spaces/[spaceId]/@contentPreviewModal/components/ContentPreviewModal';

interface ContentPreviewPageProps {
  params: Promise<{ spaceId: string; contentSourceId: string }>;
}

export default function ContentPreviewPage({ params }: ContentPreviewPageProps) {
  const { contentSourceId } = use(params);
  const router = useRouter();

  const handleClose = () => {
    router.back();
  };

  return (
    <ContentPreviewModal
      isOpen={true}
      onClose={handleClose}
      contentSourceId={contentSourceId}
    />
  );
}
