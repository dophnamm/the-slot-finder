import { LEGEND } from './constants'
import type { LegendItem, TimelineRowData } from './types'

/** The legend entries for the states the rows show, in legend order: the page has no unsaved moves. */
export function legendItems(rows: TimelineRowData[]): LegendItem[] {
  return LEGEND.filter((item) => rows.some((row) => row.cells.includes(item.state)))
}
