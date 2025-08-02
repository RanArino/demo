'use server';

import { auth } from '@clerk/nextjs/server';
import * as grpc from '@grpc/grpc-js';
import { getUserServiceClient } from '../server-client';
import {
  CreateUserRequest,
  CreateUserResponse,
  GetUserRequest,
  GetUserResponse,
  UpdateUserRequest,
  UpdateUserResponse,
  DeleteUserRequest,
  DeleteUserResponse,
  CheckUserStatusRequest,
  CheckUserStatusResponse,
  ActivateUserRequest,
  ActivateUserResponse,
} from '../generated/v1/user_pb';

/**
 * Helper function to create gRPC metadata with JWT token
 */
async function createMetadataWithAuth(): Promise<grpc.Metadata> {
  const { getToken } = await auth();
  // Get token using the custom JWT template
  const token = await getToken({ template: 'ms-user-auth' });

  const metadata = new grpc.Metadata();
  if (token) {
    metadata.add('authorization', `Bearer ${token}`);
  }
  
  return metadata;
}

/**
 * Helper function to sanitize error messages for security
 */
function sanitizeError(error: any): string {
  const isDevelopment = process.env.NODE_ENV === 'development';
  
  if (isDevelopment) {
    // In development, provide more context but still avoid exposing sensitive data
    if (error.message?.includes('token verification failed')) {
      return 'Authentication failed. Please sign in again.';
    }
    if (error.message?.includes('UNAUTHENTICATED')) {
      return 'Authentication required. Please sign in.';
    }
    return error.message || 'An error occurred';
  } else {
    // In production, return generic error messages
    if (error.code === 16 || error.message?.includes('UNAUTHENTICATED')) {
      return 'Authentication failed. Please sign in again.';
    }
    return 'An error occurred. Please try again.';
  }
}

/**
 * Server action to create a new user
 */
export async function createUser(userData: {
  email: string;
  fullName: string;
  username: string;
  role?: string;
}): Promise<{ success: boolean; user?: any; error?: string }> {
  try {
    const { userId } = await auth();
    if (!userId) {
      return { success: false, error: 'Not authenticated' };
    }

    const client = getUserServiceClient();
    const metadata = await createMetadataWithAuth();
    const request = new CreateUserRequest();
    
    request.setClerkUserId(userId);
    request.setEmail(userData.email);
    request.setFullName(userData.fullName);
    request.setUsername(userData.username);
    request.setRole(userData.role || 'user');

    return new Promise((resolve) => {
      client.createUser(request, metadata, (error: any, response: CreateUserResponse) => {
        if (error) {
          console.error('gRPC createUser error:', error);
          resolve({ success: false, error: sanitizeError(error) });
          return;
        }

        const user = response.getUser();
        if (user) {
          resolve({
            success: true,
            user: {
              id: user.getId(),
              clerkUserId: user.getClerkUserId(),
              email: user.getEmail(),
              fullName: user.getFullName(),
              username: user.getUsername(),
              role: user.getRole(),
              status: user.getStatus(),
            },
          });
        } else {
          resolve({ success: false, error: 'No user returned' });
        }
      });
    });
  } catch (error) {
    console.error('createUser action error:', error);
    return { success: false, error: sanitizeError(error) };
  }
}

/**
 * Server action to activate a user's profile
 */
export async function activateUser(userData: {
  fullName: string;
  username: string;
}): Promise<{ success: boolean; user?: any; error?: string }> {
  try {
    const { userId } = await auth();
    if (!userId) {
      return { success: false, error: 'Not authenticated' };
    }

    const client = getUserServiceClient();
    const metadata = await createMetadataWithAuth();
    const request = new ActivateUserRequest();
    
    request.setFullName(userData.fullName);
    request.setUsername(userData.username);

    return new Promise((resolve) => {
      client.activateUser(request, metadata, (error: any, response: ActivateUserResponse) => {
        if (error) {
          console.error('gRPC activateUser error:', error);
          resolve({ success: false, error: sanitizeError(error) });
          return;
        }

        const user = response.getUser();
        if (user) {
          resolve({
            success: true,
            user: {
              id: user.getId(),
              clerkUserId: user.getClerkUserId(),
              email: user.getEmail(),
              fullName: user.getFullName(),
              username: user.getUsername(),
              role: user.getRole(),
              status: user.getStatus(),
            },
          });
        } else {
          resolve({ success: false, error: 'No user returned' });
        }
      });
    });
  } catch (error) {
    console.error('activateUser action error:', error);
    return { success: false, error: sanitizeError(error) };
  }
}


/**
 * Server action to get user details
 */
