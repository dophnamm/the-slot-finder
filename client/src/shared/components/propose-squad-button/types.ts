import type { ProjectStatus } from '@/shared/programs'

export interface ProposeSquadButtonProps {
  projectId: string
  /** The button shows only for statuses a staffing draft may cover. */
  status: ProjectStatus
}
