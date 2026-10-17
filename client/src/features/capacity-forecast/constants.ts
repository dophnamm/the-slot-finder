import type { PlainKey } from '@/shared/i18n'
import type { ForecastTone, ForecastToneBands } from './types'

export const CAPACITY_FORECAST_QUERY_NAME = 'capacity-forecast'

/** Skeleton rows while the forecast loads: about one row per expertise. */
export const CAPACITY_FORECAST_LOADING_ROWS = 4

/** A cell takes the colour of the heatmap band it matches. */
export const FORECAST_TONE_BANDS: ForecastToneBands = {
  ok: 'healthy',
  short: 'near_full',
  overloaded: 'overloaded',
}

/** Legend order, from fine to worst. */
export const FORECAST_TONE_ORDER: ForecastTone[] = ['ok', 'short', 'overloaded']

export const FORECAST_TONE_LABELS: Record<ForecastTone, PlainKey> = {
  ok: 'Enough free hours',
  short: 'Short of free hours',
  overloaded: 'Overloaded',
}

/** Hours are shown with at most this many decimals: pipeline hours are weighted by a percentage. */
export const HOURS_DECIMALS = 1
