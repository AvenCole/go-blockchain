import { Box, Button, Card, CardContent, Divider, List, ListItem, ListItemText, MenuItem, Stack, TextField, Typography } from '@mui/material'
import type { MultiSigOutputView } from '../types'

type TxForm = {
  template: 'p2pkh' | 'p2pk' | 'multisig'
  from: string
  to: string
  recipients: string
  required: string
  amount: string
  fee: string
}

type SpendMultiSigForm = {
  signers: string
  sourceTxID: string
  out: string
  to: string
  amount: string
  fee: string
}

type TransactionsPageProps = {
  txForm: TxForm
  setTxForm: React.Dispatch<React.SetStateAction<TxForm>>
  spendMultiSigForm: SpendMultiSigForm
  setSpendMultiSigForm: React.Dispatch<React.SetStateAction<SpendMultiSigForm>>
  multiSigOutputs: MultiSigOutputView[]
  minerAddress: string
  setMinerAddress: React.Dispatch<React.SetStateAction<string>>
  mempool: string[]
  onQueueTransaction: () => Promise<void>
  onSpendMultiSig: () => Promise<void>
  onMine: () => Promise<void>
}

function TransactionsPage({
  txForm,
  setTxForm,
  spendMultiSigForm,
  setSpendMultiSigForm,
  multiSigOutputs,
  minerAddress,
  setMinerAddress,
  mempool,
  onQueueTransaction,
  onSpendMultiSig,
  onMine,
}: TransactionsPageProps) {
  return (
    <Box
      sx={{
        display: 'grid',
        gap: 2,
        gridTemplateColumns: 'minmax(0, 1fr) minmax(0, 1fr)',
      }}
    >
      <Box>
        <Card variant="outlined">
          <CardContent sx={{ p: 2.25 }}>
            <Typography variant="h6">发送交易</Typography>
            <Stack spacing={2} sx={{ mt: 2 }}>
              <TextField
                select
                label="Script Template"
                value={txForm.template}
                onChange={(e) => setTxForm((p) => ({ ...p, template: e.target.value as TxForm['template'] }))}
              >
                <MenuItem value="p2pkh">P2PKH</MenuItem>
                <MenuItem value="p2pk">P2PK</MenuItem>
                <MenuItem value="multisig">MultiSig</MenuItem>
              </TextField>
              <TextField label="From" value={txForm.from} onChange={(e) => setTxForm((p) => ({ ...p, from: e.target.value }))} />
              {txForm.template === 'multisig' ? (
                <>
                  <TextField
                    label="Recipients CSV"
                    value={txForm.recipients}
                    onChange={(e) => setTxForm((p) => ({ ...p, recipients: e.target.value }))}
                    helperText="例如: addr1,addr2"
                  />
                  <TextField
                    label="Required Signers"
                    value={txForm.required}
                    onChange={(e) => setTxForm((p) => ({ ...p, required: e.target.value }))}
                  />
                </>
              ) : (
                <TextField label="To" value={txForm.to} onChange={(e) => setTxForm((p) => ({ ...p, to: e.target.value }))} />
              )}
              <TextField label="Amount" value={txForm.amount} onChange={(e) => setTxForm((p) => ({ ...p, amount: e.target.value }))} />
              <TextField label="Fee" value={txForm.fee} onChange={(e) => setTxForm((p) => ({ ...p, fee: e.target.value }))} />
              <Button variant="contained" onClick={onQueueTransaction}>加入交易池</Button>
            </Stack>
          </CardContent>
        </Card>
      </Box>
      <Box>
        <Card variant="outlined">
          <CardContent sx={{ p: 2.25 }}>
            <Typography variant="h6">多签花费</Typography>
            <Stack spacing={2} sx={{ mt: 2 }}>
              <TextField
                select
                label="Unspent MultiSig Output"
                value={`${spendMultiSigForm.sourceTxID}:${spendMultiSigForm.out}`}
                onChange={(e) => {
                  const [txid, out] = e.target.value.split(':')
                  const selected = multiSigOutputs.find((item) => item.txid === txid && String(item.out) === out)
                  setSpendMultiSigForm((prev) => ({
                    ...prev,
                    sourceTxID: txid,
                    out,
                    signers: selected ? selected.participants.join(',') : prev.signers,
                  }))
                }}
              >
                {multiSigOutputs.map((item) => (
                  <MenuItem key={`${item.txid}:${item.out}`} value={`${item.txid}:${item.out}`}>
                    {item.txid.slice(0, 12)}...:{item.out} | {item.required}/{item.participants.length} | value={item.value}
                  </MenuItem>
                ))}
              </TextField>
              <TextField
                label="Signers CSV"
                value={spendMultiSigForm.signers}
                onChange={(e) => setSpendMultiSigForm((p) => ({ ...p, signers: e.target.value }))}
                helperText="顺序必须与多签输出中的参与者顺序一致"
              />
              <TextField label="To" value={spendMultiSigForm.to} onChange={(e) => setSpendMultiSigForm((p) => ({ ...p, to: e.target.value }))} />
              <TextField label="Amount" value={spendMultiSigForm.amount} onChange={(e) => setSpendMultiSigForm((p) => ({ ...p, amount: e.target.value }))} />
              <TextField label="Fee" value={spendMultiSigForm.fee} onChange={(e) => setSpendMultiSigForm((p) => ({ ...p, fee: e.target.value }))} />
              <Button variant="contained" color="secondary" onClick={onSpendMultiSig}>花费多签输出</Button>
            </Stack>
            <Divider sx={{ my: 2 }} />
            <Typography variant="h6">挖矿与交易池</Typography>
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
              <Divider />
              <Typography variant="subtitle1">当前未花费多签输出</Typography>
              {multiSigOutputs.length === 0 ? (
                <Typography color="text.secondary">当前没有可花费的多签输出</Typography>
              ) : (
                <List dense>
                  {multiSigOutputs.map((item) => (
                    <ListItem key={`${item.txid}:${item.out}`}>
                      <ListItemText
                        primary={`${item.txid.slice(0, 16)}...:${item.out}  value=${item.value}`}
                        secondary={`${item.required}/${item.participants.length} | ${item.participants.join(', ')}`}
                      />
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
