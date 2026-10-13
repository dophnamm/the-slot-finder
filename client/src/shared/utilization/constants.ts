import type { PlainKey } from '@/shared/i18n'
import type { UtilizationBand } from './types'

export const BAND_CLASSES: Record<UtilizationBand, string> = {
  unavailable: 'bg-muted text-muted-foreground',
  underused: 'bg-underused/20',
  healthy: 'bg-healthy/25',
  near_full: 'bg-warning/35',
  overloaded: 'bg-destructive/30',
}

export const BAND_LABELS: Record<UtilizationBand, PlainKey> = {
  unavailable: 'Unavailable',
  underused: 'Underused',
  healthy: 'Healthy',
  near_full: 'Near full',
  overloaded: 'Overloaded',
}

/** Legend order, from light to heavy load. */
export const BAND_ORDER: readonly UtilizationBand[] = ['underused', 'healthy', 'near_full', 'overloaded', 'unavailable']

/** Over 100 % of capacity; the calendar flags such a week. */
export const OVERLOADED_BAND: UtilizationBand = 'overloaded'
