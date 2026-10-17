import { useTranslation } from '@/shared/i18n'
import { Tabs, TabsList, TabsTrigger } from '@/shared/ui/tabs'
import { DASHBOARD_VIEW_LABELS } from '../../constants'
import type { DashboardView } from '../../types'
import type { ViewToggleProps } from './types'

/** Switches the heatmap card between people and expertises. */
export function ViewToggle({ value, onChange }: ViewToggleProps) {
  const { t } = useTranslation()

  const handleChange = (next: unknown) => onChange(next as DashboardView)

  return (
    <Tabs value={value} onValueChange={handleChange}>
      <TabsList aria-label={t('Dashboard view')}>
        {(Object.keys(DASHBOARD_VIEW_LABELS) as DashboardView[]).map((view) => (
          <TabsTrigger key={view} value={view}>
            {t(DASHBOARD_VIEW_LABELS[view])}
          </TabsTrigger>
        ))}
      </TabsList>
    </Tabs>
  )
}
