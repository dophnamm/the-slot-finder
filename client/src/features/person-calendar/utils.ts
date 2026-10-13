import type { ScheduleDay, ScheduleWeek } from '@/shared/people'
import { isoWeekDate } from '@/utils/functions'

/** The working weekdays in week order, Monday first. */
export function orderedWeekdays(weekdays: number[]): number[] {
  return [...weekdays].sort((a, b) => a - b)
}

/** The schedule's days by date, for drawing the grid. */
export function daysByDate(days: ScheduleDay[]): Map<string, ScheduleDay> {
  return new Map(days.map((day) => [day.date, day]))
}

/** One grid row: each working weekday of the week, with what is on that day. */
export function weekCells(week: string, weekdays: number[], days: Map<string, ScheduleDay>): ScheduleDay[] {
  return weekdays.map((weekday) => {
    const date = isoWeekDate(week, weekday)
    return days.get(date) ?? { date, absences: [], holidays: [] }
  })
}

/** The schedule's week with this label; nothing when no week is selected. */
export function findWeek(weeks: ScheduleWeek[], week: string | null): ScheduleWeek | undefined {
  return week === null ? undefined : weeks.find((candidate) => candidate.week === week)
}
