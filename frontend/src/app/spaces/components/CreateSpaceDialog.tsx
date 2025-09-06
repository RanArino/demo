'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { createSpace } from '@/api/actions/spaceActions';
import { Space, CreateSpaceRequest } from '@/api/generated/v1/knowledge_pb';
import { useToast } from '@/components/ui/use-toast';

interface CreateSpaceDialogProps {
  open: boolean;
  onClose: () => void;
  onSuccess: (space: Space) => void;
}

export default function CreateSpaceDialog({
  open,
  onClose,
  onSuccess,
}: CreateSpaceDialogProps) {
  const { toast } = useToast();
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState<Partial<CreateSpaceRequest>>({
    title: '',
    description: '',
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!formData.title?.trim()) {
      toast({
        title: 'Error',
        description: 'Please enter a space title',
        variant: 'destructive',
      });
      return;
    }

    setLoading(true);
    try {
      const result = await createSpace(formData as CreateSpaceRequest);
      if (result.ok && result.data) {
        toast({
          title: 'Success',
          description: 'Space created successfully',
        });
        // Reset form
        setFormData({
          title: '',
          description: '',
        });
        
        // Call the success callback first
        onSuccess(result.data);
        
        // Navigate to the space page
        router.push(`/spaces/${result.data.id}`);
      } else {
        toast({
          title: 'Error',
          description: result.error?.message || 'Failed to create space',
          variant: 'destructive',
        });
      }
    } catch (err) {
      console.error('Space creation error:', err);
      toast({
        title: 'Error',
        description: 'An unexpected error occurred',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div 
        className="absolute inset-0 bg-background/80 backdrop-blur-sm"
        onClick={onClose}
      />
      
      {/* Dialog */}
      <div className="relative bg-card border rounded-lg shadow-lg w-full max-w-md p-6">
        <h2 className="text-2xl font-semibold mb-4">Create New Space</h2>
        
        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Title */}
          <div>
            <Label htmlFor="title">Title *</Label>
            <Input
              id="title"
              value={formData.title}
              onChange={(e) => setFormData({ ...formData, title: e.target.value })}
              placeholder="Enter space title"
              disabled={loading}
              required
            />
          </div>

          {/* Description */}
          <div>
            <Label htmlFor="description">Description</Label>
            <textarea
              id="description"
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              placeholder="Enter space description"
              className="w-full px-3 py-2 rounded-md border border-input bg-background text-sm min-h-[80px] resize-none"
              disabled={loading}
            />
          </div>

          {/* Actions */}
          <div className="flex gap-2 pt-4">
            <Button
              type="button"
              variant="outline"
              onClick={onClose}
              disabled={loading}
              className="flex-1"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={loading}
              className="flex-1"
            >
              {loading ? 'Creating...' : 'Create Space'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}