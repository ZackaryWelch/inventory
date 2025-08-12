'use client';

import { authentikAuth } from '@/lib/auth/authentikAuth';

import { useEffect } from 'react';

export default function SilentCallbackPage() {
  useEffect(() => {
    const handleSilentCallback = async () => {
      try {
        await authentikAuth.completeSilentLogin();
      } catch (error) {
        console.error('Silent callback error:', error);
      }
    };

    handleSilentCallback();
  }, []);

  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="text-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mb-2"></div>
        <p className="text-sm text-gray-600">Refreshing session...</p>
      </div>
    </div>
  );
}
