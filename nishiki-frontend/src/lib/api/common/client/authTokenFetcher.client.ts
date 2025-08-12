'use client';

import { authentikAuth } from '@/lib/auth/authentikAuth';

/**
 * Get access token for API calls (client-side)
 * @returns Bearer token string or null
 */
export const getToken = async (): Promise<string | null> => {
  try {
    const accessToken = await authentikAuth.getAccessToken();
    return accessToken;
  } catch (error) {
    console.error('Failed to get access token:', error);
    return null;
  }
};

/**
 * Get authorization header for API calls
 * @returns Authorization header string or null
 */
export const getAuthHeader = async (): Promise<string | null> => {
  const token = await getToken();
  return token ? `Bearer ${token}` : null;
};

/**
 * Check if user is authenticated
 * @returns boolean indicating authentication status
 */
export const isAuthenticated = async (): Promise<boolean> => {
  return await authentikAuth.isAuthenticated();
};
