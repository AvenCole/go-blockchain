import { createFileRoute } from '@tanstack/react-router'
import NetworkRoutePage from '@/features/workbench/route-pages/NetworkRoutePage'

export const Route = createFileRoute('/network')({
  component: NetworkRoutePage,
})
