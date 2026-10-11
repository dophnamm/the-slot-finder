import type { Assignment, Draft, DraftDemand } from '@/shared/staffing'

export interface AssignmentGridProps {
  draft: Draft
  editable: boolean
}

/** A covered demand with the assignments that cover it. */
export interface DemandGroup {
  demand: DraftDemand
  assignments: Assignment[]
}
