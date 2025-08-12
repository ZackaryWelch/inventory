'use client';

import { IconPersonCircle } from '@/assets/images/icons';
import { Card, Icon } from '@/components/ui';
import { fetchUserList } from '@/lib/api/user/client';
import { IGroup, IUser } from '@/types/definition';

import { useEffect, useState } from 'react';

import { MemberCardDropdownMenuTriggerButton } from './MemberCardDropdownMenuTriggerButton';

interface IMembersPageProps {
  /**
   * an identifier of a group
   */
  groupId: IGroup['id'];
}

export const MemberCardList = ({ groupId }: IMembersPageProps) => {
  const [users, setUsers] = useState<IUser[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchUsers = async () => {
      try {
        const usersResult = await fetchUserList(groupId);
        const users: IUser[] = usersResult.ok ? usersResult.value : [];
        setUsers(users);
      } catch (error) {
        console.error('Failed to fetch users:', error);
        setUsers([]);
      } finally {
        setLoading(false);
      }
    };

    fetchUsers();
  }, [groupId]);

  if (loading) {
    return (
      <div className="flex flex-col gap-2 pb-1">
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={i} className="animate-pulse bg-gray-200 h-16 rounded-lg"></div>
        ))}
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-2 pb-1">
      {users.map((user, i) => (
        <Card key={i} asChild>
          <div className="flex justify-between gap-2">
            <div className="flex grow gap-4 items-center pl-4 py-2">
              <Icon icon={IconPersonCircle} color="gray" size={11} />
              <span className="leading-5">{user.name}</span>
            </div>
            <MemberCardDropdownMenuTriggerButton userId={user.id} groupId={groupId} />
          </div>
        </Card>
      ))}
    </div>
  );
};
