import { useEffect, useMemo, useState } from 'react'
import { Card, CardContent, Chip, Divider, Paper, Stack, Typography } from '@mui/material'
import type { ChainEventView, NodeStatus } from '../../types'
import EventFilterToolbar from './EventFilterToolbar'
import {
  buildTimeline,
  collectKinds,
  collectSources,
  filterTimelineItems,
} from '../../utils/networkView'

type NetworkTimelineCardProps = {
  nodes: NodeStatus[]
  recentEvents: ChainEventView[]
}

function NetworkTimelineCard({ nodes, recentEvents }: NetworkTimelineCardProps) {
  const [query, setQuery] = useState('')
  const [kind, setKind] = useState('')
  const [source, setSource] = useState('')
  const timeline = useMemo(() => buildTimeline(nodes, recentEvents), [nodes, recentEvents])
  const kindOptions = useMemo(() => collectKinds(timeline), [timeline])
  const sourceOptions = useMemo(() => collectSources(timeline), [timeline])
  const filteredTimeline = useMemo(
    () => filterTimelineItems(timeline, { kind, query, source }).slice(0, 24),
    [kind, query, source, timeline],
  )

  useEffect(() => {
    if (kind && !kindOptions.includes(kind)) {
      setKind('')
    }
  }, [kind, kindOptions])

  useEffect(() => {
    if (source && !sourceOptions.includes(source)) {
      setSource('')
    }
  }, [source, sourceOptions])

  return (
    <Card variant="outlined">
      <CardContent sx={{ p: 2 }}>
        <Typography variant="h6">事件时间线</Typography>
        <Typography color="text.secondary" sx={{ mt: 0.5 }}>
          按时间合并主链事件与节点网络事件。
        </Typography>
        <EventFilterToolbar
          query={query}
          onQueryChange={setQuery}
          kind={kind}
          kindOptions={kindOptions}
          onKindChange={setKind}
          source={source}
          sourceOptions={sourceOptions}
          onSourceChange={setSource}
          matchedCount={filteredTimeline.length}
          totalCount={timeline.length}
        />

        {timeline.length === 0 ? (
          <Typography color="text.secondary" sx={{ mt: 2 }}>
            暂无时间线事件。执行网络操作后会在此显示。
          </Typography>
        ) : filteredTimeline.length === 0 ? (
          <Typography color="text.secondary" sx={{ mt: 2 }}>
            当前筛选条件下没有匹配的时间线事件。
          </Typography>
        ) : (
          <Stack spacing={1.25} sx={{ mt: 2 }}>
            {filteredTimeline.map((item) => (
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
