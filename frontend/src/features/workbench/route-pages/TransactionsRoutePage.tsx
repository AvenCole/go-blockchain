import TransactionsPage from '@/pages/TransactionsPage'
import { useWorkbench } from '@/features/workbench/context/useWorkbench'

function TransactionsRoutePage() {
  const {
    txForm,
    setTxForm,
    spendMultiSigForm,
    setSpendMultiSigForm,
    multiSigOutputs,
    minerAddress,
    setMinerAddress,
    mempool,
    handleQueueTransaction,
    handleSpendMultiSig,
    handleMine,
  } = useWorkbench()

  return (
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
  )
}

export default TransactionsRoutePage
