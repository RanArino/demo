"use client";
import { useRouter } from 'next/navigation';
import ContentPreviewModal from '@/app/spaces/[spaceId]/@contentPreviewModal/components/ContentPreviewModal';

interface ContentPreviewPageProps {
  params: { spaceId: string; contentSourceId: string };
}

export default function ContentPreviewPage({ params }: ContentPreviewPageProps) {
  const { contentSourceId } = params;
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
