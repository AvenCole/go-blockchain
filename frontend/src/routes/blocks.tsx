import { createFileRoute } from '@tanstack/react-router'
import BlocksRoutePage from '@/features/workbench/route-pages/BlocksRoutePage'

export const Route = createFileRoute('/blocks')({
  component: BlocksRoutePage,
})
