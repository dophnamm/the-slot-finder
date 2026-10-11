import type { CellState, LegendItem } from './types'

export const CELL_CLASSES: Record<CellState, string> = {
  empty: 'bg-muted',
  planned: 'bg-primary',
  moved: 'bg-warning',
  vacated: 'bg-primary/25',
}

/** Outlines every cell of the deadline week. */
export const DEADLINE_CELL_CLASS = 'ring-2 ring-inset ring-destructive'

/** Shared width of the phase-name column, so labels and cells line up. */
export const LABEL_COLUMN_CLASS = 'w-36 shrink-0 truncate pr-2'

/** Shared size of one week column. */
export const WEEK_COLUMN_CLASS = 'min-w-9 flex-1'

export const LEGEND: readonly LegendItem[] = [
  { state: 'planned', label: 'Planned' },
  { state: 'moved', label: 'Moved, not saved yet' },
  { state: 'vacated', label: 'Before the change' },
]
