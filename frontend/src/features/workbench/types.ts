import type { Dispatch, SetStateAction } from 'react'
import type {
  BlockView,
  CommandResult,
  DashboardData,
  MultiSigOutputView,
  NetworkDemoResult,
  NetworkOperationProgress,
  NetworkPartitionDemoResult,
  NetworkReorgDemoResult,
  NodeStatus,
  WalletView,
} from '@/types'

export type BusyActionKey =
  | 'startNode'
  | 'stopNode'
  | 'connectNode'
  | 'initializeNodeBlockchain'
  | 'submitNodeTransaction'
  | 'mineNode'
  | 'runNetworkQuickDemo'
  | 'runNetworkReorgDemo'
  | 'runNetworkPartitionDemo'

export type BusyActions = Record<BusyActionKey, boolean>

export type TxForm = {
  template: 'p2pkh' | 'p2pk' | 'multisig'
  from: string
  to: string
  recipients: string
  required: string
  amount: string
  fee: string
}

export type SpendMultiSigForm = {
  signers: string
  sourceTxID: string
  out: string
  to: string
  amount: string
  fee: string
}

export type NodeForm = {
  address: string
  seed: string
  miner: string
}

export type ConnectForm = {
  address: string
  seed: string
}

export type NodeControlForm = {
  address: string
  rewardAddress: string
  from: string
  to: string
  amount: string
  fee: string
}

export type WorkbenchContextValue = {
  dashboard: DashboardData | null
  wallets: WalletView[]
  blocks: BlockView[]
  multiSigOutputs: MultiSigOutputView[]
  mempool: string[]
  nodes: NodeStatus[]
  message: string
  error: string
  networkDemo: NetworkDemoResult | null
  networkReorgDemo: NetworkReorgDemoResult | null
  networkPartitionDemo: NetworkPartitionDemoResult | null
  networkOperation: NetworkOperationProgress | null
  busyActions: BusyActions
  txForm: TxForm
  setTxForm: Dispatch<SetStateAction<TxForm>>
  spendMultiSigForm: SpendMultiSigForm
  setSpendMultiSigForm: Dispatch<SetStateAction<SpendMultiSigForm>>
  minerAddress: string
  setMinerAddress: Dispatch<SetStateAction<string>>
  command: string
  setCommand: Dispatch<SetStateAction<string>>
  history: CommandResult[]
  chainInitAddress: string
  setChainInitAddress: Dispatch<SetStateAction<string>>
  isInitializingBlockchain: boolean
  nodeForm: NodeForm
  setNodeForm: Dispatch<SetStateAction<NodeForm>>
  connectForm: ConnectForm
  setConnectForm: Dispatch<SetStateAction<ConnectForm>>
  nodeControlForm: NodeControlForm
  setNodeControlForm: Dispatch<SetStateAction<NodeControlForm>>
  latestBlock: BlockView | null
  isNodeActionBusy: boolean
  isDemoBusy: boolean
  clearFeedback: () => void
  refresh: () => Promise<void>
  handleCreateWallet: () => Promise<void>
  handleInitializeBlockchain: () => Promise<void>
  handleQueueTransaction: () => Promise<void>
  handleSpendMultiSig: () => Promise<void>
  handleMine: () => Promise<void>
  handleExecuteCommand: () => Promise<void>
  handleRunCommand: (commandLine: string, clearInput?: boolean) => Promise<void>
  handleStartNode: () => Promise<void>
  handleStopNode: (address: string) => Promise<void>
  handleConnectNode: () => Promise<void>
  handleInitializeNodeBlockchain: () => Promise<void>
  handleSubmitNodeTransaction: () => Promise<void>
  handleMineNode: () => Promise<void>
  handleRunNetworkQuickDemo: () => Promise<void>
  handleRunNetworkReorgDemo: () => Promise<void>
  handleRunNetworkPartitionDemo: () => Promise<void>
}
