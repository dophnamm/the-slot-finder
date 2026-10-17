import type { ReactNode } from 'react'
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
  /** Hours worked that week from imported timesheets; null when nothing was imported. */
  actual_hours: number | null
}

export interface HeatmapRow {
  person: PersonRef
  primary_expertise_id: string
  primary_expertise: string
  cells: HeatmapCell[]
  /** The first run of overloaded weeks long enough to warn about; null when there is none. */
  sustained_overload: OverloadRun | null
}

export interface OverloadRun {
  from_week: string
  weeks: number
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
  /** Program id: only the people of its team; empty for every program. */
  programId: string
}

/** Rows that share a primary expertise; the server already orders them. */
export interface ExpertiseGroup {
  expertise: string
  rows: HeatmapRow[]
}

export interface HeatmapProps {
  /** Shown on the right of the card heading, e.g. a view switch. */
  action?: ReactNode
}

/** Heatmap filters: UI state kept while moving between pages. */
export interface HeatmapFiltersState {
  expertiseId: string
  programId: string
  /** Name or expertise text; rows are filtered in the browser. */
  search: string
  setExpertiseId: (expertiseId: string) => void
  setProgramId: (programId: string) => void
  setSearch: (search: string) => void
}
