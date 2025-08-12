'use client';

import { authentikAuth } from '@/lib/auth/authentikAuth';
import { useAuth } from '@/contexts/AuthContext';

import { useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';

export default function AuthCallbackPage() {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const router = useRouter();
  const { refreshAuthState } = useAuth();

  useEffect(() => {
    const handleCallback = async () => {
      try {
        const user = await authentikAuth.completeLogin();

        if (user) {
          console.log('Login successful:', user);

          // Refresh the auth context to update the global auth state
          await refreshAuthState();

          // Give a moment for the auth state to propagate
          await new Promise(resolve => setTimeout(resolve, 100));

          // Redirect to the intended page or home (which will redirect based on auth status)
          const returnUrl = sessionStorage.getItem('returnUrl') || '/';
          sessionStorage.removeItem('returnUrl');
          console.log('Auth callback redirecting to:', returnUrl);
          router.push(returnUrl);
        } else {
          setError('Login failed. Please try again.');
        }
      } catch (err) {
        console.error('Authentication error:', err);
        setError('Authentication failed. Please try again.');
      } finally {
        setLoading(false);
      }
    };

    handleCallback();
  }, [router]);

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <h2 className="text-xl font-semibold text-gray-900">Completing login...</h2>
          <p className="text-gray-600 mt-2">Please wait while we authenticate you.</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
            <h2 className="text-xl font-semibold mb-2">Authentication Error</h2>
            <p>{error}</p>
          </div>
          <button
            onClick={() => router.push('/login')}
            className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"
          >
            Try Again
          </button>
        </div>
      </div>
    );
  }

  return null;
}
