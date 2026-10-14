import type { ChangeEvent } from '@/shared/events'
import type { PersonRef, ProjectStatus } from '@/shared/programs'

/** They match the server's constants/plan.go and the engine's codes. */
export type PlanStatus = 'draft' | 'published' | 'rejected' | 'superseded'
export type AssignmentKind = 'primary' | 'cross_expertise' | 'stretch'
export type UnmetReason =
  'NO_SKILL' | 'LEVEL_TOO_LOW' | 'NO_CAPACITY_BEFORE_DEADLINE' | 'BLOCKED_BY_LOCKED' | 'NO_ONE_IN_TEAM'
export type PlanChangeKind = 'added' | 'removed' | 'changed'

/** An assignment's hours in one ISO week. */
export interface Allocation {
  week: string
  hours: number
  locked: boolean
}

/** The demand an assignment covers, with its window. */
export interface AssignmentDemand {
  id: string
  expertise: string
  min_level: number
  headcount: number
  hours: number
  project_id: string
  project: string
  program_id: string
  program: string
  phase: string | null
  from_week: string
  to_week: string
}

export interface Assignment {
  id: string
  person: PersonRef
  demand: AssignmentDemand
  kind: AssignmentKind
  reason: string[]
  /** Added by a Planner rather than proposed by the engine. */
  by_hand: boolean
  total_hours: number
  allocations: Allocation[]
}

export interface DraftProject {
  id: string
  name: string
  program_id: string
  program: string
  /** The person who manages the program; they may act on a draft of its projects. */
  program_lead_id: string | null
  status: ProjectStatus
  priority: number
  deadline_week: string
  staffed_at: string | null
}

/** A covered demand with the hours the draft gives it. */
export interface DraftDemand {
  id: string
  project_id: string
  project: string
  expertise: string
  phase: string | null
  min_level: number
  headcount: number
  hours: number
  assigned_hours: number
  from_week: string
  to_week: string
  unmet_reason: UnmetReason | null
  /** The program team who may staff the demand; null puts no limit. */
  team_ids: string[] | null
}

/** How one person's hours on one demand change against the baseline. */
export interface PlanChange {
  person: PersonRef
  demand_id: string
  project: string
  expertise: string
  before_hours: number
  after_hours: number
  change: PlanChangeKind
}

export interface Draft {
  id: string
  status: PlanStatus
  /** The baseline version the draft was built on; sent back to accept it. */
  base_version: number
  current_version: number
  /** The baseline moved on since the draft was made, so accepting it fails. */
  stale: boolean
  created_at: string
  published_at: string | null
  weeks: string[]
  projects: DraftProject[]
  demands: DraftDemand[]
  assignments: Assignment[]
  changes: PlanChange[]
  has_stretch: boolean
  /** A re-plan draft that changes created; triggered_by lists them. */
  auto: boolean
  triggered_by: ChangeEvent[]
  /** The newer version of a re-plan draft that replaced this one. */
  superseded_by: string | null
}

export interface WeekHours {
  week: string
  hours: number
}

/** A batch of draft changes the server applies together. */
export interface DraftEdit {
  add?: { person_id: string; demand_id: string; allocations: WeekHours[] }[]
  remove?: string[]
  /** Hours 0 removes the week; leave locked out to keep the current lock. */
  set?: { assignment_id: string; week: string; hours: number; locked?: boolean }[]
}

export interface DraftAcceptance {
  base_version: number
  confirm_stretch: boolean
}
