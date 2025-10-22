'use client';

import { getContainersByGroup } from '@/lib/api/container/client';
import { fetchUserList } from '@/lib/api/user/client';
import { IContainer, IUser } from '@/types/definition';

import { FC, useEffect, useState } from 'react';

import { GroupCardContent } from './GroupCardContent';

interface IGroupCardProps {
  groupId: string;
  groupName: string;
}

export const GroupCard: FC<IGroupCardProps> = ({ groupId, groupName }) => {
  const [containerCount, setContainerCount] = useState(0);
  const [userCount, setUserCount] = useState(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [containersResult, usersResult] = await Promise.all([
          getContainersByGroup(groupId),
          fetchUserList(groupId),
        ]);

        const containers: IContainer[] = containersResult.ok ? (containersResult.value || []) : [];
        const users: IUser[] = usersResult.ok ? (usersResult.value || []) : [];

        setContainerCount(containers.length);
        setUserCount(users.length);
      } catch (error) {
        console.error('Failed to fetch group data:', error);
        // Set defaults on error
        setContainerCount(0);
        setUserCount(0);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [groupId]);

  if (loading) {
    return (
      <div className="animate-pulse bg-gray-200 h-20 rounded-lg"></div>
    );
  }

  return (
    <GroupCardContent
      groupId={groupId}
      groupName={groupName}
      containerCount={containerCount}
      userCount={userCount}
    />
  );
};
