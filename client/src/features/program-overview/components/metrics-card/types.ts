import type { ProgramOverview } from '@/shared/programs'

export interface MetricsCardProps {
  /** Undefined while it loads: the card keeps its shape with skeleton bars. */
  overview: ProgramOverview | undefined
}
