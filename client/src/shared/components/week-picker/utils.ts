import type { DateRange } from 'react-day-picker'
import { isIsoWeek, isoWeekRange } from '@/utils/functions'

/** Monday to Sunday of an ISO week, for the calendar to highlight; undefined when empty. */
export function weekDays(week: string): DateRange | undefined {
  if (!isIsoWeek(week)) return undefined
  const { start, end } = isoWeekRange(week)
  return { from: start.toDate(), to: end.toDate() }
}
