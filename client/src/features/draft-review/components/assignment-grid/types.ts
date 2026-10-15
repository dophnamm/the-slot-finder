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

/** A person's capacity in a week and the hours they carry there apart from one assignment. */
export interface WeekLoad {
  capacity: number
  rest: number
}

/** The hours being dragged: one week of one assignment. */
export interface DragSource {
  assignment: Assignment
  week: string
}
