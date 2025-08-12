import { FoodsPage } from '@/components/pages/FoodsPage';
// Note: fetchAllContainerList is not available in Go backend - containers are fetched by group
import { IContainer } from '@/types/definition';

export const dynamic = 'force-dynamic';

export default async function Foods() {
  // TODO: Need to fetch all user's groups first, then get containers for each group
  // This requires updating the Go backend or changing the UI approach
  const containers: IContainer[] = [];
  return <FoodsPage containers={containers} />;
}
