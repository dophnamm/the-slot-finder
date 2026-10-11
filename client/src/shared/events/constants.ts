import type { PlainKey } from '@/shared/i18n'
import type { ChangeEventStatus } from './types'

export const EVENTS_QUERY_KEY = ['events'] as const

/** Changes are handled in the background, so their status is fetched again this often. */
export const EVENTS_REFETCH_MS = 15_000

/** The status of an open re-plan draft. */
export const OPEN_REPLAN_DRAFT_STATUS = 'draft'

export const EVENT_STATUS_LABELS: Record<ChangeEventStatus, PlainKey> = {
  pending: 'Waiting',
  no_impact: 'No impact',
  draft_created: 'Draft created',
  failed: 'Failed',
}

export const EVENT_STATUS_BADGE: Record<ChangeEventStatus, 'secondary' | 'outline' | 'destructive'> = {
  pending: 'outline',
  no_impact: 'secondary',
  draft_created: 'secondary',
  failed: 'destructive',
}
