"use client";
import ContentPreviewModal from '@/app/spaces/[spaceId]/@contentPreviewModal/components/ContentPreviewModal';

interface ContentPreviewPageProps {
  params: { spaceId: string; contentSourceId: string };
}

export default function ContentPreviewPage({ params }: ContentPreviewPageProps) {
  const { contentSourceId } = params;
  return (
    <ContentPreviewModal
      isOpen={true}
      onClose={() => {}}
      contentSourceId={contentSourceId}
    />
  );
}
