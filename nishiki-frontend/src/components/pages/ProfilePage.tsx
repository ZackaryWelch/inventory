'use client';

import { useEffect, useState } from 'react';

import { MobileLayout } from '@/components/layouts/MobileLayout';
import { ProfileHeaderDropdownMenuTriggerButton, UserProfile } from '@/features/profile/components';
import { getCurrentUser } from '@/lib/api/auth/client';

export const ProfilePage = () => {
  const [userInfo, setUserInfo] = useState<{ userId: string; name: string } | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const getUserInfo = async () => {
      try {
        const getCurrentUserResult = await getCurrentUser();
        if (getCurrentUserResult.ok) {
          const { id: userId, name } = getCurrentUserResult.value;
          setUserInfo({ userId, name });
        } else {
          throw new Error('Failed to get user info');
        }
      } catch (err) {
        console.error('Failed to get user info:', err);
        setError(err instanceof Error ? err.message : 'Failed to get user info');
      } finally {
        setLoading(false);
      }
    };

    getUserInfo();
  }, []);

  if (loading) {
    return (
      <MobileLayout heading="Profile">
        <div className="flex items-center justify-center p-8">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        </div>
      </MobileLayout>
    );
  }

  if (error || !userInfo) {
    return (
      <MobileLayout heading="Profile">
        <div className="text-red-600 p-4 border border-red-200 rounded-lg m-4">
          Error: {error || 'Failed to load user info'}
        </div>
      </MobileLayout>
    );
  }

  return (
    <MobileLayout
      heading="Profile"
      headerRight={<ProfileHeaderDropdownMenuTriggerButton userId={userInfo.userId} />}
    >
      <UserProfile userId={userInfo.userId} name={userInfo.name} />
    </MobileLayout>
  );
};
