/** Utilization bands; they match the server's constants/insights.go. */
export type UtilizationBand = 'unavailable' | 'underused' | 'healthy' | 'near_full' | 'overloaded'

/** The limits between bands, as share of capacity; they come from the server's config. */
export interface UtilizationBands {
  underused_below: number
  near_full_from: number
}
