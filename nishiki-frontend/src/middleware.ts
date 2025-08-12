import { authRequired } from '@/utils/authUtils';

import { NextRequest, NextResponse } from 'next/server';

/**
 * Next.js middleware - simplified for client-side JWT auth
 * Authentication state is managed client-side via AuthContext
 * @param request - NextRequest
 * @returns NextResponse
 */
export const middleware = async (request: NextRequest) => {
  const response = NextResponse.next();
  const isOnLoginPage = request.nextUrl.pathname.startsWith('/login');
  const isAuthCallback = request.nextUrl.pathname.startsWith('/auth/');

  // If authentication is not required, skip the following authentication process.
  if (!authRequired()) return response;

  // Allow auth callback routes to pass through without authentication check
  if (isAuthCallback) {
    return response;
  }

  // For JWT-based auth, let client-side AuthGuard handle authentication
  // Only redirect from login page if user somehow got there
  if (isOnLoginPage && request.nextUrl.pathname === '/login') {
    return response;
  }

  return response;
};

export const config = {
  matcher: [
    /**
     * Match all request paths except for the ones starting with:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     * - auth (authentication callback routes)
     */
    '/',
    '/login/:path*',
    '/groups/:path*',
    '/foods/:path*',
    '/profile/:path*',
  ],
};
