'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { createSpace } from '@/api/actions/spaceActions';
import { Space, CreateSpaceInput } from '../types/spaces';
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
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState<CreateSpaceInput>({
    title: '',
    description: '',
    keywords: [],
    icon: '',
  });
  const [keywordInput, setKeywordInput] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!formData.title.trim()) {
      toast({
        title: 'Error',
        description: 'Please enter a space title',
        variant: 'destructive',
      });
      return;
    }

    setLoading(true);
    try {
      const result = await createSpace(formData);
      if (result.ok && result.data) {
        toast({
          title: 'Success',
          description: 'Space created successfully',
        });
        onSuccess(result.data);
        // Reset form
        setFormData({
          title: '',
          description: '',
          keywords: [],
          icon: '',
        });
        setKeywordInput('');
      } else {
        toast({
          title: 'Error',
          description: result.error?.message || 'Failed to create space',
          variant: 'destructive',
        });
      }
    } catch (error) {
      toast({
        title: 'Error',
        description: 'An unexpected error occurred',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  const handleAddKeyword = () => {
    const trimmed = keywordInput.trim();
    if (trimmed && !formData.keywords?.includes(trimmed)) {
      setFormData({
        ...formData,
        keywords: [...(formData.keywords || []), trimmed],
      });
      setKeywordInput('');
    }
  };

  const handleRemoveKeyword = (keyword: string) => {
    setFormData({
      ...formData,
      keywords: formData.keywords?.filter(k => k !== keyword),
    });
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

          {/* Icon */}
          <div>
            <Label htmlFor="icon">Icon (Emoji)</Label>
            <Input
              id="icon"
              value={formData.icon || ''}
              onChange={(e) => setFormData({ ...formData, icon: e.target.value })}
              placeholder="📚"
              disabled={loading}
              maxLength={2}
            />
          </div>

          {/* Keywords */}
          <div>
            <Label htmlFor="keywords">Keywords</Label>
            <div className="flex gap-2 mb-2">
              <Input
                id="keywords"
                value={keywordInput}
                onChange={(e) => setKeywordInput(e.target.value)}
                placeholder="Add a keyword"
                disabled={loading}
                onKeyPress={(e) => {
                  if (e.key === 'Enter') {
                    e.preventDefault();
                    handleAddKeyword();
                  }
                }}
              />
              <Button
                type="button"
                variant="outline"
                onClick={handleAddKeyword}
                disabled={loading}
              >
                Add
              </Button>
            </div>
            
            {/* Keywords list */}
            {formData.keywords && formData.keywords.length > 0 && (
              <div className="flex flex-wrap gap-2">
                {formData.keywords.map((keyword) => (
                  <span
                    key={keyword}
                    className="px-2 py-1 text-xs bg-secondary text-secondary-foreground rounded-md flex items-center gap-1"
                  >
                    {keyword}
                    <button
                      type="button"
                      onClick={() => handleRemoveKeyword(keyword)}
                      className="hover:text-destructive"
                      disabled={loading}
                    >
                      ×
                    </button>
                  </span>
                ))}
              </div>
            )}
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