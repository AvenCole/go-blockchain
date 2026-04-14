import {
  Blocks,
  ConnectNode,
  CreateWallet,
  Dashboard,
  ExecuteCLI,
  MinePending,
  Nodes,
  PendingTransactions,
  QueueTransaction,
  StartNode,
  StopNode,
  Wallets,
} from '../../wailsjs/go/main/App'
import type { BlockView, CommandResult, DashboardData, NodeStatus, WalletView } from '../types'

export const fetchDashboard = (): Promise<DashboardData> => Dashboard()
export const fetchWallets = (): Promise<WalletView[]> => Wallets()
export const fetchBlocks = (): Promise<BlockView[]> => Blocks()
export const fetchPendingTransactions = (): Promise<string[]> => PendingTransactions()
export const createWallet = (): Promise<string> => CreateWallet()
export const queueTransaction = (from: string, to: string, amount: number, fee: number): Promise<string> =>
  QueueTransaction(from, to, amount, fee)
export const minePending = (minerAddress: string): Promise<string> => MinePending(minerAddress)
export const executeCLI = (commandLine: string): Promise<CommandResult> => ExecuteCLI(commandLine)
export const fetchNodes = (): Promise<NodeStatus[]> => Nodes()
export const startNode = (address: string, seed: string, miner: string): Promise<string> => StartNode(address, seed, miner)
export const stopNode = (address: string): Promise<void> => StopNode(address)
export const connectNode = (address: string, seed: string): Promise<void> => ConnectNode(address, seed)
