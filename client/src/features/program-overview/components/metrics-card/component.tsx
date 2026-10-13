import { MetricBar } from '@/shared/components/metric-bar'
import { Section } from '@/shared/components/section'
import { useTranslation } from '@/shared/i18n'
import { activeWarnings, burnMetric } from '@/shared/programs'
import { Badge } from '@/shared/ui/badge'
import { Figure } from './components/figure'
import type { MetricsCardProps } from './types'

/** Time, Scope, Burn and Staffing with their warnings, and the cost figures. */
export function MetricsCard({ overview }: MetricsCardProps) {
  const { t } = useTranslation()
  const { metrics } = overview
  const warnings = activeWarnings(metrics)
  const burn = burnMetric(metrics)

  return (
    <Section
      card
      title={t('Progress')}
      description={
        overview.projects === 1
          ? t('Summed over {count} project that is not cancelled.', { count: overview.projects })
          : t('Summed over {count} projects that are not cancelled.', { count: overview.projects })
      }
      contentClassName="gap-5"
    >
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <MetricBar label={t('Time')} value={metrics.time_percent} warning={metrics.past_end} />
        <MetricBar label={t('Scope')} value={metrics.scope_percent} warning={metrics.behind_on_delivery} />
        <MetricBar label={t(burn.label)} value={burn.value} />
        <MetricBar label={t('Staffing')} value={metrics.staffing_percent} warning={metrics.understaffed} />
      </div>
      {warnings.length > 0 && (
        <div className="flex flex-wrap gap-2">
          {warnings.map((warning) => (
            <Badge key={warning.flag} variant={warning.badge}>
              {t(warning.label)}
            </Badge>
          ))}
        </div>
      )}
      <dl className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-6">
        <Figure label={t('Budget')} hours={metrics.budget_hours} />
        <Figure label={t('Planned')} hours={metrics.planned_hours} warning={metrics.over_planned} />
        <Figure label={t('Planned to date')} hours={metrics.planned_to_date_hours} />
        {metrics.actual_hours !== null && <Figure label={t('Actual')} hours={metrics.actual_hours} />}
        <Figure label={t('Forecast')} hours={metrics.forecast_hours} />
        <Figure label={t('Gap')} hours={metrics.gap_hours} warning={metrics.understaffed} />
      </dl>
    </Section>
  )
}
