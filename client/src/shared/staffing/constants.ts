import type { PlainKey } from '@/shared/i18n'
import type { ProjectStatus } from '@/shared/programs'
import type { AssignmentKind, PlanChangeKind, PlanStatus, UnmetReason } from './types'

export const draftQueryKey = (id: string) => ['staffing-drafts', id] as const

/** The server's code when the baseline moved on since the draft was made. */
export const PLAN_VERSION_CONFLICT = 'PLAN_VERSION_CONFLICT'

/** The server's code when a change produced a newer version of a re-plan draft. */
export const DRAFT_SUPERSEDED = 'DRAFT_SUPERSEDED'

/** The status of a draft that can still be changed, accepted or rejected. */
export const OPEN_DRAFT_STATUS: PlanStatus = 'draft'

/** The status of an accepted draft: it is (or was) the baseline. */
export const PUBLISHED_PLAN_STATUS: PlanStatus = 'published'

/** The kind of a below-level or skill-less assignment a Planner added; it stays locked. */
export const STRETCH_KIND: AssignmentKind = 'stretch'

/** Project statuses a staffing draft may cover; they match the server's rule. */
export const PLANNABLE_STATUSES: readonly ProjectStatus[] = ['incoming', 'planned', 'active']

export const PLAN_STATUS_LABELS: Record<PlanStatus, PlainKey> = {
  draft: 'Draft',
  published: 'Accepted',
  rejected: 'Rejected',
  superseded: 'Superseded',
}

export const ASSIGNMENT_KIND_LABELS: Record<AssignmentKind, PlainKey> = {
  primary: 'Primary',
  cross_expertise: 'Cross-expertise',
  stretch: 'Stretch',
}

export const UNMET_REASON_LABELS: Record<UnmetReason, PlainKey> = {
  NO_SKILL: 'Nobody has this expertise',
  LEVEL_TOO_LOW: 'Nobody has the minimum level',
  NO_CAPACITY_BEFORE_DEADLINE: 'No free capacity in the window',
  BLOCKED_BY_LOCKED: 'The free hours are held by locked allocations',
}

export const PLAN_CHANGE_LABELS: Record<PlanChangeKind, PlainKey> = {
  added: 'Added',
  removed: 'Removed',
  changed: 'Changed',
}
