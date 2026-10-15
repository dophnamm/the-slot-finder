import type { ProjectDetail } from '@/shared/programs'

export interface ScopePointsProps {
  project: ProjectDetail
  /** The viewer may change the project and it is still open. */
  editable: boolean
}
