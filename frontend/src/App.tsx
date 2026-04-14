import { useEffect, useMemo, useState } from 'react'
import type { ReactElement } from 'react'
import {
  Alert,
  AppBar,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
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
import TerminalIcon from '@mui/icons-material/Terminal'
import HubIcon from '@mui/icons-material/Hub'
import LightModeIcon from '@mui/icons-material/LightMode'
import DarkModeIcon from '@mui/icons-material/DarkMode'
import DashboardPage from './pages/DashboardPage'
import WalletsPage from './pages/WalletsPage'
import BlocksPage from './pages/BlocksPage'
import TransactionsPage from './pages/TransactionsPage'
import ConsolePage from './pages/ConsolePage'
import NetworkPage from './pages/NetworkPage'
import {
  connectNode,
  createWallet,
  executeCLI,
  fetchBlocks,
  fetchDashboard,
  fetchNodes,
  fetchPendingTransactions,
  fetchWallets,
  minePending,
  queueTransaction,
  startNode,
  stopNode,
} from './api/backend'
import type { BlockView, CommandResult, DashboardData, NodeStatus, WalletView } from './types'

type NavItem = {
  label: string
  icon: ReactElement
}

const navItems: NavItem[] = [
  { label: 'Dashboard', icon: <DashboardIcon /> },
  { label: '钱包', icon: <WalletIcon /> },
  { label: '区块', icon: <AccountTreeIcon /> },
  { label: '交易与挖矿', icon: <BuildIcon /> },
  { label: '网络', icon: <HubIcon /> },
  { label: '控制台', icon: <TerminalIcon /> },
]

function App() {
  const prefersDark = useMediaQuery('(prefers-color-scheme: dark)')
  const prefersWideNav = useMediaQuery('(min-width:1200px)')
  const [mode, setMode] = useState<'light' | 'dark'>(prefersDark ? 'dark' : 'light')
  const [tab, setTab] = useState(0)
  const [dashboard, setDashboard] = useState<DashboardData | null>(null)
  const [wallets, setWallets] = useState<WalletView[]>([])
  const [blocks, setBlocks] = useState<BlockView[]>([])
  const [mempool, setMempool] = useState<string[]>([])
  const [nodes, setNodes] = useState<NodeStatus[]>([])
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')
  const [txForm, setTxForm] = useState({ from: '', to: '', amount: '20', fee: '2' })
  const [minerAddress, setMinerAddress] = useState('')
  const [command, setCommand] = useState('')
  const [history, setHistory] = useState<CommandResult[]>([])
  const [nodeForm, setNodeForm] = useState({ address: '127.0.0.1:3010', seed: '', miner: '' })
  const [connectForm, setConnectForm] = useState({ address: '', seed: '' })

  const refresh = async () => {
    try {
      setError('')
      const [dash, walletList, blockList, pending, nodeList] = await Promise.all([
        fetchDashboard(),
        fetchWallets(),
        fetchBlocks(),
        fetchPendingTransactions(),
        fetchNodes(),
      ])

      setDashboard(dash)
      setWallets(walletList)
      setBlocks(blockList)
      setMempool(pending)
      setNodes(nodeList)

      if (!minerAddress && walletList.length > 0) {
        setMinerAddress(walletList[0].address)
      }
      if (!txForm.from && walletList.length > 0) {
        setTxForm((prev) => ({ ...prev, from: walletList[0].address }))
      }
      if (!txForm.to && walletList.length > 1) {
        setTxForm((prev) => ({ ...prev, to: walletList[1].address }))
      }
      if (!nodeForm.miner && walletList.length > 0) {
        setNodeForm((prev) => ({ ...prev, miner: walletList[0].address }))
      }
      if (!connectForm.address && nodeList.length > 0) {
        setConnectForm((prev) => ({ ...prev, address: nodeList[0].address }))
      }
    } catch (err) {
      setError(String(err))
    }
  }

  useEffect(() => {
    void refresh()
  }, [])

  useEffect(() => {
    setMode(prefersDark ? 'dark' : 'light')
  }, [prefersDark])

  const latestBlock = useMemo(() => (blocks.length > 0 ? blocks[0] : null), [blocks])
  const theme = useMemo(
    () =>
      createTheme({
        palette: {
          mode,
          ...(mode === 'dark'
            ? {
                background: { default: '#0b1020', paper: '#121a2b' },
              }
            : {
                background: { default: '#f5f7fb', paper: '#ffffff' },
              }),
        },
        shape: { borderRadius: 14 },
        components: {
          MuiCard: {
            styleOverrides: {
              root: {
                borderRadius: 18,
              },
            },
          },
          MuiPaper: {
            styleOverrides: {
              root: {
                borderRadius: 18,
              },
            },
          },
        },
      }),
    [mode],
  )

  const navOrientation = prefersWideNav ? 'vertical' : 'horizontal'

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

  const handleExecuteCommand = async () => {
    if (!command.trim()) {
      return
    }

    try {
      setError('')
      const result = await executeCLI(command)
      setHistory((prev) => [result, ...prev].slice(0, 20))
      setMessage(`命令执行完成：${result.command}`)
      setCommand('')
      await refresh()
    } catch (err) {
      setError(String(err))
    }
  }

  const handleStartNode = async () => {
    try {
      setError('')
      const addr = await startNode(nodeForm.address, nodeForm.seed, nodeForm.miner)
      setConnectForm((prev) => ({ ...prev, address: addr }))
      setMessage(`节点已启动：${addr}`)
      await refresh()
    } catch (err) {
      setError(String(err))
    }
  }

  const handleStopNode = async (address: string) => {
    try {
      setError('')
      await stopNode(address)
      setMessage(`节点已停止：${address}`)
      await refresh()
    } catch (err) {
      setError(String(err))
    }
  }

  const handleConnectNode = async () => {
    try {
      setError('')
      await connectNode(connectForm.address, connectForm.seed)
      setMessage(`节点已连接：${connectForm.address} -> ${connectForm.seed}`)
      await refresh()
    } catch (err) {
      setError(String(err))
    }
  }

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Box sx={{ minHeight: '100dvh', backgroundColor: 'background.default' }}>
        <AppBar position="sticky" color="inherit" elevation={0} sx={{ borderBottom: 1, borderColor: 'divider' }}>
          <Toolbar sx={{ gap: 2, minHeight: { xs: 64, md: 72 } }}>
            <Stack spacing={0.25} sx={{ flexGrow: 1 }}>
              <Typography variant="h5" sx={{ fontWeight: 700 }}>
                go-blockchain GUI
              </Typography>
              <Typography variant="body2" color="text.secondary">
                采用 Bitcoin 客户端风格的链状态、钱包、网络与控制台一体化演示界面
              </Typography>
            </Stack>
            <IconButton color="inherit" onClick={() => setMode((prev) => (prev === 'dark' ? 'light' : 'dark'))}>
              {mode === 'dark' ? <LightModeIcon /> : <DarkModeIcon />}
            </IconButton>
            <Button startIcon={<RefreshIcon />} variant="contained" onClick={() => void refresh()}>
              刷新
            </Button>
          </Toolbar>
        </AppBar>

        <Container maxWidth={false} sx={{ py: 3, px: { xs: 2, md: 3 } }}>
          <Stack spacing={2.5}>
            {message ? <Alert severity="success">{message}</Alert> : null}
            {error ? <Alert severity="error">{error}</Alert> : null}

            <Paper
              variant="outlined"
              sx={{
                p: { xs: 2, md: 2.5 },
                backgroundImage:
                  mode === 'dark'
                    ? 'linear-gradient(135deg, rgba(59,130,246,0.12), rgba(16,185,129,0.08))'
                    : 'linear-gradient(135deg, rgba(59,130,246,0.10), rgba(16,185,129,0.10))',
              }}
            >
              <Stack direction={{ xs: 'column', lg: 'row' }} spacing={2} justifyContent="space-between" alignItems={{ xs: 'flex-start', lg: 'center' }}>
                <Stack spacing={1}>
                  <Typography variant="h5" sx={{ fontWeight: 700 }}>
                    区块链仿真系统桌面演示层
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    当前 GUI 直接调用真实 Go 后端，适合课堂演示钱包生成、交易流转、出块、节点联通和命令行能力。
                  </Typography>
                </Stack>
                <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} flexWrap="wrap" useFlexGap>
                  <Chip label={`高度 ${dashboard?.height ?? '-'}`} color="primary" variant="outlined" />
                  <Chip label={`钱包 ${wallets.length}`} color="secondary" variant="outlined" />
                  <Chip label={`Mempool ${mempool.length}`} variant="outlined" />
                  <Chip label={`节点 ${nodes.length}`} variant="outlined" />
                </Stack>
              </Stack>
            </Paper>

            <Box sx={{ display: 'grid', gap: 2.5, gridTemplateColumns: { xs: '1fr', xl: '260px minmax(0, 1fr)' } }}>
              <Paper variant="outlined" sx={{ p: 1.25, height: 'fit-content', overflow: 'hidden' }}>
                <Stack spacing={1.5}>
                  <Box sx={{ px: 1, pt: 0.5 }}>
                    <Typography variant="subtitle2" color="text.secondary">
                      工作台导航
                    </Typography>
                    <Typography variant="body2" sx={{ mt: 0.5 }}>
                      模拟 Bitcoin 客户端左侧功能分区
                    </Typography>
                  </Box>
                  <Tabs
                    orientation={navOrientation}
                    value={tab}
                    onChange={(_, value) => setTab(value)}
                    variant="scrollable"
                    sx={{
                      '& .MuiTabs-flexContainer': {
                        alignItems: prefersWideNav ? 'stretch' : undefined,
                      },
                      '& .MuiTab-root': {
                        alignItems: 'flex-start',
                        justifyContent: 'flex-start',
                        textAlign: 'left',
                        minHeight: 48,
                        borderRadius: 2,
                      },
                    }}
                  >
                    {navItems.map((item) => (
                      <Tab key={item.label} icon={item.icon} iconPosition="start" label={item.label} />
                    ))}
                  </Tabs>
                  <Paper variant="outlined" sx={{ p: 1.5, mx: 0.5 }}>
                    <Typography variant="caption" color="text.secondary">
                      数据目录
                    </Typography>
                    <Typography variant="body2" sx={{ mt: 0.5, wordBreak: 'break-all' }}>
                      {dashboard?.dataDir ?? '加载中'}
                    </Typography>
                  </Paper>
                </Stack>
              </Paper>

              <Stack spacing={2.5} sx={{ minWidth: 0 }}>
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

                <Card variant="outlined" sx={{ display: tab === 4 ? 'block' : 'none' }}>
                  <CardContent>
                    <NetworkPage
                      nodes={nodes}
                      nodeForm={nodeForm}
                      setNodeForm={setNodeForm}
                      connectForm={connectForm}
                      setConnectForm={setConnectForm}
                      onStartNode={handleStartNode}
                      onStopNode={handleStopNode}
                      onConnectNode={handleConnectNode}
                    />
                  </CardContent>
                </Card>

                <Card variant="outlined" sx={{ display: tab === 5 ? 'block' : 'none' }}>
                  <CardContent>
                    <ConsolePage
                      command={command}
                      setCommand={setCommand}
                      history={history}
                      onExecute={handleExecuteCommand}
                    />
                  </CardContent>
                </Card>
              </Stack>
            </Box>

            <Paper variant="outlined" sx={{ p: 1.5 }}>
              <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.5} justifyContent="space-between">
                <Typography variant="body2" color="text.secondary">
                  最新区块：{latestBlock?.hash ?? '尚未初始化'}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  节点数：{nodes.length} | 待打包交易：{mempool.length} | 网络模式：{dashboard?.networkMode ?? '-'}
                </Typography>
              </Stack>
            </Paper>
          </Stack>
        </Container>
      </Box>
    </ThemeProvider>
  )
}

export default App
