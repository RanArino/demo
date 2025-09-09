/**
 * Unit tests for cached space server actions
 * Tests cache hit/miss scenarios, revalidation, and error handling
 */

import { searchSpaces, getSpace, createSpace, updateSpace, deleteSpace } from '../spaceActions';
import { SpaceFilters, Space, CreateSpaceRequest, UpdateSpaceRequest } from '../../generated/v1/knowledge_pb';
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
const mockCache = jest.fn();
const mockUnstableCache = jest.fn();
const mockRevalidateTag = jest.fn();
const mockRevalidatePath = jest.fn();

// Mock client
const mockClient = {
  listSpaces: jest.fn(),
  getSpace: jest.fn(),
  createSpace: jest.fn(),
  updateSpace: jest.fn(),
  deleteSpace: jest.fn(),
};

// Mock utility functions
const mockCreateAuthHeaders = jest.fn();
const mockSanitizeProtobufForJson = jest.fn();
const mockGenerateSpacesListCacheKey = jest.fn();
const mockSanitizeError = jest.fn();
const mockIsUnauthorizedError = jest.fn();
const mockLogAuthFailure = jest.fn();

describe('Space Actions', () => {
  beforeEach(() => {
    jest.clearAllMocks();

    // Setup mocks
    (clerkNext.auth as jest.Mock).mockResolvedValue({ userId: 'test-user-123' });
    (nextCache.revalidateTag as jest.Mock).mockImplementation(mockRevalidateTag);
    (nextCache.revalidatePath as jest.Mock).mockImplementation(mockRevalidatePath);
    (nextCache.unstable_cache as jest.Mock).mockImplementation(mockUnstableCache);
    
    (utils.createAuthHeaders as jest.Mock).mockImplementation(mockCreateAuthHeaders);
    (utils.sanitizeProtobufForJson as jest.Mock).mockImplementation(mockSanitizeProtobufForJson);
    (utils.generateSpacesListCacheKey as jest.Mock).mockImplementation(mockGenerateSpacesListCacheKey);
    (utils.sanitizeError as jest.Mock).mockImplementation(mockSanitizeError);
    (utils.isUnauthorizedError as jest.Mock).mockImplementation(mockIsUnauthorizedError);
    (utils.logAuthFailure as jest.Mock).mockImplementation(mockLogAuthFailure);

    (getKnowledgeServiceClient as jest.Mock).mockReturnValue(mockClient);

    // Setup default mock implementations
    mockCreateAuthHeaders.mockResolvedValue(new Headers({ authorization: 'Bearer test-token' }));
    mockSanitizeProtobufForJson.mockImplementation(x => x);
    mockGenerateSpacesListCacheKey.mockReturnValue('test-cache-key');
    mockSanitizeError.mockImplementation(error => ({ code: 'UNKNOWN', message: error.message }));
    mockIsUnauthorizedError.mockReturnValue(false);
  });

  describe('searchSpaces', () => {
    it('should return cached data on cache hit', async () => {
      const mockSpaces = [{ id: 'space1', title: 'Test Space' }];
      const mockResponse = { items: mockSpaces };
      
      // Mock cache hit - unstable_cache returns cached data
      mockUnstableCache.mockImplementation((fn, keys, options) => {
        // Simulate cache hit by immediately returning cached data
        return jest.fn().mockResolvedValue({
          ok: true,
          data: {
            spaces: mockSpaces,
            totalCount: 1,
            page: 1,
            pageSize: 1,
          },
        });
      });

      const result = await searchSpaces();

      expect(result.ok).toBe(true);
      expect(result.data?.spaces).toEqual(mockSpaces);
      expect(mockGenerateSpacesListCacheKey).toHaveBeenCalledWith('test-user-123', undefined);
      expect(mockUnstableCache).toHaveBeenCalledWith(
        expect.any(Function),
        ['test-cache-key'],
        {
          tags: ['spaces-list-test-user-123'],
          revalidate: 300,
        }
      );
    });

    it('should fetch fresh data on cache miss', async () => {
      const mockSpaces = [{ id: 'space1', title: 'Test Space' }];
      const mockResponse = { items: mockSpaces };
      
      mockClient.listSpaces.mockResolvedValue(mockResponse);
      
      // Mock cache miss - unstable_cache executes the function
      mockUnstableCache.mockImplementation((fn, keys, options) => {
        return fn;
      });

      const filters = new SpaceFilters({ q: 'test query' });
      const result = await searchSpaces(filters);

      expect(mockClient.listSpaces).toHaveBeenCalled();
      expect(mockSanitizeProtobufForJson).toHaveBeenCalledTimes(mockSpaces.length);
      expect(mockGenerateSpacesListCacheKey).toHaveBeenCalledWith('test-user-123', filters);
    });

    it('should handle authentication errors properly', async () => {
      (clerkNext.auth as jest.Mock).mockResolvedValue({ userId: null });

      const result = await searchSpaces();

      expect(result.ok).toBe(false);
      expect(result.error?.code).toBe('UNAUTHORIZED');
    });

    it('should handle unauthorized gRPC errors', async () => {
      const authError = new Error('Unauthorized');
      mockClient.listSpaces.mockRejectedValue(authError);
      mockIsUnauthorizedError.mockReturnValue(true);
      mockSanitizeError.mockReturnValue({ code: 'UNAUTHORIZED', message: 'Unauthorized' });

      mockUnstableCache.mockImplementation((fn) => fn);

      const result = await searchSpaces();

      expect(result.ok).toBe(false);
      expect(mockLogAuthFailure).toHaveBeenCalledWith('searchSpacesCore', authError);
      expect(mockIsUnauthorizedError).toHaveBeenCalledWith(authError);
    });

    it('should use correct cache configuration', async () => {
      mockUnstableCache.mockImplementation((fn) => fn);
      
      await searchSpaces();

      expect(mockUnstableCache).toHaveBeenCalledWith(
        expect.any(Function),
        ['test-cache-key'],
        {
          tags: ['spaces-list-test-user-123'],
          revalidate: 300, // 5 minutes TTL
        }
      );
    });
  });

  describe('getSpace', () => {
    it('should return cached space data', async () => {
      const mockSpace = { id: 'space1', title: 'Test Space' };
      mockClient.getSpace.mockResolvedValue(mockSpace);
      
      mockUnstableCache.mockImplementation((fn) => fn);

      const result = await getSpace('space1');

      expect(result.ok).toBe(true);
      expect(mockUnstableCache).toHaveBeenCalledWith(
        expect.any(Function),
        ['getSpace-space1'],
        {
          tags: ['space-space1'],
          revalidate: 300,
        }
      );
    });

    it('should handle space not found errors', async () => {
      const notFoundError = new Error('Space not found');
      mockClient.getSpace.mockRejectedValue(notFoundError);
      mockSanitizeError.mockReturnValue({ code: 'NOT_FOUND', message: 'Space not found' });

      mockUnstableCache.mockImplementation((fn) => fn);

      const result = await getSpace('nonexistent');

      expect(result.ok).toBe(false);
      expect(result.error?.code).toBe('NOT_FOUND');
    });
  });

  describe('createSpace', () => {
    it('should create space and revalidate caches', async () => {
      const mockSpace = { id: 'new-space', title: 'New Space' };
      const createRequest = new CreateSpaceRequest({ title: 'New Space', description: 'Test description' });
      
      mockClient.createSpace.mockResolvedValue(mockSpace);

      const result = await createSpace(createRequest);

      expect(result.ok).toBe(true);
      expect(result.data).toEqual(mockSpace);
      expect(mockRevalidateTag).toHaveBeenCalledWith('spaces-list-test-user-123');
      expect(mockRevalidatePath).toHaveBeenCalledWith('/spaces');
    });

    it('should handle creation errors', async () => {
      const createError = new Error('Creation failed');
      const createRequest = new CreateSpaceRequest({ title: 'New Space' });
      
      mockClient.createSpace.mockRejectedValue(createError);
      mockSanitizeError.mockReturnValue({ code: 'INTERNAL', message: 'Creation failed' });

      const result = await createSpace(createRequest);

      expect(result.ok).toBe(false);
      expect(result.error?.message).toBe('Creation failed');
    });
  });

  describe('updateSpace', () => {
    it('should update space and revalidate relevant caches', async () => {
      const mockSpace = { id: 'space1', title: 'Updated Space' };
      const updateRequest = new UpdateSpaceRequest({ title: 'Updated Space' });
      
      mockClient.updateSpace.mockResolvedValue(mockSpace);

      const result = await updateSpace('space1', updateRequest);

      expect(result.ok).toBe(true);
      expect(mockRevalidateTag).toHaveBeenCalledWith('spaces-list-test-user-123');
      expect(mockRevalidateTag).toHaveBeenCalledWith('space-space1');
      expect(mockRevalidatePath).toHaveBeenCalledWith('/spaces');
      expect(mockRevalidatePath).toHaveBeenCalledWith('/spaces/space1');
    });
  });

  describe('deleteSpace', () => {
    it('should delete space and revalidate caches', async () => {
      mockClient.deleteSpace.mockResolvedValue(undefined);

      const result = await deleteSpace('space1');

      expect(result.ok).toBe(true);
      expect(mockRevalidateTag).toHaveBeenCalledWith('spaces-list-test-user-123');
      expect(mockRevalidatePath).toHaveBeenCalledWith('/spaces');
    });

    it('should handle deletion errors', async () => {
      const deleteError = new Error('Deletion failed');
      mockClient.deleteSpace.mockRejectedValue(deleteError);
      mockSanitizeError.mockReturnValue({ code: 'INTERNAL', message: 'Deletion failed' });

      const result = await deleteSpace('space1');

      expect(result.ok).toBe(false);
      expect(result.error?.message).toBe('Deletion failed');
    });
  });
});