/** They match the server's constants/events.go. */
export type ChangeEventType = 'AbsenceAdded' | 'AbsenceRemoved' | 'PhaseSlipped' | 'ProjectSlipped'
export type ChangeEventStatus = 'pending' | 'no_impact' | 'draft_created' | 'failed'

/** Payload fields by type; a holiday has no person. */
export interface ChangeEventPayload {
  person_name?: string
  person_id?: string | null
  from_date?: string
  to_date?: string
  project?: string
  phase?: string
  new_end_week?: string
  weeks?: number
  old_deadline?: string
  new_deadline?: string
}

export interface ChangeEvent {
  id: string
  type: ChangeEventType
  payload: ChangeEventPayload
  occurred_at: string
  status: ChangeEventStatus
  attempts: number
  last_error: string | null
  project_ids: string[]
  /** The re-plan draft the change led to, at its latest version. */
  draft: { id: string; status: string } | null
}
