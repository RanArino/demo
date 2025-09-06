'use server';

import { auth } from '@clerk/nextjs/server';
import { ConnectError } from '@bufbuild/connect';

/**
 * Helper function to create headers with JWT token
 */
export async function createAuthHeaders(): Promise<Headers> {
  const { getToken } = await auth();
  const token = await getToken({ template: 'ms-user-auth' });

  const headers = new Headers();
  if (token) {
    headers.append('authorization', `Bearer ${token}`);
  }

  return headers;
}

/**
 * Helper function to sanitize error messages for security
 */
export function sanitizeError(error: unknown): { code: string; message: string } {
  if (error instanceof ConnectError) {
    return { code: error.code.toString(), message: error.message };
  }

  const isDevelopment = process.env.NODE_ENV === 'development';
  const message = error instanceof Error ? error.message : 'An unexpected error occurred';

  return {
    code: 'INTERNAL',
    message: isDevelopment ? message : 'An unexpected error occurred. Please try again.',
  };
}

/**
 * Helper function to sanitize error messages for security (string version)
 */
export function sanitizeErrorString(error: unknown): string {
  if (error instanceof ConnectError) {
    return error.message;
  }

  const isDevelopment = process.env.NODE_ENV === 'development';
  
  if (isDevelopment) {
    return error instanceof Error ? error.message : 'An error occurred';
  } else {
    return 'An error occurred. Please try again.';
  }
}
