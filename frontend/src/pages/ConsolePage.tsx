import { Button, Card, CardContent, Paper, Stack, TextField, Typography } from '@mui/material'
import type { CommandResult } from '../types'

type ConsolePageProps = {
  command: string
  setCommand: React.Dispatch<React.SetStateAction<string>>
  history: CommandResult[]
  onExecute: () => Promise<void>
}

function ConsolePage({ command, setCommand, history, onExecute }: ConsolePageProps) {
  return (
    <Stack spacing={2}>
      <Card variant="outlined">
        <CardContent sx={{ p: 2 }}>
          <Typography variant="h6">终端控制台</Typography>
          <Typography color="text.secondary" sx={{ mt: 1 }}>
            可直接输入 CLI 命令进行演示，适合答辩时展示终端链路。
          </Typography>
          <Stack direction={{ xs: 'column', md: 'row' }} spacing={2} sx={{ mt: 2 }}>
            <TextField
              fullWidth
              label="Command"
              value={command}
              onChange={(e) => setCommand(e.target.value)}
              placeholder="例如: runnetdemo / runreorgdemo / runpartitiondemo / nodeinit <node> / nodesend <node> <from> <to> 10 1"
            />
            <Button variant="contained" onClick={onExecute}>
              执行
            </Button>
          </Stack>
        </CardContent>
      </Card>

      <Paper
        variant="outlined"
        sx={{
          p: 2,
          minHeight: 420,
          bgcolor: '#0f172a',
          color: '#d6f5d6',
          fontFamily: 'Consolas, monospace',
          overflow: 'auto',
          borderRadius: 0.5,
          borderColor: 'divider',
        }}
      >
        <Stack spacing={2}>
          {history.length === 0 ? (
            <Typography sx={{ color: '#8adf8a' }}>
              尚无命令输出。可以先尝试：runnetdemo、runreorgdemo、runpartitiondemo 或 nodes
            </Typography>
          ) : (
            history.map((item, index) => (
              <Stack key={`${item.command}-${index}`} spacing={1}>
                <Typography sx={{ color: '#7dd3fc' }}>{'>'} {item.command}</Typography>
                {item.stdout ? <Typography sx={{ whiteSpace: 'pre-wrap' }}>{item.stdout}</Typography> : null}
                {item.stderr ? <Typography sx={{ whiteSpace: 'pre-wrap', color: '#fca5a5' }}>{item.stderr}</Typography> : null}
                <Typography sx={{ color: '#94a3b8' }}>exitCode={item.exitCode}</Typography>
              </Stack>
            ))
          )}
        </Stack>
      </Paper>
    </Stack>
  )
}

export default ConsolePage
