import type { Expertise } from '@/shared/expertises'
import type { ReviewDemand } from '../../types'

export type DemandRowField = 'expertise_id' | 'total_hours' | 'min_level'

export interface DemandRowProps {
  demand: ReviewDemand
  expertises: Expertise[]
  expertisesLoading: boolean
  onChange: (key: string, field: DemandRowField, value: string) => void
  onRemove: (key: string) => void
}
