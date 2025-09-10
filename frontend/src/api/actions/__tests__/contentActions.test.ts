/**
 * Unit tests for cached content server actions
 * Tests cache hit/miss scenarios, revalidation, and error handling
 */

import { 
  listContentSources, 
  createUploadURL, 
  confirmUpload, 
  deleteContentSource, 
  getContentSource 
} from '../contentActions';
import { 
  ContentSource, 
  CreateUploadURLRequest, 
  CreateUploadURLResponse,
  ContentStatus 
} from '../../generated/v1/knowledge_pb';
import * as clerkNext from '@clerk/nextjs/server';
import * as nextCache from 'next/cache';
import * as utils from '../utils';
import { getKnowledgeServiceClient } from '../../server-client';

// Mock modules
jest.mock('@clerk/nextjs/server');
jest.mock('next/cache');
jest.mock('../utils');
jest.mock('../../server-client');

// Mock cache functions
const mockUnstableCache = jest.fn();
const mockRevalidateTag = jest.fn();
const mockRevalidatePath = jest.fn();

// Mock client
const mockClient = {
  listContentSources: jest.fn(),
  createUploadURL: jest.fn(),
  confirmUpload: jest.fn(),
  deleteContentSource: jest.fn(),
  getContentSource: jest.fn(),
};

// Mock utility functions
const mockCreateAuthHeaders = jest.fn();
const mockSanitizeProtobufForJson = jest.fn();
const mockSanitizeError = jest.fn();
const mockIsUnauthorizedError = jest.fn();
const mockLogAuthFailure = jest.fn();

