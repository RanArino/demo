// =========================================================================
// ACTION RESULT INTERFACE
// =========================================================================

export interface ActionResult<T> {
    ok: boolean
    data?: T
    error?: {
      code: string
      message: string
      details?: Record<string, any>
    }
  }

export interface ServerActionResult<T> extends ActionResult<T> {
    timestamp: string
    requestId?: string
}

// =========================================================================
// TIMESTAMP UTILITIES
// =========================================================================

/**
 * Safely converts a protobuf Timestamp to a JavaScript Date
 * Handles both proper Timestamp objects and serialized objects with seconds/nanos
 */
export function safeTimestampToDate(timestamp: any): Date | null {
  if (!timestamp) return null;
  
  // If it's a proper Timestamp object with toDate method
  if (typeof timestamp.toDate === 'function') {
    try {
      return timestamp.toDate();
    } catch (error) {
      console.warn('Error calling toDate() on Timestamp:', error);
    }
  }
  
  // If it's a serialized object with seconds and nanos
  if (typeof timestamp.seconds !== 'undefined') {
    try {
      const seconds = typeof timestamp.seconds === 'bigint' 
        ? Number(timestamp.seconds) 
        : timestamp.seconds;
      const nanos = timestamp.nanos || 0;
      return new Date(seconds * 1000 + nanos / 1000000);
    } catch (error) {
      console.warn('Error converting seconds/nanos to Date:', error);
    }
  }
  
  return null;
}
