import type { PersonRef } from '@/shared/programs'
import type { UtilizationBand } from '@/shared/utilization'

export type { UtilizationBand }

/** The part of a week's load on one expertise. */
export interface ExpertiseHours {
  expertise_id: string
  expertise: string
  hours: number
  /** The person's primary skill. */
  primary: boolean
}

export interface HeatmapCell {
  week: string
  capacity: number
  load: number
  /** load ÷ capacity; null when capacity is 0. */
  utilization: number | null
  band: UtilizationBand
  /** Planned (not yet approved) absence this week. */
  tentative: boolean
  by_expertise: ExpertiseHours[]
}

export interface HeatmapRow {
  person: PersonRef
  primary_expertise_id: string
  primary_expertise: string
  cells: HeatmapCell[]
}

export interface HeatmapData {
  weeks: string[]
  rows: HeatmapRow[]
}

export interface HeatmapQuery {
  from: string
  to: string
  /** Expertise id; empty for everyone. */
  expertise: string
}

/** Rows that share a primary expertise; the server already orders them. */
export interface ExpertiseGroup {
  expertise: string
  rows: HeatmapRow[]
}

/** Heatmap filters: UI state kept while moving between pages. */
export interface HeatmapFiltersState {
  from: string
  weeks: number
  expertiseId: string
  /** Name or expertise text; rows are filtered in the browser. */
  search: string
  setFrom: (from: string) => void
  setWeeks: (weeks: number) => void
  setExpertiseId: (expertiseId: string) => void
  setSearch: (search: string) => void
}
