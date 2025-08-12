import { MembersPage } from '@/components/pages/MembersPage';

export const dynamic = 'force-dynamic';

export default async function Members({ params }: { params: Promise<{ groupId: string }> }) {
  const { groupId } = await params;
  return <MembersPage groupId={groupId} />;
}
