import { useEffect, useMemo, useState } from 'react'
import {
  Box,
  Button,
  Card,
  CardContent,
  Divider,
  MenuItem,
  Stack,
  TextField,
  Typography,
} from '@mui/material'
import EventFilterToolbar from '../components/network/EventFilterToolbar'
import FocusedNodeCard from '../components/network/FocusedNodeCard'
import NetworkOperationStatusCard from '../components/network/NetworkOperationStatusCard'
import NetworkTimelineCard from '../components/network/NetworkTimelineCard'
import NetworkTopologyCard from '../components/network/NetworkTopologyCard'
import NodeDirectoryCard from '../components/network/NodeDirectoryCard'
import type {
  ChainEventView,
  NetworkDemoResult,
  NetworkOperationProgress,
  NetworkPartitionDemoResult,
  NetworkReorgDemoResult,
  NodeStatus,
  ReorgStatusView,
  WalletView,
} from '../types'
import { collectKinds, filterChainEvents } from '../utils/networkView'

type NetworkPageProps = {
  nodes: NodeStatus[]
  wallets: WalletView[]
  networkDemo?: NetworkDemoResult | null
  networkReorgDemo?: NetworkReorgDemoResult | null
  networkPartitionDemo?: NetworkPartitionDemoResult | null
  lastReorg?: ReorgStatusView | null
  recentEvents?: ChainEventView[]
  nodeForm: { address: string; seed: string; miner: string }
  setNodeForm: React.Dispatch<
    React.SetStateAction<{ address: string; seed: string; miner: string }>
  >
  connectForm: { address: string; seed: string }
  setConnectForm: React.Dispatch<
    React.SetStateAction<{ address: string; seed: string }>
  >
  nodeControlForm: {
    address: string
    rewardAddress: string
    from: string
    to: string
    amount: string
    fee: string
  }
  setNodeControlForm: React.Dispatch<
    React.SetStateAction<{
      address: string
      rewardAddress: string
      from: string
      to: string
      amount: string
      fee: string
    }>
  >
  onStartNode: () => Promise<void>
  onStopNode: (address: string) => Promise<void>
  onConnectNode: () => Promise<void>
  onInitializeNodeBlockchain: () => Promise<void>
  onSubmitNodeTransaction: () => Promise<void>
  onMineNode: () => Promise<void>
  onRunNetworkQuickDemo: () => Promise<void>
  onRunNetworkReorgDemo: () => Promise<void>
  onRunNetworkPartitionDemo: () => Promise<void>
  operationProgress: NetworkOperationProgress | null
  isDemoBusy: boolean
  isNodeActionBusy: boolean
  busyActions: {
    startNode: boolean
    stopNode: boolean
    connectNode: boolean
    initializeNodeBlockchain: boolean
    submitNodeTransaction: boolean
    mineNode: boolean
    runNetworkQuickDemo: boolean
    runNetworkReorgDemo: boolean
    runNetworkPartitionDemo: boolean
  }
}

