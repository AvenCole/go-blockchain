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
}

export type WalletView = {
  address: string
  balance: number
}

export type InputView = {
  txid: string
  out: number
  source: string
}

export type OutputView = {
  to: string
  value: number
}

export type TransactionView = {
  id: string
  fee: number
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
}
