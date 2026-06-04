import {
  Alert,
  Box,
  Divider,
  List,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Typography,
} from '@mui/material'
import { useNavigate, useRouterState } from '@tanstack/react-router'
import { appNavItems } from '@/app/navigation/items'
import { useWorkbench } from '@/features/workbench/context/useWorkbench'

function AppSidebar() {
  const navigate = useNavigate()
  const pathname = useRouterState({
    select: (state) => state.location.pathname,
  })
  const { error, message, clearFeedback } = useWorkbench()

  return (
    <Box
      component="nav"
      sx={{
        width: 248,
        borderRight: 1,
        borderColor: 'divider',
        display: 'flex',
        flexDirection: 'column',
        minHeight: 0,
      }}
    >
      <Box sx={{ px: 2, py: 2.25 }}>
        <Typography variant="subtitle2" color="text.secondary">
          导航
        </Typography>
      </Box>
      <Divider />
      <List sx={{ flex: 1, overflowY: 'auto', py: 1 }}>
        {appNavItems.map((item) => (
          <ListItemButton
            key={item.to}
            selected={pathname === item.to}
            onClick={() => void navigate({ to: item.to })}
            sx={{
              mx: 1,
              mb: 0.5,
              minHeight: 44,
              borderRadius: 1,
              alignItems: 'center',
            }}
          >
            <ListItemIcon sx={{ minWidth: 36 }}>{item.icon}</ListItemIcon>
            <ListItemText
              primary={
                <Typography
                  sx={{ fontSize: 14, fontWeight: pathname === item.to ? 700 : 500 }}
                >
                  {item.label}
                </Typography>
              }
            />
          </ListItemButton>
        ))}
      </List>
      <Divider />
      <Box sx={{ p: 2 }}>
        {message || error ? (
          <Alert
            onClose={clearFeedback}
            severity={error ? 'error' : 'success'}
            sx={{ '& .MuiAlert-message': { wordBreak: 'break-all' } }}
          >
            {error || message}
          </Alert>
        ) : (
          <Typography variant="caption" color="text.secondary">
            就绪
          </Typography>
        )}
      </Box>
    </Box>
  )
}

export default AppSidebar
