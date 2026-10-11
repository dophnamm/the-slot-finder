import type { Project } from '@/shared/programs'

export interface ProjectFormDialogProps {
  open: boolean
  programId: string
  /** The project to edit, or null to add one to the program. */
  project: Project | null
  onOpenChange: (open: boolean) => void
}

/** Form fields as the inputs hold them. */
export interface ProjectFormValues {
  name: string
  priority: string
  start_earliest: string
  deadline: string
  scope_points: string
  done_points: string
  lead_id: string
}

export type ProjectTextField = keyof ProjectFormValues
