'use client';

import { authentikAuth, AuthUser } from '@/lib/auth/authentikAuth';

import { createContext, ReactNode, useContext, useEffect, useState } from 'react';

interface AuthContextType {
  user: AuthUser | null;
  loading: boolean;
  login: () => Promise<void>;
  logout: () => Promise<void>;
  isAuthenticated: boolean;
  getAccessToken: () => Promise<string | null>;
  hasGroup: (groupName: string) => Promise<boolean>;
  refreshToken: () => Promise<void>;
  refreshAuthState: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [loading, setLoading] = useState(true);
  const [isAuthenticated, setIsAuthenticated] = useState(false);

  // Debug user state changes
  useEffect(() => {
    console.log('AuthContext user state changed:', { user: !!user, loading, isAuthenticated });
  }, [user, loading, isAuthenticated]);

  useEffect(() => {
    initializeAuth();
    setupEventListeners();
  }, []);

  const initializeAuth = async () => {
    try {
      const currentUser = await authentikAuth.getUser();
      const authenticated = await authentikAuth.isAuthenticated();
      setUser(currentUser);
      setIsAuthenticated(authenticated);
    } catch (error) {
      console.error('Failed to initialize auth:', error);
      setUser(null);
      setIsAuthenticated(false);
    } finally {
      setLoading(false);
    }
  };

  const setupEventListeners = () => {
    authentikAuth.setupEventListeners();
  };

  const login = async () => {
    try {
      await authentikAuth.login();
    } catch (error) {
      console.error('Login failed:', error);
      throw error;
    }
  };

  const logout = async () => {
    try {
      await authentikAuth.logout();
      setUser(null);
      setIsAuthenticated(false);
    } catch (error) {
      console.error('Logout failed:', error);
      throw error;
    }
  };

  const getAccessToken = async () => {
    return await authentikAuth.getAccessToken();
  };

  const hasGroup = async (groupName: string) => {
    return await authentikAuth.hasGroup(groupName);
  };

  const refreshToken = async () => {
    try {
      const refreshedUser = await authentikAuth.refreshToken();
      const authenticated = refreshedUser ? await authentikAuth.isAuthenticated() : false;
      setUser(refreshedUser);
      setIsAuthenticated(authenticated);
    } catch (error) {
      console.error('Token refresh failed:', error);
      setUser(null);
      setIsAuthenticated(false);
    }
  };

  const refreshAuthState = async () => {
    try {
      const currentUser = await authentikAuth.getUser();
      const authenticated = await authentikAuth.isAuthenticated();
      console.log('AuthContext refreshAuthState result:', { currentUser: !!currentUser, authenticated });
      setUser(currentUser);
      setIsAuthenticated(authenticated);
    } catch (error) {
      console.error('Failed to refresh auth state:', error);
      setUser(null);
      setIsAuthenticated(false);
    }
  };

  const value: AuthContextType = {
    user,
    loading,
    login,
    logout,
    isAuthenticated,
    getAccessToken,
    hasGroup,
    refreshToken,
    refreshAuthState,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

export default AuthContext;
