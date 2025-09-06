'use client';

import { useState, useCallback } from 'react';
import { useRouter } from 'next/navigation';
interface UseUploadModalOptions {
  spaceId: string;
  onClose?: () => void;
}

interface UseUploadModalReturn {
  isOpen: boolean;
  open: (autoOpenTab?: 'file' | 'google-drive' | 'link' | 'text', objectKind?: 'original' | 'processed') => void;
  close: () => void;
  openUploadModal: (autoOpenTab?: 'file' | 'google-drive' | 'link' | 'text', objectKind?: 'original' | 'processed') => void;
}

export function useUploadModal({
  spaceId,
  onClose
}: UseUploadModalOptions): UseUploadModalReturn {
  const [isOpen, setIsOpen] = useState(false);
  const router = useRouter();

  const open = useCallback((autoOpenTab?: 'file' | 'google-drive' | 'link' | 'text', objectKind?: 'original' | 'processed') => {
    setIsOpen(true);
    const params = new URLSearchParams();
    if (autoOpenTab) params.set('tab', autoOpenTab);
    if (objectKind) params.set('objectKind', objectKind);
    const searchParams = params.toString() ? `?${params.toString()}` : '';
    router.push(`/spaces/${spaceId}/upload${searchParams}`);
  }, [spaceId, router]);

  const close = useCallback(() => {
    setIsOpen(false);
    onClose?.();
    router.back();
  }, [onClose, router]);

  const openUploadModal = useCallback((autoOpenTab?: 'file' | 'google-drive' | 'link' | 'text', objectKind?: 'original' | 'processed') => {
    open(autoOpenTab, objectKind);
  }, [open]);

  return {
    isOpen,
    open,
    close,
    openUploadModal
  };
}