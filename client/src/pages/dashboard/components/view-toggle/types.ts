import type { DashboardView } from '../../types'

export interface ViewToggleProps {
  value: DashboardView
  onChange: (view: DashboardView) => void
}
