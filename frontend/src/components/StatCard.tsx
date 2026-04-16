import { Card, CardContent, Stack, Typography } from '@mui/material'
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
        borderRadius: 1,
        borderColor: 'divider',
      }}
    >
      <CardContent sx={{ p: 2 }}>
        <Stack spacing={0.9}>
          <Typography
            variant="overline"
            color="text.secondary"
            sx={{ lineHeight: 1, letterSpacing: 0.7 }}
          >
            {title}
          </Typography>
          <Typography variant="h4" sx={{ fontWeight: 700, lineHeight: 1.05 }}>
            {value}
          </Typography>
          {secondary ? (
            <Typography
              variant="body2"
              color="text.secondary"
              sx={{ wordBreak: 'break-word' }}
            >
              {secondary}
            </Typography>
          ) : null}
        </Stack>
      </CardContent>
    </Card>
  )
}

export default StatCard
