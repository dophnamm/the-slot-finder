import type { Assignment } from '@/shared/staffing'

export interface SwapPersonDialogProps {
  draftId: string
  /** The assignment whose person is replaced. */
  assignment: Assignment
  busy: boolean
  /** Called with the chosen person; the grid moves the hours over. */
  onChoose: (personId: string) => void
  onCancel: () => void
}
