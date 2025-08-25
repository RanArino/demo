'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Space } from '@/app/spaces/types/spaces';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { EditSpaceForm } from '@/app/spaces/components/EditSpaceForm';
import ChatHistorySection from './ChatHistorySection';
import { ArrowLeft, Globe, Users, Lock, Eye, Settings, Share2, Pin, PinOff } from 'lucide-react';
import { cn } from '@/lib/utils';
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

  const getAccessIcon = () => {
    switch (space.accessLevel) {
      case 'public': return <Globe className="h-4 w-4" />;
      case 'shared': return <Users className="h-4 w-4" />;
      case 'private': return <Lock className="h-4 w-4" />;
      default: return <Eye className="h-4 w-4" />;
    }
  };

  const getAccessColor = () => {
    switch (space.accessLevel) {
      case 'public': return 'bg-green-100 text-green-800 border-green-200';
      case 'shared': return 'bg-blue-100 text-blue-800 border-blue-200';
      case 'private': return 'bg-gray-100 text-gray-800 border-gray-200';
      default: return 'bg-gray-100 text-gray-800 border-gray-200';
    }
  };

  const handleBack = () => {
    router.push('/spaces');
  };

  const handleShare = () => {
    // TODO: Implement share functionality
    console.log('Share space:', space.id);
  };

  const togglePin = () => {
    setIsPinned(!isPinned);
    if (!isPinned) {
      setIsVisible(true);
    }
  };

  return (
    <>
      {/* Hover trigger zone - only when not pinned */}
      {!isPinned && (
        <div 
          className="fixed left-0 top-0 w-8 h-full z-40"
          onMouseEnter={() => setIsVisible(true)}
        />
      )}

      {/* Backdrop - only when visible and not pinned */}
      {isVisible && !isPinned && (
        <div 
          className="fixed inset-0 bg-black bg-opacity-10 z-40 transition-opacity duration-200"
          onClick={() => setIsVisible(false)}
        />
      )}

      {/* Sidebar */}
      <div
        className={cn(
          "fixed left-0 top-0 h-full w-80 bg-gray-50 shadow-xl z-50",
          "transform transition-transform duration-300 ease-out",
          "flex flex-col",
          (isVisible || isPinned) ? "translate-x-0" : "-translate-x-full",
          className
        )}
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
              >
                <Settings className="h-4 w-4" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                onClick={handleShare}
                className="h-8 w-8"
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
              >
                {isPinned ? <PinOff className="h-4 w-4" /> : <Pin className="h-4 w-4" />}
              </Button>
            </div>
          </div>
        </div>

        {/* Scrollable Content */}
        <div className="flex-1 overflow-y-auto">
          {/* Space Cover Image */}
          {space.coverImage && (
            <div className="relative h-48 w-full bg-white">
              <Image
                src={space.coverImage}
                alt={space.title}
                fill
                className="object-cover"
              />
              {/* Access Badge on Image */}
              <div className="absolute top-3 right-3">
                <Badge 
                  variant="secondary" 
                  className={cn(
                    "text-xs capitalize flex items-center gap-1 backdrop-blur-sm",
                    getAccessColor()
                  )}
                >
                  {getAccessIcon()}
                  {space.accessLevel}
                </Badge>
              </div>
            </div>
          )}
          
          {/* Space Info */}
          <div className="p-6 bg-white border-b border-gray-200">
            {/* Space Icon and Title */}
            <div className="flex items-start gap-3 mb-4">
              <span className="text-3xl flex-shrink-0">{space.icon || '📚'}</span>
              <div className="min-w-0 flex-1">
                <h1 className="text-xl font-bold text-gray-900 break-words">
                  {space.title}
                </h1>
                {!space.coverImage && (
                  <Badge 
                    variant="secondary" 
                    className={cn(
                      "text-xs capitalize mt-2 inline-flex items-center gap-1",
                      getAccessColor()
                    )}
                  >
                    {getAccessIcon()}
                    {space.accessLevel}
                  </Badge>
                )}
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
                {new Date(space.createdAt).toLocaleDateString()}
              </div>
              <div>
                <span className="font-medium">Updated:</span>{' '}
                {new Date(space.lastUpdatedAt).toLocaleDateString()}
              </div>
            </div>

            {/* Document Statistics */}
            <div className="grid grid-cols-2 gap-3 mb-4">
              <div className="bg-gray-50 rounded-lg p-3">
                <div className="text-xl font-bold text-gray-900">
                  {space.documentCount || 0}
                </div>
                <div className="text-xs text-gray-600">Documents</div>
              </div>
              <div className="bg-gray-50 rounded-lg p-3">
                <div className="text-xl font-bold text-gray-900">
                  {((space.totalSizeBytes || 0) / (1024 * 1024)).toFixed(1)}
                </div>
                <div className="text-xs text-gray-600">MB Used</div>
              </div>
            </div>

            {/* Keywords */}
            {space.keywords && space.keywords.length > 0 && (
              <div>
                <h3 className="text-xs font-medium text-gray-900 mb-2">Keywords</h3>
                <div className="flex flex-wrap gap-1">
                  {space.keywords.map((keyword) => (
                    <span
                      key={keyword}
                      className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800"
                    >
                      {keyword}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>
          
          {/* Chat History Section */}
          <ChatHistorySection spaceId={space.id} className="border-t border-gray-200" />
        </div>
      </div>

      {/* Settings Dialog */}
      <Dialog open={isSettingsDialogOpen} onOpenChange={setIsSettingsDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Space Settings</DialogTitle>
          </DialogHeader>
          <EditSpaceForm 
            space={space} 
            onSuccess={() => setIsSettingsDialogOpen(false)}
          />
        </DialogContent>
      </Dialog>
    </>
  );
}