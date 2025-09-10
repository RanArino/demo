/**
 * Simplified unit tests for cache utility functions
 */

import { 
  sanitizeProtobufForJson, 
  generateCacheKey,
  recordCacheHit,
  recordCacheMiss,
  getCacheHitRate,
  resetCacheMetrics 
} from '../utils';

describe('Cache Utility Functions', () => {
  beforeEach(() => {
    resetCacheMetrics();
  });

  describe('sanitizeProtobufForJson', () => {
    it('should convert BigInt to number', () => {
      const input = 12345n;
      const result = sanitizeProtobufForJson(input);
      expect(result).toBe(12345);
      expect(typeof result).toBe('number');
    });

    it('should handle nested objects with BigInt', () => {
      const input = {
        id: 'test',
        count: 123n,
        nested: {
          size: 456n
        }
      };
      const result = sanitizeProtobufForJson(input);
      expect(result.id).toBe('test');
      expect(result.count).toBe(123);
      expect(result.nested.size).toBe(456);
    });

    it('should preserve primitive types', () => {
      expect(sanitizeProtobufForJson('test')).toBe('test');
      expect(sanitizeProtobufForJson(123)).toBe(123);
      expect(sanitizeProtobufForJson(true)).toBe(true);
      expect(sanitizeProtobufForJson(null)).toBeNull();
      expect(sanitizeProtobufForJson(undefined)).toBeUndefined();
    });
  });

  describe('generateCacheKey', () => {
    it('should generate consistent keys for same inputs', () => {
      const key1 = generateCacheKey('test', 'user123', { q: 'search' });
      const key2 = generateCacheKey('test', 'user123', { q: 'search' });
      expect(key1).toBe(key2);
    });

    it('should generate different keys for different inputs', () => {
      const key1 = generateCacheKey('test', 'user123', { q: 'search1' });
      const key2 = generateCacheKey('test', 'user123', { q: 'search2' });
      expect(key1).not.toBe(key2);
    });

    it('should handle undefined filters', () => {
      const key = generateCacheKey('test', 'user123');
      expect(key).toBe('test-user123');
    });
  });

  describe('Cache Metrics', () => {
    it('should record cache hits correctly', () => {
      recordCacheHit('test-operation', 50);
      const hitRate = getCacheHitRate('test-operation');
      expect(hitRate).toBe(100);
    });

    it('should record cache misses correctly', () => {
      recordCacheMiss('test-operation', 200);
      const hitRate = getCacheHitRate('test-operation');
      expect(hitRate).toBe(0);
    });

    it('should calculate hit rate correctly with mixed hits and misses', () => {
      recordCacheHit('test-operation', 50);
      recordCacheHit('test-operation', 45);
      recordCacheMiss('test-operation', 200);
      recordCacheMiss('test-operation', 180);
      
      const hitRate = getCacheHitRate('test-operation');
      expect(hitRate).toBe(50); // 2 hits out of 4 total = 50%
    });

    it('should return 0 hit rate for unknown operations', () => {
      const hitRate = getCacheHitRate('unknown-operation');
      expect(hitRate).toBe(0);
    });

    it('should reset metrics correctly', () => {
      recordCacheHit('test-operation', 50);
      expect(getCacheHitRate('test-operation')).toBe(100);
      
      resetCacheMetrics('test-operation');
      expect(getCacheHitRate('test-operation')).toBe(0);
    });
  });
});