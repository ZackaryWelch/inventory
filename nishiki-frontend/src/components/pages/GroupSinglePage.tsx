import { MobileLayout } from '@/components/layouts/MobileLayout';
import { HeaderBackButton } from '@/components/parts/Header';
import { ContainerCardList } from '@/features/groups/components/ContainerCardList';
import { GroupSingleHeaderDropdownMenu } from '@/features/groups/components/GroupSingleHeaderDropdownMenu';
import { MemberList } from '@/features/groups/components/MemberList';
// Note: getGroup function not implemented in Go backend - group info comes from containers
import { IGroup } from '@/types/definition';

interface IGroupSinglePageProps {
  /**
   * The ID of the group to display.
   */
  groupId: IGroup['id'];
}

export const GroupSinglePage = async ({ groupId }: IGroupSinglePageProps) => {
  // TODO: Get group name from first container or pass it as prop
  const groupName = 'Group'; // Temporary fallback

  return (
    <MobileLayout
      heading={groupName}
      headerLeft={<HeaderBackButton href={{ pathname: '/groups' }} />}
      headerRight={<GroupSingleHeaderDropdownMenu groupId={groupId} currentGroupName={groupName} />}
    >
      <div className="px-4 pt-6 pb-16">
        <MemberList groupId={groupId} />
        <ContainerCardList groupId={groupId} />
      </div>
    </MobileLayout>
  );
};
