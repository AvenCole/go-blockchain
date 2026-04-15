import {
  Button,
  Card,
  CardContent,
  Chip,
  Stack,
  Typography,
} from '@mui/material'
import type { CommandPresetGroup } from '../../utils/consolePresets'

type CommandPresetGroupCardProps = {
  group: CommandPresetGroup
  onFillCommand: (command: string) => void
  onRunCommand: (command: string) => Promise<void>
}

function CommandPresetGroupCard({
  group,
  onFillCommand,
  onRunCommand,
}: CommandPresetGroupCardProps) {
  return (
    <Card variant="outlined">
      <CardContent sx={{ p: 2 }}>
        <Typography variant="h6">{group.title}</Typography>
        <Stack spacing={1.5} sx={{ mt: 2 }}>
          {group.presets.map((preset) => (
            <Stack
              key={preset.id}
              spacing={1}
              sx={{
                p: 1.5,
                border: 1,
                borderColor: 'divider',
                borderRadius: 0.5,
              }}
            >
              <Stack
                direction={{ xs: 'column', md: 'row' }}
                spacing={1}
                sx={{ justifyContent: 'space-between', alignItems: { xs: 'flex-start', md: 'center' } }}
              >
                <Stack direction="row" spacing={1} sx={{ alignItems: 'center' }}>
                  <Typography variant="subtitle2">{preset.label}</Typography>
                  <Chip
                    size="small"
                    label={preset.ready ? '可直接执行' : '需补参数'}
                    color={preset.ready ? 'success' : 'default'}
                    variant="outlined"
                  />
                </Stack>
                <Stack direction="row" spacing={1}>
                  <Button
                    size="small"
                    variant="outlined"
                    onClick={() => onFillCommand(preset.command)}
                  >
                    填入
                  </Button>
                  <Button
                    size="small"
                    variant="contained"
                    disabled={!preset.ready}
                    onClick={() => void onRunCommand(preset.command)}
                  >
                    执行
                  </Button>
                </Stack>
              </Stack>
              <Typography variant="body2" color="text.secondary">
                {preset.description}
              </Typography>
              <Typography
                variant="body2"
                sx={{
                  fontFamily: 'Consolas, monospace',
                  bgcolor: 'action.hover',
                  px: 1,
                  py: 0.75,
                  borderRadius: 0.5,
                  wordBreak: 'break-all',
                }}
              >
                {preset.command}
              </Typography>
            </Stack>
          ))}
        </Stack>
      </CardContent>
    </Card>
  )
}

export default CommandPresetGroupCard
