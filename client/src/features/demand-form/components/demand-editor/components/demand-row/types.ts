import type { Expertise } from '@/shared/expertises'
import type { Phase } from '@/shared/programs'
import type { DemandDraft, WeekWindow } from '../../types'

export interface DemandRowProps {
  index: number
  draft: DemandDraft
  expertises: Expertise[]
  phases: Phase[]
  /** Weeks the row may be staffed in, from its phase or the whole project. */
  window: WeekWindow
  onChange: (index: number, draft: DemandDraft) => void
  onRemove: (index: number) => void
}

export type DemandTextField = 'expertise_id' | 'phase_id' | 'headcount' | 'total_hours'
