const NODE_ENV = process.env.NODE_ENV || '';
const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || '';
const AUTHENTIK_URL = process.env.NEXT_PUBLIC_AUTHENTIK_URL || '';
const AUTHENTIK_CLIENT_ID = process.env.NEXT_PUBLIC_AUTHENTIK_CLIENT_ID || '';

/**
 * Check if the API is a mock API.
 * @returns true if the API is a mock API, false otherwise
 */
export const isMockApi = (): boolean => {
  return API_BASE_URL.includes(':8080') || API_BASE_URL.includes(':9080');
};

/**
 * Check if authentication is required.
 * If the environment is development and Authentik is not configured, then authentication is not required.
 * This allows for development with mock APIs without authentication.
 *
 * @returns true if authentication is required, false otherwise
 */
export const authRequired = () => {
  if (NODE_ENV === 'development' && (!AUTHENTIK_URL || !AUTHENTIK_CLIENT_ID)) {
    return false;
  }
  return true;
};

/**
 * Check if Authentik is properly configured
 * @returns true if Authentik is configured, false otherwise
 */
export const isAuthentikConfigured = (): boolean => {
  return !!(AUTHENTIK_URL && AUTHENTIK_CLIENT_ID);
};

/**
 * Get Authentik configuration
 * @returns Authentik configuration object
 */
export const getAuthentikConfig = () => {
  return {
    url: AUTHENTIK_URL,
    clientId: AUTHENTIK_CLIENT_ID,
    appUrl: process.env.NEXT_PUBLIC_APP_URL || 'http://localhost:3000',
  };
};
