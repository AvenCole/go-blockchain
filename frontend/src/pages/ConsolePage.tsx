import { useMemo } from 'react'
import {
  Button,
  Card,
  CardContent,
  Chip,
  Paper,
  Stack,
  TextField,
  Typography,
} from '@mui/material'
import CommandPresetGroupCard from '../components/console/CommandPresetGroupCard'
import type {
  CommandResult,
  MultiSigOutputView,
  NodeStatus,
  WalletView,
} from '../types'
import {
  buildConsolePresetGroups,
  uniqueRecentCommands,
} from '../utils/consolePresets'

type ConsolePageProps = {
  command: string
  setCommand: React.Dispatch<React.SetStateAction<string>>
  history: CommandResult[]
  wallets: WalletView[]
  nodes: NodeStatus[]
  multiSigOutputs: MultiSigOutputView[]
  onExecute: () => Promise<void>
  onRunCommand: (commandLine: string) => Promise<void>
}

function ConsolePage({
  command,
  setCommand,
  history,
  wallets,
  nodes,
  multiSigOutputs,
  onExecute,
  onRunCommand,
}: ConsolePageProps) {
  const presetGroups = useMemo(
    () =>
      buildConsolePresetGroups({
        wallets,
        nodes,
        multiSigOutputs,
      }),
    [multiSigOutputs, nodes, wallets],
  )
  const recentCommands = useMemo(
    () => uniqueRecentCommands(history.map((item) => item.command)),
    [history],
  )

  return (
    <Stack spacing={2}>
      <Card variant="outlined">
        <CardContent sx={{ p: 2 }}>
          <Typography variant="h6">终端控制台</Typography>
          <Typography color="text.secondary" sx={{ mt: 1 }}>
            输入 CLI 命令并查看 stdout / stderr。
          </Typography>
          <Stack direction={{ xs: 'column', md: 'row' }} spacing={2} sx={{ mt: 2 }}>
            <TextField
              fullWidth
              label="命令"
              value={command}
              onChange={(e) => setCommand(e.target.value)}
              placeholder="例如: runnetdemo / runreorgdemo / runpartitiondemo / nodeinit <node> / nodesend <node> <from> <to> 10 1"
            />
            <Button variant="contained" onClick={onExecute}>
              执行
            </Button>
          </Stack>
          <Stack spacing={1.25} sx={{ mt: 2.5 }}>
            <Typography variant="subtitle2">最近命令</Typography>
            {recentCommands.length === 0 ? (
              <Typography variant="body2" color="text.secondary">
                执行过的命令会显示在这里，方便快速复用。
              </Typography>
            ) : (
              <Stack direction="row" spacing={1} useFlexGap sx={{ flexWrap: 'wrap' }}>
                {recentCommands.map((item) => (
                  <Chip
                    key={item}
                    label={item}
                    variant="outlined"
                    onClick={() => setCommand(item)}
                  />
                ))}
              </Stack>
            )}
          </Stack>
        </CardContent>
      </Card>

      <Stack direction={{ xs: 'column', xl: 'row' }} spacing={2}>
        <Stack sx={{ flex: 1, minWidth: 0 }}>
          <CommandPresetGroupCard
            group={presetGroups[0]}
            onFillCommand={setCommand}
            onRunCommand={onRunCommand}
          />
        </Stack>
        <Stack sx={{ flex: 1, minWidth: 0 }}>
          <CommandPresetGroupCard
            group={presetGroups[1]}
            onFillCommand={setCommand}
            onRunCommand={onRunCommand}
          />
        </Stack>
      </Stack>

      <Stack direction={{ xs: 'column', xl: 'row' }} spacing={2}>
        <Stack sx={{ flex: 1, minWidth: 0 }}>
          <CommandPresetGroupCard
            group={presetGroups[2]}
            onFillCommand={setCommand}
            onRunCommand={onRunCommand}
          />
        </Stack>
        <Stack sx={{ flex: 1, minWidth: 0 }}>
          <CommandPresetGroupCard
            group={presetGroups[3]}
            onFillCommand={setCommand}
            onRunCommand={onRunCommand}
          />
        </Stack>
      </Stack>

      <CommandPresetGroupCard
        group={presetGroups[4]}
        onFillCommand={setCommand}
        onRunCommand={onRunCommand}
      />

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
