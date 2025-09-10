'use server';

import { auth } from '@clerk/nextjs/server';
import { getUserServiceClient } from '../server-client';
import {
  CreateUserRequest,
  GetUserRequest,
  UpdateUserRequest,
  DeleteUserRequest,
  CheckUserStatusRequest,
  ActivateUserRequest,
  User,
} from '../generated/v1/user_pb';
import { createAuthHeaders, sanitizeErrorString, sanitizeProtobufForJson } from './utils';

// Sanitized User type with BigInt fields converted to numbers
export type SanitizedUser = Omit<User, 'storageUsedBytes' | 'storageQuotaBytes'> & {
  storageUsedBytes: number;
  storageQuotaBytes: number;
};

/**
 * Server action to create a new user
 */
export async function createUser(userData: {
  email: string;
  fullName: string;
  username: string;
  role?: string;
}): Promise<{ success: boolean; user?: SanitizedUser; error?: string }> {
  try {
    const { userId } = await auth();
    if (!userId) {
      return { success: false, error: 'Not authenticated' };
    }

    const client = getUserServiceClient();
    const headers = await createAuthHeaders();
    
    const request = new CreateUserRequest({
      clerkUserId: userId,
      email: userData.email,
      // fullName: userData.fullName,
      // username: userData.username,
      // role: userData.role || 'user',
    });

    const response = await client.createUser(request, { headers });

    if (response.user) {
      return {
        success: true,
        user: sanitizeProtobufForJson(response.user) as unknown as SanitizedUser,
      };
    } else {
      return { success: false, error: 'No user returned' };
    }
  } catch (error) {
    console.error('createUser action error:', error);
    return { success: false, error: sanitizeErrorString(error) };
  }
}

/**
 * Server action to activate a user's profile
 */
export async function activateUser(userData: {
  fullName: string;
  username: string;
}): Promise<{ success: boolean; user?: SanitizedUser; error?: string }> {
  try {
    const { userId } = await auth();
    if (!userId) {
      return { success: false, error: 'Not authenticated' };
    }

    const client = getUserServiceClient();
    const headers = await createAuthHeaders();
    
    const request = new ActivateUserRequest({
      fullName: userData.fullName,
      username: userData.username,
    });

    const response = await client.activateUser(request, { headers });

    if (response.user) {
      return {
        success: true,
        user: sanitizeProtobufForJson(response.user) as unknown as SanitizedUser,
      };
    } else {
      return { success: false, error: 'No user returned' };
    }
  } catch (error) {
    console.error('activateUser action error:', error);
    return { success: false, error: sanitizeErrorString(error) };
  }
}


/**
 * Server action to get user details
 */
export async function getUser(userId?: string): Promise<{ success: boolean; user?: SanitizedUser; error?: string }> {
  try {
    const { userId: authUserId } = await auth();
    if (!authUserId) {
      return { success: false, error: 'Not authenticated' };
    }

    const client = getUserServiceClient();
    const headers = await createAuthHeaders();
    
    // Determine which type of ID we're using
    const targetUserId = userId || authUserId;
    
    // Helper function to detect if a string is a UUID (internal user ID)
    const isUUID = (str: string): boolean => {
      const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
      return uuidRegex.test(str);
    };
    
    // Determine the identifier type:
    const isInternalUserId = userId && isUUID(userId);
    
    const request = new GetUserRequest({ 
      identifier: { 
        case: isInternalUserId ? "userId" : "clerkUserId", 
        value: targetUserId 
      } 
    });

    const response = await client.getUser(request, { headers });

    if (response.user) {
      return {
        success: true,
        user: sanitizeProtobufForJson(response.user) as unknown as SanitizedUser,
      };
    } else {
      return { success: false, error: 'User not found' };
    }
  } catch (error) {
    console.error('getUser action error:', error);
    return { success: false, error: sanitizeErrorString(error) };
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
}): Promise<{ success: boolean; user?: SanitizedUser; error?: string }> {
  try {
    const { userId: authUserId } = await auth();
    if (!authUserId) {
      return { success: false, error: 'Not authenticated' };
    }

    const client = getUserServiceClient();
    const headers = await createAuthHeaders();
    
    const request = new UpdateUserRequest({
      email: userData.email,
      fullName: userData.fullName,
      username: userData.username,
      role: userData.role,
    });

    const response = await client.updateUser(request, { headers });

    if (response.user) {
      return {
        success: true,
        user: sanitizeProtobufForJson(response.user) as unknown as SanitizedUser,
      };
    } else {
      return { success: false, error: 'No user returned' };
    }
  } catch (error) {
    console.error('updateUser action error:', error);
    return { success: false, error: sanitizeErrorString(error) };
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
  user?: SanitizedUser; 
  error?: string; 
}> {
  try {
    const { userId } = await auth();
    if (!userId) {
      return { success: false, error: 'Not authenticated' };
    }

    const client = getUserServiceClient();
    const headers = await createAuthHeaders();
    const request = new CheckUserStatusRequest();

    const response = await client.checkUserStatus(request, { headers });

    return {
      success: true,
      profileCompleted: response.profileCompleted,
      needsRedirect: response.needsRedirect,
      redirectUrl: response.redirectUrl,
      user: response.user ? (sanitizeProtobufForJson(response.user) as unknown as SanitizedUser) : undefined,
    };
  } catch (error) {
    console.error('checkUserStatus action error:', error);
    return { success: false, error: sanitizeErrorString(error) };
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
    const headers = await createAuthHeaders();
    
    // If userId is provided, use it; otherwise delete current authenticated user
    const request = new DeleteUserRequest({
      userId: userId || ""
    });

    await client.deleteUser(request, { headers });

    return { success: true };
  } catch (error) {
    console.error('deleteUser action error:', error);
    return { success: false, error: sanitizeErrorString(error) };
  }
}
