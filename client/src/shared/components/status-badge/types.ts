import type { ProgramStatus, ProjectStatus } from '@/shared/programs'

export interface StatusBadgeProps {
  status: ProgramStatus | ProjectStatus
  kind: 'program' | 'project'
}