function NetworkPage({
  nodes,
  wallets,
  networkDemo,
  networkReorgDemo,
  networkPartitionDemo,
  lastReorg,
  recentEvents = [],
  nodeForm,
  setNodeForm,
  connectForm,
  setConnectForm,
  nodeControlForm,
  setNodeControlForm,
  onStartNode,
  onStopNode,
  onConnectNode,
  onInitializeNodeBlockchain,
  onSubmitNodeTransaction,
  onMineNode,
  onRunNetworkQuickDemo,
  onRunNetworkReorgDemo,
  onRunNetworkPartitionDemo,
  operationProgress,
  isDemoBusy,
  isNodeActionBusy,
  busyActions,
}: NetworkPageProps) {
  const [focusedNodeAddress, setFocusedNodeAddress] = useState('')
  const [chainEventQuery, setChainEventQuery] = useState('')
  const [chainEventKind, setChainEventKind] = useState('')

  useEffect(() => {
    if (nodes.length === 0) {
      if (focusedNodeAddress) setFocusedNodeAddress('')
      return
    }

    if (focusedNodeAddress && nodes.some((node) => node.address === focusedNodeAddress)) {
      return
    }

    const nextFocusedAddress =
      nodeControlForm.address &&
      nodes.some((node) => node.address === nodeControlForm.address)
        ? nodeControlForm.address
        : nodes[0].address

    if (nextFocusedAddress !== focusedNodeAddress) {
      setFocusedNodeAddress(nextFocusedAddress)
    }
  }, [focusedNodeAddress, nodeControlForm.address, nodes])

  const focusedNode = useMemo(
    () => nodes.find((node) => node.address === focusedNodeAddress) ?? null,
    [focusedNodeAddress, nodes],
  )
  const chainEventKindOptions = useMemo(
    () => collectKinds(recentEvents),
    [recentEvents],
  )
  const filteredChainEvents = useMemo(
    () =>
      filterChainEvents(recentEvents, {
        kind: chainEventKind,
        query: chainEventQuery,
      }).slice(0, 12),
    [chainEventKind, chainEventQuery, recentEvents],
  )

  useEffect(() => {
    if (chainEventKind && !chainEventKindOptions.includes(chainEventKind)) {
      setChainEventKind('')
    }
  }, [chainEventKind, chainEventKindOptions])

  return (
    <Stack spacing={2}>
      <Card variant="outlined">
        <CardContent sx={{ p: 2.25 }}>
          <Typography variant="h6">网络流程</Typography>
          <Stack direction="row" spacing={1.25} sx={{ mt: 2, flexWrap: 'wrap' }}>
            <Button
              variant="contained"
              disabled={isDemoBusy}
              onClick={onRunNetworkQuickDemo}
            >
              {busyActions.runNetworkQuickDemo ? '快速同步中...' : '快速同步'}
            </Button>
            <Button
              variant="contained"
              color="secondary"
              disabled={isDemoBusy}
              onClick={onRunNetworkReorgDemo}
            >
              {busyActions.runNetworkReorgDemo ? '重组流程中...' : '分叉 / 重组'}
            </Button>
            <Button
              variant="outlined"
              disabled={isDemoBusy}
              onClick={onRunNetworkPartitionDemo}
            >
              {busyActions.runNetworkPartitionDemo ? '分区流程中...' : '分区 / 合流'}
            </Button>
          </Stack>

          <Box
            sx={{
              display: 'grid',
              gap: 2,
              gridTemplateColumns: 'minmax(0, 1fr) 320px',
              mt: 2,
            }}
          >
            <NetworkOperationStatusCard operation={operationProgress} />

            <Card variant="outlined" sx={{ bgcolor: 'background.default' }}>
              <CardContent sx={{ p: 2 }}>
                <Typography variant="subtitle2">结果</Typography>
                <Stack spacing={0.75} sx={{ mt: 1.5 }}>
                  {networkDemo ? (
                    <Typography variant="body2">
                      快速同步：{networkDemo.peerHeight}
                    </Typography>
                  ) : null}
                  {networkReorgDemo ? (
                    <Typography variant="body2">
                      重组：{networkReorgDemo.sourceOldHeight} → {networkReorgDemo.sourceNewHeight}
                    </Typography>
                  ) : null}
                  {networkPartitionDemo ? (
                    <Typography variant="body2">
                      分区：forkHeight={networkPartitionDemo.forkHeight}
                    </Typography>
                  ) : null}
                  {!networkDemo && !networkReorgDemo && !networkPartitionDemo ? (
                    <Typography variant="body2" color="text.secondary">
                      暂无结果
                    </Typography>
                  ) : null}
                </Stack>
              </CardContent>
            </Card>
          </Box>
        </CardContent>
      </Card>

      <Box
        sx={{
          display: 'grid',
          gap: 2,
          gridTemplateColumns: 'minmax(0, 1fr) minmax(0, 1fr)',
        }}
      >
        <NetworkTopologyCard
          nodes={nodes}
          selectedNodeAddress={focusedNodeAddress}
          onSelectNode={setFocusedNodeAddress}
        />
        <NetworkTimelineCard nodes={nodes} recentEvents={recentEvents} />
      </Box>

      <Box
        sx={{
          display: 'grid',
          gap: 2,
          gridTemplateColumns: '320px minmax(0, 1fr)',
        }}
      >
        <NodeDirectoryCard
          nodes={nodes}
          selectedNodeAddress={focusedNodeAddress}
          onSelectNode={setFocusedNodeAddress}
        />
        <FocusedNodeCard
          node={focusedNode}
          onUseAsConnectNode={(address) =>
            setConnectForm((prev) => ({ ...prev, address }))
          }
          onUseAsSeed={(address) => {
            setConnectForm((prev) => ({ ...prev, seed: address }))
            setNodeForm((prev) => ({ ...prev, seed: address }))
          }}
          onUseAsControlNode={(address) =>
            setNodeControlForm((prev) => ({ ...prev, address }))
          }
          onStopNode={onStopNode}
          isStopDisabled={isDemoBusy || isNodeActionBusy}
          isStopping={busyActions.stopNode}
        />
      </Box>

      <Card variant="outlined">
        <CardContent sx={{ p: 2.25 }}>
          <Typography variant="h6">链事件</Typography>
          <EventFilterToolbar
            query={chainEventQuery}
            onQueryChange={setChainEventQuery}
            kind={chainEventKind}
            kindOptions={chainEventKindOptions}
            onKindChange={setChainEventKind}
            matchedCount={filteredChainEvents.length}
            totalCount={recentEvents.length}
          />

          <Box
            sx={{
              display: 'grid',
              gap: 2,
              gridTemplateColumns: '320px minmax(0, 1fr)',
              mt: 2,
            }}
          >
            <Card variant="outlined" sx={{ bgcolor: 'background.default' }}>
              <CardContent sx={{ p: 2 }}>
                <Typography variant="subtitle2">最近重组</Typography>
                {lastReorg ? (
                  <Stack spacing={0.6} sx={{ mt: 1.5 }}>
                    <Typography variant="body2">{lastReorg.timestamp}</Typography>
                    <Typography variant="body2">
                      {lastReorg.oldHeight} → {lastReorg.newHeight}
                    </Typography>
                    <Typography variant="body2">
                      恢复：{lastReorg.restoredTxCount}
                    </Typography>
                    <Typography variant="body2">
                      清理：{lastReorg.droppedConfirmedCount}
                    </Typography>
                  </Stack>
                ) : (
                  <Typography variant="body2" color="text.secondary" sx={{ mt: 1.5 }}>
                    暂无
                  </Typography>
                )}
              </CardContent>
            </Card>

            <Card variant="outlined" sx={{ bgcolor: 'background.default' }}>
              <CardContent sx={{ p: 2 }}>
                <Stack spacing={1}>
                  {recentEvents.length > 0 ? (
                    filteredChainEvents.length > 0 ? (
                      filteredChainEvents.map((event, index) => (
                        <Stack key={`${event.timestamp}-${index}`} spacing={0.25}>
                          <Typography variant="body2">
                            {event.timestamp} · {event.kind}
                          </Typography>
                          <Typography variant="body2" color="text.secondary">
                            {event.summary}
                          </Typography>
                        </Stack>
                      ))
                    ) : (
                      <Typography variant="body2" color="text.secondary">
                        当前筛选条件下没有匹配项
                      </Typography>
                    )
                  ) : (
                    <Typography variant="body2" color="text.secondary">
                      暂无链事件
                    </Typography>
                  )}
                </Stack>
              </CardContent>
            </Card>
          </Box>
        </CardContent>
      </Card>

      <Card variant="outlined">
        <CardContent sx={{ p: 2.25 }}>
          <Typography variant="h6">节点操作</Typography>

          <Box
            sx={{
              display: 'grid',
              gap: 2,
              gridTemplateColumns: 'repeat(3, minmax(0, 1fr))',
              mt: 2,
            }}
          >
            <Stack spacing={2}>
              <Typography variant="subtitle2">启动</Typography>
              <TextField
                label="监听地址"
                value={nodeForm.address}
                onChange={(e) =>
                  setNodeForm((p) => ({ ...p, address: e.target.value }))
                }
              />
              <TextField
                label="Seed"
                value={nodeForm.seed}
                onChange={(e) =>
                  setNodeForm((p) => ({ ...p, seed: e.target.value }))
                }
              />
              <TextField
                label="矿工地址"
                value={nodeForm.miner}
                onChange={(e) =>
                  setNodeForm((p) => ({ ...p, miner: e.target.value }))
                }
              />
              <Button
                variant="contained"
                disabled={isDemoBusy || isNodeActionBusy}
                onClick={onStartNode}
              >
                {busyActions.startNode ? '启动中...' : '启动节点'}
              </Button>
            </Stack>

            <Stack spacing={2}>
              <Typography variant="subtitle2">连接</Typography>
              <TextField
                label="本地节点"
                value={connectForm.address}
                onChange={(e) =>
                  setConnectForm((p) => ({ ...p, address: e.target.value }))
                }
              />
              <TextField
                label="Seed"
                value={connectForm.seed}
                onChange={(e) =>
                  setConnectForm((p) => ({ ...p, seed: e.target.value }))
                }
              />
              <Button
                variant="contained"
                color="secondary"
                disabled={isDemoBusy || isNodeActionBusy}
                onClick={onConnectNode}
              >
                {busyActions.connectNode ? '连接中...' : '连接 Seed'}
              </Button>
            </Stack>

            <Stack spacing={2}>
              <Typography variant="subtitle2">链控制</Typography>
              <TextField
                select
                label="目标节点"
                value={nodeControlForm.address}
                onChange={(e) =>
                  setNodeControlForm((prev) => ({
                    ...prev,
                    address: e.target.value,
                  }))
                }
              >
                {nodes.length === 0 ? (
                  <MenuItem value="" disabled>
                    无节点
                  </MenuItem>
                ) : (
                  nodes.map((node) => (
                    <MenuItem key={node.address} value={node.address}>
                      {node.address}
                    </MenuItem>
                  ))
                )}
              </TextField>
              <TextField
                select
                label="创世奖励地址"
                value={nodeControlForm.rewardAddress}
                onChange={(e) =>
                  setNodeControlForm((prev) => ({
                    ...prev,
                    rewardAddress: e.target.value,
                  }))
                }
              >
                {wallets.map((wallet) => (
                  <MenuItem key={`reward-${wallet.address}`} value={wallet.address}>
                    {wallet.address}
                  </MenuItem>
                ))}
              </TextField>
              <Button
                variant="contained"
                color="secondary"
                onClick={onInitializeNodeBlockchain}
                disabled={
                  isDemoBusy ||
                  isNodeActionBusy ||
                  !nodeControlForm.address ||
                  !nodeControlForm.rewardAddress
                }
              >
                {busyActions.initializeNodeBlockchain ? '初始化中...' : '初始化节点链'}
              </Button>
            </Stack>
          </Box>

          <Divider sx={{ my: 2 }} />

          <Box
            sx={{
              display: 'grid',
              gap: 2,
              gridTemplateColumns: '1fr 1fr 140px 140px',
            }}
          >
            <TextField
              select
              label="发送方钱包"
              value={nodeControlForm.from}
              onChange={(e) =>
                setNodeControlForm((prev) => ({ ...prev, from: e.target.value }))
              }
            >
              {wallets.map((wallet) => (
                <MenuItem key={`from-${wallet.address}`} value={wallet.address}>
                  {wallet.address}
                </MenuItem>
              ))}
            </TextField>
            <TextField
              select
              label="接收方钱包"
              value={nodeControlForm.to}
              onChange={(e) =>
                setNodeControlForm((prev) => ({ ...prev, to: e.target.value }))
              }
            >
              {wallets.map((wallet) => (
                <MenuItem key={`to-${wallet.address}`} value={wallet.address}>
                  {wallet.address}
                </MenuItem>
              ))}
            </TextField>
            <TextField
              label="金额"
              value={nodeControlForm.amount}
              onChange={(e) =>
                setNodeControlForm((prev) => ({ ...prev, amount: e.target.value }))
              }
            />
            <TextField
              label="手续费"
              value={nodeControlForm.fee}
              onChange={(e) =>
                setNodeControlForm((prev) => ({ ...prev, fee: e.target.value }))
              }
            />
          </Box>

          <Stack direction="row" spacing={2} sx={{ mt: 2 }}>
            <Button
              variant="contained"
              onClick={onSubmitNodeTransaction}
              disabled={
                isDemoBusy ||
                isNodeActionBusy ||
                !nodeControlForm.address ||
                !nodeControlForm.from ||
                !nodeControlForm.to
              }
            >
              {busyActions.submitNodeTransaction ? '发送中...' : '通过节点发交易'}
            </Button>
            <Button
              variant="outlined"
              onClick={onMineNode}
              disabled={isDemoBusy || isNodeActionBusy || !nodeControlForm.address}
            >
              {busyActions.mineNode ? '挖矿中...' : '让节点挖矿'}
            </Button>
          </Stack>
        </CardContent>
      </Card>
    </Stack>
  )
}

export default NetworkPage
