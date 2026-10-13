import type { ScheduleWeek } from '@/shared/people'

export interface WeekDetailProps {
  week: ScheduleWeek
  /** Whether the viewer may open project pages; without it projects are named as plain text. */
  canOpenProjects: boolean
}
