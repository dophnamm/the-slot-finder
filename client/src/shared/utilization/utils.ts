import type { UtilizationBand, UtilizationBands } from './types'

/** Rounding noise allowed when comparing hours, as on the server. */
const HOURS_EPSILON = 1e-6

/** The band of load over capacity, as the server works it out: no capacity is unavailable (overloaded with load), over 100 % is overloaded. */
export function utilizationBand(load: number, capacity: number, bands: UtilizationBands): UtilizationBand {
  if (capacity <= HOURS_EPSILON) return load <= HOURS_EPSILON ? 'unavailable' : 'overloaded'
  const utilization = load / capacity
  if (utilization < bands.underused_below) return 'underused'
  if (utilization < bands.near_full_from) return 'healthy'
  return utilization <= 1 + HOURS_EPSILON ? 'near_full' : 'overloaded'
}
