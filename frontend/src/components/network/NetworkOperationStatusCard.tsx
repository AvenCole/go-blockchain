import {
  Card,
  CardContent,
  Chip,
  LinearProgress,
  Stack,
  Typography,
} from '@mui/material'
import type { NetworkOperationProgress } from '../../types'

type NetworkOperationStatusCardProps = {
  operation: NetworkOperationProgress | null
}

function NetworkOperationStatusCard({
  operation,
}: NetworkOperationStatusCardProps) {
  return (
    <Card variant="outlined">
      <CardContent sx={{ p: 2 }}>
        <Typography variant="h6">流程状态</Typography>

        {!operation ? (
          <Typography color="text.secondary" sx={{ mt: 1.5 }}>
            当前没有运行中的网络流程。
          </Typography>
        ) : (
          <Stack spacing={1.5} sx={{ mt: 1.5 }}>
            <Stack
              direction={{ xs: 'column', md: 'row' }}
              spacing={1}
              sx={{
                justifyContent: 'space-between',
                alignItems: { xs: 'flex-start', md: 'center' },
              }}
            >
              <Stack spacing={0.5}>
                <Typography variant="subtitle2">
                  {labelForOperation(operation.operation)}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  阶段：{operation.phase}
                </Typography>
              </Stack>
              <Stack direction="row" spacing={0.75}>
                <Chip
                  size="small"
                  label={labelForStatus(operation.status)}
                  color={colorForStatus(operation.status)}
                  variant="outlined"
                />
                <Chip
                  size="small"
                  label={`${operation.currentStep}/${operation.totalSteps}`}
                  variant="outlined"
                />
              </Stack>
            </Stack>

            <LinearProgress
              variant="determinate"
              value={progressValue(operation)}
              color={operation.status === 'failed' ? 'error' : 'primary'}
            />

            <Typography variant="body2">{operation.message}</Typography>

            {operation.error ? (
              <Typography variant="body2" color="error.main">
                错误：{operation.error}
              </Typography>
            ) : null}

            {operation.summary ? (
              <Typography variant="body2" color="text.secondary">
                摘要：{operation.summary}
              </Typography>
            ) : null}

            <Typography variant="caption" color="text.secondary">
              开始：{operation.startedAt}
              {operation.finishedAt ? ` · 结束：${operation.finishedAt}` : ''}
              {` · 耗时：${formatElapsed(operation.elapsedMs)}`}
            </Typography>
          </Stack>
        )}
      </CardContent>
    </Card>
  )
}

function labelForOperation(operation: string): string {
  switch (operation) {
    case 'network.quick-demo':
      return '快速同步流程'
    case 'network.reorg-demo':
      return '重组流程'
    case 'network.partition-demo':
      return '分区 / 合流流程'
    default:
      return operation
  }
}

function labelForStatus(status: NetworkOperationProgress['status']): string {
  switch (status) {
    case 'started':
      return '已开始'
    case 'progress':
      return '进行中'
    case 'completed':
      return '已完成'
    case 'failed':
      return '失败'
    default:
      return status
  }
}

function colorForStatus(status: NetworkOperationProgress['status']) {
  switch (status) {
    case 'completed':
      return 'success'
    case 'failed':
      return 'error'
    case 'started':
    case 'progress':
      return 'primary'
    default:
      return 'default'
  }
}

function progressValue(operation: NetworkOperationProgress): number {
  if (operation.status === 'completed') {
    return 100
  }
  if (operation.totalSteps <= 0) {
    return 0
  }
  return Math.min(
    100,
    Math.max(0, (operation.currentStep / operation.totalSteps) * 100),
  )
}

function formatElapsed(value: number): string {
  if (value < 1000) {
    return `${value}ms`
  }

  return `${(value / 1000).toFixed(1)}s`
}

export default NetworkOperationStatusCard
