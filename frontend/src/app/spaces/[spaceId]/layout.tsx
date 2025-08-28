import { ReactNode } from 'react';

interface SpaceLayoutProps {
  children: ReactNode;
  uploadModal: ReactNode;
}

export default function SpaceLayout({
  children,
  uploadModal,
}: SpaceLayoutProps) {
  return (
    <>
      {children}
      {uploadModal}
    </>
  );
}