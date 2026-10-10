import type { Person } from '@/features/people'

export interface PersonAbsencesDialogProps {
  /** The person whose absences are shown; null keeps the dialog closed. */
  person: Person | null
  canApprove: boolean
  onClose: () => void
}
