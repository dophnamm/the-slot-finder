import { can, type Role } from '@/resources/rbac'
import { OPEN_DRAFT_STATUS } from './constants'
import type { Draft, DraftDemand } from './types'

/** Planners change a draft while it is still a draft. */
export function canEditDraft(role: string | undefined, draft: Draft): boolean {
  return can(role as Role, 'staffing-drafts:manage') && draft.status === OPEN_DRAFT_STATUS
}

/** Hours a demand still lacks in the draft, rounded to hundredths. */
export function shortfall(demand: DraftDemand): number {
  return Math.max(0, Math.round((demand.hours - demand.assigned_hours) * 100) / 100)
}
