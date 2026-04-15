import { Card, CardContent, Chip, Stack, Typography } from '@mui/material'
import type { ReactNode } from 'react'

type StatCardProps = {
  title: string
  value: ReactNode
  secondary?: ReactNode
}

function StatCard({ title, value, secondary }: StatCardProps) {
  return (
    <Card
      variant="outlined"
      sx={{
        height: '100%',
        backgroundImage: 'none',
        borderRadius: 0.5,
        borderColor: 'divider',
      }}
    >
      <CardContent sx={{ p: 1.75 }}>
        <Stack spacing={1.25}>
          <Chip label={title} size="small" variant="outlined" sx={{ alignSelf: 'flex-start', fontWeight: 600, borderRadius: 0.5, height: 22 }} />
          <Typography variant="h5" sx={{ fontWeight: 700, lineHeight: 1.1 }}>{value}</Typography>
          {secondary ? (
            <Typography variant="caption" color="text.secondary">
              {secondary}
            </Typography>
          ) : null}
        </Stack>
      </CardContent>
    </Card>
  )
}

export default StatCard
