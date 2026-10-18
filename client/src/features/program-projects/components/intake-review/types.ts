import type { IntakeSuggestion } from '@/shared/ai'
import type { ProjectDetail } from '@/shared/programs'

export interface IntakeReviewProps {
  programId: string
  suggestion: IntakeSuggestion
  /** The project is created with its demand: close the dialog. */
  onClose: () => void
}

/** A demand row as the inputs hold it; `quote` is the brief's supporting text. */
export interface ReviewDemand {
  key: string
  expertise_id: string
  total_hours: string
  /** Empty means the default level. */
  min_level: string
  quote: string
}

/** The review form as the inputs hold it. */
export interface ReviewValues {
  name: string
  start_earliest: string
  deadline: string
  scope_points: string
  demands: ReviewDemand[]
}

export type ReviewTextField = 'name' | 'scope_points'
export type ReviewDateField = 'start_earliest' | 'deadline'

/** What creating the project returns: the project, and whether its demands were saved. */
export interface CreateResult {
  project: ProjectDetail
  demandsSaved: boolean
}
