'use client';

import { useState, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { ContentSource } from '@/app/spaces/types/content';

interface UseUploadModalOptions {
  spaceId: string;
  onClose?: () => void;
}

interface UseUploadModalReturn {
  isOpen: boolean;
  open: (autoOpenTab?: 'file' | 'google-drive' | 'link' | 'text') => void;
  close: () => void;
  openUploadModal: (autoOpenTab?: 'file' | 'google-drive' | 'link' | 'text') => void;
}

export function useUploadModal({
  spaceId,
  onClose
}: UseUploadModalOptions): UseUploadModalReturn {
  const [isOpen, setIsOpen] = useState(false);
  const router = useRouter();

  const open = useCallback((autoOpenTab?: 'file' | 'google-drive' | 'link' | 'text') => {
    setIsOpen(true);
    const searchParams = autoOpenTab ? `?tab=${autoOpenTab}` : '';
    router.push(`/spaces/${spaceId}/upload${searchParams}`);
  }, [spaceId, router]);

  const close = useCallback(() => {
    setIsOpen(false);
    onClose?.();
    router.back();
  }, [onClose, router]);

  const openUploadModal = useCallback((autoOpenTab?: 'file' | 'google-drive' | 'link' | 'text') => {
    open(autoOpenTab);
  }, [open]);

  return {
    isOpen,
    open,
    close,
    openUploadModal
  };
}