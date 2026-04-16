import { Box, Button, Card, CardContent, Divider, Stack, Typography } from '@mui/material'
import type { WalletView } from '../types'

type WalletsPageProps = {
  wallets: WalletView[]
  onCreateWallet: () => Promise<void>
}

function WalletsPage({ wallets, onCreateWallet }: WalletsPageProps) {
  return (
    <Box
      sx={{
        display: 'grid',
        gap: 2,
        gridTemplateColumns: { xs: '1fr', lg: '320px minmax(0, 1fr)' },
      }}
    >
      <Box>
        <Card variant="outlined">
          <CardContent sx={{ p: 2.25 }}>
            <Typography variant="h6">钱包操作</Typography>
            <Button sx={{ mt: 2 }} fullWidth variant="contained" onClick={onCreateWallet}>
              创建钱包
            </Button>
          </CardContent>
        </Card>
      </Box>
      <Box>
        <Card variant="outlined">
          <CardContent sx={{ p: 2.25 }}>
            <Typography variant="h6">地址与余额</Typography>
            <Stack spacing={1.5} sx={{ mt: 2 }}>
              {wallets.length > 0 ? wallets.map((item) => (
                <Card key={item.address} variant="outlined" sx={{ borderRadius: 0.5 }}>
                  <CardContent sx={{ p: 1.5 }}>
                    <Typography variant="body2" sx={{ wordBreak: 'break-all' }}>
                      {item.address}
                    </Typography>
                    <Divider sx={{ my: 1.5 }} />
                    <Typography color="text.secondary">余额：{item.balance}</Typography>
                    <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mt: 1 }}>
                      P2PKH 脚本
                    </Typography>
                    <Typography variant="body2" sx={{ mt: 0.5, wordBreak: 'break-word', fontFamily: 'Consolas, monospace' }}>
                      {item.lockingScript}
                    </Typography>
                  </CardContent>
                </Card>
              )) : (
                <Typography color="text.secondary">
                  当前还没有钱包。
                </Typography>
              )}
            </Stack>
          </CardContent>
        </Card>
      </Box>
    </Box>
  )
}

export default WalletsPage
