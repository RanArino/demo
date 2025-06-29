
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Pencil, Calendar, FileText, Users, Eye, Lock, Globe } from 'lucide-react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { EditSpaceForm } from './EditSpaceForm';
import { Space } from '@/lib/types';
import Image from 'next/image';
import Link from 'next/link';

interface SpaceCardProps {
  space: Space;
}

export function SpaceCard({ space }: SpaceCardProps) {
  const getAccessIcon = () => {
    switch (space.access_level) {
      case 'public': return <Globe className="h-3 w-3" />;
      case 'shared': return <Users className="h-3 w-3" />;
      case 'private': return <Lock className="h-3 w-3" />;
      default: return <Eye className="h-3 w-3" />;
    }
  };

  const getAccessColor = () => {
    switch (space.access_level) {
      case 'public': return 'bg-green-100 text-green-800 border-green-200';
      case 'shared': return 'bg-blue-100 text-blue-800 border-blue-200';
      case 'private': return 'bg-gray-100 text-gray-800 border-gray-200';
      default: return 'bg-gray-100 text-gray-800 border-gray-200';
    }
  };

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <div className="bg-white border border-gray-200 rounded-xl shadow-sm hover:shadow-md transition-all duration-200 relative group overflow-hidden">
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
                <DialogContent>
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
                {space.access_level}
              </Badge>
            </div>

            <Link href={`/spaces/${space.id}`} className="block">
              {/* Cover Image */}
              <div className="relative h-48 w-full">
                <Image
                  src={space.cover_image}
                  alt={space.title}
                  fill
                  className="object-cover transition-transform duration-200 group-hover:scale-105"
                />
                <div className="absolute inset-0 bg-gradient-to-t from-black/20 to-transparent" />
              </div>

              {/* Content */}
              <div className="p-4">
                {/* Title and Icon */}
                <div className="flex items-start justify-between mb-2">
                  <div className="flex items-center gap-2 min-w-0 flex-1">
                    <span className="text-2xl flex-shrink-0">{space.icon}</span>
                    <h3 className="font-semibold text-gray-900 truncate">{space.title}</h3>
                  </div>
                </div>

                {/* Description */}
                <p className="text-sm text-gray-600 line-clamp-2 mb-3 min-h-[2.5rem]">
                  {space.description}
                </p>

                {/* Keywords */}
                <div className="flex flex-wrap gap-1 mb-3 min-h-[1.5rem]">
                  {space.keywords.slice(0, 3).map((keyword) => (
                    <Badge key={keyword} variant="outline" className="text-xs text-gray-600">
                      {keyword}
                    </Badge>
                  ))}
                  {space.keywords.length > 3 && (
                    <Badge variant="outline" className="text-xs text-gray-400">
                      +{space.keywords.length - 3}
                    </Badge>
                  )}
                </div>

                {/* Metadata */}
                <div className="space-y-2">
                  <div className="flex items-center justify-between text-xs text-gray-500">
                    <div className="flex items-center gap-1">
                      <Calendar className="h-3 w-3" />
                      <span>{new Date(space.created_at).toLocaleDateString()}</span>
                    </div>
                    <div className="flex items-center gap-1">
                      <FileText className="h-3 w-3" />
                      <span>{space.document_count} docs</span>
                    </div>
                  </div>
                  
                  <div className="text-xs text-gray-400">
                    Last updated: {new Date(space.last_updated_at).toLocaleDateString()}
                  </div>
                </div>
              </div>
            </Link>
          </div>
        </TooltipTrigger>
        <TooltipContent side="top" className="max-w-sm">
          <div className="space-y-2">
            <p className="font-medium">{space.title}</p>
            <p className="text-sm">{space.description}</p>
            <div>
              <p className="text-xs font-medium mb-1">Keywords:</p>
              <p className="text-xs text-gray-300">{space.keywords.join(', ')}</p>
            </div>
            <div className="text-xs">
              <p>Size: {(space.total_size_bytes / (1024 * 1024)).toFixed(1)} MB</p>
            </div>
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
