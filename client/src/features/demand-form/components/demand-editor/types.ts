import type { Expertise } from '@/shared/expertises'
import type { ProjectDetail } from '@/shared/programs'

export interface DemandEditorProps {
  project: ProjectDetail
  expertises: Expertise[]
  /** The expertise catalogue is still loading. */
  expertisesLoading: boolean
  editable: boolean
}

/** A demand row as the inputs hold it; `phase_id` is empty for the whole project. */
export interface DemandDraft {
  key: string
  id?: string
  expertise_id: string
  phase_id: string
  min_level: number
  headcount: string
  total_hours: string
}

/** The demand open in the dialog: `index` is null for a new one. */
export interface DemandEdit {
  index: number | null
  draft: DemandDraft
}

export interface WeekWindow {
  from: string
  to: string
}
