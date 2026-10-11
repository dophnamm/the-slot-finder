import { MetricBar } from '@/shared/components/metric-bar'
import { useTranslation } from '@/shared/i18n'
import { activeWarnings } from '@/shared/programs'
import { Badge } from '@/shared/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/ui/card'
import { Figure } from './components/figure'
import type { MetricsCardProps } from './types'

/** Time, Scope, Planned burn and Staffing with their warnings, and the cost figures. */
export function MetricsCard({ overview }: MetricsCardProps) {
  const { t } = useTranslation()
  const { metrics } = overview
  const warnings = activeWarnings(metrics)

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('Progress')}</CardTitle>
        <CardDescription>
          {overview.projects === 1
            ? t('Summed over {count} project that is not cancelled.', { count: overview.projects })
            : t('Summed over {count} projects that are not cancelled.', { count: overview.projects })}
        </CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-5">
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <MetricBar label={t('Time')} value={metrics.time_percent} warning={metrics.past_end} />
          <MetricBar label={t('Scope')} value={metrics.scope_percent} warning={metrics.behind_on_delivery} />
          <MetricBar label={t('Planned burn')} value={metrics.planned_burn_percent} />
          <MetricBar label={t('Staffing')} value={metrics.staffing_percent} warning={metrics.understaffed} />
        </div>
        {warnings.length > 0 && (
          <div className="flex flex-wrap gap-2">
            {warnings.map((warning) => (
              <Badge key={warning.flag} variant="destructive">
                {t(warning.label)}
              </Badge>
            ))}
          </div>
        )}
        <dl className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
          <Figure label={t('Budget')} hours={metrics.budget_hours} />
          <Figure label={t('Planned')} hours={metrics.planned_hours} warning={metrics.over_planned} />
          <Figure label={t('Planned to date')} hours={metrics.planned_to_date_hours} />
          <Figure label={t('Forecast')} hours={metrics.forecast_hours} />
          <Figure label={t('Gap')} hours={metrics.gap_hours} warning={metrics.understaffed} />
        </dl>
      </CardContent>
    </Card>
  )
}
