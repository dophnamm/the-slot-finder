import type { TermName } from '@/shared/components/term'
import type { PlainKey } from '@/shared/i18n'
import type { ProjectStatus } from '@/shared/programs'
import type { BadgeVariant } from '@/shared/ui/badge'
import type { AssignmentKind, PlanChangeKind, PlanStatus, UnmetReason } from './types'

export const draftQueryKey = (id: string) => ['staffing-drafts', id] as const

export const alternativesQueryKey = (id: string, assignmentId: string) =>
  ['staffing-drafts', id, 'alternatives', assignmentId] as const

/** The drafts inbox; generating, accepting or rejecting a draft refetches it. */
export const OPEN_DRAFTS_QUERY_KEY = ['staffing-drafts', 'open'] as const

/** Re-plans add drafts in the background, so the inbox and its sidebar count are fetched again this often. */
export const OPEN_DRAFTS_REFETCH_MS = 15_000

/** The server's code when the baseline moved on since the draft was made. */
export const PLAN_VERSION_CONFLICT = 'PLAN_VERSION_CONFLICT'

/** The server's code when a change produced a newer version of a re-plan draft. */
export const DRAFT_SUPERSEDED = 'DRAFT_SUPERSEDED'

/** The status of a draft that can still be changed, accepted or rejected. */
export const OPEN_DRAFT_STATUS: PlanStatus = 'draft'

/** The status of an accepted draft: it is (or was) the baseline. */
export const PUBLISHED_PLAN_STATUS: PlanStatus = 'published'

/** The kind of an assignment on the person's primary expertise. */
export const PRIMARY_KIND: AssignmentKind = 'primary'

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

/** Badge colour per plan status: an open draft is blue, an accepted plan green, the rest gray. */
export const PLAN_STATUS_BADGE: Record<PlanStatus, BadgeVariant> = {
  draft: 'blue',
  published: 'green',
  rejected: 'gray',
  superseded: 'gray',
}

export const ASSIGNMENT_KIND_LABELS: Record<AssignmentKind, PlainKey> = {
  primary: 'Primary',
  cross_expertise: 'Cross-expertise',
  stretch: 'Stretch',
}

/** Kinds whose word has a definition to show on hover. */
export const ASSIGNMENT_KIND_TERMS: Partial<Record<AssignmentKind, TermName>> = {
  cross_expertise: 'crossExpertise',
  stretch: 'stretch',
}

/** Badge colour per assignment kind: primary blue, cross-expertise violet, stretch red. */
export const ASSIGNMENT_KIND_BADGE: Record<AssignmentKind, BadgeVariant> = {
  primary: 'blue',
  cross_expertise: 'violet',
  stretch: 'red',
}

export const UNMET_REASON_LABELS: Record<UnmetReason, PlainKey> = {
  NO_SKILL: 'Nobody has this expertise',
  LEVEL_TOO_LOW: 'Nobody has the minimum level',
  NO_CAPACITY_BEFORE_DEADLINE: 'No free capacity in the window',
  BLOCKED_BY_LOCKED: 'The free hours are held by locked allocations',
  NO_ONE_IN_TEAM: 'No one in the program team',
}

/** Badge colour per change: added green, changed amber, removed red. */
export const PLAN_CHANGE_BADGE: Record<PlanChangeKind, BadgeVariant> = {
  added: 'green',
  changed: 'amber',
  removed: 'red',
}

export const PLAN_CHANGE_LABELS: Record<PlanChangeKind, PlainKey> = {
  added: 'Added',
  removed: 'Removed',
  changed: 'Changed',
}
