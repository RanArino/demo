import { notFound } from 'next/navigation';
import { getSpace } from '@/api/actions/spaceActions';
import UploadModal from '@/app/spaces/components/UploadModal';

interface UploadModalPageProps {
  params: {
    spaceId: string;
  };
}

export default async function UploadModalPage({ params }: UploadModalPageProps) {
  const { spaceId } = params;

  // Verify space exists
  const spaceResult = await getSpace(spaceId);
  if (!spaceResult.ok || !spaceResult.data) {
    notFound();
  }

  return (
    <UploadModal
      spaceId={spaceId}
      spaceName={spaceResult.data.title}
      isOpen={true}
      onClose={() => {
        // Navigation will be handled by the modal component
      }}
    />
  );
}