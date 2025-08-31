'use client';

import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Skeleton } from '@/components/ui/skeleton';

interface UploadModalSkeletonProps {
  isOpen: boolean;
}

export default function UploadModalSkeleton({ isOpen }: UploadModalSkeletonProps) {
  return (
    <Dialog open={isOpen}>
      <DialogContent className="sm:max-w-[600px] max-h-[80vh] overflow-hidden flex flex-col">
        <DialogHeader>
          <DialogTitle className="sr-only">Uploading...</DialogTitle>
          <div className="flex items-center justify-between">
            <Skeleton className="h-6 w-48" />
            <Skeleton className="h-6 w-6" />
          </div>
        </DialogHeader>

        <div className="flex-1 overflow-hidden space-y-4">
          {/* Tabs skeleton */}
          <div className="grid w-full grid-cols-4 gap-2">
            {[...Array(4)].map((_, i) => (
              <Skeleton key={i} className="h-10" />
            ))}
          </div>

          {/* Content area skeleton */}
          <div className="space-y-4">
            <Skeleton className="h-32 w-full" />
            <div className="space-y-2">
              <Skeleton className="h-4 w-3/4" />
              <Skeleton className="h-4 w-1/2" />
            </div>
            <Skeleton className="h-10 w-full" />
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}