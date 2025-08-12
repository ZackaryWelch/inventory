'use client';

import { getContainersByGroup } from '@/lib/api/container/client';
import { IContainer, IGroup } from '@/types/definition';

import { useEffect, useState } from 'react';

import { ContainerCard } from './ContainerCard';
import { CreateContainerButton } from './CreateContainerButton';

interface IContainerListProps {
  /**
   * an identifier of a group which a list of containers belongs to
   */
  groupId: IGroup['id'];
}

export const ContainerCardList = ({ groupId }: IContainerListProps) => {
  const [containers, setContainers] = useState<IContainer[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchContainers = async () => {
      try {
        const containersResult = await getContainersByGroup(groupId);
        const containers: IContainer[] = containersResult.ok ? containersResult.value : [];
        setContainers(containers);
      } catch (error) {
        console.error('Failed to fetch containers:', error);
        setContainers([]);
      } finally {
        setLoading(false);
      }
    };

    fetchContainers();
  }, [groupId]);

  if (loading) {
    return (
      <>
        <div className="flex items-center justify-between mb-2 h-12">
          <h2 className="text-xl">Container</h2>
          <div className="flex gap-0.5">
            <CreateContainerButton groupId={groupId} />
          </div>
        </div>
        <div className="flex flex-col gap-3">
          <div className="animate-pulse bg-gray-200 h-20 rounded-lg"></div>
          <div className="animate-pulse bg-gray-200 h-20 rounded-lg"></div>
        </div>
      </>
    );
  }

  return (
    <>
      <div className="flex items-center justify-between mb-2 h-12">
        <h2 className="text-xl">Container</h2>
        <div className="flex gap-0.5">
          <CreateContainerButton groupId={groupId} />
        </div>
      </div>
      <div className="flex flex-col gap-3">
        {containers.map((container) => (
          <ContainerCard
            key={container.id}
            containerId={container.id}
            groupId={groupId}
            containerName={container.name}
          />
        ))}
      </div>
    </>
  );
};
