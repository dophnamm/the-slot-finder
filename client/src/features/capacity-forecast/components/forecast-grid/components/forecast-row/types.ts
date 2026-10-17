import type { ExpertiseForecast } from '../../../../types'

export interface ForecastRowProps {
  expertise: ExpertiseForecast
  /** This week's label, to outline its column. */
  currentWeek: string
}
