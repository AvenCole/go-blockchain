import { Box, Typography } from '@mui/material'
import { Outlet, createRootRoute } from '@tanstack/react-router'
import AppShell from '@/app/layout/AppShell'

function RootRouteComponent() {
  return (
    <AppShell>
      <Outlet />
    </AppShell>
  )
}

function NotFoundPage() {
  return (
    <Box sx={{ p: 2 }}>
      <Typography variant="h6">页面不存在</Typography>
      <Typography color="text.secondary" sx={{ mt: 1 }}>
        请从左侧导航重新进入功能页。
      </Typography>
    </Box>
  )
}

export const Route = createRootRoute({
  component: RootRouteComponent,
  notFoundComponent: NotFoundPage,
})
