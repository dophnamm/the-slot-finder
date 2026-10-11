import type { ProgressMetrics } from '@/shared/programs'
import type { PreviewFigure } from './types'

/** Budget, planned and gap hours, with the gap marked when the program is understaffed. */
export function previewFigures(metrics: ProgressMetrics): PreviewFigure[] {
  return [
    { label: 'Budget', hours: Math.round(metrics.budget_hours), warning: false },
    { label: 'Planned', hours: Math.round(metrics.planned_hours), warning: metrics.over_planned },
    { label: 'Gap', hours: Math.round(metrics.gap_hours), warning: metrics.understaffed },
  ]
}
