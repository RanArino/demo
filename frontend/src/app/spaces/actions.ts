'use server';

import { revalidatePath } from 'next/cache';
import { z } from 'zod';

const spaceSchema = z.object({
  title: z.string().min(1, 'Title is required'),
  description: z.string().optional(),
  keywords: z.array(z.string()).optional(),
});

export async function createSpace(prevState: unknown, formData: FormData) {
  const validatedFields = spaceSchema.safeParse({
    title: formData.get('title'),
    description: formData.get('description'),
    keywords: formData.getAll('keywords'),
  });

  if (!validatedFields.success) {
    return {
      errors: validatedFields.error.flatten().fieldErrors,
    };
  }

  // Here you would typically call your backend API to create the space
  console.log('Creating space:', validatedFields.data);

  revalidatePath('/spaces');
  return { errors: {} };
}

export async function updateSpace(spaceId: string, prevState: unknown, formData: FormData) {
  const validatedFields = spaceSchema.partial().safeParse({
    title: formData.get('title'),
    description: formData.get('description'),
    keywords: formData.getAll('keywords'),
  });

  if (!validatedFields.success) {
    return {
      errors: validatedFields.error.flatten().fieldErrors,
    };
  }

  // Here you would typically call your backend API to update the space
  console.log(`Updating space ${spaceId}:`, validatedFields.data);

  revalidatePath('/spaces');
  return { errors: {} };
}

export async function deleteSpace(spaceId: string) {
  // Here you would typically call your backend API to delete the space
  console.log(`Deleting space ${spaceId}`);

  revalidatePath('/spaces');
}