import { useEffect, useMemo, useState } from 'react';
import type { ReactElement } from 'react';
import {
  Alert,
  AppBar,
  Box,
  Button,
  Container,
  CssBaseline,
  Divider,
  Stack,
  Tab,
  Tabs,
  Toolbar,
  Typography,
  IconButton,
  createTheme,
  ThemeProvider,
  PaletteMode,
} from '@mui/material';
import RefreshIcon from '@mui/icons-material/Refresh';
import WalletIcon from '@mui/icons-material/AccountBalanceWallet';
import AccountTreeIcon from '@mui/icons-material/AccountTree';
import DashboardIcon from '@mui/icons-material/Dashboard';
import BuildIcon from '@mui/icons-material/Construction';
import TerminalIcon from '@mui/icons-material/Terminal';
import HubIcon from '@mui/icons-material/Hub';
import Brightness4Icon from '@mui/icons-material/Brightness4';
import Brightness7Icon from '@mui/icons-material/Brightness7';
import { EventsOn } from '../wailsjs/runtime/runtime';
import DashboardPage from './pages/DashboardPage';
import WalletsPage from './pages/WalletsPage';
import BlocksPage from './pages/BlocksPage';
import TransactionsPage from './pages/TransactionsPage';
import ConsolePage from './pages/ConsolePage';
import NetworkPage from './pages/NetworkPage';
import {
  connectNode,
  createWallet,
  executeCLI,
  fetchBlocks,
  fetchDashboard,
  fetchMultiSigOutputs,
  fetchNodes,
  fetchPendingTransactions,
  fetchWallets,
  initializeNodeBlockchain,
  minePending,
  mineNodePending,
  queueMultiSigTransaction,
  queueP2PKTransaction,
  queueSpendMultiSigTransaction,
  queueTransaction,
  runNetworkPartitionDemo,
  runNetworkQuickDemo,
  runNetworkReorgDemo,
  startNode,
  stopNode,
  submitNodeTransaction,
} from './api/backend';
import type {
  BlockView,
  CommandResult,
  DashboardData,
  MultiSigOutputView,
  NetworkDemoResult,
  NetworkPartitionDemoResult,
  NetworkOperationProgress,
  NetworkReorgDemoResult,
  NodeStatus,
  WalletView,
} from './types';

type NavItem = {
  label: string;
  icon: ReactElement;
};

const navItems: NavItem[] = [
  { label: 'Dashboard', icon: <DashboardIcon /> },
  { label: '钱包', icon: <WalletIcon /> },
  { label: '区块', icon: <AccountTreeIcon /> },
  { label: '交易与挖矿', icon: <BuildIcon /> },
  { label: '网络', icon: <HubIcon /> },
  { label: '控制台', icon: <TerminalIcon /> },
];

