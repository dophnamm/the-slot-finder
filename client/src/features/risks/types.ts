import type { PersonRef, ProgressMetrics, ProjectStatus } from '@/shared/programs'

export interface RiskProject {
  id: string
  name: string
  program_id: string
  program: string
  status: ProjectStatus
  deadline_week: string
  metrics: ProgressMetrics
}

export interface OverloadedWeek {
  week: string
  load: number
  capacity: number
  utilization: number | null
}

export interface OverloadedPerson {
  person: PersonRef
  weeks: OverloadedWeek[]
}

/** A person over 100 % for several weeks in a row, from a week. */
export interface SustainedOverload {
  person: PersonRef
  from_week: string
  weeks: number
}

export interface RisksData {
  projects: RiskProject[]
  sustained_overload: SustainedOverload[]
  people: OverloadedPerson[]
}

export interface BenchPerson {
  person_id: string
  name: string
  primary_expertise: string
  free_hours: number
}

/** People with spare hours across the next `weeks` weeks. */
export interface BenchData {
  weeks: number
  people: BenchPerson[]
}
