'use client';

import { useAuth } from '@/contexts/AuthContext';

import { useRouter } from 'next/navigation';
import { useEffect } from 'react';

export default function Home() {
  const { isAuthenticated, loading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    console.log('Root page auth state:', { isAuthenticated, loading });
    if (!loading) {
      if (isAuthenticated) {
        console.log('User is authenticated, redirecting to /groups');
        router.push('/groups');
      } else {
        console.log('User is not authenticated, redirecting to /login');
        router.push('/login');
      }
    }
  }, [isAuthenticated, loading, router]);

  // Show loading while determining auth status
  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading...</p>
        </div>
      </div>
    );
  }

  return null; // Component will redirect before rendering
}
