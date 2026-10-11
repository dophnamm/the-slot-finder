import type { Demand } from '@/shared/programs'

/** Changes whenever a saved demand changes, so the editor restarts from the saved list. */
export function demandsKey(demands: Demand[]): string {
  return demands.map((d) => [d.id, d.expertise_id, d.phase_id, d.min_level, d.headcount, d.total_hours].join(':')).join('|')
}
