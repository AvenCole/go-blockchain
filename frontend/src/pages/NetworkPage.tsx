import {
  Button,
  Card,
  CardContent,
  Chip,
  Divider,
  List,
  ListItem,
  ListItemText,
  Stack,
  TextField,
  Typography,
} from '@mui/material'
import type { NodeStatus } from '../types'

type NetworkPageProps = {
  nodes: NodeStatus[]
  nodeForm: { address: string; seed: string; miner: string }
  setNodeForm: React.Dispatch<React.SetStateAction<{ address: string; seed: string; miner: string }>>
  connectForm: { address: string; seed: string }
  setConnectForm: React.Dispatch<React.SetStateAction<{ address: string; seed: string }>>
  onStartNode: () => Promise<void>
  onStopNode: (address: string) => Promise<void>
  onConnectNode: () => Promise<void>
}

function NetworkPage({
  nodes,
  nodeForm,
  setNodeForm,
  connectForm,
  setConnectForm,
  onStartNode,
  onStopNode,
  onConnectNode,
}: NetworkPageProps) {
  return (
    <Stack spacing={2.5}>
      <Stack direction={{ xs: 'column', xl: 'row' }} spacing={2.5}>
        <Card variant="outlined" sx={{ flex: 1 }}>
          <CardContent>
            <Typography variant="h6">启动本地节点</Typography>
            <Typography color="text.secondary" sx={{ mt: 1 }}>
              在 GUI 中直接拉起 TCP 节点，便于演示节点监听、矿工地址绑定和种子接入。
            </Typography>
            <Stack spacing={2} sx={{ mt: 2.5 }}>
              <TextField
                fullWidth
                label="监听地址"
                value={nodeForm.address}
                onChange={(e) => setNodeForm((p) => ({ ...p, address: e.target.value }))}
                placeholder="127.0.0.1:3010 或 127.0.0.1:0"
              />
              <TextField
                fullWidth
                label="Seed 节点"
                value={nodeForm.seed}
                onChange={(e) => setNodeForm((p) => ({ ...p, seed: e.target.value }))}
                placeholder="可选：127.0.0.1:3011"
              />
              <TextField
                fullWidth
                label="矿工地址"
                value={nodeForm.miner}
                onChange={(e) => setNodeForm((p) => ({ ...p, miner: e.target.value }))}
                placeholder="可选，不填则只做普通节点"
              />
              <Button variant="contained" onClick={onStartNode}>
                启动节点
              </Button>
            </Stack>
          </CardContent>
        </Card>

        <Card variant="outlined" sx={{ flex: 1 }}>
          <CardContent>
            <Typography variant="h6">连接已有节点</Typography>
            <Typography color="text.secondary" sx={{ mt: 1 }}>
              用于演示节点发现和同步入口，先选择一个本地节点，再指定种子地址。
            </Typography>
            <Stack spacing={2} sx={{ mt: 2.5 }}>
              <TextField
                fullWidth
                label="本地节点地址"
                value={connectForm.address}
                onChange={(e) => setConnectForm((p) => ({ ...p, address: e.target.value }))}
                placeholder="例如 127.0.0.1:3010"
              />
              <TextField
                fullWidth
                label="Seed 地址"
                value={connectForm.seed}
                onChange={(e) => setConnectForm((p) => ({ ...p, seed: e.target.value }))}
                placeholder="例如 127.0.0.1:3011"
              />
              <Button variant="contained" color="secondary" onClick={onConnectNode}>
                连接 Seed
              </Button>
            </Stack>
          </CardContent>
        </Card>
      </Stack>

      <Card variant="outlined">
        <CardContent>
          <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.5} justifyContent="space-between" alignItems={{ xs: 'flex-start', md: 'center' }}>
            <div>
              <Typography variant="h6">节点状态</Typography>
              <Typography color="text.secondary" sx={{ mt: 0.5 }}>
                展示当前由 GUI 托管的节点、区块高度、矿工配置与已知 Peer。
              </Typography>
            </div>
            <Chip label={`GUI 节点数 ${nodes.length}`} variant="outlined" color="primary" />
          </Stack>

          <List sx={{ mt: 2 }}>
            {nodes.length === 0 ? (
              <ListItem>
                <ListItemText primary="当前还没有运行中的 GUI 节点" secondary="可先使用上方表单启动一个本地节点，再执行连接演示。" />
              </ListItem>
            ) : (
              nodes.map((node) => (
                <ListItem
                  key={node.address}
                  divider
                  alignItems="flex-start"
                  secondaryAction={
                    <Button color="error" onClick={() => void onStopNode(node.address)}>
                      停止
                    </Button>
                  }
                  sx={{ pr: 12 }}
                >
                  <ListItemText
                    primary={`${node.address}  (height=${node.height})`}
                    secondary={
                      <Stack spacing={0.75} sx={{ mt: 1 }}>
                        <Typography variant="body2">miner={node.minerAddress || '(none)'}</Typography>
                        <Typography variant="body2">running={String(node.running)}</Typography>
                        <Divider />
                        <Typography variant="body2">peers:</Typography>
                        {node.peers.length === 0 ? (
                          <Typography variant="body2" color="text.secondary">
                            暂无 peer
                          </Typography>
                        ) : (
                          node.peers.map((peer) => (
                            <Typography key={`${node.address}-${peer}`} variant="body2" color="text.secondary">
                              {peer}
                            </Typography>
                          ))
                        )}
                      </Stack>
                    }
                  />
                </ListItem>
              ))
            )}
          </List>
        </CardContent>
      </Card>
    </Stack>
  )
}

export default NetworkPage
