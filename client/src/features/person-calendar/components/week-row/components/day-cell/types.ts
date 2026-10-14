import type { ScheduleDay } from '@/shared/people'

export interface DayCellProps {
  day: ScheduleDay
  /** A day off (outside the working weekdays) is shaded. */
  working: boolean
  /** The month on show; days outside it are dimmed. */
  month: string
}
