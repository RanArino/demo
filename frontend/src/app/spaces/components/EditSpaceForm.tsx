import { Space } from '@/api/generated/v1/knowledge_pb';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

interface EditSpaceFormProps {
  space: Space;
  onSuccess?: () => void;
}

export function EditSpaceForm({ space, onSuccess }: EditSpaceFormProps) {
  return (
    <div className="space-y-4">
      <div>
        <Label htmlFor="title">Title</Label>
        <Input id="title" defaultValue={space.title} />
      </div>
      <div>
        <Label htmlFor="description">Description</Label>
        <Input id="description" defaultValue={space.description} />
      </div>
      <div className="flex justify-end gap-2">
        <Button variant="outline">Cancel</Button>
        <Button onClick={() => onSuccess?.()}>Save Changes</Button>
      </div>
    </div>
  );
}