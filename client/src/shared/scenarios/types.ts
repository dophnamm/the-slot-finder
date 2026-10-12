import type { PersonRef } from '@/shared/programs'
import type { PlanChangeKind, UnmetReason } from '@/shared/staffing'

/** Hypothetical changes a scenario can hold; they match the server's constants/plan.go. */
export type AssumptionType = 'shift_project' | 'add_absence'

/** One assumption as the server returns it, with the names of what it refers to. */
export interface Assumption {
  type: AssumptionType
  project_id?: string
  project?: string
  /** Positive is later, negative earlier. */
  weeks?: number
  person_id?: string
  person?: string
  from_date?: string
  to_date?: string
  hours_per_day?: number
}

export interface ScenarioSummary {
  people_changed: number
  /** Hours placed in a week beyond what that week had in the baseline. */
  hours_moved: number
  unmet_hours: number
}

/** A what-if in the list (`GET /scenarios`). */
export interface Scenario {
  id: string
  name: string
  assumptions: Assumption[]
  base_version: number
  current_version: number
  /** The baseline moved on since the scenario was computed. */
  stale: boolean
  summary: ScenarioSummary
  created_at: string
  updated_at: string
}

export interface ScenarioProject {
  id: string
  name: string
  program_id: string
}

/** One person on one demand, before and after the scenario. Weeks are '' when there are none. */
export interface ScenarioChange {
  person: PersonRef
  demand_id: string
  project_id: string
  project: string
  expertise: string
  before_hours: number
  after_hours: number
  moved_hours: number
  before_from: string
  before_to: string
  after_from: string
  after_to: string
  change: PlanChangeKind
}

export interface ScenarioUnmet {
  demand_id: string
  project_id: string
  project: string
  expertise: string
  hours: number
  reason: UnmetReason
}

/** One scenario with its result (`GET /scenarios/:id`). */
export interface ScenarioDetail extends Scenario {
  /** The projects planned again. */
  projects: ScenarioProject[]
  changes: ScenarioChange[]
  unmet: ScenarioUnmet[]
}

/** One assumption to send: shift_project needs project_id and weeks; add_absence the person, dates and hours. */
export interface AssumptionInput {
  type: AssumptionType
  project_id?: string
  weeks?: number
  person_id?: string
  from_date?: string
  to_date?: string
  hours_per_day?: number
}

export interface ScenarioInput {
  name: string
  assumptions: AssumptionInput[]
}
