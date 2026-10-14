import type { Demand } from '@/shared/programs'

export interface DemandViewRowProps {
  index: number
  demand: Demand
  /** Name of the demand's phase, or null for the whole project. */
  phase: string | null
  /** Set for people who may edit the project: the row then has edit and remove buttons. */
  onEdit?: (index: number) => void
  onRemove?: (index: number) => void
}
