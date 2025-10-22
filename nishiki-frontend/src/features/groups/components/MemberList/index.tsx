'use client';

import { IconPersonCircle } from '@/assets/images/icons';
import { Icon } from '@/components/ui';
import { fetchUserList } from '@/lib/api/user/client';
import { IGroup, IUser } from '@/types/definition';

import Link from 'next/link';
import { useEffect, useState } from 'react';

import { InviteMemberDialogTrigger } from './InviteMemberDialogTrigger';

interface IMemberListProps {
  /**
   * an identifier of a group which members belong to
   */
  groupId: IGroup['id'];
}

export const MemberList = ({ groupId }: IMemberListProps) => {
  const [users, setUsers] = useState<IUser[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchUsers = async () => {
      try {
        const usersResult = await fetchUserList(groupId);
        const users: IUser[] = usersResult.ok ? (usersResult.value || []) : [];
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

  return (
    <div className="mb-6">
      <div className="flex items-center justify-between mb-2 h-12">
        <h2 className="text-xl">Members</h2>
        <InviteMemberDialogTrigger groupId={groupId} />
      </div>
      <Link href={`/groups/${groupId}/members`} className="flex flex-row gap-2">
        {loading ? (
          // Show placeholder icons while loading
          Array.from({ length: 3 }).map((_, idx) => (
            <div key={idx} className="animate-pulse bg-gray-200 w-10 h-10 rounded-full"></div>
          ))
        ) : (
          users.map((_, idx) => (
            <Icon key={idx} icon={IconPersonCircle} color="gray" size={10} />
          ))
        )}
      </Link>
    </div>
  );
};
