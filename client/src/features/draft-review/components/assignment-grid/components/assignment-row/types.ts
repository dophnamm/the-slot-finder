import type { Assignment } from '@/shared/staffing'

export interface AssignmentRowProps {
  assignment: Assignment
  weeks: string[]
  editable: boolean
  busy: boolean
  onSetHours: (assignment: Assignment, week: string, hours: number) => void
  onToggleLock: (assignment: Assignment) => void
  onRemove: (assignment: Assignment) => void
  /** Opens the dialog that sets many weeks at once. */
  onAllocate: (assignment: Assignment) => void
}
