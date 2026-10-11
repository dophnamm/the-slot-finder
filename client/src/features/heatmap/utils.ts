import { toSearchText } from '@/utils/functions'
import type { ExpertiseGroup, HeatmapRow } from './types'

/** Rows whose person name or primary expertise contains the search text, ignoring case and accents. */
export function filterRows(rows: HeatmapRow[], search: string): HeatmapRow[] {
  const query = toSearchText(search)
  if (!query) return rows
  return rows.filter((row) => toSearchText(`${row.person.name} ${row.primary_expertise}`).includes(query))
}

/** Consecutive rows with the same primary expertise, as the server orders them. */
export function groupByExpertise(rows: HeatmapRow[]): ExpertiseGroup[] {
  const groups: ExpertiseGroup[] = []
  for (const row of rows) {
    const last = groups.at(-1)
    if (last && last.expertise === row.primary_expertise) last.rows.push(row)
    else groups.push({ expertise: row.primary_expertise, rows: [row] })
  }
  return groups
}

/** A utilization ratio as a whole percentage, e.g. 0.875 → 88. */
export function percent(utilization: number): number {
  return Math.round(utilization * 100)
}
