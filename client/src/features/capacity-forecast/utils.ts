import { HOURS_DECIMALS } from './constants'
import type { ForecastTone, ForecastWeek } from './types'

/** Overloaded beats short: the people already work above their capacity. */
export function forecastTone(cell: ForecastWeek): ForecastTone {
  if (cell.overloaded) return 'overloaded'
  return cell.shortfall_hours > 0 ? 'short' : 'ok'
}

/** Hours rounded for display, e.g. 7.5 stays 7.5 and 6.0000001 becomes 6. */
export function roundHours(hours: number): number {
  const factor = 10 ** HOURS_DECIMALS
  return Math.round(hours * factor) / factor
}

/** A utilization ratio as a whole percentage, e.g. 0.875 → 88. */
export function percent(utilization: number): number {
  return Math.round(utilization * 100)
}
