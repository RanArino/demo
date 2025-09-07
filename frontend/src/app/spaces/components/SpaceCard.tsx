import React from 'react';
import { Space, SpaceStats } from '@/api/generated/v1/knowledge_pb';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Pencil, Calendar, FileText, Users, Eye, Lock, Globe } from 'lucide-react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { EditSpaceForm } from './EditSpaceForm';
import Image from 'next/image';
import Link from 'next/link';
import { cn, formatSize } from '@/lib/utils';
import { safeTimestampToDate } from '@/lib/types';

// TODO: Move this to a shared types file
interface SpaceWithStats extends Space {
  stats?: SpaceStats;
}

interface SpaceCardProps {
  space: Space | SpaceWithStats;
  className?: string;
  onSelect?: (spaceId: string) => void;
  onEdit?: (spaceId: string) => void;
  onDelete?: (spaceId: string) => void;
  isDeleting?: boolean;
}

export default function SpaceCard({
  space,
  className,
  onSelect,
  onEdit,
  onDelete,
  isDeleting,
}: SpaceCardProps) {
  const stats = space.stats;
  
  const getAccessIcon = () => {
    // TODO: Implement access level logic based on new data model
    return <Globe className="h-3 w-3" />;
  };

  const getAccessColor = () => {
    // TODO: Implement access level logic based on new data model
    return 'bg-green-100 text-green-800 border-green-200';
  };

  const formatDate = (date: Date | string | undefined) => {
    if (!date) return '';
    const d = typeof date === 'string' ? new Date(date) : date;
    return d.toLocaleDateString();
  };

  const formatTimestamp = (timestamp: any) => {
    const date = safeTimestampToDate(timestamp);
    return date ? date.toLocaleDateString() : '';
  };

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <div className={cn(
            "bg-white border border-gray-200 rounded-xl shadow-sm hover:shadow-md transition-all duration-200 relative group overflow-hidden hover:-translate-y-1",
            isDeleting && "opacity-50 pointer-events-none",
            className
          )}>
            {/* Settings Button */}
            <div className="absolute top-3 right-3 z-10 opacity-0 group-hover:opacity-100 transition-opacity">
              <Dialog>
                <DialogTrigger asChild>
                  <Button 
                    variant="secondary" 
                    size="icon" 
                    className="h-8 w-8 bg-white/90 hover:bg-white shadow-sm"
                    onClick={(e) => e.stopPropagation()}
                  >
                    <Pencil className="h-4 w-4" />
                  </Button>
                </DialogTrigger>
                <DialogContent aria-describedby={undefined}>
                  <DialogHeader>
                    <DialogTitle>Edit Space</DialogTitle>
                  </DialogHeader>
                  <EditSpaceForm space={space} />
                </DialogContent>
              </Dialog>
            </div>

            {/* Access Level Badge */}
            <div className="absolute top-3 left-3 z-10">
              <Badge 
                variant="secondary" 
                className={`text-xs capitalize ${getAccessColor()} flex items-center gap-1`}
              >
                {getAccessIcon()}
                {/* {space.accessLevel} */}
              </Badge>
            </div>

            <div className="block cursor-pointer" onClick={() => onSelect?.(space.id)}>
              {/* Cover Image */}
              <div className="relative h-48 w-full">
                <div className="h-full w-full bg-gradient-to-br from-primary/20 to-primary/10 flex items-center justify-center transition-transform duration-200 group-hover:scale-105">
                  <span className="text-6xl">📚</span>
                </div>
                <div className="absolute inset-0 bg-gradient-to-t from-black/20 to-transparent" />
              </div>

              {/* Content */}
              <div className="p-4">
                {/* Title and Icon */}
                <div className="flex items-start justify-between mb-2">
                  <div className="flex items-center gap-2 min-w-0 flex-1">
                    <span className="text-2xl flex-shrink-0">📚</span>
                    <h3 className="font-semibold text-gray-900 truncate">{space.title}</h3>
                  </div>
                </div>

                {/* Description */}
                <p className="text-sm text-gray-600 line-clamp-2 mb-3 min-h-[2.5rem]">
                  {space.description || 'No description'}
                </p>

                {/* Keywords */}
                <div className="flex flex-wrap gap-1 mb-3 min-h-[1.5rem]">
                  {/* space.keywords.slice(0, 3).map((keyword) => (
                    <Badge key={keyword} variant="outline" className="text-xs text-gray-600">
                      {keyword}
                    </Badge>
                  )) */}
                </div>

                {/* Metadata */}
                <div className="space-y-2">
                  <div className="flex items-center justify-between text-xs text-gray-500">
                    <div className="flex items-center gap-1">
                      <Calendar className="h-3 w-3" />
                      <span>{formatTimestamp(space.createdAt)}</span>
                    </div>
                    <div className="flex items-center gap-1">
                      <FileText className="h-3 w-3" />
                      <span>{stats?.contentCount.toString() || 0} docs</span>
                    </div>
                  </div>
                  
                  <div className="text-xs text-gray-400">
                    Last updated: {formatTimestamp(space.updatedAt)}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </TooltipTrigger>
        <TooltipContent side="top" className="max-w-sm">
          <div className="space-y-2">
            <p className="font-medium">{space.title}</p>
            <p className="text-sm">{space.description}</p>
            {/* <div>
              <p className="text-xs font-medium mb-1">Keywords:</p>
              <p className="text-xs text-gray-300">{space.keywords.join(', ')}</p>
            </div> */}
            <div className="text-xs">
              {/* <p>Size: {formatSize(space.totalSizeBytes || 0)}</p> */}
            </div>
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}