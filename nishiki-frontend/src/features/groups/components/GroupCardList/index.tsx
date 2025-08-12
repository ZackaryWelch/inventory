'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';

import { getGroupList } from '@/lib/api/group/client';
import { IGroup } from '@/types/definition';

import { GroupCard } from './GroupCard';

export const GroupCardList = () => {
  const [groups, setGroups] = useState<IGroup[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const router = useRouter();

  useEffect(() => {
    const fetchGroups = async () => {
      try {
        const groupsResult = await getGroupList();
        if (!groupsResult.ok) {
          // If authentication failed, redirect to login
          if (groupsResult.error?.includes('Authentication token is not available')) {
            router.push('/login');
            return;
          }
          throw new Error(groupsResult.error || 'Failed to fetch groups');
        }
        setGroups(groupsResult.value);
      } catch (err) {
        console.error('Failed to fetch groups:', err);
        setError(err instanceof Error ? err.message : 'Failed to fetch groups');
      } finally {
        setLoading(false);
      }
    };

    fetchGroups();
  }, [router]);

  if (loading) {
    return (
      <div className="flex flex-col gap-2">
        <div className="animate-pulse bg-gray-200 h-20 rounded-lg"></div>
        <div className="animate-pulse bg-gray-200 h-20 rounded-lg"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-red-600 p-4 border border-red-200 rounded-lg">
        Error: {error}
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-2">
      {groups.map((group) => (
        <GroupCard key={group.id} groupId={group.id} groupName={group.name} />
      ))}
    </div>
  );
};
