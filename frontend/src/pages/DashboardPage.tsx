import {
  Box,
  Button,
  Card,
  CardContent,
  Divider,
  List,
  ListItem,
  ListItemText,
  MenuItem,
  Stack,
  TextField,
  Typography,
} from '@mui/material'
import StatCard from '../components/StatCard'
import type {
  BlockView,
  DashboardData,
  MultiSigOutputView,
  NodeStatus,
  WalletView,
} from '../types'
import { shortHash } from '../utils/format'

type DashboardPageProps = {
  dashboard: DashboardData | null
  latestBlock: BlockView | null
  wallets: WalletView[]
  mempool: string[]
  multiSigOutputs: MultiSigOutputView[]
  nodes: NodeStatus[]
  chainInitAddress: string
  setChainInitAddress: React.Dispatch<React.SetStateAction<string>>
  onInitializeBlockchain: () => Promise<void>
  isInitializingBlockchain: boolean
}

function DashboardPage({
  dashboard,
  latestBlock,
  wallets,
  mempool,
  multiSigOutputs,
  nodes,
  chainInitAddress,
  setChainInitAddress,
  onInitializeBlockchain,
  isInitializingBlockchain,
}: DashboardPageProps) {
  if (!dashboard) return null

  const recentEvents = dashboard.recentEvents ?? []
  const latestNodeEvents = nodes
    .flatMap((node) =>
      (node.recentEvents ?? []).slice(0, 2).map((event) => ({
        node: node.address,
        ...event,
      })),
    )
    .slice(0, 6)

  return (
    <Stack spacing={2}>
      {dashboard.height < 0 ? (
        <Card variant="outlined">
          <CardContent sx={{ p: 2.25 }}>
            <Typography variant="h6">主链初始化</Typography>
            <Stack
              direction="row"
              spacing={2}
              sx={{ mt: 2, alignItems: 'center' }}
            >
              <TextField
                select
                fullWidth
                label="钱包地址"
                value={chainInitAddress}
                onChange={(e) => setChainInitAddress(e.target.value)}
              >
                {wallets.length === 0 ? (
                  <MenuItem value="" disabled>
                    请先创建钱包
                  </MenuItem>
                ) : (
                  wallets.map((wallet) => (
                    <MenuItem key={wallet.address} value={wallet.address}>
                      {wallet.address}
                    </MenuItem>
                  ))
                )}
              </TextField>
              <Button
                variant="contained"
                disabled={!chainInitAddress || isInitializingBlockchain}
                onClick={onInitializeBlockchain}
              >
                {isInitializingBlockchain ? '初始化中...' : '初始化主链'}
              </Button>
            </Stack>
          </CardContent>
        </Card>
      ) : null}

      <Box
        sx={{
          display: 'grid',
          gap: 2,
          gridTemplateColumns: 'repeat(4, minmax(0, 1fr))',
        }}
      >
        <StatCard
          title="区块高度"
          value={dashboard.height >= 0 ? dashboard.height : '--'}
          secondary={shortHash(dashboard.latestHash) || '尚未建链'}
        />
        <StatCard
          title="待打包交易"
          value={dashboard.pendingTxCount}
          secondary="Mempool"
        />
        <StatCard
          title="钱包数量"
          value={dashboard.walletCount}
          secondary={`${wallets.length} 个地址`}
        />
        <StatCard
          title="网络模式"
          value={dashboard.networkMode}
          secondary={`难度 ${dashboard.difficulty ?? '-'}`}
        />
      </Box>

      <Box
        sx={{
          display: 'grid',
          gap: 2,
          gridTemplateColumns: 'minmax(0, 1.3fr) 360px',
        }}
      >
        <Stack spacing={2}>
          <Box
            sx={{
              display: 'grid',
              gap: 2,
              gridTemplateColumns: 'minmax(0, 1.1fr) minmax(320px, 0.9fr)',
            }}
          >
            <Card variant="outlined">
              <CardContent sx={{ p: 2.25 }}>
                <Typography variant="h6">最新区块</Typography>
                {latestBlock ? (
                  <Stack spacing={1.1} sx={{ mt: 2 }}>
                    <Typography variant="body2">Hash: {shortHash(latestBlock.hash, 16, 12)}</Typography>
                    <Typography variant="body2">Prev: {shortHash(latestBlock.prevHash, 16, 12)}</Typography>
                    <Typography variant="body2">Merkle: {shortHash(latestBlock.merkleRoot, 16, 12)}</Typography>
                    <Typography variant="body2">
                      PoW: difficulty={latestBlock.difficulty} nonce={latestBlock.nonce} valid={String(latestBlock.powValid)}
                    </Typography>
                    <Typography variant="body2">Time: {latestBlock.timestamp}</Typography>
                    <Typography variant="body2">交易数: {latestBlock.transactionCount}</Typography>
                  </Stack>
                ) : (
                  <Typography sx={{ mt: 2 }} color="text.secondary">
                    当前还没有区块数据。
                  </Typography>
                )}
              </CardContent>
            </Card>

            <Card variant="outlined">
              <CardContent sx={{ p: 2.25 }}>
                <Typography variant="h6">钱包总览</Typography>
                {wallets.length > 0 ? (
                  <List
                    dense
                    disablePadding
                    sx={{ mt: 1.5, maxHeight: 260, overflow: 'auto' }}
                  >
                    {wallets.slice(0, 6).map((wallet) => (
                      <ListItem key={wallet.address} disablePadding sx={{ py: 0.7 }}>
                        <ListItemText
                          primary={shortHash(wallet.address, 14, 10)}
                          secondary={`balance=${wallet.balance}`}
                        />
                      </ListItem>
                    ))}
                  </List>
                ) : (
                  <Typography sx={{ mt: 2 }} color="text.secondary">
                    当前还没有钱包。
                  </Typography>
                )}
              </CardContent>
            </Card>
          </Box>

          <Box
            sx={{
              display: 'grid',
              gap: 2,
              gridTemplateColumns: '1fr 1fr',
            }}
          >
            <Card variant="outlined">
              <CardContent sx={{ p: 2.25 }}>
                <Typography variant="h6">最近链切换 / 重组状态</Typography>
                {dashboard.lastReorg ? (
                  <Stack spacing={1.1} sx={{ mt: 2 }}>
                    <Typography variant="body2">时间：{dashboard.lastReorg.timestamp}</Typography>
                    <Typography variant="body2">
                      高度：{dashboard.lastReorg.oldHeight} → {dashboard.lastReorg.newHeight}
                    </Typography>
                    <Typography variant="body2">
                      旧 Tip：{shortHash(dashboard.lastReorg.oldTip, 14, 12)}
                    </Typography>
                    <Typography variant="body2">
                      新 Tip：{shortHash(dashboard.lastReorg.newTip, 14, 12)}
                    </Typography>
                    <Typography variant="body2">
                      恢复交易：{dashboard.lastReorg.restoredTxCount}
                    </Typography>
                    <Typography variant="body2">
                      清理已确认：{dashboard.lastReorg.droppedConfirmedCount}
                    </Typography>
                  </Stack>
                ) : (
                  <Typography sx={{ mt: 2 }} color="text.secondary">
                    当前还没有记录到链重组事件。
                  </Typography>
                )}
              </CardContent>
            </Card>

            <Card variant="outlined">
              <CardContent sx={{ p: 2.25 }}>
                <Typography variant="h6">最近链事件</Typography>
                {recentEvents.length > 0 ? (
                  <Stack spacing={1.2} sx={{ mt: 2 }}>
                    {recentEvents.map((event, index) => (
                      <Stack key={`${event.timestamp}-${index}`} spacing={0.35}>
                        <Typography variant="body2">
                          {event.timestamp} · {event.kind}
                        </Typography>
                        <Typography variant="body2" color="text.secondary">
                          {event.summary}
                        </Typography>
                      </Stack>
                    ))}
                  </Stack>
                ) : (
                  <Typography sx={{ mt: 2 }} color="text.secondary">
                    当前还没有记录到链事件。
                  </Typography>
                )}
              </CardContent>
            </Card>
          </Box>
        </Stack>

        <Stack spacing={2}>
            <Card variant="outlined">
              <CardContent sx={{ p: 2.25 }}>
                <Typography variant="h6">Mempool</Typography>
                {mempool.length > 0 ? (
                <List dense disablePadding sx={{ mt: 1.5, maxHeight: 220, overflow: 'auto' }}>
                  {mempool.slice(0, 8).map((txid) => (
                    <ListItem key={txid} disablePadding sx={{ py: 0.7 }}>
                      <ListItemText primary={shortHash(txid, 14, 10)} />
                    </ListItem>
                  ))}
                </List>
              ) : (
                <Typography sx={{ mt: 2 }} color="text.secondary">
                  当前没有待打包交易。
                </Typography>
              )}
            </CardContent>
          </Card>

            <Card variant="outlined">
              <CardContent sx={{ p: 2.25 }}>
                <Typography variant="h6">未花费多签输出</Typography>
                {multiSigOutputs.length > 0 ? (
                <List dense disablePadding sx={{ mt: 1.5, maxHeight: 220, overflow: 'auto' }}>
                  {multiSigOutputs.slice(0, 6).map((item) => (
                    <ListItem key={`${item.txid}:${item.out}`} disablePadding sx={{ py: 0.7 }}>
                      <ListItemText
                        primary={`${shortHash(item.txid, 12, 8)}:${item.out} value=${item.value}`}
                        secondary={`${item.required}/${item.participants.length} | ${item.participants.join(', ')}`}
                      />
                    </ListItem>
                  ))}
                </List>
              ) : (
                <Typography sx={{ mt: 2 }} color="text.secondary">
                  当前没有未花费多签输出。
                </Typography>
              )}
            </CardContent>
          </Card>

            <Card variant="outlined">
              <CardContent sx={{ p: 2.25 }}>
                <Typography variant="h6">节点活动摘要</Typography>
                {latestNodeEvents.length > 0 ? (
                <List dense disablePadding sx={{ mt: 1.5, maxHeight: 220, overflow: 'auto' }}>
                  {latestNodeEvents.map((event, index) => (
                    <ListItem
                      key={`${event.node}-${event.timestamp}-${index}`}
                      disablePadding
                      sx={{ py: 0.7 }}
                    >
                      <ListItemText
                        primary={`${shortHash(event.node, 10, 8)} · ${event.kind}`}
                        secondary={event.detail}
                      />
                    </ListItem>
                  ))}
                </List>
              ) : (
                <Typography sx={{ mt: 2 }} color="text.secondary">
                  当前没有节点事件。
                </Typography>
              )}
            </CardContent>
          </Card>

          <Card variant="outlined">
            <CardContent sx={{ p: 2.25 }}>
              <Typography variant="h6">数据目录</Typography>
              <Divider sx={{ my: 1.5 }} />
              <Typography variant="body2" color="text.secondary" sx={{ wordBreak: 'break-all' }}>
                {dashboard.dataDir}
              </Typography>
            </CardContent>
          </Card>
        </Stack>
      </Box>
    </Stack>
  )
}

export default DashboardPage
