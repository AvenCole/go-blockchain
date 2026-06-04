import {
  createContext,
  type PropsWithChildren,
} from 'react'
import { useWorkbenchController } from '@/features/workbench/hooks/useWorkbenchController'
import type { WorkbenchContextValue } from '@/features/workbench/types'

export const WorkbenchContext =
  createContext<WorkbenchContextValue | undefined>(undefined)

export function WorkbenchProvider({ children }: PropsWithChildren) {
  const value = useWorkbenchController()

  return (
    <WorkbenchContext.Provider value={value}>
      {children}
    </WorkbenchContext.Provider>
  )
}
