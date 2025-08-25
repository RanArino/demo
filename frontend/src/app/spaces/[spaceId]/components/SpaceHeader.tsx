'use client';

import { useState } from 'react';
import { Space } from '@/app/spaces/types/spaces';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { EditSpaceForm } from '@/app/spaces/components/EditSpaceForm';
import { ArrowLeft, Pencil, Share2, Trash2, Globe, Users, Lock, Eye } from 'lucide-react';
import { useRouter } from 'next/navigation';
import Image from 'next/image';
import { cn } from '@/lib/utils';

interface SpaceHeaderProps {
  space: Space;
  className?: string;
}

export default function SpaceHeader({ space, className }: SpaceHeaderProps) {
  const router = useRouter();
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);

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

  const handleEdit = () => {
    setIsEditDialogOpen(true);
  };

  const handleShare = () => {
    // TODO: Implement share functionality
    console.log('Share space:', space.id);
  };

  const handleDelete = () => {
    // TODO: Implement delete functionality with confirmation
    console.log('Delete space:', space.id);
  };

  const handleBack = () => {
    router.push('/spaces');
  };

  return (
    <div className={cn("bg-white border-b border-gray-200", className)}>
      {/* Header Actions */}
      <div className="flex items-center justify-between p-4 border-b border-gray-100">
        <Button
          variant="ghost"
          size="sm"
          onClick={handleBack}
          className="flex items-center gap-2 text-gray-600 hover:text-gray-900"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to Spaces
        </Button>

        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={handleShare}
            className="flex items-center gap-2"
          >
            <Share2 className="h-4 w-4" />
            Share
          </Button>
          
          <Dialog open={isEditDialogOpen} onOpenChange={setIsEditDialogOpen}>
            <DialogTrigger asChild>
              <Button
                variant="ghost"
                size="sm"
                onClick={handleEdit}
                className="flex items-center gap-2"
              >
                <Pencil className="h-4 w-4" />
                Edit
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Edit Space</DialogTitle>
              </DialogHeader>
              <EditSpaceForm 
                space={space} 
                onSuccess={() => setIsEditDialogOpen(false)}
              />
            </DialogContent>
          </Dialog>

          <Button
            variant="ghost"
            size="sm"
            onClick={handleDelete}
            className="flex items-center gap-2 text-red-600 hover:text-red-700 hover:bg-red-50"
          >
            <Trash2 className="h-4 w-4" />
            Delete
          </Button>
        </div>
      </div>

      {/* Space Info */}
      <div className="p-6">
        {/* Cover Image or Icon */}
        <div className="relative mb-4">
          {space.coverImage ? (
            <div className="relative h-32 w-full rounded-lg overflow-hidden">
              <Image
                src={space.coverImage}
                alt={space.title}
                fill
                className="object-cover"
              />
              <div className="absolute inset-0 bg-gradient-to-t from-black/20 to-transparent" />
            </div>
          ) : (
            <div className="h-32 w-full bg-gradient-to-br from-primary/20 to-primary/10 rounded-lg flex items-center justify-center">
              <span className="text-6xl">{space.icon || '📚'}</span>
            </div>
          )}
          
          {/* Access Level Badge */}
          <div className="absolute top-3 right-3">
            <Badge 
              variant="secondary" 
              className={`text-xs capitalize ${getAccessColor()} flex items-center gap-1`}
            >
              {getAccessIcon()}
              {space.accessLevel}
            </Badge>
          </div>
        </div>

        {/* Title and Description */}
        <div className="space-y-3">
          <div className="flex items-start gap-3">
            <span className="text-3xl flex-shrink-0">{space.icon || '📚'}</span>
            <div className="min-w-0 flex-1">
              <h1 className="text-2xl font-bold text-gray-900 break-words">
                {space.title}
              </h1>
            </div>
          </div>

          {space.description && (
            <p className="text-gray-600 leading-relaxed">
              {space.description}
            </p>
          )}
        </div>
      </div>
    </div>
  );
}