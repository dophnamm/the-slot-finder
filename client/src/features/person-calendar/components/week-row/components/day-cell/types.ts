import type { ScheduleDay } from '@/shared/people'

export interface DayCellProps {
  day: ScheduleDay
  /** The month on show; days outside it are dimmed. */
  month: string
}
