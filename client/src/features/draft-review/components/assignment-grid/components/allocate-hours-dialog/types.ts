import type { Assignment, DraftEdit } from '@/shared/staffing'
import type { ALLOCATE_MODE } from './constants'

export type AllocateMode = (typeof ALLOCATE_MODE)[keyof typeof ALLOCATE_MODE]

export interface AllocateHoursDialogProps {
  assignment: Assignment
  /** The draft's weeks inside the demand's window; only these are offered. */
  weeks: string[]
  busy: boolean
  onSave: (changes: NonNullable<DraftEdit['set']>) => void
  onCancel: () => void
}
