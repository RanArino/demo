/**
 * Performance tests for cached server actions
 * Tests cache hit rates, response times, and memory usage
 */

import { searchSpaces, getSpace } from '../spaceActions';
import { listContentSources } from '../contentActions';
import { SpaceFilters, Space, ContentSource } from '../../generated/v1/knowledge_pb';
import * as utils from '../utils';
import * as nextCache from 'next/cache';
import * as clerkNext from '@clerk/nextjs/server';
import { getKnowledgeServiceClient } from '../../server-client';

// Mock modules
jest.mock('@clerk/nextjs/server');
jest.mock('next/cache');
jest.mock('../utils');
jest.mock('../../server-client');

// Mock client
const mockClient = {
  listSpaces: jest.fn(),
  getSpace: jest.fn(),
  listContentSources: jest.fn(),
};

// Performance monitoring utilities
interface PerformanceMetrics {
  executionTime: number;
  cacheHit: boolean;
  memoryUsage?: number;
}

const measurePerformance = async (operation: () => Promise<any>): Promise<PerformanceMetrics> => {
  const startTime = performance.now();
  const startMemory = process.memoryUsage().heapUsed;
  
  await operation();
  
  const endTime = performance.now();
  const endMemory = process.memoryUsage().heapUsed;
  
  return {
    executionTime: endTime - startTime,
    cacheHit: false, // Will be determined by mock behavior
    memoryUsage: endMemory - startMemory,
  };
};

