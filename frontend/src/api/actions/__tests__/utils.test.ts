/**
 * Unit tests for sanitizeProtobufForJson utility function
 * Tests BigInt handling, Timestamp conversion, and protobuf message sanitization
 */

import { sanitizeProtobufForJson, normalizeFilters, generateCacheKey } from '../utils';
import { SpaceFilters } from '../../generated/v1/knowledge_pb';

describe('sanitizeProtobufForJson', () => {
  describe('BigInt handling', () => {
    it('should convert small BigInt to number', () => {
      const input = 12345n;
      const result = sanitizeProtobufForJson(input);
      expect(result).toBe(12345);
      expect(typeof result).toBe('number');
    });

    it('should convert BigInt within safe integer range to number', () => {
      const input = BigInt(Number.MAX_SAFE_INTEGER);
      const result = sanitizeProtobufForJson(input);
      expect(result).toBe(Number.MAX_SAFE_INTEGER);
      expect(typeof result).toBe('number');
    });

    it('should convert large BigInt to number (temporary behavior)', () => {
      const input = BigInt(Number.MAX_SAFE_INTEGER) + 1n;
      const result = sanitizeProtobufForJson(input);
      expect(result).toBe(Number(BigInt(Number.MAX_SAFE_INTEGER) + 1n));
      expect(typeof result).toBe('number');
      // Note: This may lose precision, but maintains UI compatibility
    });

    it('should convert negative large BigInt to number (temporary behavior)', () => {
      const input = -(BigInt(Number.MAX_SAFE_INTEGER) + 1n);
      const result = sanitizeProtobufForJson(input);
      expect(result).toBe(Number(-(BigInt(Number.MAX_SAFE_INTEGER) + 1n)));
      expect(typeof result).toBe('number');
      // Note: This may lose precision, but maintains UI compatibility
    });
  });

  describe('Timestamp handling', () => {
    it('should convert protobuf Timestamp to ISO string', () => {
      const timestamp = {
        seconds: 1609459200, // 2021-01-01T00:00:00Z
        nanos: 0
      };
      const result = sanitizeProtobufForJson(timestamp);
      expect(result).toBe('2021-01-01T00:00:00.000Z');
    });

    it('should handle Timestamp with nanoseconds', () => {
      const timestamp = {
        seconds: 1609459200,
        nanos: 123456789 // 123.456789 ms
      };
      const result = sanitizeProtobufForJson(timestamp);
      expect(result).toBe('2021-01-01T00:00:00.123Z');
    });

    it('should handle malformed Timestamp gracefully', () => {
      const timestamp = {
        seconds: 'invalid',
        nanos: 'invalid'
      };
      const result = sanitizeProtobufForJson(timestamp);
      // Should fallback to object handling instead of crashing
      expect(typeof result).toBe('object');
    });
  });

  describe('Nested objects and arrays', () => {
    it('should sanitize nested BigInt values in objects', () => {
      const input = {
        id: 'test',
        count: 123n,
        largeCount: BigInt(Number.MAX_SAFE_INTEGER) + 1n,
        nested: {
          size: 456n
        }
      };
      const result = sanitizeProtobufForJson(input);
      expect(result.id).toBe('test');
      expect(result.count).toBe(123);
      expect(result.largeCount).toBe(Number(BigInt(Number.MAX_SAFE_INTEGER) + 1n));
      expect(result.nested.size).toBe(456);
    });

    it('should sanitize BigInt values in arrays', () => {
      const input = [123n, BigInt(Number.MAX_SAFE_INTEGER) + 1n, 'string', { count: 789n }];
      const result = sanitizeProtobufForJson(input);
      expect(result[0]).toBe(123);
      expect(result[1]).toBe(Number(BigInt(Number.MAX_SAFE_INTEGER) + 1n));
      expect(result[2]).toBe('string');
      expect(result[3].count).toBe(789);
    });
  });

  describe('Primitive types', () => {
    it('should preserve null and undefined', () => {
      expect(sanitizeProtobufForJson(null)).toBeNull();
      expect(sanitizeProtobufForJson(undefined)).toBeUndefined();
    });

    it('should preserve strings, numbers, and booleans', () => {
      expect(sanitizeProtobufForJson('test')).toBe('test');
      expect(sanitizeProtobufForJson(123)).toBe(123);
      expect(sanitizeProtobufForJson(true)).toBe(true);
      expect(sanitizeProtobufForJson(false)).toBe(false);
    });

    it('should preserve Date objects', () => {
      const date = new Date('2021-01-01T00:00:00Z');
      const result = sanitizeProtobufForJson(date);
      expect(result).toBe(date);
      expect(result instanceof Date).toBe(true);
    });
  });

  describe('Protobuf Message objects', () => {
    it('should use toJson() method when available', () => {
      const mockMessage = {
        field: 123n,
        toJson: jest.fn(() => ({ field: 123 }))
      };
      const result = sanitizeProtobufForJson(mockMessage);
      expect(mockMessage.toJson).toHaveBeenCalled();
      expect(result.field).toBe(123);
    });

    it('should use toObject() method when toJson() not available', () => {
      const mockMessage = {
        field: 123n,
        toObject: jest.fn(() => ({ field: 123 }))
      };
      const result = sanitizeProtobufForJson(mockMessage);
      expect(mockMessage.toObject).toHaveBeenCalled();
      expect(result.field).toBe(123);
    });

    it('should handle constructor cloning when no proto methods available', () => {
      // Mock a protobuf-like constructor
      function MockConstructor(this: any, data: any) {
        Object.assign(this, data);
      }
      MockConstructor.prototype.constructor = MockConstructor;

      const input = new (MockConstructor as any)({ field: 123n });
      const result = sanitizeProtobufForJson(input);
      expect(result.field).toBe(123);
    });
  });

  describe('Edge cases', () => {
    it('should handle circular references gracefully', () => {
      const obj: any = { field: 123n };
      obj.circular = obj;
      
      // Should not throw but may have some reasonable behavior
      expect(() => sanitizeProtobufForJson(obj)).not.toThrow();
    });

    it('should handle complex nested structures', () => {
      const input = {
        spaces: [
          {
            id: 'space1',
            documentCount: 100n,
            stats: {
              contentCount: 50n,
              linkCount: 25n,
              lastActivityAt: {
                seconds: 1609459200,
                nanos: 0
              }
            }
          }
        ],
        totalCount: BigInt(Number.MAX_SAFE_INTEGER) + 1n
      };

      const result = sanitizeProtobufForJson(input);
      expect(result.spaces[0].documentCount).toBe(100);
      expect(result.spaces[0].stats.contentCount).toBe(50);
      expect(result.spaces[0].stats.linkCount).toBe(25);
      expect(result.spaces[0].stats.lastActivityAt).toBe('2021-01-01T00:00:00.000Z');
      expect(result.totalCount).toBe(Number(BigInt(Number.MAX_SAFE_INTEGER) + 1n));
    });
  });
});

