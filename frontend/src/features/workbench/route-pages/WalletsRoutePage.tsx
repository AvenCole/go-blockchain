import WalletsPage from '@/pages/WalletsPage'
import { useWorkbench } from '@/features/workbench/context/useWorkbench'

function WalletsRoutePage() {
  const { wallets, handleCreateWallet } = useWorkbench()

  return (
    <WalletsPage wallets={wallets} onCreateWallet={handleCreateWallet} />
  )
}

export default WalletsRoutePage
