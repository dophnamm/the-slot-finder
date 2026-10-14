import type { ReactNode } from 'react'
import type { PROGRAM_TABS } from '@/resources/routes'

export type ProgramTab = (typeof PROGRAM_TABS)[keyof typeof PROGRAM_TABS]

export interface ProgramOverviewProps {
  programId: string
  /** The program's projects and team, from their own features, each shown in its tab. */
  projects: ReactNode
  team: ReactNode
}
