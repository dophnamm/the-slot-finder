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
  /** Called with the new hours when the value changed; 0 removes the week. */
  onCommit: (week: string, hours: number) => void
}
