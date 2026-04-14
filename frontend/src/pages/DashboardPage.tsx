import { Box, Card, CardContent, Divider, Stack, Typography } from '@mui/material'
import StatCard from '../components/StatCard'
import type { BlockView, DashboardData } from '../types'
import { shortHash } from '../utils/format'

type DashboardPageProps = {
  dashboard: DashboardData | null
  latestBlock: BlockView | null
}

function DashboardPage({ dashboard, latestBlock }: DashboardPageProps) {
  if (!dashboard) return null
  const recentEvents = dashboard.recentEvents ?? []

  return (
    <Box sx={{ display: 'grid', gap: 2, gridTemplateColumns: { xs: '1fr', md: 'repeat(4, 1fr)' } }}>
      <Box>
        <StatCard title="区块高度" value={dashboard.height} secondary={shortHash(dashboard.latestHash) || '尚未建链'} />
      </Box>
      <Box>
        <StatCard title="待打包交易" value={dashboard.pendingTxCount} secondary="Mempool 中交易数" />
      </Box>
      <Box>
        <StatCard title="钱包数量" value={dashboard.walletCount} secondary={dashboard.dataDir} />
      </Box>
      <Box>
        <StatCard title="网络模式" value={dashboard.networkMode} secondary={`难度 ${dashboard.difficulty ?? '-'}`} />
      </Box>
      <Box sx={{ gridColumn: '1 / -1' }}>
        <Card variant="outlined">
          <CardContent>
            <Typography variant="h6">最新区块</Typography>
            {latestBlock ? (
              <Stack spacing={1.25} sx={{ mt: 2 }}>
                <Typography>Hash: {shortHash(latestBlock.hash, 14, 12)}</Typography>
                <Typography>MerkleRoot: {shortHash(latestBlock.merkleRoot, 14, 12)}</Typography>
                <Typography>PoW: difficulty={latestBlock.difficulty} nonce={latestBlock.nonce}</Typography>
                <Typography>交易数: {latestBlock.transactionCount}</Typography>
                <Divider sx={{ my: 1 }} />
                <Typography color="text.secondary">最新区块用于快速确认当前链状态、MerkleRoot 和 PoW 是否正常。</Typography>
              </Stack>
            ) : (
              <Typography sx={{ mt: 2 }} color="text.secondary">
                当前还没有区块数据。
              </Typography>
            )}
          </CardContent>
        </Card>
      </Box>
      <Box sx={{ gridColumn: '1 / -1' }}>
        <Card variant="outlined">
          <CardContent>
            <Typography variant="h6">最近链切换 / 重组状态</Typography>
            {dashboard.lastReorg ? (
              <Stack spacing={1.25} sx={{ mt: 2 }}>
                <Typography>时间: {dashboard.lastReorg.timestamp}</Typography>
                <Typography>旧高度: {dashboard.lastReorg.oldHeight} / 新高度: {dashboard.lastReorg.newHeight}</Typography>
                <Typography>旧 Tip: {shortHash(dashboard.lastReorg.oldTip, 14, 12)}</Typography>
                <Typography>新 Tip: {shortHash(dashboard.lastReorg.newTip, 14, 12)}</Typography>
                <Typography>恢复交易数: {dashboard.lastReorg.restoredTxCount}</Typography>
                <Typography>清理已确认交易数: {dashboard.lastReorg.droppedConfirmedCount}</Typography>
              </Stack>
            ) : (
              <Typography sx={{ mt: 2 }} color="text.secondary">
                当前还没有记录到链重组事件。
              </Typography>
            )}
          </CardContent>
        </Card>
      </Box>
      <Box sx={{ gridColumn: '1 / -1' }}>
        <Card variant="outlined">
          <CardContent>
            <Typography variant="h6">最近链事件</Typography>
            {recentEvents.length > 0 ? (
              <Stack spacing={1.25} sx={{ mt: 2 }}>
                {recentEvents.map((event, index) => (
                  <Stack key={`${event.timestamp}-${index}`} spacing={0.4}>
                    <Typography>{event.timestamp} · {event.kind}</Typography>
                    <Typography color="text.secondary">{event.summary}</Typography>
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
    </Box>
  )
}

export default DashboardPage
