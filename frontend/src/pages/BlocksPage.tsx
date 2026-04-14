import { Accordion, AccordionDetails, AccordionSummary, Card, CardContent, Chip, Divider, Stack, Typography } from '@mui/material'
import ExpandMoreIcon from '@mui/icons-material/ExpandMore'
import type { BlockView } from '../types'
import { shortHash } from '../utils/format'

type BlocksPageProps = {
  blocks: BlockView[]
}

function BlocksPage({ blocks }: BlocksPageProps) {
  return (
    <Stack spacing={2}>
      {blocks.map((block) => (
        <Card key={block.hash} variant="outlined">
          <CardContent>
            <Stack direction="row" spacing={1} sx={{ justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap' }}>
              <Typography variant="h6">区块 #{block.height}</Typography>
              <Stack direction="row" spacing={1} sx={{ flexWrap: 'wrap' }}>
                <Chip label={`Txs: ${block.transactionCount}`} color="primary" size="small" />
                <Chip label={`Diff: ${block.difficulty}`} size="small" />
                <Chip label={`Nonce: ${block.nonce}`} size="small" />
              </Stack>
            </Stack>
            <Divider sx={{ my: 2 }} />
            <Typography variant="body2">Hash: {shortHash(block.hash, 16, 12)}</Typography>
            <Typography variant="body2">Prev: {shortHash(block.prevHash, 16, 12)}</Typography>
            <Typography variant="body2">Merkle: {shortHash(block.merkleRoot, 16, 12)}</Typography>
            <Typography variant="body2">PoW: difficulty={block.difficulty} nonce={block.nonce} valid={String(block.powValid)}</Typography>
            <Typography variant="body2">Time: {block.timestamp}</Typography>
            <Stack spacing={1} sx={{ mt: 2 }}>
              {block.transactions.map((tx) => (
                <Accordion key={tx.id} disableGutters elevation={0}>
                  <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                    <Stack direction="row" spacing={1} sx={{ alignItems: 'center', flexWrap: 'wrap' }}>
                      <Typography variant="subtitle2">TxID: {shortHash(tx.id, 14, 10)}</Typography>
                      <Chip size="small" label={`V${tx.version}`} />
                      <Chip size="small" label={`Fee: ${tx.fee}`} />
                      <Chip size="small" color={tx.usesScriptVM ? 'success' : 'default'} label={tx.usesScriptVM ? 'Script VM' : 'Legacy'} />
                    </Stack>
                  </AccordionSummary>
                  <AccordionDetails>
                    <Typography variant="body2" sx={{ mt: 1 }}>Inputs:</Typography>
                    {tx.inputs.map((input, index) => (
                      <Stack key={`${tx.id}-in-${index}`} spacing={0.5} sx={{ mb: 1 }}>
                        <Typography variant="body2" color="text.secondary">
                          txid={shortHash(input.txid || '(coinbase)', 12, 8)} out={input.out} source={shortHash(input.source, 12, 8)}
                        </Typography>
                        <Typography variant="caption" sx={{ fontFamily: 'Consolas, monospace', wordBreak: 'break-word' }}>
                          ScriptSig: {input.scriptSig}
                        </Typography>
                      </Stack>
                    ))}
                    <Typography variant="body2" sx={{ mt: 1 }}>Outputs:</Typography>
                    {tx.outputs.map((output, index) => (
                      <Stack key={`${tx.id}-out-${index}`} spacing={0.5} sx={{ mb: 1 }}>
                        <Typography variant="body2" color="text.secondary">
                          to={shortHash(output.to, 12, 8)} value={output.value}
                        </Typography>
                        <Typography variant="caption" sx={{ fontFamily: 'Consolas, monospace', wordBreak: 'break-word' }}>
                          ScriptPubKey: {output.scriptPubKey}
                        </Typography>
                      </Stack>
                    ))}
                  </AccordionDetails>
                </Accordion>
              ))}
            </Stack>
          </CardContent>
        </Card>
      ))}
    </Stack>
  )
}

export default BlocksPage
