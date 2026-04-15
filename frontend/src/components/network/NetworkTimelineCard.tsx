import { Card, CardContent, Chip, Divider, Paper, Stack, Typography } from '@mui/material'
import type { ChainEventView, NodeStatus } from '../../types'
import { buildTimeline } from '../../utils/networkView'

type NetworkTimelineCardProps = {
  nodes: NodeStatus[]
  recentEvents: ChainEventView[]
}

function NetworkTimelineCard({ nodes, recentEvents }: NetworkTimelineCardProps) {
  const timeline = buildTimeline(nodes, recentEvents)

  return (
    <Card variant="outlined">
      <CardContent sx={{ p: 2 }}>
        <Typography variant="h6">事件时间线</Typography>
        <Typography color="text.secondary" sx={{ mt: 0.5 }}>
          按时间合并主链事件与节点网络事件。
        </Typography>

        {timeline.length === 0 ? (
          <Typography color="text.secondary" sx={{ mt: 2 }}>
            暂无时间线事件。执行网络操作后会在此显示。
          </Typography>
        ) : (
          <Stack spacing={1.25} sx={{ mt: 2 }}>
            {timeline.map((item) => (
              <Paper key={item.id} variant="outlined" sx={{ p: 1.25 }}>
                <Stack spacing={0.75}>
                  <Stack direction={{ xs: 'column', md: 'row' }} spacing={1} sx={{ justifyContent: 'space-between', alignItems: { xs: 'flex-start', md: 'center' } }}>
                    <Stack direction="row" spacing={0.75} useFlexGap sx={{ flexWrap: 'wrap' }}>
                      <Chip size="small" label={item.source} variant="outlined" />
                      <Chip size="small" label={item.kind} color={item.tone} variant="outlined" />
                    </Stack>
                    <Typography variant="caption" color="text.secondary">
                      {item.timestamp}
                    </Typography>
                  </Stack>
                  <Divider />
                  <Typography variant="body2" color="text.secondary">
                    {item.detail}
                  </Typography>
                </Stack>
              </Paper>
            ))}
          </Stack>
        )}
      </CardContent>
    </Card>
  )
}

export default NetworkTimelineCard
