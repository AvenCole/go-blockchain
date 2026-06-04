import RefreshIcon from '@mui/icons-material/Refresh'
import Brightness4Icon from '@mui/icons-material/Brightness4'
import Brightness7Icon from '@mui/icons-material/Brightness7'
import {
  AppBar,
  Box,
  Button,
  Stack,
  Toolbar,
  Typography,
} from '@mui/material'
import { useThemeMode } from '@/app/providers/ThemeModeProvider'
import { useWorkbench } from '@/features/workbench/context/useWorkbench'

function AppHeader() {
  const { mode, toggleColorMode } = useThemeMode()
  const { refresh } = useWorkbench()

  return (
    <AppBar
      position="static"
      color="inherit"
      elevation={0}
      sx={{ borderBottom: 1, borderColor: 'divider' }}
    >
      <Toolbar sx={{ minHeight: 64, gap: 2 }}>
        <Stack spacing={0.25} sx={{ flexGrow: 1, minWidth: 0 }}>
          <Typography variant="h6" sx={{ fontWeight: 700 }}>
            go-blockchain
          </Typography>
          <Typography variant="caption" color="text.secondary">
            desktop client
          </Typography>
        </Stack>
        <Box sx={{ display: 'flex', gap: 1 }}>
          <Button
            color="inherit"
            onClick={toggleColorMode}
            startIcon={
              mode === 'dark' ? <Brightness7Icon /> : <Brightness4Icon />
            }
            variant="text"
          >
            {mode === 'dark' ? '浅色' : '深色'}
          </Button>
          <Button
            onClick={() => void refresh()}
            startIcon={<RefreshIcon />}
            variant="outlined"
          >
            刷新
          </Button>
        </Box>
      </Toolbar>
    </AppBar>
  )
}

export default AppHeader
