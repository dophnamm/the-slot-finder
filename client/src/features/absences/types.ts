/** Absence types and statuses; they match the server's constants/absence.go. */
export type AbsenceType = 'leave' | 'training' | 'holiday'
export type AbsenceStatus = 'planned' | 'approved'

/** An absence as the server returns it; dates are inclusive, e.g. `2026-10-19`. */
export interface Absence {
  id: string
  person_id: string | null
  from_date: string
  to_date: string
  hours_per_day: number
  type: AbsenceType
  status: AbsenceStatus
}

export interface CreateAbsenceInput {
  from_date: string
  to_date: string
  hours_per_day: number
  type: AbsenceType
  /** Honoured for planners; others always report planned. */
  status?: AbsenceStatus
}

export interface AbsenceListProps {
  personId: string
  /** Planners may choose the status; others report planned absence. */
  canApprove: boolean
}
