import { can, type Role } from '@/resources/rbac'
import type { EditorUser } from '@/shared/programs'
import { OPEN_DRAFT_STATUS } from './constants'
import type { Allocation, Assignment, Draft, DraftDemand, DraftEdit } from './types'

/** Planners staff any program; a manager staffs only when they manage every one of these programs. */
export function canStaff(user: EditorUser | undefined, programLeadIds: (string | null | undefined)[]): boolean {
  if (!user || !can(user.role as Role, 'staffing-drafts:manage')) return false
  if (can(user.role as Role, 'staffing-drafts:manage-all')) return true
  return user.person_id !== null && programLeadIds.length > 0 && programLeadIds.every((id) => id === user.person_id)
}

/** A draft can change while it is still a draft, by someone who may staff all its projects. */
export function canEditDraft(user: EditorUser | undefined, draft: Draft): boolean {
  return (
    draft.status === OPEN_DRAFT_STATUS &&
    canStaff(
      user,
      draft.projects.map((project) => project.program_lead_id),
    )
  )
}

/** Hours a demand still lacks in the draft, rounded to hundredths. */
export function shortfall(demand: DraftDemand): number {
  return Math.max(0, Math.round((demand.hours - demand.assigned_hours) * 100) / 100)
}

const roundHours = (hours: number) => Math.round(hours * 100) / 100

function withSetHours(assignment: Assignment, set: NonNullable<DraftEdit['set']>): Assignment {
  let allocations: Allocation[] = assignment.allocations
  for (const change of set.filter((entry) => entry.assignment_id === assignment.id)) {
    const current = allocations.find((alloc) => alloc.week === change.week)
    const rest = allocations.filter((alloc) => alloc.week !== change.week)
    allocations =
      change.hours > 0
        ? [...rest, { week: change.week, hours: change.hours, locked: change.locked ?? current?.locked ?? false }]
        : rest
    allocations = allocations.toSorted((a, b) => a.week.localeCompare(b.week))
  }
  return {
    ...assignment,
    allocations,
    total_hours: roundHours(allocations.reduce((sum, alloc) => sum + alloc.hours, 0)),
  }
}

/** The draft as it will be once the server applies these hour changes; used to show an edit before the answer arrives. */
export function applySetHours(draft: Draft, set: NonNullable<DraftEdit['set']>): Draft {
  const assignments = draft.assignments.map((assignment) => withSetHours(assignment, set))
  const demands = draft.demands.map((demand) => ({
    ...demand,
    assigned_hours: roundHours(
      assignments.filter((assignment) => assignment.demand.id === demand.id).reduce((sum, a) => sum + a.total_hours, 0),
    ),
  }))
  return { ...draft, assignments, demands }
}
