import type { Assignment } from '@/shared/staffing'
import type { UtilizationBands } from '@/shared/utilization'
import type { WeekLoad } from '../../types'

export interface AssignmentRowProps {
  assignment: Assignment
  weeks: string[]
  /** Capacity and the hours carried apart from this assignment, by week. */
  loads: Map<string, WeekLoad>
  bands: UtilizationBands
  editable: boolean
  busy: boolean
  onSetHours: (assignment: Assignment, week: string, hours: number) => void
  onToggleLock: (assignment: Assignment) => void
  onRemove: (assignment: Assignment) => void
  /** Opens the dialog that sets many weeks at once. */
  onAllocate: (assignment: Assignment) => void
}
