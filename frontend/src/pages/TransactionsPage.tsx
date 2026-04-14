import { Box, Button, Card, CardContent, Divider, List, ListItem, ListItemText, Stack, TextField, Typography } from '@mui/material'

type TxForm = {
  from: string
  to: string
  amount: string
  fee: string
}

type TransactionsPageProps = {
  txForm: TxForm
  setTxForm: React.Dispatch<React.SetStateAction<TxForm>>
  minerAddress: string
  setMinerAddress: React.Dispatch<React.SetStateAction<string>>
  mempool: string[]
  onQueueTransaction: () => Promise<void>
  onMine: () => Promise<void>
}

function TransactionsPage({
  txForm,
  setTxForm,
  minerAddress,
  setMinerAddress,
  mempool,
  onQueueTransaction,
  onMine,
}: TransactionsPageProps) {
  return (
    <Box sx={{ display: 'grid', gap: 2, gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' } }}>
      <Box>
        <Card variant="outlined">
          <CardContent>
            <Typography variant="h6">发送交易</Typography>
            <Typography color="text.secondary" sx={{ mt: 1 }}>
              这里的发送动作会把交易排队进入 mempool，而不是立刻成块。
            </Typography>
            <Stack spacing={2} sx={{ mt: 2 }}>
              <TextField label="From" value={txForm.from} onChange={(e) => setTxForm((p) => ({ ...p, from: e.target.value }))} />
              <TextField label="To" value={txForm.to} onChange={(e) => setTxForm((p) => ({ ...p, to: e.target.value }))} />
              <TextField label="Amount" value={txForm.amount} onChange={(e) => setTxForm((p) => ({ ...p, amount: e.target.value }))} />
              <TextField label="Fee" value={txForm.fee} onChange={(e) => setTxForm((p) => ({ ...p, fee: e.target.value }))} />
              <Button variant="contained" onClick={onQueueTransaction}>加入交易池</Button>
            </Stack>
          </CardContent>
        </Card>
      </Box>
      <Box>
        <Card variant="outlined">
          <CardContent>
            <Typography variant="h6">挖矿与交易池</Typography>
            <Typography color="text.secondary" sx={{ mt: 1 }}>
              指定矿工地址后，当前待打包交易会被统一出块。
            </Typography>
            <Stack spacing={2} sx={{ mt: 2 }}>
              <TextField label="Miner Address" value={minerAddress} onChange={(e) => setMinerAddress(e.target.value)} />
              <Button variant="contained" color="secondary" onClick={onMine}>打包并挖矿</Button>
              <Divider />
              <Typography variant="subtitle1">当前 Mempool</Typography>
              {mempool.length === 0 ? (
                <Typography color="text.secondary">当前没有待打包交易</Typography>
              ) : (
                <List dense>
                  {mempool.map((txid) => (
                    <ListItem key={txid}>
                      <ListItemText primary={txid} />
                    </ListItem>
                  ))}
                </List>
              )}
            </Stack>
          </CardContent>
        </Card>
      </Box>
    </Box>
  )
}

export default TransactionsPage
