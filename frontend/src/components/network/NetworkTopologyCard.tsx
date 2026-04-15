import { Card, CardContent, Chip, Divider, Paper, Stack, Typography } from '@mui/material'
import type { NodeStatus } from '../../types'
import { buildTopology, shortAddress } from '../../utils/networkView'

type NetworkTopologyCardProps = {
  nodes: NodeStatus[]
}

function NetworkTopologyCard({ nodes }: NetworkTopologyCardProps) {
  const topology = buildTopology(nodes)

  return (
    <Card variant="outlined">
      <CardContent sx={{ p: 2 }}>
        <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.5} sx={{ justifyContent: 'space-between', alignItems: { xs: 'flex-start', md: 'center' } }}>
          <div>
            <Typography variant="h6">网络拓扑总览</Typography>
            <Typography color="text.secondary" sx={{ mt: 0.5 }}>
              显示节点连接关系与当前 tip 收敛状态。
            </Typography>
          </div>
          <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} useFlexGap sx={{ flexWrap: 'wrap' }}>
            <Chip label={`节点 ${topology.nodes.length}`} variant="outlined" />
            <Chip label={`连接 ${topology.links.length}`} variant="outlined" />
            <Chip label={`孤立节点 ${topology.isolatedCount}`} variant="outlined" />
            <Chip
              label={topology.converged ? `已收敛 ${topology.sharedTipHash}` : `未收敛 ${topology.uniqueTipCount || 0} 条 tip`}
              color={topology.converged ? 'success' : 'warning'}
              variant="outlined"
            />
          </Stack>
        </Stack>

        {topology.nodes.length === 0 ? (
          <Typography color="text.secondary" sx={{ mt: 2 }}>
            当前还没有运行中的节点，启动节点后这里会自动形成拓扑视图。
          </Typography>
        ) : (
          <Stack spacing={2} sx={{ mt: 2 }}>
            <Stack
              direction={{ xs: 'column', xl: 'row' }}
              spacing={1.5}
              sx={{
                '& > *': {
                  flex: 1,
                  minWidth: 0,
                },
              }}
            >
              {topology.nodes.map((node) => (
                <Paper key={node.address} variant="outlined" sx={{ p: 1.5 }}>
                  <Stack spacing={1}>
                    <Stack direction="row" spacing={1} sx={{ alignItems: 'center', justifyContent: 'space-between' }}>
                      <Typography variant="subtitle2">{node.shortAddress}</Typography>
                      <Stack direction="row" spacing={0.75}>
                        {node.isMiner ? <Chip size="small" label="Miner" color="secondary" variant="outlined" /> : null}
                        {node.hasReorg ? <Chip size="small" label="Reorg" color="warning" variant="outlined" /> : null}
                      </Stack>
                    </Stack>
                    <Typography variant="body2" color="text.secondary" sx={{ wordBreak: 'break-all' }}>
                      {node.address}
                    </Typography>
                    <Divider />
                    <Typography variant="body2">height={node.initialized ? node.height : '未初始化'}</Typography>
                    <Typography variant="body2">mempool={node.mempoolCount}</Typography>
                    <Typography variant="body2">orphans={node.orphanCount}</Typography>
                    <Typography variant="body2" sx={{ wordBreak: 'break-all' }}>
                      tip={node.shortTipHash}
                    </Typography>
                    <Divider />
                    <Typography variant="body2">connected peers={node.peerCount}</Typography>
                    {node.connectedPeers.length === 0 ? (
                      <Typography variant="body2" color="text.secondary">
                        当前没有连到其他 GUI 节点
                      </Typography>
                    ) : (
                      <Stack direction="row" spacing={0.75} useFlexGap sx={{ flexWrap: 'wrap' }}>
                        {node.connectedPeers.map((peer) => (
                          <Chip key={`${node.address}-${peer}`} size="small" label={shortAddress(peer)} variant="outlined" />
                        ))}
                      </Stack>
                    )}
                  </Stack>
                </Paper>
              ))}
            </Stack>

            <Paper variant="outlined" sx={{ p: 1.5 }}>
              <Typography variant="subtitle2">连接关系</Typography>
              {topology.links.length === 0 ? (
                <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                  当前还没有可视化连接关系。
                </Typography>
              ) : (
                <Stack spacing={1} sx={{ mt: 1.25 }}>
                  {topology.links.map((link) => (
                    <Stack
                      key={link.key}
                      direction={{ xs: 'column', md: 'row' }}
                      spacing={1}
                      sx={{ justifyContent: 'space-between', alignItems: { xs: 'flex-start', md: 'center' } }}
                    >
                      <Typography variant="body2">
                        {link.fromShort} ⇄ {link.toShort}
                      </Typography>
                      <Chip size="small" label={link.mutual ? '双向已知' : '单向可见'} color={link.mutual ? 'success' : 'warning'} variant="outlined" />
                    </Stack>
                  ))}
                </Stack>
              )}
            </Paper>
          </Stack>
        )}
      </CardContent>
    </Card>
  )
}

export default NetworkTopologyCard
