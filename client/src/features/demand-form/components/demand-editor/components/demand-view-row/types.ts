import type { Demand } from '@/shared/programs'

export interface DemandViewRowProps {
  demand: Demand
  /** Name of the demand's phase, or null for the whole project. */
  phase: string | null
}
