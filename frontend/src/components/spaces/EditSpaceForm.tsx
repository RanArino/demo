'use client';

import { useActionState } from 'react';
import { updateSpace } from '@/app/spaces/actions';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Space } from '@/lib/types';

type FormState = {
  errors: {
    title?: string[];
    description?: string[];
    keywords?: string[];
  };
};

interface EditSpaceFormProps {
  space: Space;
}

export function EditSpaceForm({ space }: EditSpaceFormProps) {
  const initialState: FormState = { errors: {} };
  const updateSpaceWithId = (prevState: unknown, formData: FormData) => updateSpace(space.id, prevState, formData);
  const [state, dispatch] = useActionState(updateSpaceWithId, initialState);

  return (
    <form action={dispatch}>
      <div className="grid gap-4 py-4">
        <div className="grid grid-cols-4 items-center gap-4">
          <label htmlFor="title" className="text-right">Title</label>
          <Input id="title" name="title" defaultValue={space.title} className="col-span-3" />
          {state.errors?.title && <p className="col-span-4 text-red-500 text-sm">{state.errors.title}</p>}
        </div>
        <div className="grid grid-cols-4 items-center gap-4">
          <label htmlFor="description" className="text-right">Description</label>
          <Input id="description" name="description" defaultValue={space.description} className="col-span-3" />
        </div>
        <div className="grid grid-cols-4 items-center gap-4">
          <label htmlFor="keywords" className="text-right">Keywords</label>
          <Input id="keywords" name="keywords" defaultValue={space.keywords.join(', ')} className="col-span-3" />
        </div>
      </div>
      <Button type="submit">Save Changes</Button>
    </form>
  );
}