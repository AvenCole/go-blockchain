import { createFileRoute } from '@tanstack/react-router'
import ConsoleRoutePage from '@/features/workbench/route-pages/ConsoleRoutePage'

export const Route = createFileRoute('/console')({
  component: ConsoleRoutePage,
})
