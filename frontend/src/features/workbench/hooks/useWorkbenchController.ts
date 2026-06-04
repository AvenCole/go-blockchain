import { useEffect, useMemo, useState } from 'react'
import { EventsOn } from '@wails/runtime/runtime'
import {
  connectNode,
  createWallet,
  executeCLI,
  fetchBlocks,
  fetchDashboard,
  fetchMultiSigOutputs,
  fetchNodes,
  fetchPendingTransactions,
  fetchWallets,
  initializeNodeBlockchain,
  minePending,
  mineNodePending,
  queueMultiSigTransaction,
  queueP2PKTransaction,
  queueSpendMultiSigTransaction,
  queueTransaction,
  runNetworkPartitionDemo,
  runNetworkQuickDemo,
  runNetworkReorgDemo,
  startNode,
  stopNode,
  submitNodeTransaction,
} from '@/api/backend'
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
import type {
  BusyActionKey,
  BusyActions,
  ConnectForm,
  NodeControlForm,
  NodeForm,
  SpendMultiSigForm,
  TxForm,
  WorkbenchContextValue,
} from '@/features/workbench/types'

const defaultBusyActions: BusyActions = {
  startNode: false,
  stopNode: false,
  connectNode: false,
  initializeNodeBlockchain: false,
  submitNodeTransaction: false,
  mineNode: false,
  runNetworkQuickDemo: false,
  runNetworkReorgDemo: false,
  runNetworkPartitionDemo: false,
}

const defaultTxForm: TxForm = {
  template: 'p2pkh',
  from: '',
  to: '',
  recipients: '',
  required: '2',
  amount: '20',
  fee: '2',
}

const defaultSpendMultiSigForm: SpendMultiSigForm = {
  signers: '',
  sourceTxID: '',
  out: '0',
  to: '',
  amount: '10',
  fee: '1',
}

const defaultNodeForm: NodeForm = {
  address: '127.0.0.1:3010',
  seed: '',
  miner: '',
}

const defaultConnectForm: ConnectForm = {
  address: '',
  seed: '',
}

const defaultNodeControlForm: NodeControlForm = {
  address: '',
  rewardAddress: '',
  from: '',
  to: '',
  amount: '10',
  fee: '1',
}

