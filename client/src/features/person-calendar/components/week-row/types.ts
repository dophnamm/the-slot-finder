import type { ScheduleWeek } from '@/shared/people'
import type { CalendarCell } from '../../types'

export interface WeekRowProps {
  week: ScheduleWeek
  /** The week's seven days, in column order, each marked as worked or off. */
  days: CalendarCell[]
  /** The month on show; days outside it are dimmed. */
  month: string
  selected: boolean
  onSelect: (week: string) => void
}
