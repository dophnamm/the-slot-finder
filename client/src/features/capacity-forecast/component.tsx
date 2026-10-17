import { SearchX } from 'lucide-react'
import { EmptyState } from '@/shared/components/empty-state'
import { LoadError } from '@/shared/components/load-error'
import { Section } from '@/shared/components/section'
import { TableLoading } from '@/shared/components/table-loading'
import { WeekRangeFilters } from '@/shared/components/week-range-filters'
import { useTranslation } from '@/shared/i18n'
import { Table, TableBody } from '@/shared/ui/table'
import { useWeekRange } from '@/shared/week-range'
import { addIsoWeeks } from '@/utils/functions'
import { ForecastGrid } from './components/forecast-grid'
import { ForecastLegend } from './components/forecast-legend'
import { CAPACITY_FORECAST_LOADING_ROWS } from './constants'
import type { CapacityForecastProps } from './types'
import { useCapacityForecast } from './use-capacity-forecast'

/** Free hours and unmet demand per expertise and week, to see a crunch before it comes. */
export function CapacityForecast({ action }: CapacityForecastProps) {
  const { t } = useTranslation()
  const from = useWeekRange((state) => state.from)
  const weeks = useWeekRange((state) => state.weeks)
  const { data, error, isPending, isPlaceholderData, refetch } = useCapacityForecast({
    from,
    to: addIsoWeeks(from, weeks - 1),
  })

  const handleRetry = () => void refetch()

  return (
    <Section
      action={action}
      refreshing={isPlaceholderData}
      title={t('Capacity forecast')}
      description={t(
        'Free hours per expertise and week, against the demand the baseline does not cover yet, including incoming projects at their expected share.',
      )}
      className="h-full min-h-0"
    >
      <div className="flex flex-wrap items-end gap-4">
        <WeekRangeFilters />
      </div>
      {data && data.expertises.length > 0 && <ForecastGrid data={data} />}
      {isPending && (
        <Table>
          <TableBody>
            <TableLoading columns={weeks + 1} rows={CAPACITY_FORECAST_LOADING_ROWS} />
          </TableBody>
        </Table>
      )}
      {error && !data && <LoadError framed error={error} onRetry={handleRetry} />}
      {data && data.expertises.length === 0 && (
        <EmptyState
          framed
          icon={SearchX}
          title={t('No expertise to forecast.')}
          description={t('Add people with a primary expertise to see their capacity.')}
        />
      )}
      <ForecastLegend />
    </Section>
  )
}
