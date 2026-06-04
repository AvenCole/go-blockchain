import { createFileRoute } from '@tanstack/react-router'
import TransactionsRoutePage from '@/features/workbench/route-pages/TransactionsRoutePage'

export const Route = createFileRoute('/transactions')({
  component: TransactionsRoutePage,
})
