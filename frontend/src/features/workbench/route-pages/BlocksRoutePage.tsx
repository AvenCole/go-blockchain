import BlocksPage from '@/pages/BlocksPage'
import { useWorkbench } from '@/features/workbench/context/useWorkbench'

function BlocksRoutePage() {
  const { blocks } = useWorkbench()

  return <BlocksPage blocks={blocks} />
}

export default BlocksRoutePage
