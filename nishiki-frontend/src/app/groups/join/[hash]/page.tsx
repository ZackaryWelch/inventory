// Note: putJoinRequest not implemented in Go backend yet

import { redirect } from 'next/navigation';

export default async function Join({ params }: { params: Promise<{ hash: string }> }) {
  const { hash } = await params;
  // TODO: Implement group join functionality in Go backend
  console.log('Join hash:', hash);
  redirect('/groups');
}
