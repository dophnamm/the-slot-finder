import type { Absence } from '../../types'

export interface AbsenceRowProps {
  absence: Absence
  onDelete: (id: string) => void
  /** Approves a planned absence; without it the row has no Approve button. */
  onApprove?: (id: string) => void
  /** This absence's approval is running; Approve is locked until it ends. */
  approving?: boolean
}