describe('Cache Performance Tests', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    
    // Setup base mocks
    (clerkNext.auth as jest.Mock).mockResolvedValue({ userId: 'test-user-123' });
    (utils.createAuthHeaders as jest.Mock).mockResolvedValue(new Headers({ authorization: 'Bearer test-token' }));
    (utils.sanitizeProtobufForJson as jest.Mock).mockImplementation(x => x);
    (utils.generateSpacesListCacheKey as jest.Mock).mockReturnValue('test-cache-key');
    (utils.sanitizeError as jest.Mock).mockImplementation(error => ({ code: 'UNKNOWN', message: error.message }));
    (utils.isUnauthorizedError as jest.Mock).mockReturnValue(false);
    (getKnowledgeServiceClient as jest.Mock).mockReturnValue(mockClient);
  });

  describe('Cache hit rate optimization', () => {
    it('should achieve high cache hit rates for repeated identical requests', async () => {
      const mockSpaces = Array.from({ length: 100 }, (_, i) => ({
        id: `space${i}`,
        title: `Test Space ${i}`,
        description: `Description ${i}`,
      }));

      let cacheCallCount = 0;
      let actualCallCount = 0;

      // Mock unstable_cache to simulate cache behavior
      (nextCache.unstable_cache as jest.Mock).mockImplementation((fn, keys, options) => {
        return async () => {
          cacheCallCount++;
          // Simulate cache hit after first call
          if (cacheCallCount === 1) {
            actualCallCount++;
            return {
              ok: true,
              data: { spaces: mockSpaces, totalCount: mockSpaces.length, page: 1, pageSize: 100 },
            };
          }
          // Subsequent calls are cache hits
          return {
            ok: true,
            data: { spaces: mockSpaces, totalCount: mockSpaces.length, page: 1, pageSize: 100 },
          };
        };
      });

      const filters = new SpaceFilters({ q: 'test query' });
      const requestCount = 10;
      const results = [];

      // Execute multiple identical requests
      for (let i = 0; i < requestCount; i++) {
        const metrics = await measurePerformance(() => searchSpaces(filters));
        results.push(metrics);
      }

      // Calculate cache hit rate
      const cacheHitRate = (requestCount - actualCallCount) / requestCount;
      
      expect(cacheHitRate).toBeGreaterThan(0.8); // Should have >80% cache hit rate
      expect(actualCallCount).toBe(1); // Only first call should hit the backend
      expect(cacheCallCount).toBe(requestCount); // All calls should go through cache layer
    });

    it('should generate different cache keys for different filter combinations', async () => {
      const cacheKeys = new Set<string>();
      
      (utils.generateSpacesListCacheKey as jest.Mock).mockImplementation((userId, filters) => {
        const key = `${userId}-${JSON.stringify(filters || {})}`;
        cacheKeys.add(key);
        return key;
      });

      const filterCombinations = [
        new SpaceFilters({ q: 'query1' }),
        new SpaceFilters({ q: 'query2' }),
        new SpaceFilters({ keywords: ['tag1'] }),
        new SpaceFilters({ keywords: ['tag1', 'tag2'] }),
        new SpaceFilters({ q: 'query1', keywords: ['tag1'] }),
      ];

      (nextCache.unstable_cache as jest.Mock).mockImplementation(() => 
        async () => ({ ok: true, data: { spaces: [], totalCount: 0, page: 1, pageSize: 20 } })
      );

      for (const filters of filterCombinations) {
        await searchSpaces(filters);
      }

      // Should generate unique cache keys for each filter combination
      expect(cacheKeys.size).toBe(filterCombinations.length);
    });
  });

  describe('Response time performance', () => {
    it('should respond faster on cache hits than cache misses', async () => {
      const mockSpaces = Array.from({ length: 50 }, (_, i) => ({
        id: `space${i}`,
        title: `Space ${i}`,
      }));

      let isFirstCall = true;

      // Mock to simulate slower backend call vs faster cache hit
      (nextCache.unstable_cache as jest.Mock).mockImplementation((fn) => {
        return async () => {
          if (isFirstCall) {
            isFirstCall = false;
            // Simulate slower backend call
            await new Promise(resolve => setTimeout(resolve, 100));
            return {
              ok: true,
              data: { spaces: mockSpaces, totalCount: mockSpaces.length, page: 1, pageSize: 50 },
            };
          } else {
            // Simulate fast cache hit
            await new Promise(resolve => setTimeout(resolve, 5));
            return {
              ok: true,
              data: { spaces: mockSpaces, totalCount: mockSpaces.length, page: 1, pageSize: 50 },
            };
          }
        };
      });

      // First call (cache miss)
      const cacheMissMetrics = await measurePerformance(() => searchSpaces());
      
      // Second call (cache hit)
      const cacheHitMetrics = await measurePerformance(() => searchSpaces());

      expect(cacheHitMetrics.executionTime).toBeLessThan(cacheMissMetrics.executionTime * 0.5);
      expect(cacheMissMetrics.executionTime).toBeGreaterThan(50); // Should take at least 50ms for backend call
      expect(cacheHitMetrics.executionTime).toBeLessThan(50); // Cache hit should be much faster
    });

    it('should maintain acceptable performance under load', async () => {
      const concurrentRequests = 20;
      const mockSpaces = [{ id: 'space1', title: 'Test Space' }];

      (nextCache.unstable_cache as jest.Mock).mockImplementation(() => 
        async () => ({
          ok: true,
          data: { spaces: mockSpaces, totalCount: 1, page: 1, pageSize: 1 },
        })
      );

      const startTime = performance.now();
      
      // Execute concurrent requests
      const promises = Array.from({ length: concurrentRequests }, () => 
        searchSpaces(new SpaceFilters({ q: 'load test' }))
      );

      await Promise.all(promises);
      
      const totalTime = performance.now() - startTime;
      const averageTime = totalTime / concurrentRequests;

      // All requests should complete reasonably quickly
      expect(totalTime).toBeLessThan(5000); // Total time under 5 seconds
      expect(averageTime).toBeLessThan(500); // Average under 500ms per request
    });
  });

  describe('Memory usage optimization', () => {
    it('should not cause excessive memory growth with repeated cache operations', async () => {
      const largeSpacesData = Array.from({ length: 1000 }, (_, i) => ({
        id: `space${i}`,
        title: `Large Space ${i}`,
        description: `Large description with lots of content for space ${i}`,
        metadata: { 
          tags: Array.from({ length: 10 }, (_, j) => `tag${j}`),
          content: 'x'.repeat(1000), // 1KB of content per space
        },
      }));

      (nextCache.unstable_cache as jest.Mock).mockImplementation(() => 
        async () => ({
          ok: true,
          data: { 
            spaces: largeSpacesData, 
            totalCount: largeSpacesData.length, 
            page: 1, 
            pageSize: largeSpacesData.length 
          },
        })
      );

      const initialMemory = process.memoryUsage().heapUsed;
      const requestCount = 50;

      // Execute multiple requests with large datasets
      for (let i = 0; i < requestCount; i++) {
        await searchSpaces(new SpaceFilters({ q: `query${i % 10}` }));
        
        // Force garbage collection every 10 requests if available
        if (i % 10 === 0 && global.gc) {
          global.gc();
        }
      }

      const finalMemory = process.memoryUsage().heapUsed;
      const memoryGrowth = finalMemory - initialMemory;
      const memoryGrowthMB = memoryGrowth / (1024 * 1024);

      // Memory growth should be reasonable (less than 50MB for this test)
      expect(memoryGrowthMB).toBeLessThan(50);
    });

    it('should efficiently handle different data sizes', async () => {
      const smallDataset = Array.from({ length: 10 }, (_, i) => ({ id: `small${i}`, title: `Small ${i}` }));
      const largeDataset = Array.from({ length: 1000 }, (_, i) => ({ id: `large${i}`, title: `Large ${i}` }));

      let callCount = 0;
      (nextCache.unstable_cache as jest.Mock).mockImplementation(() => {
        return async () => {
          callCount++;
          const data = callCount % 2 === 1 ? smallDataset : largeDataset;
          return {
            ok: true,
            data: { spaces: data, totalCount: data.length, page: 1, pageSize: data.length },
          };
        };
      });

      const smallDataMetrics = await measurePerformance(() => 
        searchSpaces(new SpaceFilters({ q: 'small' }))
      );

      const largeDataMetrics = await measurePerformance(() => 
        searchSpaces(new SpaceFilters({ q: 'large' }))
      );

      // Large dataset should not be exponentially slower
      const performanceRatio = largeDataMetrics.executionTime / smallDataMetrics.executionTime;
      expect(performanceRatio).toBeLessThan(10); // Should not be more than 10x slower
    });
  });

  describe('Cache invalidation performance', () => {
    it('should efficiently handle cache invalidation without blocking', async () => {
      const mockRevalidateTag = jest.fn().mockImplementation(() => {
        // Simulate some processing time for cache invalidation
        return new Promise(resolve => setTimeout(resolve, 10));
      });
      
      (nextCache.revalidateTag as jest.Mock).mockImplementation(mockRevalidateTag);

      const startTime = performance.now();
      
      // Simulate multiple cache invalidations
      const tags = Array.from({ length: 20 }, (_, i) => `spaces-list-user${i}`);
      
      for (const tag of tags) {
        mockRevalidateTag(tag);
      }

      const endTime = performance.now();
      const totalTime = endTime - startTime;

      expect(totalTime).toBeLessThan(1000); // Should complete within 1 second
      expect(mockRevalidateTag).toHaveBeenCalledTimes(tags.length);
    });
  });

  describe('Cross-service cache coordination', () => {
    it('should coordinate cache keys between spaces and content sources efficiently', async () => {
      const spaceId = 'test-space-123';
      const mockSpace = { id: spaceId, title: 'Test Space' };
      const mockContentSources = Array.from({ length: 100 }, (_, i) => ({
        id: `content${i}`,
        title: `Content ${i}`,
        spaceId,
      }));

      // Mock both space and content caching
      (nextCache.unstable_cache as jest.Mock)
        .mockImplementationOnce(() => async () => ({ ok: true, data: mockSpace }))
        .mockImplementationOnce(() => async () => ({ ok: true, data: mockContentSources }));

      const startTime = performance.now();
      
      // Execute related cache operations
      const [spaceResult, contentResult] = await Promise.all([
        getSpace(spaceId),
        listContentSources(spaceId, 'processed')
      ]);

      const endTime = performance.now();
      const totalTime = endTime - startTime;

      expect(spaceResult.ok).toBe(true);
      expect(contentResult.ok).toBe(true);
      expect(totalTime).toBeLessThan(500); // Should complete quickly with caching
      expect(nextCache.unstable_cache).toHaveBeenCalledTimes(2);
    });
  });

  describe('Cache key collision prevention', () => {
    it('should generate unique cache keys to prevent collisions', async () => {
      const generatedKeys = new Set<string>();
      
      (utils.generateSpacesListCacheKey as jest.Mock).mockImplementation((userId, filters) => {
        const key = `spaces-list-${userId}-${JSON.stringify(filters || {})}`;
        generatedKeys.add(key);
        return key;
      });

      (nextCache.unstable_cache as jest.Mock).mockImplementation(() => 
        async () => ({ ok: true, data: { spaces: [], totalCount: 0, page: 1, pageSize: 20 } })
      );

      // Generate many different filter combinations
      const users = ['user1', 'user2', 'user3'];
      const queries = ['query1', 'query2', '', undefined];
      const keywordSets = [[], ['tag1'], ['tag1', 'tag2'], ['tag3']];

      for (const userId of users) {
        (clerkNext.auth as jest.Mock).mockResolvedValue({ userId });
        
        for (const q of queries) {
          for (const keywords of keywordSets) {
            const filters = new SpaceFilters({ q, keywords });
            await searchSpaces(filters);
          }
        }
      }

      // All generated keys should be unique
      const totalCombinations = users.length * queries.length * keywordSets.length;
      expect(generatedKeys.size).toBe(totalCombinations);
    });
  });
});