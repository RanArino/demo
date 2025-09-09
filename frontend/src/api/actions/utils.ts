import { auth } from '@clerk/nextjs/server';
import { ConnectError } from '@bufbuild/connect';
import { SpaceFilters } from '../generated/v1/knowledge_pb';

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

/**
 * Utility: detect likely unauthorized/auth-related errors
 */
export function isUnauthorizedError(error: unknown): boolean {
  try {
    if (error instanceof ConnectError) {
      const codeStr = error.code?.toString?.().toUpperCase?.() || '';
      const msg = (error.message || '').toUpperCase();
      return (
        codeStr.includes('UNAUTH') ||
        codeStr.includes('PERMISSION') ||
        msg.includes('UNAUTH') ||
        msg.includes('PERMISSION')
      );
    }
  } catch {}
  return false;
}

/**
 * Utility: log auth failures with minimal, non-sensitive info
 */
export function logAuthFailure(context: string, error: unknown): void {
  try {
    const message = sanitizeErrorString(error);
    console.warn(`[auth-failure] ${context}: ${message}`);
  } catch {}
}

/**
 * Normalize SpaceFilters to handle undefined/null values consistently
 * This ensures stable cache key generation regardless of input variations
 */
export function normalizeFilters(filters?: SpaceFilters): SpaceFilters {
  return new SpaceFilters({
    q: filters?.q || '',
    keywords: filters?.keywords || [],
    // Add default values for any additional filter properties as they are added
  });
}

/**
 * Generic protobuf sanitizer that converts BigInt fields to numbers for JSON serialization
 * Handles nested objects, arrays, protobuf Timestamps, and prevents precision loss
 */
export function sanitizeProtobufForJson<T>(obj: T): T {
  if (obj === null || obj === undefined) {
    return obj;
  }

  // Handle BigInt with precision safety
  if (typeof obj === 'bigint') {
    const n = Number(obj);
    // Use string if value exceeds safe integer range to preserve exactness
    return (Math.abs(n) <= Number.MAX_SAFE_INTEGER ? n : obj.toString()) as unknown as T;
  }

  if (typeof obj !== 'object') {
    return obj;
  }

  // Handle arrays
  if (Array.isArray(obj)) {
    return obj.map(item => sanitizeProtobufForJson(item)) as T;
  }

  // Handle Date objects (preserve them)
  if (obj instanceof Date) {
    return obj;
  }

  // Detect and convert protobuf Timestamp-like objects to ISO strings
  if (obj && typeof obj === 'object' && 'seconds' in obj && ('nanos' in obj || 'nanos' in (obj as any))) {
    try {
      const seconds = Number((obj as any).seconds || 0);
      const nanos = Number((obj as any).nanos || 0);
      const ms = seconds * 1000 + Math.round(nanos / 1e6);
      return new Date(ms).toISOString() as unknown as T;
    } catch {
      // Fallback to regular object handling if conversion fails
    }
  }

  // Handle protobuf Message objects - prefer toJson/toObject methods
  if (typeof (obj as any).constructor === 'function' && (obj as any).constructor.name) {
    try {
      // Prefer protobuf's toJson() method for safe serialization
      if (typeof (obj as any).toJson === 'function') {
        return sanitizeProtobufForJson((obj as any).toJson()) as T;
      }
      
      // Fallback to toObject() method if available
      if (typeof (obj as any).toObject === 'function') {
        return sanitizeProtobufForJson((obj as any).toObject()) as T;
      }
      
      // Last resort: try to create a new instance of the same type
      const Constructor = (obj as any).constructor;
      const clone = new Constructor(obj);
      
      // Recursively sanitize all enumerable properties
      for (const key in clone) {
        if (Object.prototype.hasOwnProperty.call(clone, key)) {
          clone[key] = sanitizeProtobufForJson(clone[key]);
        }
      }
      
      return clone as T;
    } catch {
      // Fallback to plain object handling
    }
  }

  // Handle plain objects
  const result = {} as T;
  for (const key in obj) {
    if (Object.prototype.hasOwnProperty.call(obj, key)) {
      (result as any)[key] = sanitizeProtobufForJson((obj as any)[key]);
    }
  }

  return result;
}

/**
 * Generate consistent cache key for any protobuf type with filters
 * Uses sanitized objects to ensure stable JSON serialization
 */
export function generateCacheKey(prefix: string, userId: string, filters?: any): string {
  if (!filters) {
    return `${prefix}-${userId}`;
  }
  
  const sanitizedFilters = sanitizeProtobufForJson(filters);
  // Use JSON from protobuf message if available, otherwise use direct JSON.stringify
  const filtersAsJson = typeof sanitizedFilters?.toJson === 'function'
    ? (sanitizedFilters as any).toJson()
    : sanitizedFilters;
    
  return `${prefix}-${userId}-${JSON.stringify(filtersAsJson)}`;
}

/**
 * Generate consistent cache key for spaces list
 * Combines userId with normalized filter parameters for stable caching
 */
export function generateSpacesListCacheKey(userId: string, filters?: SpaceFilters): string {
  const normalizedFilters = normalizeFilters(filters);
  return generateCacheKey('spaces-list', userId, normalizedFilters);
}
