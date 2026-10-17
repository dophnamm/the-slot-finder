import type { PlainKey } from '@/shared/i18n'
import type { BadgeVariant } from '@/shared/ui/badge'
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

export const EVENT_STATUS_BADGE: Record<ChangeEventStatus, BadgeVariant> = {
  pending: 'amber',
  no_impact: 'gray',
  draft_created: 'blue',
  failed: 'red',
}

/** The capacity fields a PersonCapacityChanged event can list. */
export const CAPACITY_FIELD_LABELS: Record<string, PlainKey> = {
  weekly_hours: 'weekly hours',
  non_project_hours: 'non-project hours',
  start_date: 'start date',
  end_date: 'end date',
  active: 'active status',
}

/** The `change` values of a TeamChanged event. */
export const TEAM_CHANGE_ADDED = 'added'
