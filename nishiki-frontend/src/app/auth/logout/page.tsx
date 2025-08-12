'use client';

import { useRouter } from 'next/navigation';
import { useEffect } from 'react';

export default function LogoutPage() {
  const router = useRouter();

  useEffect(() => {
    // Clear any local storage or session data
    localStorage.clear();
    sessionStorage.clear();

    // Redirect to home page after a short delay
    const timer = setTimeout(() => {
      router.push('/');
    }, 2000);

    return () => clearTimeout(timer);
  }, [router]);

  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="text-center">
        <div className="bg-green-100 border border-green-400 text-green-700 px-6 py-4 rounded-lg">
          <h2 className="text-xl font-semibold mb-2">Logged Out Successfully</h2>
          <p>You have been successfully logged out. Redirecting to home...</p>
        </div>
      </div>
    </div>
  );
}
