import {
  createContext,
  useContext,
  useMemo,
  useState,
  type PropsWithChildren,
} from 'react'
import {
  GlobalStyles,
  PaletteMode,
  ThemeProvider,
  alpha,
  createTheme,
} from '@mui/material'

type ThemeModeContextValue = {
  mode: PaletteMode
  toggleColorMode: () => void
}

const STORAGE_KEY = 'themeMode'

const ThemeModeContext = createContext<ThemeModeContextValue | undefined>(
  undefined,
)

function getInitialMode(): PaletteMode {
  if (typeof window === 'undefined') {
    return 'light'
  }

  const storedMode = window.localStorage.getItem(STORAGE_KEY)
  if (storedMode === 'light' || storedMode === 'dark') {
    return storedMode
  }

  if (typeof window.matchMedia === 'function') {
    return window.matchMedia('(prefers-color-scheme: dark)').matches
      ? 'dark'
      : 'light'
  }

  return 'light'
}

export function ThemeModeProvider({ children }: PropsWithChildren) {
  const [mode, setMode] = useState<PaletteMode>(getInitialMode)

  const theme = useMemo(
    () =>
      createTheme({
        palette: {
          mode,
        },
      }),
    [mode],
  )

  const value = useMemo<ThemeModeContextValue>(
    () => ({
      mode,
      toggleColorMode: () => {
        setMode((previousMode) => {
          const nextMode = previousMode === 'light' ? 'dark' : 'light'

          if (typeof window !== 'undefined') {
            window.localStorage.setItem(STORAGE_KEY, nextMode)
          }

          return nextMode
        })
      },
    }),
    [mode],
  )

  return (
    <ThemeModeContext.Provider value={value}>
      <ThemeProvider theme={theme}>
        <GlobalStyles
          styles={(currentTheme) => {
            const thumbColor = alpha(
              currentTheme.palette.text.primary,
              currentTheme.palette.mode === 'dark' ? 0.28 : 0.18,
            )
            const thumbHoverColor = alpha(
              currentTheme.palette.text.primary,
              currentTheme.palette.mode === 'dark' ? 0.44 : 0.3,
            )
            const trackColor = alpha(
              currentTheme.palette.divider,
              currentTheme.palette.mode === 'dark' ? 0.16 : 0.08,
            )

            return {
              'html, body': {
                scrollbarWidth: 'thin',
                scrollbarColor: `${thumbColor} ${trackColor}`,
              },
              '*': {
                scrollbarWidth: 'thin',
                scrollbarColor: `${thumbColor} ${trackColor}`,
              },
              '*::-webkit-scrollbar': {
                width: 8,
                height: 8,
              },
              '*::-webkit-scrollbar-track': {
                backgroundColor: trackColor,
              },
              '*::-webkit-scrollbar-thumb': {
                backgroundColor: thumbColor,
                borderRadius: 999,
                border: `2px solid ${trackColor}`,
              },
              '*::-webkit-scrollbar-thumb:hover': {
                backgroundColor: thumbHoverColor,
              },
              '*::-webkit-scrollbar-corner': {
                backgroundColor: 'transparent',
              },
            }
          }}
        />
        {children}
      </ThemeProvider>
    </ThemeModeContext.Provider>
  )
}

export function useThemeMode() {
  const context = useContext(ThemeModeContext)

  if (!context) {
    throw new Error('useThemeMode must be used within ThemeModeProvider')
  }

  return context
}
