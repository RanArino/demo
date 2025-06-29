
import { Modal } from '@/components/ui/modal';

export default async function DocumentsModal({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return (
    <Modal title={`Documents for Space ${id}`}>
      <p>List of documents for space {id} will be displayed here.</p>
    </Modal>
  );
}
