import {
  Button,
  Card,
  CardContent,
  Chip,
  Divider,
  List,
  ListItem,
  ListItemText,
  MenuItem,
  Stack,
  TextField,
  Typography,
} from '@mui/material'
import type { ChainEventView, NetworkDemoResult, NodeStatus, ReorgStatusView, WalletView } from '../types'

type NetworkPageProps = {
  nodes: NodeStatus[]
  wallets: WalletView[]
  networkDemo?: NetworkDemoResult | null
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
}

function NetworkPage({
  nodes,
  wallets,
  networkDemo,
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
}: NetworkPageProps) {
  return (
    <Stack spacing={2.5}>
      <Card variant="outlined">
        <CardContent sx={{ p: 2 }}>
          <Typography variant="h6">一键网络演示</Typography>
          <Typography color="text.secondary" sx={{ mt: 1 }}>
            自动创建双节点同步场景：准备钱包、启动两个节点、初始化主节点链、连接 peer、发送交易并挖矿。
          </Typography>
          <Stack direction={{ xs: 'column', md: 'row' }} spacing={2} sx={{ mt: 2.5, alignItems: { md: 'center' } }}>
            <Button variant="contained" color="primary" onClick={onRunNetworkQuickDemo}>
              运行快速演示
            </Button>
            {networkDemo ? (
              <Stack spacing={0.5}>
                <Typography variant="body2">source={networkDemo.sourceNode}</Typography>
                <Typography variant="body2">peer={networkDemo.peerNode} · peerHeight={networkDemo.peerHeight}</Typography>
                <Typography variant="body2">tipAnnounced={String(networkDemo.tipAnnounced)}</Typography>
              </Stack>
            ) : (
              <Typography variant="body2" color="text.secondary">
                适合答辩时快速搭建“节点同步 + 交易广播 + 出块同步”演示链路。
              </Typography>
            )}
          </Stack>
        </CardContent>
      </Card>

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
              在 GUI 中直接拉起 TCP 节点，便于演示节点监听、矿工地址绑定和种子接入。
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
              用于演示节点发现和同步入口，先选择一个本地节点，再指定种子地址。
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
            直接在 GUI 中初始化节点链、通过指定节点发送交易，并触发该节点挖矿，适合演示网络节点完整生命周期。
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

      <Card variant="outlined">
        <CardContent sx={{ p: 2 }}>
          <Stack
            direction={{ xs: 'column', md: 'row' }}
            spacing={1.5}
            sx={{ justifyContent: 'space-between', alignItems: { xs: 'flex-start', md: 'center' } }}
          >
            <div>
              <Typography variant="h6">节点状态</Typography>
              <Typography color="text.secondary" sx={{ mt: 0.5 }}>
                展示当前由 GUI 托管的节点、区块高度、矿工配置与已知 Peer。
              </Typography>
            </div>
            <Chip label={`GUI 节点数 ${nodes.length}`} variant="outlined" color="primary" />
          </Stack>

          <List sx={{ mt: 2 }}>
            {nodes.length === 0 ? (
              <ListItem>
                <ListItemText primary="当前还没有运行中的 GUI 节点" secondary="可先使用上方表单启动一个本地节点，再执行连接演示。" />
              </ListItem>
            ) : (
              nodes.map((node) => (
                <ListItem
                  key={node.address}
                  divider
                  alignItems="flex-start"
                  secondaryAction={
                    <Button color="error" onClick={() => void onStopNode(node.address)}>
                      停止
                    </Button>
                  }
                  sx={{ pr: 12 }}
                >
                  <ListItemText
                    primary={`${node.address}  (${node.initialized ? `height=${node.height}` : '未初始化'})`}
                    secondary={
                      <Stack spacing={0.75} sx={{ mt: 1, alignItems: 'flex-start' }}>
                        <Typography variant="body2">miner={node.minerAddress || '(none)'}</Typography>
                        <Typography variant="body2">initialized={String(node.initialized)}</Typography>
                        <Typography variant="body2">mempool={node.mempoolCount}</Typography>
                        <Typography variant="body2">running={String(node.running)}</Typography>
                        <Typography variant="body2">height={node.initialized ? node.height : '未初始化'}</Typography>
                        <Typography variant="body2">orphans={node.orphanCount}</Typography>
                        <Divider />
                        <Typography variant="body2">recent events:</Typography>
                        {(node.recentEvents ?? []).length === 0 ? (
                          <Typography variant="body2" color="text.secondary">
                            暂无网络事件
                          </Typography>
                        ) : (
                          (node.recentEvents ?? []).slice(0, 4).map((event, idx) => (
                            <Typography key={`${node.address}-event-${idx}`} variant="body2" color="text.secondary">
                              {event.timestamp} · {event.kind} · {event.detail}
                            </Typography>
                          ))
                        )}
                        <Divider />
                        <Typography variant="body2">peers:</Typography>
                        {node.peers.length === 0 ? (
                          <Typography variant="body2" color="text.secondary">
                            暂无 peer
                          </Typography>
                        ) : (
                          node.peers.map((peer) => (
                            <Typography key={`${node.address}-${peer}`} variant="body2" color="text.secondary">
                              {peer}
                            </Typography>
                          ))
                        )}
                      </Stack>
                    }
                  />
                </ListItem>
              ))
            )}
          </List>
        </CardContent>
      </Card>
    </Stack>
  )
}

export default NetworkPage
