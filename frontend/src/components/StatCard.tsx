import { Card, CardContent, Chip, Stack, Typography } from '@mui/material'
import type { ReactNode } from 'react'

type StatCardProps = {
  title: string
  value: ReactNode
  secondary?: ReactNode
}

function StatCard({ title, value, secondary }: StatCardProps) {
  return (
    <Card variant="outlined" sx={{ height: '100%' }}>
      <CardContent>
        <Stack spacing={1.25}>
          <Chip label={title} size="small" sx={{ alignSelf: 'flex-start' }} />
          <Typography variant="h5">{value}</Typography>
          {secondary ? (
            <Typography variant="body2" color="text.secondary">
              {secondary}
            </Typography>
          ) : null}
        </Stack>
      </CardContent>
    </Card>
  )
}

export default StatCard
