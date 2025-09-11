import { auth } from '@clerk/nextjs/server';
import { ConnectError } from '@bufbuild/connect';
import { SpaceFilters, CreateUploadURLRequest, DownloadObjectKind } from '../generated/v1/knowledge_pb';
import { protoInt64 } from '@bufbuild/protobuf';

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
    // Use string if value exceeds safe integer range to preserve exactness
    const maxSafe = BigInt(Number.MAX_SAFE_INTEGER);
    const minSafe = BigInt(Number.MIN_SAFE_INTEGER);
    return (obj <= maxSafe && obj >= minSafe ? Number(obj) : obj.toString()) as unknown as T;
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
 * Rebuilds a CreateUploadURLRequest to ensure numeric and enum fields are properly typed
 * before binary serialization. Accepts a typed message that may have been de-serialized
 * across server action boundaries and returns a normalized instance.
 */
export function rebuildCreateUploadURLRequest(input: CreateUploadURLRequest): CreateUploadURLRequest {
  const rebuilt = new CreateUploadURLRequest({
    spaceId: String((input as any).spaceId ?? input.spaceId ?? ''),
    filename: String((input as any).filename ?? input.filename ?? ''),
    mimeType: String((input as any).mimeType ?? input.mimeType ?? ''),
    title: String((input as any).title ?? input.title ?? ''),
  });

  // Normalize size_bytes (int64)
  try {
    const rawSize = (input as any).sizeBytes ?? 0;
    rebuilt.sizeBytes = protoInt64.parse(rawSize);
  } catch (e) {
    rebuilt.sizeBytes = (CreateUploadURLRequest as any).prototype.sizeBytes ?? 0;
    console.warn('rebuildCreateUploadURLRequest: failed to normalize sizeBytes', e);
  }

  // Normalize object_kind (enum)
  const rawKind = (input as any).objectKind ?? (input as any).object_kind ?? input.objectKind;
  if (typeof rawKind === 'number') {
    rebuilt.objectKind = rawKind as any;
  } else if (typeof rawKind === 'string') {
    const exact = (DownloadObjectKind as any)[rawKind];
    const prefixed = (DownloadObjectKind as any)[`DOWNLOAD_OBJECT_KIND_${rawKind}`];
    if (typeof exact === 'number') rebuilt.objectKind = exact;
    else if (typeof prefixed === 'number') rebuilt.objectKind = prefixed;
    else rebuilt.objectKind = DownloadObjectKind.ORIGINAL;
  } else {
    rebuilt.objectKind = input.objectKind ?? DownloadObjectKind.ORIGINAL;
  }

  return rebuilt;
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
  const filtersAsJson = sanitizedFilters && typeof sanitizedFilters === 'object' && 'toJson' in sanitizedFilters && typeof sanitizedFilters.toJson === 'function'
    ? (sanitizedFilters as { toJson(): unknown }).toJson()
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

// Cache metrics tracking
interface CacheMetrics {
  hits: number;
  misses: number;
  errors: number;
  totalRequests: number;
  avgResponseTime: number;
  lastUpdated: number;
}

// In-memory cache metrics storage (use Redis/external storage in production)
const cacheMetrics = new Map<string, CacheMetrics>();

/**
 * Initialize cache metrics for a given operation
 */
function initializeCacheMetrics(_operation: string): CacheMetrics {
  return {
    hits: 0,
    misses: 0,
    errors: 0,
    totalRequests: 0,
    avgResponseTime: 0,
    lastUpdated: Date.now(),
  };
}

/**
 * Record cache hit for monitoring
 */
export function recordCacheHit(operation: string, responseTime: number = 0): void {
  try {
    const metrics = cacheMetrics.get(operation) || initializeCacheMetrics(operation);
    
    metrics.hits++;
    metrics.totalRequests++;
    
    // Update average response time (exponential moving average)
    metrics.avgResponseTime = metrics.avgResponseTime === 0 
      ? responseTime 
      : (metrics.avgResponseTime * 0.8) + (responseTime * 0.2);
    
    metrics.lastUpdated = Date.now();
    cacheMetrics.set(operation, metrics);
    
    // Log significant cache events in development
    if (process.env.NODE_ENV === 'development') {
      console.log(`[cache-hit] ${operation}: ${responseTime.toFixed(2)}ms`);
    }
  } catch (error) {
    console.warn('[cache-metrics] Failed to record cache hit:', error);
  }
}

/**
 * Record cache miss for monitoring
 */
export function recordCacheMiss(operation: string, responseTime: number = 0): void {
  try {
    const metrics = cacheMetrics.get(operation) || initializeCacheMetrics(operation);
    
    metrics.misses++;
    metrics.totalRequests++;
    
    // Update average response time
    metrics.avgResponseTime = metrics.avgResponseTime === 0 
      ? responseTime 
      : (metrics.avgResponseTime * 0.8) + (responseTime * 0.2);
    
    metrics.lastUpdated = Date.now();
    cacheMetrics.set(operation, metrics);
    
    // Log cache misses in development
    if (process.env.NODE_ENV === 'development') {
      console.log(`[cache-miss] ${operation}: ${responseTime.toFixed(2)}ms`);
    }
  } catch (error) {
    console.warn('[cache-metrics] Failed to record cache miss:', error);
  }
}

/**
 * Record cache error for monitoring
 */
export function recordCacheError(operation: string, error: unknown): void {
  try {
    const metrics = cacheMetrics.get(operation) || initializeCacheMetrics(operation);
    
    metrics.errors++;
    metrics.totalRequests++;
    metrics.lastUpdated = Date.now();
    cacheMetrics.set(operation, metrics);
    
    // Log cache errors
    const errorMessage = error instanceof Error ? error.message : 'Unknown cache error';
    console.error(`[cache-error] ${operation}: ${errorMessage}`);
  } catch (logError) {
    console.warn('[cache-metrics] Failed to record cache error:', logError);
  }
}

/**
 * Get cache metrics for a specific operation
 */
export function getCacheMetrics(operation: string): CacheMetrics | null {
  return cacheMetrics.get(operation) || null;
}

/**
 * Get all cache metrics
 */
export function getAllCacheMetrics(): Record<string, CacheMetrics> {
  const result: Record<string, CacheMetrics> = {};
  for (const [operation, metrics] of cacheMetrics.entries()) {
    result[operation] = { ...metrics };
  }
  return result;
}

/**
 * Calculate cache hit rate for an operation
 */
export function getCacheHitRate(operation: string): number {
  const metrics = cacheMetrics.get(operation);
  if (!metrics || metrics.totalRequests === 0) {
    return 0;
  }
  return (metrics.hits / metrics.totalRequests) * 100;
}

/**
 * Reset cache metrics (useful for testing or periodic resets)
 */
export function resetCacheMetrics(operation?: string): void {
  if (operation) {
    cacheMetrics.delete(operation);
  } else {
    cacheMetrics.clear();
  }
}

/**
 * Log cache performance summary
 */
export function logCachePerformanceSummary(): void {
  try {
    const allMetrics = getAllCacheMetrics();
    const operations = Object.keys(allMetrics);
    
    if (operations.length === 0) {
      console.log('[cache-summary] No cache metrics available');
      return;
    }
    
    console.log('[cache-summary] Cache Performance Summary:');
    console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
    
    for (const operation of operations) {
      const metrics = allMetrics[operation];
      const hitRate = getCacheHitRate(operation);
      const avgTime = metrics.avgResponseTime.toFixed(2);
      
      console.log(`${operation}:`);
      console.log(`  Hit Rate: ${hitRate.toFixed(1)}% (${metrics.hits}/${metrics.totalRequests})`);
      console.log(`  Avg Response: ${avgTime}ms`);
      console.log(`  Errors: ${metrics.errors}`);
      console.log(`  Last Updated: ${new Date(metrics.lastUpdated).toLocaleString()}`);
      console.log('');
    }
    
    // Calculate overall statistics
    const totalRequests = operations.reduce((sum, op) => sum + allMetrics[op].totalRequests, 0);
    const totalHits = operations.reduce((sum, op) => sum + allMetrics[op].hits, 0);
    const totalErrors = operations.reduce((sum, op) => sum + allMetrics[op].errors, 0);
    const overallHitRate = totalRequests > 0 ? (totalHits / totalRequests) * 100 : 0;
    
    console.log('Overall Statistics:');
    console.log(`  Total Requests: ${totalRequests}`);
    console.log(`  Overall Hit Rate: ${overallHitRate.toFixed(1)}%`);
    console.log(`  Total Errors: ${totalErrors}`);
    console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  } catch (error) {
    console.warn('[cache-summary] Failed to log performance summary:', error);
  }
}

/**
 * Monitor cache performance and log alerts
 */
export function monitorCachePerformance(): void {
  try {
    const allMetrics = getAllCacheMetrics();
    const now = Date.now();
    
    for (const [operation, metrics] of Object.entries(allMetrics)) {
      const hitRate = getCacheHitRate(operation);
      const errorRate = metrics.totalRequests > 0 ? (metrics.errors / metrics.totalRequests) * 100 : 0;
      const minutesSinceUpdate = (now - metrics.lastUpdated) / (1000 * 60);
      
      // Alert on low hit rate
      if (metrics.totalRequests >= 10 && hitRate < 50) {
        console.warn(`[cache-alert] Low hit rate for ${operation}: ${hitRate.toFixed(1)}%`);
      }
      
      // Alert on high error rate
      if (metrics.totalRequests >= 5 && errorRate > 10) {
        console.warn(`[cache-alert] High error rate for ${operation}: ${errorRate.toFixed(1)}%`);
      }
      
      // Alert on slow response times
      if (metrics.avgResponseTime > 1000) {
        console.warn(`[cache-alert] Slow responses for ${operation}: ${metrics.avgResponseTime.toFixed(2)}ms avg`);
      }
      
      // Alert on stale metrics (no updates in 30 minutes)
      if (minutesSinceUpdate > 30) {
        console.warn(`[cache-alert] Stale metrics for ${operation}: ${minutesSinceUpdate.toFixed(1)} minutes since last update`);
      }
    }
  } catch (error) {
    console.warn('[cache-monitor] Failed to monitor cache performance:', error);
  }
}

/**
 * Export cache metrics for external monitoring systems
 */
export function exportCacheMetricsForMonitoring(): string {
  try {
    const allMetrics = getAllCacheMetrics();
    const timestamp = new Date().toISOString();
    
    const exportData = {
      timestamp,
      metrics: allMetrics,
      summary: {
        totalOperations: Object.keys(allMetrics).length,
        totalRequests: Object.values(allMetrics).reduce((sum, m) => sum + m.totalRequests, 0),
        overallHitRate: Object.values(allMetrics).reduce((sum, m) => sum + m.hits, 0) / 
                       Math.max(1, Object.values(allMetrics).reduce((sum, m) => sum + m.totalRequests, 0)) * 100,
        totalErrors: Object.values(allMetrics).reduce((sum, m) => sum + m.errors, 0),
      },
    };
    
    return JSON.stringify(exportData, null, 2);
  } catch (error) {
    console.warn('[cache-export] Failed to export metrics:', error);
    return JSON.stringify({ error: 'Failed to export metrics', timestamp: new Date().toISOString() });
  }
}

/**
 * Start periodic cache monitoring. Returns a cleanup function to clear intervals.
 * Call this from your server entrypoint or initialization code.
 */
export function startCacheMonitoring(): () => void {
  if (typeof window !== 'undefined' || process.env.NODE_ENV === 'test') {
    // Do not start monitoring in browser or test environments
    return () => {};
  }
  
  // Log summary every 5 minutes
  const summaryInterval = setInterval(() => {
    logCachePerformanceSummary();
  }, 5 * 60 * 1000);
  
  // Monitor performance every minute
  const monitorInterval = setInterval(() => {
    monitorCachePerformance();
  }, 60 * 1000);
  
  // Return cleanup function
  return () => {
    clearInterval(summaryInterval);
    clearInterval(monitorInterval);
  };
}
