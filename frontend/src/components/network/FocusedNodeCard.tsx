import { useEffect, useMemo, useState } from 'react'
import {
  Button,
  Card,
  CardContent,
  Chip,
  Divider,
  Paper,
  Stack,
  Typography,
} from '@mui/material'
import type { NodeStatus } from '../../types'
import {
  collectKinds,
  filterNodeEvents,
  shortAddress,
  shortHash,
} from '../../utils/networkView'
import EventFilterToolbar from './EventFilterToolbar'

type FocusedNodeCardProps = {
  node: NodeStatus | null
  onUseAsConnectNode: (address: string) => void
  onUseAsSeed: (address: string) => void
  onUseAsControlNode: (address: string) => void
  onStopNode: (address: string) => Promise<void>
  isStopDisabled: boolean
  isStopping: boolean
}

function FocusedNodeCard({
  node,
  onUseAsConnectNode,
  onUseAsSeed,
  onUseAsControlNode,
  onStopNode,
  isStopDisabled,
  isStopping,
}: FocusedNodeCardProps) {
  const [eventQuery, setEventQuery] = useState('')
  const [eventKind, setEventKind] = useState('')
  const nodeEvents = node?.recentEvents ?? []
  const eventKindOptions = useMemo(() => collectKinds(nodeEvents), [nodeEvents])
  const filteredNodeEvents = useMemo(
    () => filterNodeEvents(nodeEvents, { kind: eventKind, query: eventQuery }).slice(0, 12),
    [eventKind, eventQuery, nodeEvents],
  )

  useEffect(() => {
    setEventQuery('')
    setEventKind('')
  }, [node?.address])

  useEffect(() => {
    if (eventKind && !eventKindOptions.includes(eventKind)) {
      setEventKind('')
    }
  }, [eventKind, eventKindOptions])

  return (
    <Card variant="outlined">
      <CardContent sx={{ p: 2 }}>
        <Typography variant="h6">节点焦点面板</Typography>

        {!node ? (
          <Typography color="text.secondary" sx={{ mt: 2 }}>
            选择一个节点后可查看详细状态并将地址快速填入操作表单。
          </Typography>
        ) : (
          <Stack spacing={2} sx={{ mt: 2 }}>
            <Stack
              direction={{ xs: 'column', md: 'row' }}
              spacing={1.5}
              sx={{ justifyContent: 'space-between', alignItems: { xs: 'flex-start', md: 'center' } }}
            >
              <Stack spacing={0.75}>
                <Typography variant="subtitle1" sx={{ fontWeight: 600 }}>
                  {shortAddress(node.address, 14, 8)}
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ wordBreak: 'break-all' }}>
                  {node.address}
                </Typography>
              </Stack>
              <Stack direction="row" spacing={0.75} useFlexGap sx={{ flexWrap: 'wrap' }}>
                <Chip
                  size="small"
                  label={node.running ? '运行中' : '已停止'}
                  color={node.running ? 'success' : 'default'}
                  variant="outlined"
                />
                <Chip
                  size="small"
                  label={node.initialized ? `height=${node.height}` : '未初始化'}
                  color={node.initialized ? 'primary' : 'default'}
                  variant="outlined"
                />
                {node.minerAddress ? (
                  <Chip size="small" label="Miner" color="secondary" variant="outlined" />
                ) : null}
              </Stack>
            </Stack>

            <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.25} useFlexGap sx={{ flexWrap: 'wrap' }}>
              <Button variant="contained" onClick={() => onUseAsConnectNode(node.address)}>
                填入连接节点
              </Button>
              <Button variant="outlined" onClick={() => onUseAsSeed(node.address)}>
                填入 Seed
              </Button>
              <Button variant="outlined" color="secondary" onClick={() => onUseAsControlNode(node.address)}>
                填入链控制
              </Button>
              <Button
                variant="outlined"
                color="error"
                disabled={isStopDisabled}
                onClick={() => void onStopNode(node.address)}
              >
                {isStopping ? '停止中...' : '停止节点'}
              </Button>
            </Stack>

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
              <Paper variant="outlined" sx={{ p: 1.5 }}>
                <Stack spacing={0.75}>
                  <Typography variant="subtitle2">链状态</Typography>
                  <Typography variant="body2">height={node.initialized ? node.height : '未初始化'}</Typography>
                  <Typography variant="body2">mempool={node.mempoolCount}</Typography>
                  <Typography variant="body2">orphans={node.orphanCount}</Typography>
                  <Typography variant="body2">peers={node.peers.length}</Typography>
                  <Typography variant="body2">miner={node.minerAddress || '(none)'}</Typography>
                  <Typography variant="body2" sx={{ wordBreak: 'break-all' }}>
                    tip={shortHash(node.tipHash)}
                  </Typography>
                </Stack>
              </Paper>

              <Paper variant="outlined" sx={{ p: 1.5 }}>
                <Stack spacing={0.75}>
                  <Typography variant="subtitle2">最近重组</Typography>
                  {node.lastReorg ? (
                    <>
                      <Typography variant="body2">
                        {node.lastReorg.oldHeight} → {node.lastReorg.newHeight}
                      </Typography>
                      <Typography variant="body2" color="text.secondary" sx={{ wordBreak: 'break-all' }}>
                        {shortHash(node.lastReorg.oldTip)} → {shortHash(node.lastReorg.newTip)}
                      </Typography>
                      <Typography variant="body2">
                        restored={node.lastReorg.restoredTxCount} · dropped={node.lastReorg.droppedConfirmedCount}
                      </Typography>
                    </>
                  ) : (
                    <Typography variant="body2" color="text.secondary">
                      暂无重组记录
                    </Typography>
                  )}
                </Stack>
              </Paper>
            </Stack>

            <Paper variant="outlined" sx={{ p: 1.5 }}>
              <Stack spacing={1}>
                <Typography variant="subtitle2">最近网络事件</Typography>
                <EventFilterToolbar
                  query={eventQuery}
                  onQueryChange={setEventQuery}
                  kind={eventKind}
                  kindOptions={eventKindOptions}
                  onKindChange={setEventKind}
                  matchedCount={filteredNodeEvents.length}
                  totalCount={nodeEvents.length}
                />
                <Divider />
                {nodeEvents.length === 0 ? (
                  <Typography variant="body2" color="text.secondary">
                    暂无网络事件
                  </Typography>
                ) : filteredNodeEvents.length === 0 ? (
                  <Typography variant="body2" color="text.secondary">
                    当前筛选条件下没有匹配的节点事件
                  </Typography>
                ) : (
                  filteredNodeEvents.map((event, index) => (
                    <Stack key={`${node.address}-recent-${index}`} spacing={0.25}>
                      <Typography variant="body2">
                        {event.timestamp} · {event.kind}
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        {event.detail}
                      </Typography>
                    </Stack>
                  ))
                )}
              </Stack>
            </Paper>

            <Paper variant="outlined" sx={{ p: 1.5 }}>
              <Stack spacing={1}>
                <Typography variant="subtitle2">Peer 列表</Typography>
                <Divider />
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
            </Paper>
          </Stack>
        )}
      </CardContent>
    </Card>
  )
}

export default FocusedNodeCard
