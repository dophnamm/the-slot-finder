import type { ProjectStatus } from '@/shared/programs'

export interface ProposeSquadButtonProps {
  projectId: string
  /** The button shows only for statuses a staffing draft may cover. */
  status: ProjectStatus
  /** The manager of the project's program; the button shows for them and for planners. */
  programLeadId: string | null | undefined
}
