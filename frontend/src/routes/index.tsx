import { createFileRoute } from '@tanstack/react-router'
import DashboardRoutePage from '@/features/workbench/route-pages/DashboardRoutePage'

export const Route = createFileRoute('/')({
  component: DashboardRoutePage,
})
