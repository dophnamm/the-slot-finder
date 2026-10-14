import type { Expertise } from '@/shared/expertises'
import type { ProjectDetail } from '@/shared/programs'
import type { DemandDraft } from '../../types'

export interface DemandDialogProps {
  /** Already translated: "Add demand" or "Edit demand". */
  title: string
  /** The demand the form starts from; the dialog is remounted for each one. */
  initial: DemandDraft
  expertises: Expertise[]
  project: ProjectDetail
  saving: boolean
  onSubmit: (draft: DemandDraft) => void
  onOpenChange: (open: boolean) => void
}

/** Text fields of the form, matched by their name attribute. */
export type DemandTextField = 'expertise_id' | 'phase_id' | 'headcount' | 'total_hours'
