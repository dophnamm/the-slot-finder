import type { PlainKey } from '@/shared/i18n'
import type { BadgeVariant } from '@/shared/ui/badge'
import { INSIGHTS_QUERY_KEY } from '@/utils/constants'
import type { AbsenceSource, AbsenceStatus, AbsenceType } from './types'

export const PEOPLE_QUERY_KEY = ['people'] as const

/** One person, under the people prefix, so saving a person refreshes it. */
export const personQueryKey = (id: string) => [...PEOPLE_QUERY_KEY, 'detail', id] as const

/** A person's schedule, under the insights prefix, so accepting a staffing draft refreshes it. */
export const scheduleQueryKey = (personId: string, from: string, to: string) =>
  [...INSIGHTS_QUERY_KEY, 'schedule', personId, from, to] as const

export const ABSENCE_TYPE_LABELS: Record<AbsenceType, PlainKey> = {
  leave: 'Leave',
  training: 'Training',
  holiday: 'Holiday',
}

export const ABSENCE_STATUS_LABELS: Record<AbsenceStatus, PlainKey> = {
  planned: 'Planned',
  approved: 'Approved',
}

/** Reported or read from a calendar, not approved yet; a planner may approve it. */
export const PLANNED_ABSENCE_STATUS: AbsenceStatus = 'planned'

/** The source that gets a mark; it matches the server's constants/absence.go. */
export const CALENDAR_SOURCE = 'calendar' satisfies AbsenceSource

export const ABSENCE_SOURCE_BADGE: Record<AbsenceSource, BadgeVariant> = {
  manual: 'gray',
  calendar: 'teal',
}
