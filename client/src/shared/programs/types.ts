import type { AssignmentKind } from '@/shared/staffing'

/** Statuses and actions; they match the server's constants/program.go. */
export type ProgramStatus = 'open' | 'pending' | 'closing' | 'closed' | 'cancelled'
export type ProgramAction = 'hold' | 'resume' | 'start-closing' | 'close' | 'cancel'
export type ProjectStatus = 'incoming' | 'planned' | 'active' | 'on_hold' | 'done' | 'cancelled'
export type ProjectAction = 'hold' | 'resume' | 'complete' | 'cancel'

export interface PersonRef {
  id: string
  name: string
}

/** A program as the server returns it; dates are `2026-10-12`, weeks `2026-W42`. */
export interface Program {
  id: string
  name: string
  client: string
  status: ProgramStatus
  start_date: string
  end_date: string
  start_week: string
  end_week: string
  goal: string
  objectives: string[]
  charter_url: string | null
  lead: PersonRef | null
}

export interface Phase {
  id: string
  name: string
  sequence: number
  start_week: string
  end_week: string
}

export interface Demand {
  id: string
  expertise_id: string
  expertise: string
  phase_id: string | null
  min_level: number
  headcount: number
  total_hours: number
  profile: string
  /** Weeks the demand may be staffed in: its phase, or the project's start through its deadline. */
  from_week: string
  to_week: string
}

/**
 * Time, Scope, Planned burn and Staffing with their warnings, computed by the server.
 * Percentages are 0–100 and null when their denominator is 0; Time % may pass 100.
 */
export interface ProgressMetrics {
  time_percent: number | null
  scope_percent: number | null
  staffing_percent: number | null
  planned_burn_percent: number | null
  budget_hours: number
  planned_hours: number
  forecast_hours: number
  gap_hours: number
  planned_to_date_hours: number
  behind_on_delivery: boolean
  understaffed: boolean
  at_risk: boolean
  over_planned: boolean
  past_end: boolean
}

/** Budget against planned hours for one expertise of a program. */
export interface ExpertiseStaffing {
  expertise_id: string
  expertise: string
  budget_hours: number
  planned_hours: number
  gap_hours: number
  staffing_percent: number | null
  understaffed: boolean
}

export interface ProgramOverview {
  program_id: string
  /** How many projects (not cancelled) the figures sum over. */
  projects: number
  metrics: ProgressMetrics
  staffing_by_expertise: ExpertiseStaffing[]
}

export interface Project {
  id: string
  program_id: string
  program_name?: string
  name: string
  priority: number
  status: ProjectStatus
  start_earliest: string
  deadline: string
  start_week: string
  deadline_week: string
  scope_points: number
  done_points: number
  lead: PersonRef | null
  created_at: string
  /** The last phase ends after the deadline week. */
  at_risk: boolean
  current_phase: Phase | null
  /** Set on a program's project list. */
  metrics?: ProgressMetrics
}

export interface ProjectDetail extends Project {
  program_lead: PersonRef | null
  phases: Phase[]
  demands: Demand[]
}

export interface ProgramInput {
  name: string
  client: string
  start_date: string
  end_date: string
  goal: string
  objectives: string[]
  charter_url?: string | null
  lead_id?: string | null
}

export interface ProjectInput {
  name: string
  priority: number
  start_earliest: string
  deadline: string
  scope_points: number
  done_points: number
  lead_id?: string | null
}

export interface PhaseInput {
  id?: string
  name: string
  start_week: string
  end_week: string
}

export interface PhaseChanges {
  name?: string
  start_week?: string
  end_week?: string
}

export interface DemandInput {
  id?: string
  expertise_id: string
  phase_id: string | null
  min_level: number
  headcount: number
  total_hours: number
}

/** Someone who can lead a program or project. */
export interface LeadOption {
  id: string
  name: string
}

/** The pieces of the signed-in user that edit rights depend on. */
export interface EditorUser {
  role: string
  person_id: string | null
}

/** Names a project, e.g. in a member's list. */
export interface ProjectRef {
  id: string
  name: string
}

/** A member's hours in one expertise of the program, as one kind of assignment. */
export interface MemberExpertise {
  expertise_id: string
  expertise: string
  kind: AssignmentKind
  hours: number
}

/** A person with baseline hours on a program's projects (`GET /programs/:id/members`). */
export interface ProgramMember {
  person: PersonRef
  total_hours: number
  from_week: string
  to_week: string
  expertises: MemberExpertise[]
  projects: ProjectRef[]
}
