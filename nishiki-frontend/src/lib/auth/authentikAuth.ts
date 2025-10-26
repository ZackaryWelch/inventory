import { User, UserManager, WebStorageStateStore } from 'oidc-client-ts';

export interface AuthentikConfig {
  authority: string;
  client_id: string;
  metadataUrl: string;
  redirect_uri: string;
  post_logout_redirect_uri: string;
  response_type: string;
  scope: string;
}

export interface AuthUser {
  id: string;
  username: string;
  email: string;
  groups: string[];
  accessToken: string;
  refreshToken?: string;
}

class AuthentikAuthService {
  private userManager: UserManager;
  private config: AuthentikConfig;

  constructor() {
    const authentikUrl = process.env.NEXT_PUBLIC_AUTHENTIK_URL;
    const clientId = process.env.NEXT_PUBLIC_AUTHENTIK_CLIENT_ID || 'nishiki';
    const apiBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:3001';

    this.config = {
      authority: `${authentikUrl}/application/o/nishiki/`,
      client_id: clientId,
      metadataUrl: `${apiBaseUrl}/auth/oidc-config?client_id=${clientId}`,
      redirect_uri: `${process.env.NEXT_PUBLIC_APP_URL || 'http://localhost:3000'}/auth/callback`,
      post_logout_redirect_uri: `${process.env.NEXT_PUBLIC_APP_URL || 'http://localhost:3000'}/auth/logout`,
      response_type: 'code',
      scope: 'openid profile email groups',
    };

    this.userManager = new UserManager({
      ...this.config,
      userStore:
        typeof window !== 'undefined'
          ? new WebStorageStateStore({ store: window.localStorage })
          : undefined,
      automaticSilentRenew: true,
      silent_redirect_uri: `${process.env.NEXT_PUBLIC_APP_URL || 'http://localhost:3000'}/auth/silent-callback`,
    });
  }

  /**
   * Initiate login flow
   */
  async login(): Promise<void> {
    await this.userManager.signinRedirect();
  }

  /**
   * Complete login flow after redirect
   */
  async completeLogin(): Promise<AuthUser | null> {
    try {
      const user = await this.userManager.signinRedirectCallback();
      return this.mapUserToAuthUser(user);
    } catch (error) {
      console.error('Login completion failed:', error);
      return null;
    }
  }

  /**
   * Logout user
   */
  async logout(): Promise<void> {
    await this.userManager.signoutRedirect();
  }

  /**
   * Get current user
   */
  async getUser(): Promise<AuthUser | null> {
    try {
      const user = await this.userManager.getUser();
      return user ? this.mapUserToAuthUser(user) : null;
    } catch (error) {
      console.error('Failed to get user:', error);
      return null;
    }
  }

  /**
   * Check if user is authenticated
   */
  async isAuthenticated(): Promise<boolean> {
    const user = await this.getUser();
    return user !== null && !this.isTokenExpired(user.accessToken);
  }

  /**
   * Get access token for API calls
   */
  async getAccessToken(): Promise<string | null> {
    const user = await this.getUser();
    return user?.accessToken || null;
  }

  /**
   * Refresh token silently
   */
  async refreshToken(): Promise<AuthUser | null> {
    try {
      const user = await this.userManager.signinSilent();
      return user ? this.mapUserToAuthUser(user) : null;
    } catch (error) {
      console.error('Silent refresh failed:', error);
      return null;
    }
  }

  /**
   * Handle silent callback
   */
  async completeSilentLogin(): Promise<void> {
    try {
      await this.userManager.signinSilentCallback();
    } catch (error) {
      console.error('Silent callback failed:', error);
    }
  }

  /**
   * Check if user has specific group
   */
  async hasGroup(groupName: string): Promise<boolean> {
    const user = await this.getUser();
    return user?.groups.includes(groupName) || false;
  }

  /**
   * Get user groups
   */
  async getUserGroups(): Promise<string[]> {
    const user = await this.getUser();
    return user?.groups || [];
  }

  /**
   * Map OIDC User to AuthUser
   */
  private mapUserToAuthUser(user: User): AuthUser {
    return {
      id: user.profile.sub || '',
      username: user.profile.preferred_username || '',
      email: user.profile.email || '',
      groups: Array.isArray(user.profile.groups) ? user.profile.groups : [],
      accessToken: user.access_token,
      refreshToken: user.refresh_token,
    };
  }

  
  /**
   * Check if token is expired
   */
  private isTokenExpired(token: string): boolean {
    try {
      const parts = token.split('.');
      if (parts.length !== 3) return true;
      const payload = JSON.parse(atob(parts[1] as string));
      const now = Math.floor(Date.now() / 1000);
      return payload.exp < now;
    } catch {
      return true;
    }
  }

  /**
   * Setup event listeners for authentication events
   */
  setupEventListeners() {
    this.userManager.events.addUserLoaded((user) => {
      console.log('User loaded:', user.profile);
    });

    this.userManager.events.addUserUnloaded(() => {
      console.log('User unloaded');
    });

    this.userManager.events.addAccessTokenExpiring(() => {
      console.log('Access token expiring');
    });

    this.userManager.events.addAccessTokenExpired(() => {
      console.log('Access token expired');
    });

    this.userManager.events.addSilentRenewError((error) => {
      console.error('Silent renew error:', error);
    });
  }
}

// Export singleton instance
export const authentikAuth = new AuthentikAuthService();
export default AuthentikAuthService;
