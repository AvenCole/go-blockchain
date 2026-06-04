import NetworkPage from '@/pages/NetworkPage'
import { useWorkbench } from '@/features/workbench/context/useWorkbench'

function NetworkRoutePage() {
  const {
    nodes,
    wallets,
    networkDemo,
    networkReorgDemo,
    networkPartitionDemo,
    dashboard,
    nodeForm,
    setNodeForm,
    connectForm,
    setConnectForm,
    nodeControlForm,
    setNodeControlForm,
    handleStartNode,
    handleStopNode,
    handleConnectNode,
    handleInitializeNodeBlockchain,
    handleSubmitNodeTransaction,
    handleMineNode,
    handleRunNetworkQuickDemo,
    handleRunNetworkReorgDemo,
    handleRunNetworkPartitionDemo,
    networkOperation,
    isDemoBusy,
    busyActions,
    isNodeActionBusy,
  } = useWorkbench()

  return (
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
      onInitializeNodeBlockchain={handleInitializeNodeBlockchain}
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
  )
}

export default NetworkRoutePage