describe('Content Actions', () => {
  beforeEach(() => {
    jest.clearAllMocks();

    // Setup mocks
    (clerkNext.auth as jest.Mock).mockResolvedValue({ userId: 'test-user-123' });
    (nextCache.revalidateTag as jest.Mock).mockImplementation(mockRevalidateTag);
    (nextCache.revalidatePath as jest.Mock).mockImplementation(mockRevalidatePath);
    (nextCache.unstable_cache as jest.Mock).mockImplementation(mockUnstableCache);
    
    (utils.createAuthHeaders as jest.Mock).mockImplementation(mockCreateAuthHeaders);
    (utils.sanitizeProtobufForJson as jest.Mock).mockImplementation(mockSanitizeProtobufForJson);
    (utils.sanitizeError as jest.Mock).mockImplementation(mockSanitizeError);
    (utils.isUnauthorizedError as jest.Mock).mockImplementation(mockIsUnauthorizedError);
    (utils.logAuthFailure as jest.Mock).mockImplementation(mockLogAuthFailure);

    (getKnowledgeServiceClient as jest.Mock).mockReturnValue(mockClient);

    // Setup default mock implementations
    mockCreateAuthHeaders.mockResolvedValue(new Headers({ authorization: 'Bearer test-token' }));
    mockSanitizeProtobufForJson.mockImplementation(x => x);
    mockSanitizeError.mockImplementation(error => ({ code: 'UNKNOWN', message: error.message }));
    mockIsUnauthorizedError.mockReturnValue(false);
  });

  describe('listContentSources', () => {
    it('should return cached content sources on cache hit', async () => {
      const mockContentSources = [
        { id: 'content1', title: 'Test Content', status: ContentStatus.PROCESSED }
      ];
      const mockResponse = { items: mockContentSources };
      
      // Mock cache hit - unstable_cache returns cached data
      mockUnstableCache.mockImplementation((fn, keys, options) => {
        return jest.fn().mockResolvedValue({
          ok: true,
          data: mockContentSources,
        });
      });

      const result = await listContentSources('space1', 'processed');

      expect(result.ok).toBe(true);
      expect(result.data).toEqual(mockContentSources);
      expect(mockUnstableCache).toHaveBeenCalledWith(
        expect.any(Function),
        ['listContentSources-space1-processed'],
        {
          tags: ['content-sources-space1'],
          revalidate: 60, // 60 second TTL
        }
      );
    });

    it('should fetch fresh data on cache miss', async () => {
      const mockContentSources = [
        { id: 'content1', title: 'Test Content', status: ContentStatus.PROCESSED }
      ];
      const mockResponse = { items: mockContentSources };
      
      mockClient.listContentSources.mockResolvedValue(mockResponse);
      
      // Mock cache miss - unstable_cache executes the function
      mockUnstableCache.mockImplementation((fn, keys, options) => {
        return fn;
      });

      const result = await listContentSources('space1');

      expect(mockClient.listContentSources).toHaveBeenCalledWith(
        { spaceId: 'space1' },
        { headers: expect.any(Headers) }
      );
      expect(mockSanitizeProtobufForJson).toHaveBeenCalledTimes(mockContentSources.length);
    });

    it('should include status filter in request when provided', async () => {
      mockUnstableCache.mockImplementation((fn) => fn);
      mockClient.listContentSources.mockResolvedValue({ items: [] });

      await listContentSources('space1', 'processed');

      expect(mockClient.listContentSources).toHaveBeenCalledWith(
        { spaceId: 'space1', status: ContentStatus.PROCESSED },
        { headers: expect.any(Headers) }
      );
    });

    it('should handle different cache keys for different status filters', async () => {
      mockUnstableCache.mockImplementation((fn) => fn);
      mockClient.listContentSources.mockResolvedValue({ items: [] });

      await listContentSources('space1', 'uploading');

      expect(mockUnstableCache).toHaveBeenCalledWith(
        expect.any(Function),
        ['listContentSources-space1-uploading'],
        {
          tags: ['content-sources-space1'],
          revalidate: 60,
        }
      );
    });

    it('should handle authentication errors', async () => {
      (clerkNext.auth as jest.Mock).mockResolvedValue({ userId: null });

      const result = await listContentSources('space1');

      expect(result.ok).toBe(false);
      expect(result.error?.code).toBe('UNAUTHORIZED');
    });

    it('should handle unauthorized gRPC errors', async () => {
      const authError = new Error('Unauthorized');
      mockClient.listContentSources.mockRejectedValue(authError);
      mockIsUnauthorizedError.mockReturnValue(true);
      mockSanitizeError.mockReturnValue({ code: 'UNAUTHORIZED', message: 'Unauthorized' });

      mockUnstableCache.mockImplementation((fn) => fn);

      const result = await listContentSources('space1');

      expect(result.ok).toBe(false);
      expect(mockLogAuthFailure).toHaveBeenCalledWith('listContentSourcesCore', authError);
    });
  });

  describe('createUploadURL', () => {
    it('should create upload URL and sanitize response', async () => {
      const mockRequest = new CreateUploadURLRequest({
        spaceId: 'space1',
        filename: 'test.pdf',
        contentType: 'application/pdf'
      });
      const mockResponse = new CreateUploadURLResponse({
        uploadUrl: 'https://s3.example.com/upload',
        contentSource: { id: 'content1', sizeBytes: BigInt(1024) } as any
      });

      mockClient.createUploadURL.mockResolvedValue(mockResponse);

      const result = await createUploadURL(mockRequest);

      expect(result.ok).toBe(true);
      expect(mockSanitizeProtobufForJson).toHaveBeenCalledWith(mockResponse);
    });

    it('should handle creation errors', async () => {
      const mockRequest = new CreateUploadURLRequest({
        spaceId: 'space1',
        filename: 'test.pdf'
      });
      const createError = new Error('Upload URL creation failed');
      
      mockClient.createUploadURL.mockRejectedValue(createError);
      mockSanitizeError.mockReturnValue({ code: 'INTERNAL', message: 'Upload URL creation failed' });

      const result = await createUploadURL(mockRequest);

      expect(result.ok).toBe(false);
      expect(result.error?.message).toBe('Upload URL creation failed');
    });
  });

  describe('confirmUpload', () => {
    it('should confirm upload and revalidate content cache', async () => {
      const mockContentSource = { 
        id: 'content1', 
        spaceId: 'space1', 
        sizeBytes: BigInt(1024) 
      } as any;
      
      mockClient.confirmUpload.mockResolvedValue(mockContentSource);

      const result = await confirmUpload('content1', 'blob-hash-123');

      expect(result.ok).toBe(true);
      expect(mockSanitizeProtobufForJson).toHaveBeenCalledWith(mockContentSource);
      expect(mockRevalidateTag).toHaveBeenCalledWith('content-sources-space1');
      expect(mockRevalidatePath).toHaveBeenCalledWith('/spaces/space1');
    });

    it('should handle confirmation errors', async () => {
      const confirmError = new Error('Confirmation failed');
      
      mockClient.confirmUpload.mockRejectedValue(confirmError);
      mockSanitizeError.mockReturnValue({ code: 'INTERNAL', message: 'Confirmation failed' });

      const result = await confirmUpload('content1', 'blob-hash-123');

      expect(result.ok).toBe(false);
      expect(result.error?.message).toBe('Confirmation failed');
    });
  });

  describe('deleteContentSource', () => {
    it('should delete content source', async () => {
      mockClient.deleteContentSource.mockResolvedValue(undefined);

      const result = await deleteContentSource('content1');

      expect(result.ok).toBe(true);
      expect(mockClient.deleteContentSource).toHaveBeenCalledWith(
        { id: 'content1' },
        { headers: expect.any(Headers) }
      );
    });

    it('should handle deletion errors', async () => {
      const deleteError = new Error('Deletion failed');
      
      mockClient.deleteContentSource.mockRejectedValue(deleteError);
      mockSanitizeError.mockReturnValue({ code: 'INTERNAL', message: 'Deletion failed' });

      const result = await deleteContentSource('content1');

      expect(result.ok).toBe(false);
      expect(result.error?.message).toBe('Deletion failed');
    });
  });

  describe('getContentSource', () => {
    it('should get content source and sanitize response', async () => {
      const mockContentSource = { 
        id: 'content1', 
        title: 'Test Content',
        sizeBytes: BigInt(2048)
      } as any;
      
      mockClient.getContentSource.mockResolvedValue(mockContentSource);

      const result = await getContentSource('content1');

      expect(result.ok).toBe(true);
      expect(mockSanitizeProtobufForJson).toHaveBeenCalledWith(mockContentSource);
    });

    it('should handle content source not found', async () => {
      const notFoundError = new Error('Content source not found');
      
      mockClient.getContentSource.mockRejectedValue(notFoundError);
      mockSanitizeError.mockReturnValue({ code: 'NOT_FOUND', message: 'Content source not found' });

      const result = await getContentSource('nonexistent');

      expect(result.ok).toBe(false);
      expect(result.error?.code).toBe('NOT_FOUND');
    });
  });

  describe('Cache behavior validation', () => {
    it('should use short TTL for content sources (60s)', async () => {
      mockUnstableCache.mockImplementation((fn) => fn);
      mockClient.listContentSources.mockResolvedValue({ items: [] });

      await listContentSources('space1');

      expect(mockUnstableCache).toHaveBeenCalledWith(
        expect.any(Function),
        expect.any(Array),
        expect.objectContaining({
          revalidate: 60, // Content sources have shorter TTL due to frequent updates
        })
      );
    });

    it('should generate different cache keys for different spaces', async () => {
      mockUnstableCache.mockImplementation((fn) => fn);
      mockClient.listContentSources.mockResolvedValue({ items: [] });

      await listContentSources('space1');
      await listContentSources('space2');

      expect(mockUnstableCache).toHaveBeenNthCalledWith(1,
        expect.any(Function),
        ['listContentSources-space1-any'],
        expect.any(Object)
      );
      
      expect(mockUnstableCache).toHaveBeenNthCalledWith(2,
        expect.any(Function),
        ['listContentSources-space2-any'],
        expect.any(Object)
      );
    });
  });
});