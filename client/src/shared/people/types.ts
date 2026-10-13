import type { Role } from '@/resources/rbac'
import type { UtilizationBand } from '@/shared/utilization'

export interface PersonSkill {
  expertise_id: string
  expertise: string
  level: number
  primary: boolean
}

/** A person as `GET /people` returns it. Dates use the server format `2026-10-12`. */
export interface Person {
  id: string
  name: string
  email: string
  role: Role
  weekly_hours: number
  non_project_hours: number
  start_date: string | null
  end_date: string | null
  active: boolean
  skills: PersonSkill[]
}

/** Absence types and statuses; they match the server's constants/absence.go. */
export type AbsenceType = 'leave' | 'training' | 'holiday'
export type AbsenceStatus = 'planned' | 'approved'

/** Where an absence comes from: entered by hand, or read from the person's calendar. */
export type AbsenceSource = 'manual' | 'calendar'

/** A person's baseline hours on one project and expertise in a week. */
export interface ScheduleProject {
  program_id: string
  program_name: string
  project_id: string
  project_name: string
  expertise: string
  hours: number
}

/** A person's load against capacity in one ISO week, computed as on the heatmap. */
export interface ScheduleWeek {
  week: string
  capacity: number
  load: number
  /** load ÷ capacity; null when capacity is 0. */
  utilization: number | null
  band: UtilizationBand
  /** Planned (not yet approved) absence this week. */
  tentative: boolean
  /** Hours a timesheet import recorded for the week; null when none are recorded. */
  actual_hours: number | null
  projects: ScheduleProject[]
}

/** One of the person's absences on a day. */
export interface ScheduleAbsence {
  id: string
  type: AbsenceType
  status: AbsenceStatus
  hours_per_day: number
  source: AbsenceSource
}

/** A company-wide holiday on a day; holidays have no name. */
export interface ScheduleHoliday {
  id: string
}

/** A working day with the person's absences and the holidays on it. */
export interface ScheduleDay {
  /** `2026-10-21` */
  date: string
  absences: ScheduleAbsence[]
  holidays: ScheduleHoliday[]
}

/** `GET /people/{id}/schedule`: weeks in the range, and only the days with something on them. */
export interface PersonSchedule {
  /** ISO day numbers that count: 1 is Monday, 7 is Sunday. */
  working_weekdays: number[]
  weeks: ScheduleWeek[]
  days: ScheduleDay[]
}
