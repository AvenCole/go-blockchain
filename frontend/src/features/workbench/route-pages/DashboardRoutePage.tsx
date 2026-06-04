import { CircularProgress, Stack, Typography } from '@mui/material'
import DashboardPage from '@/pages/DashboardPage'
import { useWorkbench } from '@/features/workbench/context/useWorkbench'

function DashboardRoutePage() {
  const {
    dashboard,
    latestBlock,
    wallets,
    mempool,
    multiSigOutputs,
    nodes,
    chainInitAddress,
    setChainInitAddress,
    handleInitializeBlockchain,
    isInitializingBlockchain,
  } = useWorkbench()

  if (!dashboard) {
    return (
      <Stack
        spacing={1}
        sx={{ alignItems: 'center', justifyContent: 'center', height: '100%' }}
      >
        <CircularProgress size={24} />
        <Typography variant="body2" color="text.secondary">
          正在加载链状态...
        </Typography>
      </Stack>
    )
  }

  return (
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
  )
}

export default DashboardRoutePage
