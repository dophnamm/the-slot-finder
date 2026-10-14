import type { ReactNode } from 'react'

export type ProgramTab = 'overview' | 'projects' | 'team'

export interface ProgramOverviewProps {
  programId: string
  /** The program's projects and team, from their own features, each shown in its tab. */
  projects: ReactNode
  team: ReactNode
}
