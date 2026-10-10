import type { PlainKey } from '@/shared/i18n'
import type { AbsenceStatus, AbsenceType } from './types'

export const absencesQueryKey = (personId: string) => ['absences', personId] as const

/** Types a person's own absence may have (holidays come from the holiday list). */
export const PERSONAL_ABSENCE_TYPES = ['leave', 'training'] as const satisfies readonly AbsenceType[]

export const ABSENCE_STATUSES = ['planned', 'approved'] as const satisfies readonly AbsenceStatus[]

export const ABSENCE_TYPE_LABELS: Record<AbsenceType, PlainKey> = {
  leave: 'Leave',
  training: 'Training',
  holiday: 'Holiday',
}

export const ABSENCE_STATUS_LABELS: Record<AbsenceStatus, PlainKey> = {
  planned: 'Planned',
  approved: 'Approved',
}
