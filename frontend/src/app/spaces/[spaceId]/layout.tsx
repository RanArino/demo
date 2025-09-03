import { ReactNode } from 'react';

interface SpaceLayoutProps {
  children: ReactNode;
  uploadModal: ReactNode;
  contentPreviewModal: ReactNode;
}

export default function SpaceLayout({
  children,
  uploadModal,
  contentPreviewModal,
}: SpaceLayoutProps) {
  return (
    <>
      {children}
      {uploadModal}
      {contentPreviewModal}
    </>
  );
}