export function useWorkbenchController(): WorkbenchContextValue {
  const [dashboard, setDashboard] = useState<DashboardData | null>(null)
  const [wallets, setWallets] = useState<WalletView[]>([])
  const [blocks, setBlocks] = useState<BlockView[]>([])
  const [multiSigOutputs, setMultiSigOutputs] = useState<MultiSigOutputView[]>(
    [],
  )
  const [mempool, setMempool] = useState<string[]>([])
  const [nodes, setNodes] = useState<NodeStatus[]>([])
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')
  const [networkDemo, setNetworkDemo] = useState<NetworkDemoResult | null>(null)
  const [networkReorgDemo, setNetworkReorgDemo] =
    useState<NetworkReorgDemoResult | null>(null)
  const [networkPartitionDemo, setNetworkPartitionDemo] =
    useState<NetworkPartitionDemoResult | null>(null)
  const [networkOperation, setNetworkOperation] =
    useState<NetworkOperationProgress | null>(null)
  const [busyActions, setBusyActions] = useState<BusyActions>(defaultBusyActions)
  const [txForm, setTxForm] = useState<TxForm>(defaultTxForm)
  const [spendMultiSigForm, setSpendMultiSigForm] =
    useState<SpendMultiSigForm>(defaultSpendMultiSigForm)
  const [minerAddress, setMinerAddress] = useState('')
  const [command, setCommand] = useState('')
  const [history, setHistory] = useState<CommandResult[]>([])
  const [chainInitAddress, setChainInitAddress] = useState('')
  const [isInitializingBlockchain, setIsInitializingBlockchain] =
    useState(false)
  const [nodeForm, setNodeForm] = useState<NodeForm>(defaultNodeForm)
  const [connectForm, setConnectForm] = useState<ConnectForm>(defaultConnectForm)
  const [nodeControlForm, setNodeControlForm] =
    useState<NodeControlForm>(defaultNodeControlForm)

  const latestBlock = useMemo(
    () => (blocks.length > 0 ? blocks[0] : null),
    [blocks],
  )

  const isNodeActionBusy = useMemo(
    () =>
      busyActions.startNode ||
      busyActions.stopNode ||
      busyActions.connectNode ||
      busyActions.initializeNodeBlockchain ||
      busyActions.submitNodeTransaction ||
      busyActions.mineNode,
    [busyActions],
  )

  const isDemoBusy = useMemo(
    () =>
      busyActions.runNetworkQuickDemo ||
      busyActions.runNetworkReorgDemo ||
      busyActions.runNetworkPartitionDemo ||
      Boolean(
        networkOperation &&
          (networkOperation.status === 'started' ||
            networkOperation.status === 'progress'),
      ),
    [busyActions, networkOperation],
  )

  const clearFeedback = () => {
    setMessage('')
    setError('')
  }

  const refresh = async () => {
    try {
      setError('')
      const [dash, walletList, blockList, pending, nodeList, multiSigList] =
        await Promise.all([
          fetchDashboard(),
          fetchWallets(),
          fetchBlocks(),
          fetchPendingTransactions(),
          fetchNodes(),
          fetchMultiSigOutputs(),
        ])

      setDashboard(dash)
      setWallets(walletList)
      setBlocks(blockList)
      setMultiSigOutputs(multiSigList)
      setMempool(pending)
      setNodes(nodeList)

      if (!minerAddress && walletList.length > 0) {
        setMinerAddress(walletList[0].address)
      }

      if (!chainInitAddress && walletList.length > 0) {
        setChainInitAddress(walletList[0].address)
      }

      if (!txForm.from && walletList.length > 0) {
        setTxForm((previousForm) => ({
          ...previousForm,
          from: walletList[0].address,
        }))
      }

      if (!txForm.to && walletList.length > 1) {
        setTxForm((previousForm) => ({
          ...previousForm,
          to: walletList[1].address,
        }))
      }

      if (!txForm.recipients && walletList.length > 1) {
        setTxForm((previousForm) => ({
          ...previousForm,
          recipients: `${walletList[0].address},${walletList[1].address}`,
        }))
      }

      if (!nodeForm.miner && walletList.length > 0) {
        setNodeForm((previousForm) => ({
          ...previousForm,
          miner: walletList[0].address,
        }))
      }

      if (!connectForm.address && nodeList.length > 0) {
        setConnectForm((previousForm) => ({
          ...previousForm,
          address: nodeList[0].address,
        }))
      }

      if (
        (!nodeControlForm.address ||
          !nodeList.some((node) => node.address === nodeControlForm.address)) &&
        nodeList.length > 0
      ) {
        setNodeControlForm((previousForm) => ({
          ...previousForm,
          address: nodeList[0].address,
        }))
      }

      if (!nodeControlForm.rewardAddress && walletList.length > 0) {
        setNodeControlForm((previousForm) => ({
          ...previousForm,
          rewardAddress: walletList[0].address,
        }))
      }

      if (!nodeControlForm.from && walletList.length > 0) {
        setNodeControlForm((previousForm) => ({
          ...previousForm,
          from: walletList[0].address,
        }))
      }

      if (!nodeControlForm.to && walletList.length > 1) {
        setNodeControlForm((previousForm) => ({
          ...previousForm,
          to: walletList[1].address,
        }))
      }

      if (!spendMultiSigForm.to && walletList.length > 0) {
        setSpendMultiSigForm((previousForm) => ({
          ...previousForm,
          to: walletList[0].address,
        }))
      }

      if (!spendMultiSigForm.sourceTxID && multiSigList.length > 0) {
        setSpendMultiSigForm((previousForm) => ({
          ...previousForm,
          sourceTxID: multiSigList[0].txid,
          out: String(multiSigList[0].out),
          signers: multiSigList[0].participants.join(','),
        }))
      }
    } catch (err) {
      setError(String(err))
    }
  }

  useEffect(() => {
    void refresh()
  }, [])

  useEffect(() => {
    const unsubscribe = EventsOn(
      'network:operation',
      (payload: NetworkOperationProgress) => {
        setNetworkOperation(payload)
      },
    )

    return () => {
      unsubscribe()
    }
  }, [])

  const runBusyAction = async (
    key: BusyActionKey,
    action: () => Promise<void>,
  ) => {
    setBusyActions((previousBusyActions) => ({
      ...previousBusyActions,
      [key]: true,
    }))

    try {
      await action()
    } finally {
      setBusyActions((previousBusyActions) => ({
        ...previousBusyActions,
        [key]: false,
      }))
    }
  }

  const handleCreateWallet = async () => {
    try {
      setError('')
      const address = await createWallet()
      setMessage(`已创建钱包：${address}`)

      if (!chainInitAddress) {
        setChainInitAddress(address)
      }

      await refresh()
    } catch (err) {
      setError(String(err))
    }
  }

  const handleInitializeBlockchain = async () => {
    if (!chainInitAddress) {
      return
    }

    try {
      setIsInitializingBlockchain(true)
      setError('')
      const result = await executeCLI(`createblockchain ${chainInitAddress}`)
      setHistory((previousHistory) => [result, ...previousHistory].slice(0, 20))

      if (result.exitCode !== 0) {
        setError(result.stderr || result.stdout || '初始化主链失败')
        return
      }

      setMessage(`主链已初始化：${chainInitAddress}`)
      await refresh()
    } catch (err) {
      setError(String(err))
    } finally {
      setIsInitializingBlockchain(false)
    }
  }

  const handleQueueTransaction = async () => {
    try {
      setError('')
      let txid = ''

      if (txForm.template === 'p2pk') {
        txid = await queueP2PKTransaction(
          txForm.from,
          txForm.to,
          Number(txForm.amount),
          Number(txForm.fee || '0'),
        )
      } else if (txForm.template === 'multisig') {
        txid = await queueMultiSigTransaction(
          txForm.from,
          txForm.recipients,
          Number(txForm.required || '0'),
          Number(txForm.amount),
          Number(txForm.fee || '0'),
        )
      } else {
        txid = await queueTransaction(
          txForm.from,
          txForm.to,
          Number(txForm.amount),
          Number(txForm.fee || '0'),
        )
      }

      setMessage(`交易已进入 Mempool：${txid}`)
      await refresh()
    } catch (err) {
      setError(String(err))
    }
  }

  const handleSpendMultiSig = async () => {
    try {
      setError('')
      const txid = await queueSpendMultiSigTransaction(
        spendMultiSigForm.signers,
        spendMultiSigForm.sourceTxID,
        Number(spendMultiSigForm.out),
        spendMultiSigForm.to,
        Number(spendMultiSigForm.amount),
        Number(spendMultiSigForm.fee || '0'),
      )
      setMessage(`多签花费交易已进入 Mempool：${txid}`)
      await refresh()
    } catch (err) {
      setError(String(err))
    }
  }

  const handleMine = async () => {
    try {
      setError('')
      const hash = await minePending(minerAddress)
      setMessage(`已挖出新区块：${hash}`)
      await refresh()
    } catch (err) {
      setError(String(err))
    }
  }

  const handleRunCommand = async (
    commandLine: string,
    clearInput = false,
  ) => {
    if (!commandLine.trim()) {
      return
    }

    try {
      setError('')

      if (!clearInput) {
        setCommand(commandLine)
      }

      const result = await executeCLI(commandLine)
      setHistory((previousHistory) => [result, ...previousHistory].slice(0, 20))
      setMessage(`命令执行完成：${result.command}`)

      if (clearInput) {
        setCommand('')
      }

      await refresh()
    } catch (err) {
      setError(String(err))
    }
  }

  const handleExecuteCommand = async () => {
    await handleRunCommand(command, true)
  }

  const handleStartNode = async () => {
    await runBusyAction('startNode', async () => {
      try {
        setError('')
        const address = await startNode(
          nodeForm.address,
          nodeForm.seed,
          nodeForm.miner,
        )
        setConnectForm((previousForm) => ({ ...previousForm, address }))
        setMessage(`节点已启动：${address}`)
        await refresh()
      } catch (err) {
        setError(String(err))
      }
    })
  }

  const handleStopNode = async (address: string) => {
    await runBusyAction('stopNode', async () => {
      try {
        setError('')
        await stopNode(address)
        setMessage(`节点已停止：${address}`)
        await refresh()
      } catch (err) {
        setError(String(err))
      }
    })
  }

  const handleConnectNode = async () => {
    await runBusyAction('connectNode', async () => {
      try {
        setError('')
        await connectNode(connectForm.address, connectForm.seed)
        setMessage(`节点已连接：${connectForm.address} -> ${connectForm.seed}`)
        await refresh()
      } catch (err) {
        setError(String(err))
      }
    })
  }

  const handleInitializeNodeBlockchain = async () => {
    await runBusyAction('initializeNodeBlockchain', async () => {
      try {
        setError('')
        await initializeNodeBlockchain(
          nodeControlForm.address,
          nodeControlForm.rewardAddress,
        )
        setMessage(`节点区块链已就绪：${nodeControlForm.address}`)
        await refresh()
      } catch (err) {
        setError(String(err))
      }
    })
  }

  const handleSubmitNodeTransaction = async () => {
    await runBusyAction('submitNodeTransaction', async () => {
      try {
        setError('')
        const txid = await submitNodeTransaction(
          nodeControlForm.address,
          nodeControlForm.from,
          nodeControlForm.to,
          Number(nodeControlForm.amount),
          Number(nodeControlForm.fee || '0'),
        )
        setMessage(`节点交易已进入 Mempool：${txid}`)
        await refresh()
      } catch (err) {
        setError(String(err))
      }
    })
  }

  const handleMineNode = async () => {
    await runBusyAction('mineNode', async () => {
      try {
        setError('')
        const hash = await mineNodePending(nodeControlForm.address)
        setMessage(`节点已挖出新区块：${hash}`)
        await refresh()
      } catch (err) {
        setError(String(err))
      }
    })
  }

  const handleRunNetworkQuickDemo = async () => {
    await runBusyAction('runNetworkQuickDemo', async () => {
      try {
        setError('')
        const result = await runNetworkQuickDemo()
        setNetworkDemo(result)
        setMessage(`快速同步已完成：${result.sourceNode} -> ${result.peerNode}`)
        await refresh()
      } catch (err) {
        setError(String(err))
      }
    })
  }

  const handleRunNetworkReorgDemo = async () => {
    await runBusyAction('runNetworkReorgDemo', async () => {
      try {
        setError('')
        const result = await runNetworkReorgDemo()
        setNetworkReorgDemo(result)
        setMessage(`重组流程已完成：${result.sourceNode} -> ${result.peerNode}`)
        await refresh()
      } catch (err) {
        setError(String(err))
      }
    })
  }

  const handleRunNetworkPartitionDemo = async () => {
    await runBusyAction('runNetworkPartitionDemo', async () => {
      try {
        setError('')
        const result = await runNetworkPartitionDemo()
        setNetworkPartitionDemo(result)
        setMessage(
          `三节点分区流程已完成：${result.sourceNode} / ${result.peerNode} / ${result.forkNode}`,
        )
        await refresh()
      } catch (err) {
        setError(String(err))
      }
    })
  }

  return {
    dashboard,
    wallets,
    blocks,
    multiSigOutputs,
    mempool,
    nodes,
    message,
    error,
    networkDemo,
    networkReorgDemo,
    networkPartitionDemo,
    networkOperation,
    busyActions,
    txForm,
    setTxForm,
    spendMultiSigForm,
    setSpendMultiSigForm,
    minerAddress,
    setMinerAddress,
    command,
    setCommand,
    history,
    chainInitAddress,
    setChainInitAddress,
    isInitializingBlockchain,
    nodeForm,
    setNodeForm,
    connectForm,
    setConnectForm,
    nodeControlForm,
    setNodeControlForm,
    latestBlock,
    isNodeActionBusy,
    isDemoBusy,
    clearFeedback,
    refresh,
    handleCreateWallet,
    handleInitializeBlockchain,
    handleQueueTransaction,
    handleSpendMultiSig,
    handleMine,
    handleExecuteCommand,
    handleRunCommand,
    handleStartNode,
    handleStopNode,
    handleConnectNode,
    handleInitializeNodeBlockchain,
    handleSubmitNodeTransaction,
    handleMineNode,
    handleRunNetworkQuickDemo,
    handleRunNetworkReorgDemo,
    handleRunNetworkPartitionDemo,
  }
}
