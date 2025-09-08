'use client';

import { useEffect, useRef, useState } from 'react';
import { useRouter } from 'next/navigation';
import { Space } from '@/api/generated/v1/knowledge_pb';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { EditSpaceForm } from '@/app/spaces/components/EditSpaceForm';
import { AccessLevelModal } from '@/app/spaces/components/AccessLevelModal';
import ChatHistorySection from './ChatHistorySection';
import { ArrowLeft, Settings, Share2, PanelLeftOpen, PanelRightOpen } from 'lucide-react';

import { cn, formatSize } from '@/lib/utils';
import { safeTimestampToDate } from '@/lib/types';
import Image from 'next/image';

interface LeftSidebarProps {
  space: Space;
  className?: string;
}

export default function LeftSidebar({ space, className }: LeftSidebarProps) {
  const router = useRouter();
  const [isVisible, setIsVisible] = useState(false);
  const [isPinned, setIsPinned] = useState(false);
  const [isSettingsDialogOpen, setIsSettingsDialogOpen] = useState(false);
  const [isAccessModalOpen, setIsAccessModalOpen] = useState(false);
  const [sidebarWidthPx, setSidebarWidthPx] = useState<number>(320);
  const isResizingRef = useRef(false);
  const sidebarWidthRef = useRef<number>(320);


  const handleBack = () => {
    router.push('/spaces');
  };

  const handleShare = () => {
    setIsAccessModalOpen(true);
  };

  // Load persisted width and pin state on mount
  useEffect(() => {
    try {
      const savedWidth = localStorage.getItem('spaceSidebarWidthPx');
      if (savedWidth) {
        const parsed = parseInt(savedWidth, 10);
        if (!Number.isNaN(parsed)) {
          setSidebarWidthPx(Math.min(Math.max(parsed, 240), 560));
        }
      }
      const savedPinned = localStorage.getItem('spaceSidebarPinned');
      if (savedPinned === 'true') {
        setIsPinned(true);
        setIsVisible(true);
      }
    } catch {
      // ignore storage errors
    }
  }, []);

  // Sync CSS variables for layout adjustment
  useEffect(() => {
    // Sidebar width variable (always useful for transitions)
    document.documentElement.style.setProperty('--left-sidebar-width', `${sidebarWidthPx}px`);
    // Sidebar offset only when pinned
    const offset = isPinned ? `${sidebarWidthPx}px` : '0px';
    document.documentElement.style.setProperty('--sidebar-offset', offset);
    sidebarWidthRef.current = sidebarWidthPx;
  }, [sidebarWidthPx, isPinned]);

  const togglePin = () => {
    const nextPinned = !isPinned;
    setIsPinned(nextPinned);
    try {
      localStorage.setItem('spaceSidebarPinned', String(nextPinned));
    } catch {
      // ignore storage errors
    }
    if (nextPinned) {
      setIsVisible(true);
    }
  };

  const handleResizeStart = (e: React.MouseEvent<HTMLDivElement>) => {
    e.preventDefault();
    isResizingRef.current = true;
    const onMouseMove = (ev: MouseEvent) => {
      if (!isResizingRef.current) return;
      const minWidth = 240;
      const maxWidth = 560;
      const newWidth = Math.min(Math.max(ev.clientX, minWidth), maxWidth);
      setSidebarWidthPx(newWidth);
    };
    const onMouseUp = () => {
      isResizingRef.current = false;
      window.removeEventListener('mousemove', onMouseMove);
      window.removeEventListener('mouseup', onMouseUp);
      try {
        localStorage.setItem('spaceSidebarWidthPx', String(sidebarWidthRef.current));
      } catch {
        // ignore storage errors
      }
    };
    window.addEventListener('mousemove', onMouseMove);
    window.addEventListener('mouseup', onMouseUp);
  };

  return (
    <>
      {/* Sidebar display button - only when not visible and not pinned */}
      {!isPinned && !isVisible && (
        <Button
          variant="ghost"
          size="icon"
          className="fixed left-4 top-4 z-50 h-8 w-8 bg-white/80 backdrop-blur-sm rounded-full shadow-md hover:bg-white"
          onClick={() => setIsVisible(true)}
          aria-label="Open sidebar"
        >
          <PanelRightOpen className="h-4 w-4" />
        </Button>
      )}

      {/* Hover trigger zone - only when not pinned */}
      {!isPinned && (
        <div 
          className="fixed left-0 top-0 w-8 h-full z-40"
          onMouseEnter={() => setIsVisible(true)}
        />
      )}

      {/* Backdrop - only when visible and not pinned (transparent overlay) */}
      {isVisible && !isPinned && (
        <div 
          className="fixed inset-0 bg-black/10 z-40 transition-opacity duration-200"
          onClick={() => setIsVisible(false)}
        />
      )}

      {/* Sidebar */}
      <div
        className={cn(
          "fixed left-0 top-0 h-full bg-gray-50 shadow-xl z-50",
          "transform transition-transform duration-300 ease-out",
          "flex flex-col",
          (isVisible || isPinned) ? "translate-x-0" : "-translate-x-full",
          className
        )}
        style={{ width: sidebarWidthPx }}
        onMouseLeave={() => !isPinned && setIsVisible(false)}
      >
        {/* Sidebar Header with Back Button and Actions */}
        <div className="p-4 border-b border-gray-200 bg-white">
          <div className="flex items-center justify-between">
            {/* Back Button */}
            <Button
              variant="ghost"
              size="sm"
              onClick={handleBack}
              className="flex items-center gap-2 text-gray-600 hover:text-gray-900"
            >
              <ArrowLeft className="h-4 w-4" />
              Back to Spaces
            </Button>
            
            {/* Action Buttons */}
            <div className="flex items-center gap-1">
              <Button
                variant="ghost"
                size="icon"
                onClick={() => setIsSettingsDialogOpen(true)}
                className="h-8 w-8"
                aria-label="Open space settings"
              >
                <Settings className="h-4 w-4" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                onClick={handleShare}
                className="h-8 w-8"
                aria-label="Share space"
              >
                <Share2 className="h-4 w-4" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                onClick={togglePin}
                className={cn(
                  "h-8 w-8",
                  isPinned && "bg-blue-50 text-blue-600"
                )}
                aria-label={isPinned ? "Unpin sidebar" : "Pin sidebar"}
              >
                {isPinned ? <PanelLeftOpen className="h-4 w-4" /> : <PanelRightOpen className="h-4 w-4" />}
              </Button>
            </div>
          </div>
        </div>

        {/* Scrollable Content */}
        <div className="flex-1 overflow-y-auto">
          {/* Space Cover Image */}
          <div className="relative h-48 w-full bg-white">
          </div>
          
          {/* Space Info */}
          <div className="p-6 bg-white border-b border-gray-200">
            {/* Space Icon and Title */}
            <div className="flex items-start gap-3 mb-4">
              <span className="text-3xl flex-shrink-0">📚</span>
              <div className="min-w-0 flex-1">
                <h1 className="text-xl font-bold text-gray-900 break-words">
                  {space.title}
                </h1>
              </div>
            </div>

          {/* Description */}
          {space.description && (
            <p className="text-gray-600 text-sm leading-relaxed mb-6">
              {space.description}
            </p>
          )}

            {/* Timestamps */}
            <div className="flex flex-col gap-1 text-xs text-gray-600 mb-4">
              <div>
                <span className="font-medium">Created:</span>{' '}
                {safeTimestampToDate(space.createdAt)?.toLocaleDateString() || 'Unknown'}
              </div>
              <div>
                <span className="font-medium">Updated:</span>{' '}
                {safeTimestampToDate(space.updatedAt)?.toLocaleDateString() || 'Unknown'}
              </div>
            </div>

            {/* Document Statistics */}
            <div className="grid grid-cols-2 gap-3 mb-4">
              <div className="bg-gray-50 rounded-lg p-3">
                <div className="text-xl font-bold text-gray-900">
                  {space.stats?.contentCount.toString() || 0}
                </div>
                <div className="text-xs text-gray-600">Documents</div>
              </div>
              <div className="bg-gray-50 rounded-lg p-3">
                <div className="text-xl font-bold text-gray-900">
                  {formatSize(Number(space.totalSizeBytes) || 0)}
                </div>
                <div className="text-xs text-gray-600">Storage Used</div>
              </div>
            </div>

            {/* Keywords */}
            <div>
              <h3 className="text-xs font-medium text-gray-900 mb-2">Keywords</h3>
              <div className="flex flex-wrap gap-1">
                {/* {space.keywords.map((keyword) => (
                  <span
                    key={keyword}
                    className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800"
                  >
                    {keyword}
                  </span>
                ))} */}
              </div>
            </div>
          </div>
          
          {/* Chat History Section */}
          <ChatHistorySection spaceId={space.id} className="border-t border-gray-200" />
        </div>
      </div>

      {/* Resize Handle */}
      {(isVisible || isPinned) && (
        <div
          onMouseDown={handleResizeStart}
          className="fixed top-0 z-50 h-full"
          style={{
            left: sidebarWidthPx - 3,
            width: 6,
            cursor: 'col-resize'
          }}
        >
          <div className="w-full h-full" />
        </div>
      )}

      {/* Settings Dialog */}
      <Dialog open={isSettingsDialogOpen} onOpenChange={setIsSettingsDialogOpen}>
        <DialogContent aria-describedby={undefined}>
          <DialogHeader>
            <DialogTitle>Space Settings</DialogTitle>
          </DialogHeader>
          <EditSpaceForm 
            space={space} 
            onSuccess={() => setIsSettingsDialogOpen(false)}
          />
        </DialogContent>
      </Dialog>

      {/* Access Level Modal */}
      <AccessLevelModal
        space={space}
        open={isAccessModalOpen}
        onOpenChange={setIsAccessModalOpen}
      />
    </>
  );
}