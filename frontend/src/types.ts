export type DashboardData = {
  height: number
  latestHash: string
  merkleRoot: string
  difficulty: number
  nonce: number
  pendingTxCount: number
  walletCount: number
  dataDir: string
  networkMode: string
  lastReorg?: ReorgStatusView | null
  recentEvents: ChainEventView[]
}

export type ReorgStatusView = {
  timestamp: string
  oldHeight: number
  newHeight: number
  oldTip: string
  newTip: string
  restoredTxCount: number
  droppedConfirmedCount: number
}

export type ChainEventView = {
  timestamp: string
  kind: string
  summary: string
  oldHeight: number
  newHeight: number
  oldTip: string
  newTip: string
  restoredTxCount: number
  droppedConfirmedCount: number
}

export type WalletView = {
  address: string
  balance: number
  lockingScript: string
}

export type InputView = {
  txid: string
  out: number
  source: string
  scriptSig: string
}

export type OutputView = {
  to: string
  value: number
  scriptPubKey: string
}

export type TransactionView = {
  id: string
  version: number
  fee: number
  usesScriptVM: boolean
  inputs: InputView[]
  outputs: OutputView[]
}

export type BlockView = {
  height: number
  hash: string
  prevHash: string
  merkleRoot: string
  difficulty: number
  nonce: number
  powValid: boolean
  timestamp: string
  transactionCount: number
  transactions: TransactionView[]
}

export type CommandResult = {
  command: string
  stdout: string
  stderr: string
  exitCode: number
}

export type NodeStatus = {
  address: string
  minerAddress: string
  peers: string[]
  height: number
  running: boolean
  orphanCount: number
  recentEvents: NodeEventView[]
}

export type NodeEventView = {
  timestamp: string
  kind: string
  detail: string
}
