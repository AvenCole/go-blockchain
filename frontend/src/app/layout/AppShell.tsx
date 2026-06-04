import type { PropsWithChildren } from 'react'
import { Box } from '@mui/material'
import AppHeader from '@/app/layout/AppHeader'
import AppSidebar from '@/app/layout/AppSidebar'

function AppShell({ children }: PropsWithChildren) {
  return (
    <Box
      sx={{
        height: '100dvh',
        display: 'grid',
        gridTemplateColumns: '248px minmax(0, 1fr)',
        overflow: 'hidden',
      }}
    >
      <AppSidebar />
      <Box
        sx={{
          minWidth: 0,
          display: 'grid',
          gridTemplateRows: '64px minmax(0, 1fr)',
          overflow: 'hidden',
        }}
      >
        <AppHeader />
        <Box component="main" sx={{ minWidth: 0, overflow: 'hidden', p: 2.5 }}>
          <Box sx={{ height: '100%', overflowY: 'auto', pr: 0.5 }}>
            {children}
          </Box>
        </Box>
      </Box>
    </Box>
  )
}

export default AppShell
