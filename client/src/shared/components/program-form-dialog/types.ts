import type { Program } from '@/shared/programs'

export interface ProgramFormDialogProps {
  open: boolean
  /** The program to edit, or null to create one. */
  program: Program | null
  onOpenChange: (open: boolean) => void
  /** Called with the saved program, e.g. to open it. */
  onSaved?: (program: Program) => void
}

/** Form fields as the inputs hold them; objectives are one per line. */
export interface ProgramFormValues {
  name: string
  client: string
  start_date: string
  end_date: string
  goal: string
  objectives: string
  charter_url: string
  lead_id: string
}

export type ProgramTextField = keyof ProgramFormValues
