import type { PlainKey } from '@/shared/i18n'
import type { DashboardView } from './types'

export const DEFAULT_DASHBOARD_VIEW: DashboardView = 'person'

/** The heatmap card's views, in toggle order. */
export const DASHBOARD_VIEW_LABELS: Record<DashboardView, PlainKey> = {
  person: 'By person',
  expertise: 'By expertise',
}
