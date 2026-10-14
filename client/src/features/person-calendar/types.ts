import type { ScheduleDay } from '@/shared/people'

/** One day in the grid, and whether it is a working day or a day off. */
export interface CalendarCell {
  day: ScheduleDay
  working: boolean
}

export interface PersonCalendarProps {
  personId: string
  /** Whether the viewer may open project pages; without it projects are named as plain text. */
  canOpenProjects: boolean
}
