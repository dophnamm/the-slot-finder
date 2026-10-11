import type { CellState, LegendItem } from './types'

/** Pastel fills: planned weeks blue, weeks a change would move amber, the weeks they leave faint blue. */
export const CELL_CLASSES: Record<CellState, string> = {
  empty: 'bg-muted',
  planned: 'bg-underused/55',
  moved: 'bg-warning/60',
  vacated: 'bg-underused/20',
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
