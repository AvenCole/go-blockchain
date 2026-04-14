import { useEffect, useMemo, useState } from 'react'
import {
  Alert,
  AppBar,
  Box,
  Button,
  Card,
  CardContent,
  Container,
  CssBaseline,
  IconButton,
  Paper,
  Stack,
  Tab,
  Tabs,
  ThemeProvider,
  Toolbar,
  Typography,
  createTheme,
  useMediaQuery,
} from '@mui/material'
import RefreshIcon from '@mui/icons-material/Refresh'
import WalletIcon from '@mui/icons-material/AccountBalanceWallet'
import AccountTreeIcon from '@mui/icons-material/AccountTree'
import DashboardIcon from '@mui/icons-material/Dashboard'
import BuildIcon from '@mui/icons-material/Construction'
import LightModeIcon from '@mui/icons-material/LightMode'
import DarkModeIcon from '@mui/icons-material/DarkMode'
import DashboardPage from './pages/DashboardPage'
import WalletsPage from './pages/WalletsPage'
import BlocksPage from './pages/BlocksPage'
import TransactionsPage from './pages/TransactionsPage'
import {
  createWallet,
  fetchBlocks,
  fetchDashboard,
  fetchPendingTransactions,
  fetchWallets,
  minePending,
  queueTransaction,
} from './api/backend'
import type { BlockView, DashboardData, WalletView } from './types'

function App() {
  const prefersDark = useMediaQuery('(prefers-color-scheme: dark)')
  const [mode, setMode] = useState<'light' | 'dark'>(prefersDark ? 'dark' : 'light')
  const [tab, setTab] = useState(0)
  const [dashboard, setDashboard] = useState<DashboardData | null>(null)
  const [wallets, setWallets] = useState<WalletView[]>([])
  const [blocks, setBlocks] = useState<BlockView[]>([])
  const [mempool, setMempool] = useState<string[]>([])
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')
  const [txForm, setTxForm] = useState({ from: '', to: '', amount: '20', fee: '2' })
  const [minerAddress, setMinerAddress] = useState('')

  const refresh = async () => {
    try {
      setError('')
      const dash = await fetchDashboard()
      const walletList = await fetchWallets()
      const blockList = await fetchBlocks()
      const pending = await fetchPendingTransactions()

      setDashboard(dash)
      setWallets(walletList)
      setBlocks(blockList)
      setMempool(pending)
      if (!minerAddress && walletList.length > 0) {
        setMinerAddress(walletList[0].address)
      }
      if (!txForm.from && walletList.length > 0) {
        setTxForm((prev) => ({ ...prev, from: walletList[0].address }))
      }
      if (!txForm.to && walletList.length > 1) {
        setTxForm((prev) => ({ ...prev, to: walletList[1].address }))
      }
    } catch (err) {
      setError(String(err))
    }
  }

  useEffect(() => {
    refresh()
  }, [])

  useEffect(() => {
    setMode(prefersDark ? 'dark' : 'light')
  }, [prefersDark])

  const latestBlock = useMemo(() => (blocks.length > 0 ? blocks[0] : null), [blocks])
  const theme = useMemo(
    () =>
      createTheme({
        palette: { mode },
      }),
    [mode],
  )

  const handleCreateWallet = async () => {
    try {
      setError('')
      const address = await createWallet()
      setMessage(`已创建钱包：${address}`)
      await refresh()
    } catch (err) {
      setError(String(err))
    }
  }

  const handleQueueTransaction = async () => {
    try {
      setError('')
      const txid = await queueTransaction(txForm.from, txForm.to, Number(txForm.amount), Number(txForm.fee || '0'))
      setMessage(`交易已进入 Mempool：${txid}`)
      await refresh()
    } catch (err) {
      setError(String(err))
    }
  }

  const handleMine = async () => {
    try {
      setError('')
      const hash = await minePending(minerAddress)
      setMessage(`已挖出新区块：${hash}`)
      await refresh()
    } catch (err) {
      setError(String(err))
    }
  }

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Box sx={{ minHeight: '100vh', backgroundColor: 'background.default' }}>
        <AppBar position="sticky" color="default" elevation={1}>
          <Toolbar sx={{ gap: 2 }}>
            <Typography variant="h5" sx={{ flexGrow: 1, fontWeight: 700 }}>
              go-blockchain GUI
            </Typography>
            <IconButton color="inherit" onClick={() => setMode((prev) => (prev === 'dark' ? 'light' : 'dark'))}>
              {mode === 'dark' ? <LightModeIcon /> : <DarkModeIcon />}
            </IconButton>
            <Button startIcon={<RefreshIcon />} variant="contained" onClick={refresh}>
              刷新
            </Button>
          </Toolbar>
        </AppBar>

        <Container maxWidth="lg" sx={{ py: 3 }}>
          <Stack spacing={2}>
            {message ? <Alert severity="success">{message}</Alert> : null}
            {error ? <Alert severity="error">{error}</Alert> : null}

            <Paper variant="outlined" sx={{ p: 2.5 }}>
              <Stack spacing={1}>
                <Typography variant="h5" sx={{ fontWeight: 700 }}>
                  区块链仿真系统桌面演示层
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  当前 GUI 直接连接真实 Go 后端，可展示区块、钱包、交易池、挖矿与链状态。
                </Typography>
              </Stack>
            </Paper>

            <Paper variant="outlined" sx={{ px: 1 }}>
              <Tabs value={tab} onChange={(_, value) => setTab(value)} variant="scrollable">
                <Tab icon={<DashboardIcon />} iconPosition="start" label="Dashboard" />
                <Tab icon={<WalletIcon />} iconPosition="start" label="钱包" />
                <Tab icon={<AccountTreeIcon />} iconPosition="start" label="区块" />
                <Tab icon={<BuildIcon />} iconPosition="start" label="交易与挖矿" />
              </Tabs>
            </Paper>

            <Card variant="outlined" sx={{ display: tab === 0 ? 'block' : 'none' }}>
              <CardContent>
                <DashboardPage dashboard={dashboard} latestBlock={latestBlock} />
              </CardContent>
            </Card>

            <Card variant="outlined" sx={{ display: tab === 1 ? 'block' : 'none' }}>
              <CardContent>
                <WalletsPage wallets={wallets} onCreateWallet={handleCreateWallet} />
              </CardContent>
            </Card>

            <Card variant="outlined" sx={{ display: tab === 2 ? 'block' : 'none' }}>
              <CardContent>
                <BlocksPage blocks={blocks} />
              </CardContent>
            </Card>

            <Card variant="outlined" sx={{ display: tab === 3 ? 'block' : 'none' }}>
              <CardContent>
                <TransactionsPage
                  txForm={txForm}
                  setTxForm={setTxForm}
                  minerAddress={minerAddress}
                  setMinerAddress={setMinerAddress}
                  mempool={mempool}
                  onQueueTransaction={handleQueueTransaction}
                  onMine={handleMine}
                />
              </CardContent>
            </Card>
          </Stack>
        </Container>
      </Box>
    </ThemeProvider>
  )
}

export default App
