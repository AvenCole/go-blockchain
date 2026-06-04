export const appRoutePaths = {
  dashboard: '/',
  wallets: '/wallets',
  blocks: '/blocks',
  transactions: '/transactions',
  network: '/network',
  console: '/console',
} as const

export type AppRoutePath =
  (typeof appRoutePaths)[keyof typeof appRoutePaths]