describe('normalizeFilters', () => {
  it('should normalize undefined SpaceFilters', () => {
    const result = normalizeFilters(undefined);
    expect(result.q).toBe('');
    expect(result.keywords).toEqual([]);
  });

  it('should normalize partial SpaceFilters', () => {
    const input = new SpaceFilters({ q: 'test' });
    const result = normalizeFilters(input);
    expect(result.q).toBe('test');
    expect(result.keywords).toEqual([]);
  });

  it('should preserve complete SpaceFilters', () => {
    const input = new SpaceFilters({ 
      q: 'test query', 
      keywords: ['tag1', 'tag2'] 
    });
    const result = normalizeFilters(input);
    expect(result.q).toBe('test query');
    expect(result.keywords).toEqual(['tag1', 'tag2']);
  });
});

describe('generateCacheKey', () => {
  it('should generate consistent keys for the same inputs', () => {
    const userId = 'user123';
    const filters = new SpaceFilters({ q: 'test', keywords: ['tag1'] });
    
    const key1 = generateCacheKey('spaces-list', userId, filters);
    const key2 = generateCacheKey('spaces-list', userId, filters);
    
    expect(key1).toBe(key2);
  });

  it('should generate different keys for different filters', () => {
    const userId = 'user123';
    const filters1 = new SpaceFilters({ q: 'test1' });
    const filters2 = new SpaceFilters({ q: 'test2' });
    
    const key1 = generateCacheKey('spaces-list', userId, filters1);
    const key2 = generateCacheKey('spaces-list', userId, filters2);
    
    expect(key1).not.toBe(key2);
  });

  it('should handle undefined filters', () => {
    const userId = 'user123';
    const key = generateCacheKey('spaces-list', userId);
    
    expect(key).toBe('spaces-list-user123');
  });

  it('should sanitize BigInt values in filters', () => {
    const userId = 'user123';
    const filtersWithBigInt = { count: 123n, q: 'test' };
    
    const key = generateCacheKey('spaces-list', userId, filtersWithBigInt);
    
    // Should not throw and should produce stable key
    expect(typeof key).toBe('string');
    expect(key).toContain('user123');
  });
});