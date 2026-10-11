import type { PlainKey } from '@/shared/i18n'
import type { UtilizationBand } from './types'

/** Weeks shown until the user picks another range. */
export const DEFAULT_WEEK_COUNT = 12

/** Range lengths the filter offers; the server caps the range with HEATMAP_MAX_WEEKS. */
export const WEEK_COUNT_OPTIONS = [4, 8, 12, 26] as const

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

export const HEATMAP_QUERY_NAME = 'heatmap'
export const PERSON_WEEK_QUERY_NAME = 'person-week'
