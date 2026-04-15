import { Card, CardContent, Chip, List, ListItemButton, Stack, Typography } from '@mui/material'
import type { NodeStatus } from '../../types'
import { shortAddress, shortHash } from '../../utils/networkView'

type NodeDirectoryCardProps = {
  nodes: NodeStatus[]
  selectedNodeAddress: string
  onSelectNode: (address: string) => void
}

function NodeDirectoryCard({
  nodes,
  selectedNodeAddress,
  onSelectNode,
}: NodeDirectoryCardProps) {
  return (
    <Card variant="outlined">
      <CardContent sx={{ p: 2 }}>
        <Stack
          direction={{ xs: 'column', sm: 'row' }}
          spacing={1.5}
          sx={{ justifyContent: 'space-between', alignItems: { xs: 'flex-start', sm: 'center' } }}
        >
          <Typography variant="h6">节点目录</Typography>
          <Chip label={`节点 ${nodes.length}`} variant="outlined" color="primary" />
        </Stack>

        {nodes.length === 0 ? (
          <Typography color="text.secondary" sx={{ mt: 2 }}>
            当前没有运行中的节点。
          </Typography>
        ) : (
          <List disablePadding sx={{ mt: 2 }}>
            {nodes.map((node) => {
              const selected = node.address === selectedNodeAddress

              return (
                <ListItemButton
                  key={node.address}
                  selected={selected}
                  onClick={() => onSelectNode(node.address)}
                  sx={{
                    mb: 1,
                    display: 'block',
                    border: 1,
                    borderRadius: 0.5,
                    borderColor: selected ? 'primary.main' : 'divider',
                    alignItems: 'flex-start',
                  }}
                >
                  <Stack
                    direction={{ xs: 'column', sm: 'row' }}
                    spacing={1.25}
                    sx={{ justifyContent: 'space-between', alignItems: { xs: 'flex-start', sm: 'center' } }}
                  >
                    <Typography variant="subtitle2">
                      {shortAddress(node.address, 10, 6)}
                    </Typography>
                    <Stack direction="row" spacing={0.75} useFlexGap sx={{ flexWrap: 'wrap' }}>
                      <Chip
                        size="small"
                        label={node.initialized ? `h=${node.height}` : '未初始化'}
                        color={node.initialized ? 'success' : 'default'}
                        variant="outlined"
                      />
                      <Chip
                        size="small"
                        label={`mempool ${node.mempoolCount}`}
                        variant="outlined"
                      />
                    </Stack>
                  </Stack>

                  <Typography
                    variant="body2"
                    color="text.secondary"
                    sx={{ mt: 0.75, wordBreak: 'break-all' }}
                  >
                    {node.address}
                  </Typography>
                  <Typography variant="caption" color="text.secondary" sx={{ mt: 1, display: 'block' }}>
                    tip {shortHash(node.tipHash)}
                  </Typography>
                </ListItemButton>
              )
            })}
          </List>
        )}
      </CardContent>
    </Card>
  )
}

export default NodeDirectoryCard
