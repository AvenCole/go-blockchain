import { useEffect, useMemo, useState } from 'react'
import {
  Button,
  Card,
  CardContent,
  MenuItem,
  Stack,
  TextField,
  Typography,
} from '@mui/material'
import FocusedNodeCard from '../components/network/FocusedNodeCard'
import NodeDirectoryCard from '../components/network/NodeDirectoryCard'
import NetworkTimelineCard from '../components/network/NetworkTimelineCard'
import NetworkTopologyCard from '../components/network/NetworkTopologyCard'
import type {
  ChainEventView,
  NetworkDemoResult,
  NetworkPartitionDemoResult,
  NetworkReorgDemoResult,
  NodeStatus,
  ReorgStatusView,
  WalletView,
} from '../types'

type NetworkPageProps = {
  nodes: NodeStatus[]
  wallets: WalletView[]
  networkDemo?: NetworkDemoResult | null
  networkReorgDemo?: NetworkReorgDemoResult | null
  networkPartitionDemo?: NetworkPartitionDemoResult | null
  lastReorg?: ReorgStatusView | null
  recentEvents?: ChainEventView[]
  nodeForm: { address: string; seed: string; miner: string }
  setNodeForm: React.Dispatch<React.SetStateAction<{ address: string; seed: string; miner: string }>>
  connectForm: { address: string; seed: string }
  setConnectForm: React.Dispatch<React.SetStateAction<{ address: string; seed: string }>>
  nodeControlForm: { address: string; rewardAddress: string; from: string; to: string; amount: string; fee: string }
  setNodeControlForm: React.Dispatch<
    React.SetStateAction<{ address: string; rewardAddress: string; from: string; to: string; amount: string; fee: string }>
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
}: NetworkPageProps) {
  const [focusedNodeAddress, setFocusedNodeAddress] = useState('')

  useEffect(() => {
    if (nodes.length === 0) {
      if (focusedNodeAddress) {
        setFocusedNodeAddress('')
      }
      return
    }

    if (
      focusedNodeAddress &&
      nodes.some((node) => node.address === focusedNodeAddress)
    ) {
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

  return (
    <Stack spacing={2.5}>
      <Stack direction={{ xs: 'column', xl: 'row' }} spacing={2.5}>
        <Card variant="outlined" sx={{ flex: 1 }}>
          <CardContent sx={{ p: 2 }}>
            <Typography variant="h6">快速同步场景</Typography>
            <Typography color="text.secondary" sx={{ mt: 1 }}>
              自动创建双节点同步流程：准备钱包、启动节点、初始化主节点链、连接 peer、发送交易并挖矿。
            </Typography>
            <Stack direction={{ xs: 'column', md: 'row' }} spacing={2} sx={{ mt: 2.5, alignItems: { md: 'center' } }}>
              <Button variant="contained" color="primary" onClick={onRunNetworkQuickDemo}>
                运行快速同步
              </Button>
              {networkDemo ? (
                <Stack spacing={0.5}>
                  <Typography variant="body2">source={networkDemo.sourceNode}</Typography>
                  <Typography variant="body2">peer={networkDemo.peerNode} · peerHeight={networkDemo.peerHeight}</Typography>
                  <Typography variant="body2">tipAnnounced={String(networkDemo.tipAnnounced)}</Typography>
                </Stack>
              ) : (
                <Typography variant="body2" color="text.secondary">
                  执行后会返回 source、peer、peerHeight 和 tipAnnounced。
                </Typography>
              )}
            </Stack>
          </CardContent>
        </Card>

        <Card variant="outlined" sx={{ flex: 1 }}>
          <CardContent sx={{ p: 2 }}>
            <Typography variant="h6">分叉 / 重组场景</Typography>
            <Typography color="text.secondary" sx={{ mt: 1 }}>
              创建已确认交易后注入更长分叉链，并触发 peer 重新同步，用于检查 reorg 与交易恢复。
            </Typography>
            <Stack direction={{ xs: 'column', md: 'row' }} spacing={2} sx={{ mt: 2.5, alignItems: { md: 'center' } }}>
              <Button variant="contained" color="secondary" onClick={onRunNetworkReorgDemo}>
                运行重组流程
              </Button>
              {networkReorgDemo ? (
                <Stack spacing={0.5}>
                  <Typography variant="body2">source={networkReorgDemo.sourceNode}</Typography>
                  <Typography variant="body2">
                    sourceHeight={networkReorgDemo.sourceOldHeight} → {networkReorgDemo.sourceNewHeight}
                  </Typography>
                  <Typography variant="body2">
                    restored={String(networkReorgDemo.restored)} · peerReorged={String(networkReorgDemo.peerReorged)}
                  </Typography>
                </Stack>
              ) : (
                <Typography variant="body2" color="text.secondary">
                  执行后可查看 sourceHeight、restored 和 peerReorged。
                </Typography>
              )}
            </Stack>
          </CardContent>
        </Card>

        <Card variant="outlined" sx={{ flex: 1 }}>
          <CardContent sx={{ p: 2 }}>
            <Typography variant="h6">三节点分区 / 合流</Typography>
            <Typography color="text.secondary" sx={{ mt: 1 }}>
              构造 source、peer、fork 三节点分区，再在合流后检查所有节点是否收敛到同一 tip。
            </Typography>
            <Stack direction={{ xs: 'column', md: 'row' }} spacing={2} sx={{ mt: 2.5, alignItems: { md: 'center' } }}>
              <Button variant="contained" color="inherit" onClick={onRunNetworkPartitionDemo}>
                运行分区流程
              </Button>
              {networkPartitionDemo ? (
                <Stack spacing={0.5}>
                  <Typography variant="body2">
                    nodes={networkPartitionDemo.sourceNode} / {networkPartitionDemo.peerNode} / {networkPartitionDemo.forkNode}
                  </Typography>
                  <Typography variant="body2">
                    restored={String(networkPartitionDemo.restored)} · converged={String(networkPartitionDemo.allConverged)}
                  </Typography>
                  <Typography variant="body2">forkHeight={networkPartitionDemo.forkHeight}</Typography>
                </Stack>
              ) : (
                <Typography variant="body2" color="text.secondary">
                  执行后可查看 restored、converged 和 forkHeight。
                </Typography>
              )}
            </Stack>
          </CardContent>
        </Card>
      </Stack>

      <Stack direction={{ xs: 'column', xl: 'row' }} spacing={2.5}>
        <Stack sx={{ flex: 1, minWidth: 0 }}>
          <NetworkTopologyCard
            nodes={nodes}
            selectedNodeAddress={focusedNodeAddress}
            onSelectNode={setFocusedNodeAddress}
          />
        </Stack>
        <Stack sx={{ flex: 1, minWidth: 0 }}>
          <NetworkTimelineCard nodes={nodes} recentEvents={recentEvents} />
        </Stack>
      </Stack>

      <Stack direction={{ xs: 'column', xl: 'row' }} spacing={2.5}>
        <Stack sx={{ flex: 0.9, minWidth: 0 }}>
          <NodeDirectoryCard
            nodes={nodes}
            selectedNodeAddress={focusedNodeAddress}
            onSelectNode={setFocusedNodeAddress}
          />
        </Stack>
        <Stack sx={{ flex: 1.1, minWidth: 0 }}>
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
          />
        </Stack>
      </Stack>

      <Card variant="outlined">
        <CardContent sx={{ p: 2 }}>
          <Typography variant="h6">链切换观测</Typography>
          {lastReorg ? (
            <Stack spacing={0.75} sx={{ mt: 1.5 }}>
              <Typography variant="body2">最近重组时间：{lastReorg.timestamp}</Typography>
              <Typography variant="body2">高度变化：{lastReorg.oldHeight} → {lastReorg.newHeight}</Typography>
              <Typography variant="body2">恢复交易：{lastReorg.restoredTxCount}</Typography>
              <Typography variant="body2">清理已确认：{lastReorg.droppedConfirmedCount}</Typography>
            </Stack>
          ) : (
            <Typography color="text.secondary" sx={{ mt: 1.5 }}>
              当前还没有记录到链重组事件。
            </Typography>
          )}
        </CardContent>
      </Card>

      <Card variant="outlined">
        <CardContent sx={{ p: 2 }}>
          <Typography variant="h6">最近链事件</Typography>
          {recentEvents.length > 0 ? (
            <Stack spacing={1} sx={{ mt: 1.5 }}>
              {recentEvents.map((event, index) => (
                <Stack key={`${event.timestamp}-${index}`} spacing={0.25}>
                  <Typography variant="body2">{event.timestamp} · {event.kind}</Typography>
                  <Typography variant="body2" color="text.secondary">{event.summary}</Typography>
                </Stack>
              ))}
            </Stack>
          ) : (
            <Typography color="text.secondary" sx={{ mt: 1.5 }}>
              当前还没有记录到链事件。
            </Typography>
          )}
        </CardContent>
      </Card>

      <Stack direction={{ xs: 'column', xl: 'row' }} spacing={2.5}>
        <Card variant="outlined" sx={{ flex: 1 }}>
          <CardContent sx={{ p: 2 }}>
            <Typography variant="h6">启动本地节点</Typography>
            <Typography color="text.secondary" sx={{ mt: 1 }}>
              在当前窗口启动 TCP 节点，并设置监听地址、种子节点与矿工地址。
            </Typography>
            <Stack spacing={2} sx={{ mt: 2.5 }}>
              <TextField
                fullWidth
                label="监听地址"
                value={nodeForm.address}
                onChange={(e) => setNodeForm((p) => ({ ...p, address: e.target.value }))}
                placeholder="127.0.0.1:3010 或 127.0.0.1:0"
              />
              <TextField
                fullWidth
                label="Seed 节点"
                value={nodeForm.seed}
                onChange={(e) => setNodeForm((p) => ({ ...p, seed: e.target.value }))}
                placeholder="可选：127.0.0.1:3011"
              />
              <TextField
                fullWidth
                label="矿工地址"
                value={nodeForm.miner}
                onChange={(e) => setNodeForm((p) => ({ ...p, miner: e.target.value }))}
                placeholder="可选，不填则只做普通节点"
              />
              <Button variant="contained" onClick={onStartNode}>
                启动节点
              </Button>
            </Stack>
          </CardContent>
        </Card>

        <Card variant="outlined" sx={{ flex: 1 }}>
          <CardContent sx={{ p: 2 }}>
            <Typography variant="h6">连接已有节点</Typography>
            <Typography color="text.secondary" sx={{ mt: 1 }}>
              选择一个本地节点并指定种子地址，建立连接。
            </Typography>
            <Stack spacing={2} sx={{ mt: 2.5 }}>
              <TextField
                fullWidth
                label="本地节点地址"
                value={connectForm.address}
                onChange={(e) => setConnectForm((p) => ({ ...p, address: e.target.value }))}
                placeholder="例如 127.0.0.1:3010"
              />
              <TextField
                fullWidth
                label="Seed 地址"
                value={connectForm.seed}
                onChange={(e) => setConnectForm((p) => ({ ...p, seed: e.target.value }))}
                placeholder="例如 127.0.0.1:3011"
              />
              <Button variant="contained" color="secondary" onClick={onConnectNode}>
                连接 Seed
              </Button>
            </Stack>
          </CardContent>
        </Card>
      </Stack>

      <Card variant="outlined">
        <CardContent sx={{ p: 2 }}>
          <Typography variant="h6">节点链控制</Typography>
          <Typography color="text.secondary" sx={{ mt: 1 }}>
            对指定节点执行链初始化、交易发送和挖矿。
          </Typography>

          <Stack spacing={2} sx={{ mt: 2.5 }}>
            <TextField
              select
              fullWidth
              label="目标节点"
              value={nodeControlForm.address}
              onChange={(e) => setNodeControlForm((prev) => ({ ...prev, address: e.target.value }))}
              helperText="先选择一个 GUI 托管节点"
            >
              {nodes.length === 0 ? (
                <MenuItem value="" disabled>
                  当前没有运行中的节点
                </MenuItem>
              ) : (
                nodes.map((node) => (
                  <MenuItem key={node.address} value={node.address}>
                    {node.address}
                  </MenuItem>
                ))
              )}
            </TextField>

            <Stack direction={{ xs: 'column', lg: 'row' }} spacing={2}>
              <TextField
                select
                fullWidth
                label="创世奖励地址"
                value={nodeControlForm.rewardAddress}
                onChange={(e) => setNodeControlForm((prev) => ({ ...prev, rewardAddress: e.target.value }))}
                helperText="如果节点还没有链数据，可用该地址初始化本地链"
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
                disabled={!nodeControlForm.address || !nodeControlForm.rewardAddress}
              >
                初始化节点链
              </Button>
            </Stack>

            <Stack direction={{ xs: 'column', lg: 'row' }} spacing={2}>
              <TextField
                select
                fullWidth
                label="发送方钱包"
                value={nodeControlForm.from}
                onChange={(e) => setNodeControlForm((prev) => ({ ...prev, from: e.target.value }))}
              >
                {wallets.map((wallet) => (
                  <MenuItem key={`from-${wallet.address}`} value={wallet.address}>
                    {wallet.address}
                  </MenuItem>
                ))}
              </TextField>
              <TextField
                select
                fullWidth
                label="接收方钱包"
                value={nodeControlForm.to}
                onChange={(e) => setNodeControlForm((prev) => ({ ...prev, to: e.target.value }))}
              >
                {wallets.map((wallet) => (
                  <MenuItem key={`to-${wallet.address}`} value={wallet.address}>
                    {wallet.address}
                  </MenuItem>
                ))}
              </TextField>
            </Stack>

            <Stack direction={{ xs: 'column', lg: 'row' }} spacing={2}>
              <TextField
                fullWidth
                label="金额"
                value={nodeControlForm.amount}
                onChange={(e) => setNodeControlForm((prev) => ({ ...prev, amount: e.target.value }))}
              />
              <TextField
                fullWidth
                label="手续费"
                value={nodeControlForm.fee}
                onChange={(e) => setNodeControlForm((prev) => ({ ...prev, fee: e.target.value }))}
              />
            </Stack>

            <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
              <Button
                variant="contained"
                onClick={onSubmitNodeTransaction}
                disabled={!nodeControlForm.address || !nodeControlForm.from || !nodeControlForm.to}
              >
                通过节点发交易
              </Button>
              <Button variant="outlined" onClick={onMineNode} disabled={!nodeControlForm.address}>
                让节点挖矿
              </Button>
            </Stack>
          </Stack>
        </CardContent>
      </Card>

    </Stack>
  )
}

export default NetworkPage
