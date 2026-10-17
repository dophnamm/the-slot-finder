import type { ReactNode } from 'react'
import type { UtilizationBand } from '@/shared/utilization'

/** One person of an expertise in one week; hours, with the heatmap's utilization and band. */
export interface ForecastPerson {
  person_id: string
  name: string
  capacity: number
  load: number
  free_hours: number
  /** Load over capacity; null when capacity is 0. */
  utilization: number | null
  band: UtilizationBand
  overloaded: boolean
}

/** One expertise in one week; all figures are hours. */
export interface ForecastWeek {
  week: string
  capacity: number
  load: number
  /** Capacity − load per person, never below 0. */
  free_hours: number
  /** Confirmed demand the baseline does not cover. */
  unmet_hours: number
  /** Demand of incoming projects × the pipeline confidence. */
  pipeline_hours: number
  /** Unmet plus pipeline hours that the free capacity cannot take. */
  shortfall_hours: number
  overloaded: boolean
  /** The expertise's people in the week, highest load first. */
  people: ForecastPerson[]
}

export interface OverloadRun {
  from_week: string
  weeks: number
}

export interface ExpertiseForecast {
  expertise_id: string
  expertise: string
  weeks: ForecastWeek[]
  /** The first run of overloaded weeks long enough to warn about; null when there is none. */
  sustained_overload: OverloadRun | null
}

export interface CapacityForecastData {
  weeks: string[]
  expertises: ExpertiseForecast[]
}

export interface CapacityForecastQuery {
  from: string
  to: string
}

/** How a cell looks: overloaded, short of free hours for the open demand, or fine. */
export type ForecastTone = 'overloaded' | 'short' | 'ok'

export type ForecastToneBands = Record<ForecastTone, UtilizationBand>

export interface CapacityForecastProps {
  /** Shown on the right of the card heading, e.g. a view switch. */
  action?: ReactNode
}
