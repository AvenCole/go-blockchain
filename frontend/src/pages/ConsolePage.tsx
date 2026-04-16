import { useMemo } from 'react'
import {
  Box,
  Button,
  Divider,
  List,
  ListItemButton,
  ListItemText,
  Paper,
  Stack,
  TextField,
  Typography,
} from '@mui/material'
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
    <Box
      sx={{
        display: 'grid',
        gap: 2,
        gridTemplateColumns: '280px minmax(0, 1fr)',
        minHeight: 760,
      }}
    >
      <Paper
        variant="outlined"
        sx={{
          overflow: 'hidden',
          display: 'flex',
          flexDirection: 'column',
          minHeight: 0,
        }}
      >
        <Box sx={{ px: 2, py: 1.5 }}>
          <Typography variant="h6">命令列表</Typography>
        </Box>
        <Divider />
        <Box sx={{ overflow: 'auto', minHeight: 0 }}>
          {recentCommands.length > 0 ? (
            <>
              <Box sx={{ px: 2, py: 1.25 }}>
                <Typography variant="subtitle2" color="text.secondary">
                  最近
                </Typography>
              </Box>
              <List dense disablePadding>
                {recentCommands.map((item) => (
                  <ListItemButton key={item} onClick={() => setCommand(item)}>
                    <ListItemText primary={item} />
                  </ListItemButton>
                ))}
              </List>
              <Divider />
            </>
          ) : null}

          {presetGroups.map((group) => (
            <Box key={group.id}>
              <Box sx={{ px: 2, py: 1.25 }}>
                <Typography variant="subtitle2" color="text.secondary">
                  {group.title}
                </Typography>
              </Box>
              <List dense disablePadding>
                {group.presets.map((preset) => (
                  <ListItemButton
                    key={preset.id}
                    onClick={() => setCommand(preset.command)}
                    onDoubleClick={() => void onRunCommand(preset.command)}
                  >
                    <ListItemText
                      primary={preset.label}
                      secondary={
                        <Typography
                          component="span"
                          variant="body2"
                          sx={{
                            fontFamily: 'Consolas, monospace',
                            wordBreak: 'break-all',
                            color: 'text.secondary',
                          }}
                        >
                          {preset.command}
                        </Typography>
                      }
                    />
                  </ListItemButton>
                ))}
              </List>
              <Divider />
            </Box>
          ))}
        </Box>
      </Paper>

      <Paper
        variant="outlined"
        sx={{
          minHeight: 0,
          display: 'flex',
          flexDirection: 'column',
          overflow: 'hidden',
          bgcolor: '#0b1220',
          color: '#d7e2f0',
          borderColor: '#1f2937',
        }}
      >
        <Stack
          direction="row"
          spacing={1.5}
          sx={{
            px: 2,
            py: 1.5,
            alignItems: 'center',
            borderBottom: '1px solid #1f2937',
          }}
        >
          <Typography variant="h6" sx={{ minWidth: 96 }}>
            终端
          </Typography>
          <TextField
            fullWidth
            size="small"
            value={command}
            onChange={(e) => setCommand(e.target.value)}
            placeholder="输入命令"
            variant="outlined"
            slotProps={{
              input: {
                sx: {
                  color: '#e5eefb',
                  fontFamily: 'Consolas, monospace',
                },
              },
            }}
            sx={{
              '& .MuiOutlinedInput-root': {
                bgcolor: '#111827',
                borderRadius: 1,
              },
              '& .MuiOutlinedInput-notchedOutline': {
                borderColor: '#334155',
              },
            }}
          />
          <Button variant="contained" onClick={onExecute}>
            执行
          </Button>
        </Stack>

        <Box
          sx={{
            flex: 1,
            minHeight: 0,
            overflow: 'auto',
            px: 2,
            py: 1.5,
            fontFamily: 'Consolas, monospace',
          }}
        >
          {history.length === 0 ? (
            <Typography sx={{ color: '#7dd3fc', fontFamily: 'Consolas, monospace' }}>
              {'>'} 等待输入
            </Typography>
          ) : (
            <Stack spacing={2}>
              {history.map((item, index) => (
                <Stack key={`${item.command}-${index}`} spacing={0.9}>
                  <Typography sx={{ color: '#7dd3fc', fontFamily: 'Consolas, monospace' }}>
                    {'>'} {item.command}
                  </Typography>
                  {item.stdout ? (
                    <Typography
                      sx={{
                        whiteSpace: 'pre-wrap',
                        fontFamily: 'Consolas, monospace',
                      }}
                    >
                      {item.stdout}
                    </Typography>
                  ) : null}
                  {item.stderr ? (
                    <Typography
                      sx={{
                        whiteSpace: 'pre-wrap',
                        color: '#fda4af',
                        fontFamily: 'Consolas, monospace',
                      }}
                    >
                      {item.stderr}
                    </Typography>
                  ) : null}
                  <Typography
                    variant="caption"
                    sx={{ color: '#94a3b8', fontFamily: 'Consolas, monospace' }}
                  >
                    exitCode={item.exitCode}
                  </Typography>
                </Stack>
              ))}
            </Stack>
          )}
        </Box>
      </Paper>
    </Box>
  )
}

export default ConsolePage
