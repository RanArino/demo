'use client';

import { useActionState } from 'react';
import { createSpace } from '@/app/spaces/actions';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';

type FormState = {
  errors: {
    title?: string[];
    description?: string[];
    keywords?: string[];
  };
};

export function CreateSpaceForm() {
  const initialState: FormState = { errors: {} };
  const [state, dispatch] = useActionState(createSpace, initialState);

  return (
    <form action={dispatch}>
      <div className="grid gap-4 py-4">
        <div className="grid grid-cols-4 items-center gap-4">
          <label htmlFor="title" className="text-right">Title</label>
          <Input id="title" name="title" className="col-span-3" />
          {state.errors?.title && <p className="col-span-4 text-red-500 text-sm">{state.errors.title}</p>}
        </div>
        <div className="grid grid-cols-4 items-center gap-4">
          <label htmlFor="description" className="text-right">Description</label>
          <Input id="description" name="description" className="col-span-3" />
        </div>
        <div className="grid grid-cols-4 items-center gap-4">
          <label htmlFor="keywords" className="text-right">Keywords</label>
          <Input id="keywords" name="keywords" className="col-span-3" />
        </div>
      </div>
      <Button type="submit">Create Space</Button>
    </form>
  );
}