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

  // Proper Timestamp instance with toDate()
  if (typeof timestamp.toDate === 'function') {
    try {
      return timestamp.toDate();
    } catch (error) {
      console.warn('Error calling toDate() on Timestamp:', error);
    }
  }

  // Serialized object with seconds and nanos (seconds may be bigint | number | string)
  if (typeof (timestamp as any).seconds !== 'undefined') {
    try {
      const rawSeconds = (timestamp as any).seconds as unknown;
      const seconds =
        typeof rawSeconds === 'bigint' ? Number(rawSeconds) :
        typeof rawSeconds === 'string' ? Number(rawSeconds) :
        (rawSeconds as number);
      const nanos = (timestamp as any).nanos || 0;
      if (!Number.isFinite(seconds)) return null;
      return new Date(seconds * 1000 + nanos / 1_000_000);
    } catch (error) {
      console.warn('Error converting seconds/nanos to Date:', error);
    }
  }

  // ISO string
  if (typeof timestamp === 'string') {
    const d = new Date(timestamp);
    return isNaN(d.getTime()) ? null : d;
  }

  return null;
}
