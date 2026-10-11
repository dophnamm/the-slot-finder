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

export interface RisksData {
  projects: RiskProject[]
  people: OverloadedPerson[]
}
