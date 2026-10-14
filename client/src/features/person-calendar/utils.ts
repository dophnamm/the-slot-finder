import type { ScheduleDay, ScheduleWeek } from '@/shared/people'
import { isoWeekDate } from '@/utils/functions'
import { CALENDAR_WEEKDAYS } from './constants'
import type { CalendarCell } from './types'

/** The schedule's days by date, for drawing the grid. */
export function daysByDate(days: ScheduleDay[]): Map<string, ScheduleDay> {
  return new Map(days.map((day) => [day.date, day]))
}

/** One grid row: all seven days of the week, Monday first, with what is on each and whether it is worked. */
export function weekCells(week: string, workingWeekdays: number[], days: Map<string, ScheduleDay>): CalendarCell[] {
  return CALENDAR_WEEKDAYS.map((weekday) => {
    const date = isoWeekDate(week, weekday)
    return { day: days.get(date) ?? { date, absences: [], holidays: [] }, working: workingWeekdays.includes(weekday) }
  })
}

/** The schedule's week with this label; nothing when no week is selected. */
export function findWeek(weeks: ScheduleWeek[], week: string | null): ScheduleWeek | undefined {
  return week === null ? undefined : weeks.find((candidate) => candidate.week === week)
}
