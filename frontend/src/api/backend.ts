import {
  Blocks,
  CreateWallet,
  Dashboard,
  MinePending,
  PendingTransactions,
  QueueTransaction,
  Wallets,
} from '../../wailsjs/go/main/App'
import type { BlockView, DashboardData, WalletView } from '../types'

export const fetchDashboard = (): Promise<DashboardData> => Dashboard()
export const fetchWallets = (): Promise<WalletView[]> => Wallets()
export const fetchBlocks = (): Promise<BlockView[]> => Blocks()
export const fetchPendingTransactions = (): Promise<string[]> => PendingTransactions()
export const createWallet = (): Promise<string> => CreateWallet()
export const queueTransaction = (from: string, to: string, amount: number, fee: number): Promise<string> =>
  QueueTransaction(from, to, amount, fee)
export const minePending = (minerAddress: string): Promise<string> => MinePending(minerAddress)
