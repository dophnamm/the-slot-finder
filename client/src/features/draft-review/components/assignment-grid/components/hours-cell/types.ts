import type { UtilizationBands } from '@/shared/utilization'

export interface HoursCellProps {
  week: string
  hours: number
  locked: boolean
  /** Inside the demand's window; outside it the cell is empty. */
  inWindow: boolean
  /** The person's capacity that week. */
  capacity: number
  /** Hours the person carries that week apart from this cell. */
  otherHours: number
  bands: UtilizationBands
  editable: boolean
  /** Hours that can be dragged to another week or person. */
  draggable: boolean
  /** Another cell's hours are being dragged and may be dropped here. */
  dropTarget: boolean
  onDragStart: (week: string) => void
  onDragEnd: () => void
  /** Called when the dragged hours are dropped on this cell. */
  onDrop: (week: string) => void
  /** Called with the new hours when the value changed; 0 removes the week. */
  onCommit: (week: string, hours: number) => void
}
