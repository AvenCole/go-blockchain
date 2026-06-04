import { CssBaseline } from '@mui/material'
import { RouterProvider } from '@tanstack/react-router'
import { router } from '@/app/router'
import { ThemeModeProvider } from '@/app/providers/ThemeModeProvider'
import { WorkbenchProvider } from '@/features/workbench/context/WorkbenchProvider'

function App() {
  return (
    <ThemeModeProvider>
      <CssBaseline />
      <WorkbenchProvider>
        <RouterProvider router={router} />
      </WorkbenchProvider>
    </ThemeModeProvider>
  )
}

export default App
