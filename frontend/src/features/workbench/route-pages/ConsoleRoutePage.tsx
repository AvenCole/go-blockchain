import { Box } from '@mui/material'
import ConsolePage from '@/pages/ConsolePage'
import { useWorkbench } from '@/features/workbench/context/useWorkbench'

function ConsoleRoutePage() {
  const {
    command,
    setCommand,
    history,
    wallets,
    nodes,
    multiSigOutputs,
    handleExecuteCommand,
    handleRunCommand,
  } = useWorkbench()

  return (
    <Box sx={{ height: '100%', minHeight: 0, overflow: 'hidden' }}>
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
  )
}

export default ConsoleRoutePage
