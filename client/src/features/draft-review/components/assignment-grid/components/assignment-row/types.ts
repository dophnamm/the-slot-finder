import type { Assignment } from '@/shared/staffing'
import type { UtilizationBands } from '@/shared/utilization'
import type { DragSource, WeekLoad } from '../../types'

export interface AssignmentRowProps {
  assignment: Assignment
  weeks: string[]
  /** Capacity and the hours carried apart from this assignment, by week. */
  loads: Map<string, WeekLoad>
  bands: UtilizationBands
  editable: boolean
  /** The hours being dragged in the grid, or null. */
  drag: DragSource | null
  onDragStart: (source: DragSource) => void
  onDragEnd: () => void
  /** Called with this assignment and week when the dragged hours are dropped there. */
  onDropHours: (target: DragSource) => void
  busy: boolean
  onSetHours: (assignment: Assignment, week: string, hours: number) => void
  onToggleLock: (assignment: Assignment) => void
  onRemove: (assignment: Assignment) => void
  /** Opens the dialog that sets many weeks at once. */
  onAllocate: (assignment: Assignment) => void
  /** Opens the dialog that swaps the person for another from the program team. */
  onSwap: (assignment: Assignment) => void
}