function App() {
  type BusyActionKey =
    | 'startNode'
    | 'stopNode'
    | 'connectNode'
    | 'initializeNodeBlockchain'
    | 'submitNodeTransaction'
    | 'mineNode'
    | 'runNetworkQuickDemo'
    | 'runNetworkReorgDemo'
    | 'runNetworkPartitionDemo';

  const [mode, setMode] = useState<PaletteMode>(
    (localStorage.getItem('themeMode') as PaletteMode) || 'light',
  );

  const theme = useMemo(
    () =>
      createTheme({
        palette: {
          mode,
        },
      }),
    [mode],
  );

  const toggleColorMode = () => {
    const newMode = mode === 'light' ? 'dark' : 'light';
    setMode(newMode);
    localStorage.setItem('themeMode', newMode);
  };

  const [tab, setTab] = useState(0);
  const [dashboard, setDashboard] = useState<DashboardData | null>(null);
  const [wallets, setWallets] = useState<WalletView[]>([]);
  const [blocks, setBlocks] = useState<BlockView[]>([]);
  const [multiSigOutputs, setMultiSigOutputs] = useState<MultiSigOutputView[]>(
    [],
  );
  const [mempool, setMempool] = useState<string[]>([]);
  const [nodes, setNodes] = useState<NodeStatus[]>([]);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');
  const [networkDemo, setNetworkDemo] = useState<NetworkDemoResult | null>(
    null,
  );
  const [networkReorgDemo, setNetworkReorgDemo] =
    useState<NetworkReorgDemoResult | null>(null);
  const [networkPartitionDemo, setNetworkPartitionDemo] =
    useState<NetworkPartitionDemoResult | null>(null);
  const [networkOperation, setNetworkOperation] =
    useState<NetworkOperationProgress | null>(null);
  const [busyActions, setBusyActions] = useState<Record<BusyActionKey, boolean>>({
    startNode: false,
    stopNode: false,
    connectNode: false,
    initializeNodeBlockchain: false,
    submitNodeTransaction: false,
    mineNode: false,
    runNetworkQuickDemo: false,
    runNetworkReorgDemo: false,
    runNetworkPartitionDemo: false,
  });
  const [txForm, setTxForm] = useState({
    template: 'p2pkh' as 'p2pkh' | 'p2pk' | 'multisig',
    from: '',
    to: '',
    recipients: '',
    required: '2',
    amount: '20',
    fee: '2',
  });
  const [spendMultiSigForm, setSpendMultiSigForm] = useState({
    signers: '',
    sourceTxID: '',
    out: '0',
    to: '',
    amount: '10',
    fee: '1',
  });
  const [minerAddress, setMinerAddress] = useState('');
  const [command, setCommand] = useState('');
  const [history, setHistory] = useState<CommandResult[]>([]);
  const [chainInitAddress, setChainInitAddress] = useState('');
  const [isInitializingBlockchain, setIsInitializingBlockchain] = useState(false);
  const [nodeForm, setNodeForm] = useState({
    address: '127.0.0.1:3010',
    seed: '',
    miner: '',
  });
  const [connectForm, setConnectForm] = useState({ address: '', seed: '' });
  const [nodeControlForm, setNodeControlForm] = useState({
    address: '',
    rewardAddress: '',
    from: '',
    to: '',
    amount: '10',
    fee: '1',
  });

  const refresh = async () => {
    try {
      setError('');
      const [dash, walletList, blockList, pending, nodeList, multiSigList] =
        await Promise.all([
          fetchDashboard(),
          fetchWallets(),
          fetchBlocks(),
          fetchPendingTransactions(),
          fetchNodes(),
          fetchMultiSigOutputs(),
        ]);

      setDashboard(dash);
      setWallets(walletList);
      setBlocks(blockList);
      setMultiSigOutputs(multiSigList);
      setMempool(pending);
      setNodes(nodeList);

      if (!minerAddress && walletList.length > 0) {
        setMinerAddress(walletList[0].address);
      }
      if (!chainInitAddress && walletList.length > 0) {
        setChainInitAddress(walletList[0].address);
      }
      if (!txForm.from && walletList.length > 0) {
        setTxForm((prev) => ({ ...prev, from: walletList[0].address }));
      }
      if (!txForm.to && walletList.length > 1) {
        setTxForm((prev) => ({ ...prev, to: walletList[1].address }));
      }
      if (!txForm.recipients && walletList.length > 1) {
        setTxForm((prev) => ({
          ...prev,
          recipients: `${walletList[0].address},${walletList[1].address}`,
        }));
      }
      if (!nodeForm.miner && walletList.length > 0) {
        setNodeForm((prev) => ({ ...prev, miner: walletList[0].address }));
      }
      if (!connectForm.address && nodeList.length > 0) {
        setConnectForm((prev) => ({ ...prev, address: nodeList[0].address }));
      }
      if (
        (!nodeControlForm.address ||
          !nodeList.some((node) => node.address === nodeControlForm.address)) &&
        nodeList.length > 0
      ) {
        setNodeControlForm((prev) => ({
          ...prev,
          address: nodeList[0].address,
        }));
      }
      if (!nodeControlForm.rewardAddress && walletList.length > 0) {
        setNodeControlForm((prev) => ({
          ...prev,
          rewardAddress: walletList[0].address,
        }));
      }
      if (!nodeControlForm.from && walletList.length > 0) {
        setNodeControlForm((prev) => ({
          ...prev,
          from: walletList[0].address,
        }));
      }
      if (!nodeControlForm.to && walletList.length > 1) {
        setNodeControlForm((prev) => ({ ...prev, to: walletList[1].address }));
      }
      if (!spendMultiSigForm.to && walletList.length > 0) {
        setSpendMultiSigForm((prev) => ({
          ...prev,
          to: walletList[0].address,
        }));
      }
      if (!spendMultiSigForm.sourceTxID && multiSigList.length > 0) {
        setSpendMultiSigForm((prev) => ({
          ...prev,
          sourceTxID: multiSigList[0].txid,
          out: String(multiSigList[0].out),
          signers: multiSigList[0].participants.join(','),
        }));
      }
    } catch (err) {
      setError(String(err));
    }
  };

  useEffect(() => {
    void refresh();
  }, []);

  useEffect(() => {
    const unsubscribe = EventsOn(
      'network:operation',
      (payload: NetworkOperationProgress) => {
        setNetworkOperation(payload);
      },
    );

    return () => {
      unsubscribe();
    };
  }, []);

  const latestBlock = useMemo(
    () => (blocks.length > 0 ? blocks[0] : null),
    [blocks],
  );

  const isNodeActionBusy =
    busyActions.startNode ||
    busyActions.stopNode ||
    busyActions.connectNode ||
    busyActions.initializeNodeBlockchain ||
    busyActions.submitNodeTransaction ||
    busyActions.mineNode;

  const isDemoBusy =
    busyActions.runNetworkQuickDemo ||
    busyActions.runNetworkReorgDemo ||
    busyActions.runNetworkPartitionDemo ||
    Boolean(
      networkOperation &&
      (networkOperation.status === 'started' ||
        networkOperation.status === 'progress'),
    );

  const runBusyAction = async (
    key: BusyActionKey,
    action: () => Promise<void>,
  ) => {
    setBusyActions((prev) => ({ ...prev, [key]: true }));
    try {
      await action();
    } finally {
      setBusyActions((prev) => ({ ...prev, [key]: false }));
    }
  };

  const handleCreateWallet = async () => {
    try {
      setError('');
      const address = await createWallet();
      setMessage(`已创建钱包：${address}`);
      if (!chainInitAddress) {
        setChainInitAddress(address);
      }
      await refresh();
    } catch (err) {
      setError(String(err));
    }
  };

  const handleInitializeBlockchain = async () => {
    if (!chainInitAddress) {
      return;
    }

    try {
      setIsInitializingBlockchain(true);
      setError('');
      const result = await executeCLI(`createblockchain ${chainInitAddress}`);
      setHistory((prev) => [result, ...prev].slice(0, 20));

      if (result.exitCode !== 0) {
        setError(result.stderr || result.stdout || '初始化主链失败');
        return;
      }

      setMessage(`主链已初始化：${chainInitAddress}`);
      await refresh();
    } catch (err) {
      setError(String(err));
    } finally {
      setIsInitializingBlockchain(false);
    }
  };

  const handleQueueTransaction = async () => {
    try {
      setError('');
      let txid = '';
      if (txForm.template === 'p2pk') {
        txid = await queueP2PKTransaction(
          txForm.from,
          txForm.to,
          Number(txForm.amount),
          Number(txForm.fee || '0'),
        );
      } else if (txForm.template === 'multisig') {
        txid = await queueMultiSigTransaction(
          txForm.from,
          txForm.recipients,
          Number(txForm.required || '0'),
          Number(txForm.amount),
          Number(txForm.fee || '0'),
        );
      } else {
        txid = await queueTransaction(
          txForm.from,
          txForm.to,
          Number(txForm.amount),
          Number(txForm.fee || '0'),
        );
      }
      setMessage(`交易已进入 Mempool：${txid}`);
      await refresh();
    } catch (err) {
      setError(String(err));
    }
  };

  const handleMine = async () => {
    try {
      setError('');
      const hash = await minePending(minerAddress);
      setMessage(`已挖出新区块：${hash}`);
      await refresh();
    } catch (err) {
      setError(String(err));
    }
  };

  const handleExecuteCommand = async () => {
    await handleRunCommand(command, true);
  };

  const handleRunCommand = async (
    commandLine: string,
    clearInput = false,
  ) => {
    if (!commandLine.trim()) {
      return;
    }

    try {
      setError('');
      if (!clearInput) {
        setCommand(commandLine);
      }
      const result = await executeCLI(commandLine);
      setHistory((prev) => [result, ...prev].slice(0, 20));
      setMessage(`命令执行完成：${result.command}`);
      if (clearInput) {
        setCommand('');
      }
      await refresh();
    } catch (err) {
      setError(String(err));
    }
  };

  const handleStartNode = async () => {
    await runBusyAction('startNode', async () => {
      try {
        setError('');
        const addr = await startNode(
          nodeForm.address,
          nodeForm.seed,
          nodeForm.miner,
        );
        setConnectForm((prev) => ({ ...prev, address: addr }));
        setMessage(`节点已启动：${addr}`);
        await refresh();
      } catch (err) {
        setError(String(err));
      }
    });
  };

  const handleStopNode = async (address: string) => {
    await runBusyAction('stopNode', async () => {
      try {
        setError('');
        await stopNode(address);
        setMessage(`节点已停止：${address}`);
        await refresh();
      } catch (err) {
        setError(String(err));
      }
    });
  };

  const handleConnectNode = async () => {
    await runBusyAction('connectNode', async () => {
      try {
        setError('');
        await connectNode(connectForm.address, connectForm.seed);
        setMessage(`节点已连接：${connectForm.address} -> ${connectForm.seed}`);
        await refresh();
      } catch (err) {
        setError(String(err));
      }
    });
  };

  const handleInitializeNodeBlockchain = async () => {
    await runBusyAction('initializeNodeBlockchain', async () => {
      try {
        setError('');
        await initializeNodeBlockchain(
          nodeControlForm.address,
          nodeControlForm.rewardAddress,
        );
        setMessage(`节点区块链已就绪：${nodeControlForm.address}`);
        await refresh();
      } catch (err) {
        setError(String(err));
      }
    });
  };

  const handleSubmitNodeTransaction = async () => {
    await runBusyAction('submitNodeTransaction', async () => {
      try {
        setError('');
        const txid = await submitNodeTransaction(
          nodeControlForm.address,
          nodeControlForm.from,
          nodeControlForm.to,
          Number(nodeControlForm.amount),
          Number(nodeControlForm.fee || '0'),
        );
        setMessage(`节点交易已进入 Mempool：${txid}`);
        await refresh();
      } catch (err) {
        setError(String(err));
      }
    });
  };

  const handleMineNode = async () => {
    await runBusyAction('mineNode', async () => {
      try {
        setError('');
        const hash = await mineNodePending(nodeControlForm.address);
        setMessage(`节点已挖出新区块：${hash}`);
        await refresh();
      } catch (err) {
        setError(String(err));
      }
    });
  };

  const handleRunNetworkQuickDemo = async () => {
    await runBusyAction('runNetworkQuickDemo', async () => {
      try {
        setError('');
        const result = await runNetworkQuickDemo();
        setNetworkDemo(result);
        setMessage(`快速同步已完成：${result.sourceNode} -> ${result.peerNode}`);
        await refresh();
      } catch (err) {
        setError(String(err));
      }
    });
  };

  const handleRunNetworkReorgDemo = async () => {
    await runBusyAction('runNetworkReorgDemo', async () => {
      try {
        setError('');
        const result = await runNetworkReorgDemo();
        setNetworkReorgDemo(result);
        setMessage(
          `重组流程已完成：${result.sourceNode} -> ${result.peerNode}`,
        );
        await refresh();
      } catch (err) {
        setError(String(err));
      }
    });
  };

  const handleRunNetworkPartitionDemo = async () => {
    await runBusyAction('runNetworkPartitionDemo', async () => {
      try {
        setError('');
        const result = await runNetworkPartitionDemo();
        setNetworkPartitionDemo(result);
        setMessage(
          `三节点分区流程已完成：${result.sourceNode} / ${result.peerNode} / ${result.forkNode}`,
        );
        await refresh();
      } catch (err) {
        setError(String(err));
      }
    });
  };

  const handleSpendMultiSig = async () => {
    try {
      setError('');
      const txid = await queueSpendMultiSigTransaction(
        spendMultiSigForm.signers,
        spendMultiSigForm.sourceTxID,
        Number(spendMultiSigForm.out),
        spendMultiSigForm.to,
        Number(spendMultiSigForm.amount),
        Number(spendMultiSigForm.fee || '0'),
      );
      setMessage(`多签花费交易已进入 Mempool：${txid}`);
      await refresh();
    } catch (err) {
      setError(String(err));
    }
  };

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Box sx={{ height: '100dvh', display: 'flex', flexDirection: 'column' }}>
        <AppBar
          position="static"
          color="inherit"
          elevation={0}
          sx={{ borderBottom: 1, borderColor: 'divider' }}
        >
          <Toolbar sx={{ gap: 2, minHeight: 64 }}>
            <Stack spacing={0.25} sx={{ flexGrow: 1 }}>
              <Typography variant="h6" sx={{ fontWeight: 800, letterSpacing: -0.5 }}>
                go-blockchain GUI
              </Typography>
            </Stack>
            <Button
              onClick={toggleColorMode}
              color="inherit"
              startIcon={mode === 'dark' ? <Brightness7Icon /> : <Brightness4Icon />}
              variant="text"
            >
              {mode === 'dark' ? '浅色' : '深色'}
            </Button>
            <Button startIcon={<RefreshIcon />} variant="outlined" onClick={() => void refresh()}>
              刷新
            </Button>
          </Toolbar>
        </AppBar>

        <Box
          sx={{
            flex: 1,
            display: 'flex',
            overflow: 'hidden',
          }}
        >
          {/* 左侧导航栏 */}
          <Box
            component="nav"
            sx={{
              width: 260,
              flexShrink: 0,
              borderRight: 1,
              borderColor: 'divider',
              bgcolor: 'background.paper',
              display: 'flex',
              flexDirection: 'column',
            }}
          >
            <Box sx={{ p: 2, pb: 1 }}>
              <Typography
                variant="overline"
                color="text.secondary"
                sx={{ ml: 1, fontWeight: 700, letterSpacing: 1 }}
              >
                主菜单
              </Typography>
            </Box>
            <Box sx={{ flex: 1, overflowY: 'auto' }}>
              <Tabs
                orientation="vertical"
                value={tab}
                onChange={(_, value) => setTab(value)}
                variant="scrollable"
                sx={{
                  minHeight: 0,
                  '& .MuiTabs-indicator': {
                    display: 'none',
                  },
                  '& .MuiTabs-flexContainer': {
                    alignItems: 'stretch',
                    px: 1,
                  },
                  '& .MuiTab-root': {
                    alignItems: 'center',
                    justifyContent: 'flex-start',
                    textAlign: 'left',
                    minHeight: 44,
                    px: 2,
                    py: 1.5,
                    borderRadius: 1,
                    textTransform: 'none',
                    fontWeight: 500,
                  },
                  '& .MuiTab-root.Mui-selected': {
                    fontWeight: 700,
                    bgcolor: 'action.selected',
                  },
                }}
              >
                {navItems.map((item) => (
                  <Tab
                    key={item.label}
                    icon={item.icon}
                    iconPosition="start"
                    label={item.label}
                    sx={{ gap: 1.5 }}
                  />
                ))}
              </Tabs>
            </Box>
            <Box sx={{ p: 2, borderTop: 1, borderColor: 'divider' }}>
              {message || error ? (
                <Alert
                  severity={error ? 'error' : 'success'}
                  sx={{ borderRadius: 2, '& .MuiAlert-message': { wordBreak: 'break-all' } }}
                  onClose={() => { setMessage(''); setError(''); }}
                >
                  {error || message}
                </Alert>
              ) : (
                <Typography variant="caption" color="text.secondary">
                  节点状态正常
                </Typography>
              )}
            </Box>
          </Box>

          {/* 右侧主内容区 */}
          <Box
            component="main"
            sx={{
              flex: 1,
              overflowY: 'auto',
              p: 3,
            }}
          >
            <Box sx={{ maxWidth: 1400, mx: 'auto' }}>
              <Box sx={{ display: tab === 0 ? 'block' : 'none' }}>
                <DashboardPage
                  dashboard={dashboard}
                  latestBlock={latestBlock}
                  wallets={wallets}
                  mempool={mempool}
                  multiSigOutputs={multiSigOutputs}
                  nodes={nodes}
                  chainInitAddress={chainInitAddress}
                  setChainInitAddress={setChainInitAddress}
                  onInitializeBlockchain={handleInitializeBlockchain}
                  isInitializingBlockchain={isInitializingBlockchain}
                />
              </Box>

              <Box sx={{ display: tab === 1 ? 'block' : 'none' }}>
                <WalletsPage
                  wallets={wallets}
                  onCreateWallet={handleCreateWallet}
                />
              </Box>

              <Box sx={{ display: tab === 2 ? 'block' : 'none' }}>
                <BlocksPage blocks={blocks} />
              </Box>

              <Box sx={{ display: tab === 3 ? 'block' : 'none' }}>
                <TransactionsPage
                  txForm={txForm}
                  setTxForm={setTxForm}
                  spendMultiSigForm={spendMultiSigForm}
                  setSpendMultiSigForm={setSpendMultiSigForm}
                  multiSigOutputs={multiSigOutputs}
                  minerAddress={minerAddress}
                  setMinerAddress={setMinerAddress}
                  mempool={mempool}
                  onQueueTransaction={handleQueueTransaction}
                  onSpendMultiSig={handleSpendMultiSig}
                  onMine={handleMine}
                />
              </Box>

              <Box sx={{ display: tab === 4 ? 'block' : 'none' }}>
                <NetworkPage
                  nodes={nodes}
                  wallets={wallets}
                  networkDemo={networkDemo}
                  networkReorgDemo={networkReorgDemo}
                  networkPartitionDemo={networkPartitionDemo}
                  lastReorg={dashboard?.lastReorg ?? null}
                  recentEvents={dashboard?.recentEvents ?? []}
                  nodeForm={nodeForm}
                  setNodeForm={setNodeForm}
                  connectForm={connectForm}
                  setConnectForm={setConnectForm}
                  nodeControlForm={nodeControlForm}
                  setNodeControlForm={setNodeControlForm}
                  onStartNode={handleStartNode}
                  onStopNode={handleStopNode}
                  onConnectNode={handleConnectNode}
                  onInitializeNodeBlockchain={
                    handleInitializeNodeBlockchain
                  }
                  onSubmitNodeTransaction={handleSubmitNodeTransaction}
                  onMineNode={handleMineNode}
                  onRunNetworkQuickDemo={handleRunNetworkQuickDemo}
                  onRunNetworkReorgDemo={handleRunNetworkReorgDemo}
                  onRunNetworkPartitionDemo={handleRunNetworkPartitionDemo}
                  operationProgress={networkOperation}
                  isDemoBusy={isDemoBusy}
                  busyActions={busyActions}
                  isNodeActionBusy={isNodeActionBusy}
                />
              </Box>

              <Box sx={{ display: tab === 5 ? 'block' : 'none' }}>
                <ConsolePage
                  command={command}
                  setCommand={setCommand}
                  history={history}
                  wallets={wallets}
                  nodes={nodes}
                  multiSigOutputs={multiSigOutputs}
                  onExecute={handleExecuteCommand}
                  onRunCommand={handleRunCommand}
                />
              </Box>
            </Box>
          </Box>
        </Box>
      </Box>
    </ThemeProvider>
  );
}

export default App;
