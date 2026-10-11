import type { PlainKey } from '@/shared/i18n'

/**
 * How one week of a phase looks: inside the saved phase (planned), inside it after an
 * unsaved change (moved), left by an unsaved change (vacated), or outside it (empty).
 */
export type CellState = 'empty' | 'planned' | 'moved' | 'vacated'

export interface TimelineRowData {
  key: string
  name: string
  /** One state per week of the timeline, in order. */
  cells: CellState[]
}

export interface LegendItem {
  state: CellState
  label: PlainKey
}

export interface PhaseTimelineProps {
  weeks: string[]
  deadlineWeek: string
  rows: TimelineRowData[]
}
