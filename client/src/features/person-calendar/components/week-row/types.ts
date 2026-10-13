import type { ScheduleDay, ScheduleWeek } from '@/shared/people'

export interface WeekRowProps {
  week: ScheduleWeek
  /** The week's working days, in column order. */
  days: ScheduleDay[]
  /** The month on show; days outside it are dimmed. */
  month: string
  selected: boolean
  onSelect: (week: string) => void
}
