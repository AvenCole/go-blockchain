import { Chip, MenuItem, Stack, TextField } from '@mui/material'

type EventFilterToolbarProps = {
  query: string
  onQueryChange: (value: string) => void
  kind: string
  kindOptions: string[]
  onKindChange: (value: string) => void
  matchedCount: number
  totalCount: number
  source?: string
  sourceOptions?: string[]
  onSourceChange?: (value: string) => void
}

function EventFilterToolbar({
  query,
  onQueryChange,
  kind,
  kindOptions,
  onKindChange,
  matchedCount,
  totalCount,
  source = '',
  sourceOptions = [],
  onSourceChange,
}: EventFilterToolbarProps) {
  return (
    <Stack
      direction={{ xs: 'column', lg: 'row' }}
      spacing={1.25}
      sx={{ mt: 1.5, alignItems: { lg: 'center' } }}
    >
      <TextField
        size="small"
        label="搜索"
        value={query}
        onChange={(event) => onQueryChange(event.target.value)}
        placeholder="按类型、节点、摘要筛选"
        sx={{ minWidth: { lg: 220 } }}
      />
      <TextField
        select
        size="small"
        label="类型"
        value={kind}
        onChange={(event) => onKindChange(event.target.value)}
        sx={{ minWidth: { lg: 160 } }}
      >
        <MenuItem value="">全部类型</MenuItem>
        {kindOptions.map((option) => (
          <MenuItem key={option} value={option}>
            {option}
          </MenuItem>
        ))}
      </TextField>
      {onSourceChange ? (
        <TextField
          select
          size="small"
          label="来源"
          value={source}
          onChange={(event) => onSourceChange(event.target.value)}
          sx={{ minWidth: { lg: 180 } }}
        >
          <MenuItem value="">全部来源</MenuItem>
          {sourceOptions.map((option) => (
            <MenuItem key={option} value={option}>
              {option}
            </MenuItem>
          ))}
        </TextField>
      ) : null}
      <Chip
        label={`显示 ${matchedCount} / ${totalCount}`}
        variant="outlined"
        color={matchedCount === totalCount ? 'default' : 'primary'}
      />
    </Stack>
  )
}

export default EventFilterToolbar
