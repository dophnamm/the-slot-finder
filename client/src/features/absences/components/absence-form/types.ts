import type { AbsenceStatus, AbsenceType } from '../../types'

export interface AbsenceFormProps {
  personId: string
  canApprove: boolean
  /** Called after the absence is saved, e.g. to close the dialog. */
  onCreated?: () => void
}

/** Form fields as the inputs hold them. */
export interface AbsenceFormValues {
  from_date: string
  to_date: string
  hours_per_day: string
  type: AbsenceType
  status: AbsenceStatus
}

/** Text inputs of the form, matched by their name attribute. */
export type AbsenceTextField = 'from_date' | 'to_date' | 'hours_per_day'
