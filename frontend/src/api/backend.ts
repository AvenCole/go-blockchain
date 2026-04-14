import {
  Blocks,
  ConnectNode,
  CreateWallet,
  Dashboard,
  ExecuteCLI,
  InitializeNodeBlockchain,
  MinePending,
  MineNodePending,
  MultiSigOutputs,
  Nodes,
  PendingTransactions,
  QueueSpendMultiSigTransaction,
  QueueTransaction,
  QueueP2PKTransaction,
  QueueMultiSigTransaction,
  StartNode,
  StopNode,
  SubmitNodeTransaction,
  Wallets,
} from '../../wailsjs/go/main/App'
import type { BlockView, CommandResult, DashboardData, MultiSigOutputView, NodeStatus, WalletView } from '../types'

export const fetchDashboard = (): Promise<DashboardData> => Dashboard()
export const fetchWallets = (): Promise<WalletView[]> => Wallets()
export const fetchBlocks = (): Promise<BlockView[]> => Blocks()
export const fetchPendingTransactions = (): Promise<string[]> => PendingTransactions()
export const fetchMultiSigOutputs = (): Promise<MultiSigOutputView[]> => MultiSigOutputs()
export const createWallet = (): Promise<string> => CreateWallet()
export const queueTransaction = (from: string, to: string, amount: number, fee: number): Promise<string> =>
  QueueTransaction(from, to, amount, fee)
export const queueP2PKTransaction = (from: string, to: string, amount: number, fee: number): Promise<string> =>
  QueueP2PKTransaction(from, to, amount, fee)
export const queueMultiSigTransaction = (
  from: string,
  recipientsCSV: string,
  required: number,
  amount: number,
  fee: number,
): Promise<string> => QueueMultiSigTransaction(from, recipientsCSV, required, amount, fee)
export const queueSpendMultiSigTransaction = (
  signersCSV: string,
  sourceTxID: string,
  out: number,
  to: string,
  amount: number,
  fee: number,
): Promise<string> => QueueSpendMultiSigTransaction(signersCSV, sourceTxID, out, to, amount, fee)
export const minePending = (minerAddress: string): Promise<string> => MinePending(minerAddress)
export const executeCLI = (commandLine: string): Promise<CommandResult> => ExecuteCLI(commandLine)
export const fetchNodes = (): Promise<NodeStatus[]> => Nodes()
export const startNode = (address: string, seed: string, miner: string): Promise<string> => StartNode(address, seed, miner)
export const stopNode = (address: string): Promise<void> => StopNode(address)
export const connectNode = (address: string, seed: string): Promise<void> => ConnectNode(address, seed)
export const initializeNodeBlockchain = (address: string, rewardAddress: string): Promise<void> =>
  InitializeNodeBlockchain(address, rewardAddress)
export const submitNodeTransaction = (
  nodeAddress: string,
  from: string,
  to: string,
  amount: number,
  fee: number,
): Promise<string> => SubmitNodeTransaction(nodeAddress, from, to, amount, fee)
export const mineNodePending = (address: string): Promise<string> => MineNodePending(address)
