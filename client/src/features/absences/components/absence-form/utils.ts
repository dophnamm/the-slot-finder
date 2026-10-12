import { ABSENCE_STATUSES, PERSONAL_ABSENCE_TYPES } from '../../constants'
import type { AbsenceStatus, AbsenceType, CreateAbsenceInput } from '../../types'
import type { AbsenceFormValues } from './types'

export function isAbsenceType(value: string): value is AbsenceType {
  return (PERSONAL_ABSENCE_TYPES as readonly string[]).includes(value)
}

export function isAbsenceStatus(value: string): value is AbsenceStatus {
  return (ABSENCE_STATUSES as readonly string[]).includes(value)
}

/** Both dates set, the end not before the start, and positive hours. */
export function canSubmit(values: AbsenceFormValues): boolean {
  return (
    values.from_date !== '' &&
    values.to_date !== '' &&
    values.to_date >= values.from_date &&
    Number(values.hours_per_day) > 0
  )
}

export function toInput(values: AbsenceFormValues, canApprove: boolean): CreateAbsenceInput {
  return {
    from_date: values.from_date,
    to_date: values.to_date,
    hours_per_day: Number(values.hours_per_day),
    type: values.type,
    status: canApprove ? values.status : undefined,
  }
}
