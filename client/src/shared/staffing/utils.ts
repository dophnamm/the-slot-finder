import { can, type Role } from '@/resources/rbac'
import type { EditorUser } from '@/shared/programs'
import { OPEN_DRAFT_STATUS } from './constants'
import type { Draft, DraftDemand } from './types'

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
