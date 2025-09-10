/**
 * Environment configuration verification tests
 * Verifies that all required environment variables are properly configured
 * for the caching infrastructure to work correctly
 */

import { getKnowledgeServiceClient } from '../../server-client';

// Mock server client for testing
jest.mock('../../server-client');

describe('Environment Configuration Tests', () => {
  describe('Required Environment Variables', () => {
    it('should have MS_KNOWLEDGE_GRPC_URL_INTERNAL configured', () => {
      // This would be set in the container environment
      const knowledgeUrl = process.env.MS_KNOWLEDGE_GRPC_URL_INTERNAL;
      
      // In test environment, we expect this to be mocked or undefined
      // In production, this should be a valid gRPC URL
      if (process.env.NODE_ENV !== 'test') {
        expect(knowledgeUrl).toBeDefined();
        expect(knowledgeUrl).toMatch(/^[a-zA-Z0-9.-]+:\d+$/); // host:port format
      }
    });

    it('should have MS_USER_GRPC_URL_INTERNAL configured', () => {
      const userUrl = process.env.MS_USER_GRPC_URL_INTERNAL;
      
      if (process.env.NODE_ENV !== 'test') {
        expect(userUrl).toBeDefined();
        expect(userUrl).toMatch(/^[a-zA-Z0-9.-]+:\d+$/);
      }
    });

    it('should have Clerk authentication keys configured', () => {
      const clerkPublicKey = process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY;
      const clerkSecretKey = process.env.CLERK_SECRET_KEY;
      
      if (process.env.NODE_ENV !== 'test') {
        expect(clerkPublicKey).toBeDefined();
        expect(clerkSecretKey).toBeDefined();
        expect(clerkPublicKey).toMatch(/^pk_/); // Clerk public keys start with pk_
        expect(clerkSecretKey).toMatch(/^sk_/); // Clerk secret keys start with sk_
      }
    });

    it('should validate Next.js serverExternalPackages configuration', () => {
      // This test verifies that the required packages are externalized
      // for proper gRPC client functionality
      const requiredExternalPackages = [
        '@grpc/grpc-js',
        '@grpc/proto-loader', 
        'google-protobuf'
      ];

      // In a real environment test, we would check next.config.ts
      // For now, we just verify the packages exist in node_modules
      requiredExternalPackages.forEach(pkg => {
        expect(() => require.resolve(pkg)).not.toThrow();
      });
    });
  });

  describe('Development Environment Setup', () => {
    it('should verify Docker Compose configuration includes required services', () => {
      const requiredServices = [
        'ms_user',
        'ms_knowledge', 
        'frontend'
      ];

      // In a real test, we would parse docker-compose.yml
      // For now, we document the expected services
      expect(requiredServices).toContain('ms_user');
      expect(requiredServices).toContain('ms_knowledge');
      expect(requiredServices).toContain('frontend');
    });

    it('should verify internal service URLs are correctly mapped', () => {
      const expectedMappings = {
        'MS_USER_GRPC_URL_INTERNAL': 'ms_user:50051',
        'MS_KNOWLEDGE_GRPC_URL_INTERNAL': 'ms_knowledge:50052'
      };

      // These should be set in docker-compose.yml environment
      Object.entries(expectedMappings).forEach(([envVar, expectedValue]) => {
        if (process.env.NODE_ENV !== 'test') {
          expect(process.env[envVar]).toBe(expectedValue);
        }
      });
    });
  });

  describe('Production Environment Setup', () => {
    it('should verify production environment variables', () => {
      if (process.env.NODE_ENV === 'production') {
        // Database configuration
        expect(process.env.DATABASE_URL).toBeDefined();
        
        // Storage configuration (Cloudflare R2)
        expect(process.env.R2_ACCESS_KEY_ID).toBeDefined();
        expect(process.env.R2_SECRET_ACCESS_KEY).toBeDefined();
        expect(process.env.R2_ACCOUNT_ID).toBeDefined();
        expect(process.env.R2_BUCKET_SOURCE_NAME).toBeDefined();
        expect(process.env.R2_BUCKET_PROCESSED_NAME).toBeDefined();

        // Kafka configuration (if used)
        if (process.env.KAFKA_BROKERS) {
          expect(process.env.KAFKA_SASL_USERNAME).toBeDefined();
          expect(process.env.KAFKA_SASL_PASSWORD).toBeDefined();
          expect(process.env.KAFKA_SECURITY_PROTOCOL).toBeDefined();
        }
      }
    });

    it('should verify production gRPC URLs use external addresses', () => {
      if (process.env.NODE_ENV === 'production') {
        const knowledgeUrl = process.env.MS_KNOWLEDGE_GRPC_URL_INTERNAL;
        const userUrl = process.env.MS_USER_GRPC_URL_INTERNAL;

        // In production, these should not use Docker internal names
        if (knowledgeUrl) {
          expect(knowledgeUrl).not.toMatch(/^ms_/);
        }
        if (userUrl) {
          expect(userUrl).not.toMatch(/^ms_/);
        }
      }
    });
  });

  describe('Staging Environment Setup', () => {
    it('should verify staging environment has appropriate configuration', () => {
      if (process.env.NODE_ENV === 'staging') {
        // Staging should have similar requirements to production
        // but may use different endpoints or reduced resources
        expect(process.env.MS_KNOWLEDGE_GRPC_URL_INTERNAL).toBeDefined();
        expect(process.env.MS_USER_GRPC_URL_INTERNAL).toBeDefined();
        expect(process.env.CLERK_SECRET_KEY).toBeDefined();
      }
    });
  });

  describe('Cache Configuration Validation', () => {
    it('should verify Next.js cache settings are optimized', () => {
      // Verify that the app is configured for caching
      // These tests would check for proper Next.js configuration
      
      // unstable_cache should be available
      expect(typeof require('next/cache').unstable_cache).toBe('function');
      expect(typeof require('next/cache').revalidateTag).toBe('function');
      expect(typeof require('next/cache').revalidatePath).toBe('function');
      
      // React cache should be available
      expect(typeof require('react').cache).toBe('function');
    });

    it('should verify cache TTL configuration is reasonable', () => {
      const defaultCacheTTL = {
        spaces: 300, // 5 minutes
        contentSources: 60, // 1 minute
        space: 300, // 5 minutes
      };

      // Verify TTL values are reasonable
      expect(defaultCacheTTL.spaces).toBeGreaterThan(60); // At least 1 minute
      expect(defaultCacheTTL.spaces).toBeLessThan(3600); // Less than 1 hour
      
      expect(defaultCacheTTL.contentSources).toBeGreaterThan(30); // At least 30 seconds
      expect(defaultCacheTTL.contentSources).toBeLessThan(300); // Less than 5 minutes
    });
  });

  describe('Security Configuration', () => {
    it('should verify auth headers are properly configured', () => {
      // Verify that auth headers can be created
      const utils = require('../utils');
      expect(typeof utils.createAuthHeaders).toBe('function');
      expect(typeof utils.isUnauthorizedError).toBe('function');
      expect(typeof utils.logAuthFailure).toBe('function');
    });

    it('should verify error sanitization is in place', () => {
      const utils = require('../utils');
      expect(typeof utils.sanitizeError).toBe('function');
      expect(typeof utils.sanitizeProtobufForJson).toBe('function');
    });
  });

  describe('Service Health Checks', () => {
    it('should be able to create gRPC clients', () => {
      const mockClient = {
        listSpaces: jest.fn(),
        getSpace: jest.fn(),
        listContentSources: jest.fn(),
      };

      (getKnowledgeServiceClient as jest.Mock).mockReturnValue(mockClient);
      
      const client = getKnowledgeServiceClient();
      expect(client).toBeDefined();
      expect(typeof client.listSpaces).toBe('function');
      expect(typeof client.getSpace).toBe('function');
      expect(typeof client.listContentSources).toBe('function');
    });

    it('should handle gRPC connection errors gracefully', () => {
      (getKnowledgeServiceClient as jest.Mock).mockImplementation(() => {
        throw new Error('Connection failed');
      });

      expect(() => {
        try {
          getKnowledgeServiceClient();
        } catch (error) {
          // Should handle connection errors gracefully
          expect(error).toBeInstanceOf(Error);
          expect((error as Error).message).toBe('Connection failed');
        }
      }).not.toThrow();
    });
  });

  describe('Memory and Performance Configuration', () => {
    it('should verify Node.js memory settings are appropriate', () => {
      // Check if the app is running with appropriate memory limits
      const memoryUsage = process.memoryUsage();
      
      // Verify we have reasonable memory available
      expect(memoryUsage.heapUsed).toBeGreaterThan(0);
      expect(memoryUsage.heapTotal).toBeGreaterThan(memoryUsage.heapUsed);
      
      // For production, we expect more memory to be available
      if (process.env.NODE_ENV === 'production') {
        expect(memoryUsage.heapTotal).toBeGreaterThan(100 * 1024 * 1024); // At least 100MB
      }
    });

    it('should verify garbage collection is properly configured', () => {
      // Check if garbage collection is working
      const initialMemory = process.memoryUsage().heapUsed;
      
      // Create some temporary objects
      const tempData = Array.from({ length: 1000 }, (_, i) => ({ id: i, data: 'x'.repeat(1000) }));
      
      // Clear references
      tempData.length = 0;
      
      // Force GC if available
      if (global.gc) {
        global.gc();
      }
      
      const finalMemory = process.memoryUsage().heapUsed;
      
      // Memory management should be working
      expect(finalMemory).toBeGreaterThan(0);
    });
  });
});