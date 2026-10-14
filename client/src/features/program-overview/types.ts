import type { ReactNode } from 'react'

export type ProgramTab = 'overview' | 'projects' | 'members'

export interface ProgramOverviewProps {
  programId: string
  /** The program's projects and members, from their own features, each shown in its tab. */
  projects: ReactNode
  members: ReactNode
}
