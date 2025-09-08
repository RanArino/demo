'use client';

import React, { useState } from 'react';
import { Space, UpdateSpaceRequest } from '@/api/generated/v1/knowledge_pb';
import { updateSpace } from '@/api/actions/spaceActions';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Lock, Globe, Users } from 'lucide-react';

interface AccessLevelModalProps {
  space: Space;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

const ACCESS_LEVELS = [
  {
    value: 'private',
    label: 'Private',
    description: 'Only you can access this space',
    icon: Lock,
  },
  {
    value: 'public',
    label: 'Public',
    description: 'Anyone can view and access this space',
    icon: Globe,
  },
  {
    value: 'shared',
    label: 'Shared',
    description: 'People you invite can access this space',
    icon: Users,
  },
];

export function AccessLevelModal({ space, open, onOpenChange }: AccessLevelModalProps) {
  const [selectedLevel, setSelectedLevel] = useState(space.accessLevel || 'private');
  const [isSaving, setIsSaving] = useState(false);

  const handleSave = async () => {
    if (selectedLevel === space.accessLevel) {
      onOpenChange(false);
      return;
    }

    setIsSaving(true);
    try {
      const updateRequest = new UpdateSpaceRequest({
        id: space.id,
        title: space.title,
        description: space.description,
        keywords: space.keywords || [],
        icon: space.icon || '',
        accessLevel: selectedLevel,
      });

      const result = await updateSpace(space.id, updateRequest);
      if (result.ok) {
        onOpenChange(false);
        // TODO: Show success toast/notification
      } else {
        console.error('Failed to update access level:', result.error);
        // TODO: Show error toast/notification
      }
    } catch (error) {
      console.error('Failed to update access level:', error);
      // TODO: Show error toast/notification
    } finally {
      setIsSaving(false);
    }
  };

  const handleOpenChange = (newOpen: boolean) => {
    if (!newOpen) {
      // Reset to current value when closing
      setSelectedLevel(space.accessLevel || 'private');
    }
    onOpenChange(newOpen);
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Change Access Level</DialogTitle>
          <DialogDescription>
            Control who can access "{space.title}". Choose the appropriate access level for your space.
          </DialogDescription>
        </DialogHeader>

        <div className="py-4 space-y-4">
          <div>
            <Label htmlFor="access-level">Access Level</Label>
            <Select value={selectedLevel} onValueChange={setSelectedLevel}>
              <SelectTrigger className="w-full mt-2">
                <SelectValue placeholder="Select access level" />
              </SelectTrigger>
              <SelectContent className="bg-white">
                {ACCESS_LEVELS.map((level) => {
                  const IconComponent = level.icon;
                  return (
                    <SelectItem 
                      key={level.value} 
                      value={level.value}
                      className="bg-white hover:bg-gray-50 focus:bg-gray-50"
                    >
                      <div className="flex items-center gap-2">
                        <IconComponent className="h-4 w-4" />
                        {level.label}
                      </div>
                    </SelectItem>
                  );
                })}
              </SelectContent>
            </Select>
          </div>

          {/* Show description for selected level */}
          {selectedLevel && (
            <div className="p-3 rounded-lg bg-gray-50 border">
              <p className="text-sm text-gray-600">
                {ACCESS_LEVELS.find(level => level.value === selectedLevel)?.description}
              </p>
            </div>
          )}
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => handleOpenChange(false)}
            disabled={isSaving}
          >
            Cancel
          </Button>
          <Button
            onClick={handleSave}
            disabled={isSaving || selectedLevel === space.accessLevel}
          >
            {isSaving ? 'Saving...' : 'Save Changes'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}