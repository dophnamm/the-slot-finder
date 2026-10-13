import { ABSENCES_QUERY_KEY } from '@/utils/constants'
import type { AbsenceStatus, AbsenceType } from './types'

export { ABSENCE_STATUS_LABELS, ABSENCE_TYPE_LABELS } from '@/shared/people'

export const absencesQueryKey = (personId: string) => [...ABSENCES_QUERY_KEY, personId] as const

/** Types a person's own absence may have (holidays come from the holiday list). */
export const PERSONAL_ABSENCE_TYPES = ['leave', 'training'] as const satisfies readonly AbsenceType[]

export const ABSENCE_STATUSES = ['planned', 'approved'] as const satisfies readonly AbsenceStatus[]

/** Columns of the table; the loading and empty rows span them all. */
export const ABSENCE_TABLE_COLUMNS = 5