export async function getUser(userId?: string): Promise<{ success: boolean; user?: any; error?: string }> {
  try {
    const { userId: authUserId } = await auth();
    if (!authUserId) {
      return { success: false, error: 'Not authenticated' };
    }

    const client = getUserServiceClient();
    const metadata = await createMetadataWithAuth();
    const request = new GetUserRequest();
    request.setUserId(userId || authUserId);

    return new Promise((resolve) => {
      client.getUser(request, metadata, (error: any, response: GetUserResponse) => {
        if (error) {
          console.error('gRPC getUser error:', error);
          resolve({ success: false, error: sanitizeError(error) });
          return;
        }

        const user = response.getUser();
        if (user) {
          resolve({
            success: true,
            user: {
              id: user.getId(),
              clerkUserId: user.getClerkUserId(),
              email: user.getEmail(),
              fullName: user.getFullName(),
              username: user.getUsername(),
              role: user.getRole(),
              status: user.getStatus(),
              storageUsedBytes: user.getStorageUsedBytes(),
              storageQuotaBytes: user.getStorageQuotaBytes(),
              createdAt: user.getCreatedAt()?.toDate(),
              updatedAt: user.getUpdatedAt()?.toDate(),
            },
          });
        } else {
          resolve({ success: false, error: 'User not found' });
        }
      });
    });
  } catch (error) {
    console.error('getUser action error:', error);
    return { success: false, error: sanitizeError(error) };
  }
}

/**
 * Server action to update user information
 */
export async function updateUser(userData: {
  userId?: string;
  email?: string;
  fullName?: string;
  username?: string;
  role?: string;
}): Promise<{ success: boolean; user?: any; error?: string }> {
  try {
    const { userId: authUserId } = await auth();
    if (!authUserId) {
      return { success: false, error: 'Not authenticated' };
    }

    const client = getUserServiceClient();
    const metadata = await createMetadataWithAuth();
    const request = new UpdateUserRequest();
    
    request.setUserId(userData.userId || authUserId);
    if (userData.email) request.setEmail(userData.email);
    if (userData.fullName) request.setFullName(userData.fullName);
    if (userData.username) request.setUsername(userData.username);
    if (userData.role) request.setRole(userData.role);

    return new Promise((resolve) => {
      client.updateUser(request, metadata, (error: any, response: UpdateUserResponse) => {
        if (error) {
          console.error('gRPC updateUser error:', error);
          resolve({ success: false, error: sanitizeError(error) });
          return;
        }

        const user = response.getUser();
        if (user) {
          resolve({
            success: true,
            user: {
              id: user.getId(),
              clerkUserId: user.getClerkUserId(),
              email: user.getEmail(),
              fullName: user.getFullName(),
              username: user.getUsername(),
              role: user.getRole(),
              status: user.getStatus(),
              updatedAt: user.getUpdatedAt()?.toDate(),
            },
          });
        } else {
          resolve({ success: false, error: 'No user returned' });
        }
      });
    });
  } catch (error) {
    console.error('updateUser action error:', error);
    return { success: false, error: sanitizeError(error) };
  }
}

/**
 * Server action to check user status
 */
export async function checkUserStatus(): Promise<{ 
  success: boolean; 
  profileCompleted?: boolean; 
  needsRedirect?: boolean; 
  redirectUrl?: string; 
  user?: any; 
  error?: string; 
}> {
  try {
    const { userId } = await auth();
    if (!userId) {
      return { success: false, error: 'Not authenticated' };
    }

    const client = getUserServiceClient();
    const metadata = await createMetadataWithAuth();
    const request = new CheckUserStatusRequest();

    return new Promise((resolve) => {
      client.checkUserStatus(request, metadata, (error: any, response: CheckUserStatusResponse) => {
        if (error) {
          console.error('gRPC checkUserStatus error:', error);
          resolve({ success: false, error: sanitizeError(error) });
          return;
        }

        const user = response.getUser();
        resolve({
          success: true,
          profileCompleted: response.getProfileCompleted(),
          needsRedirect: response.getNeedsRedirect(),
          redirectUrl: response.getRedirectUrl(),
          user: user ? {
            id: user.getId(),
            clerkUserId: user.getClerkUserId(),
            email: user.getEmail(),
            fullName: user.getFullName(),
            username: user.getUsername(),
            role: user.getRole(),
            status: user.getStatus(),
          } : undefined,
        });
      });
    });
  } catch (error) {
    console.error('checkUserStatus action error:', error);
    return { success: false, error: sanitizeError(error) };
  }
}

/**
 * Server action to delete user (soft delete)
 */
export async function deleteUser(userId?: string): Promise<{ success: boolean; error?: string }> {
  try {
    const { userId: authUserId } = await auth();
    if (!authUserId) {
      return { success: false, error: 'Not authenticated' };
    }

    const client = getUserServiceClient();
    const metadata = await createMetadataWithAuth();
    const request = new DeleteUserRequest();
    request.setUserId(userId || authUserId);

    return new Promise((resolve) => {
      client.deleteUser(request, metadata, (error: any, response: DeleteUserResponse) => {
        if (error) {
          console.error('gRPC deleteUser error:', error);
          resolve({ success: false, error: sanitizeError(error) });
          return;
        }

        resolve({ success: true });
      });
    });
  } catch (error) {
    console.error('deleteUser action error:', error);
    return { success: false, error: sanitizeError(error) };
  }
}