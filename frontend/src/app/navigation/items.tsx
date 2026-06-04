import type { ReactElement } from 'react'
import AccountTreeIcon from '@mui/icons-material/AccountTree'
import BuildIcon from '@mui/icons-material/Construction'
import DashboardIcon from '@mui/icons-material/Dashboard'
import HubIcon from '@mui/icons-material/Hub'
import TerminalIcon from '@mui/icons-material/Terminal'
import WalletIcon from '@mui/icons-material/AccountBalanceWallet'
import {
  appRoutePaths,
  type AppRoutePath,
} from '@/app/navigation/paths'

type AppNavItem = {
  label: string
  to: AppRoutePath
  icon: ReactElement
}

export const appNavItems: AppNavItem[] = [
  {
    label: 'Dashboard',
    to: appRoutePaths.dashboard,
    icon: <DashboardIcon fontSize="small" />,
  },
  {
    label: '钱包',
    to: appRoutePaths.wallets,
    icon: <WalletIcon fontSize="small" />,
  },
  {
    label: '区块',
    to: appRoutePaths.blocks,
    icon: <AccountTreeIcon fontSize="small" />,
  },
  {
    label: '交易与挖矿',
    to: appRoutePaths.transactions,
    icon: <BuildIcon fontSize="small" />,
  },
  {
    label: '网络',
    to: appRoutePaths.network,
    icon: <HubIcon fontSize="small" />,
  },
  {
    label: '控制台',
    to: appRoutePaths.console,
    icon: <TerminalIcon fontSize="small" />,
  },
]
