import { createFileRoute } from '@tanstack/react-router'
import WalletsRoutePage from '@/features/workbench/route-pages/WalletsRoutePage'

export const Route = createFileRoute('/wallets')({
  component: WalletsRoutePage,
})
