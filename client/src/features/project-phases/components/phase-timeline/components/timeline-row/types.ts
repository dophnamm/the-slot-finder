import type { TimelineRowData } from '../../types'

export interface TimelineRowProps {
  row: TimelineRowData
  weeks: string[]
  deadlineWeek: string
  currentWeek: string
}
