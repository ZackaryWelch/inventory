import { GroupSinglePage } from '@/components/pages/GroupSinglePage';

export const dynamic = 'force-dynamic';

export default async function GroupSingle({ params }: { params: Promise<{ groupId: string }> }) {
  const { groupId } = await params;
  return <GroupSinglePage groupId={groupId} />;
}
