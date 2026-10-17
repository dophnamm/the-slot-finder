import { SearchX } from 'lucide-react'
import { EmptyState } from '@/shared/components/empty-state'
import { LoadError } from '@/shared/components/load-error'
import { Section } from '@/shared/components/section'
import { TableLoading } from '@/shared/components/table-loading'
import { useDebouncedValue } from '@/shared/hooks'
import { useTranslation } from '@/shared/i18n'
import { Table, TableBody } from '@/shared/ui/table'
import { useWeekRange } from '@/shared/week-range'
import { SEARCH_DEBOUNCE_MS } from '@/utils/constants'
import { addIsoWeeks } from '@/utils/functions'
import { HeatmapFilters } from './components/heatmap-filters'
import { HeatmapGrid } from './components/heatmap-grid'
import { HeatmapLegend } from './components/heatmap-legend'
import { HEATMAP_LOADING_ROWS } from './constants'
import type { HeatmapProps } from './types'
import { useHeatmap } from './use-heatmap'
import { useHeatmapFilters } from './use-heatmap-filters'
import { filterRows } from './utils'

/** Who is overloaded or underused in the weeks ahead, from the baseline. */
export function Heatmap({ action }: HeatmapProps) {
  const { t } = useTranslation()
  const from = useWeekRange((state) => state.from)
  const weeks = useWeekRange((state) => state.weeks)
  const expertiseId = useHeatmapFilters((state) => state.expertiseId)
  const programId = useHeatmapFilters((state) => state.programId)
  const search = useHeatmapFilters((state) => state.search)
  const { data, error, isPending, isPlaceholderData, refetch } = useHeatmap({
    from,
    to: addIsoWeeks(from, weeks - 1),
    expertise: expertiseId,
    programId,
  })
  const query = useDebouncedValue(search, SEARCH_DEBOUNCE_MS)
  const rows = filterRows(data?.rows ?? [], query)

  const handleRetry = () => void refetch()

  return (
    <Section
      action={action}
      refreshing={isPlaceholderData}
      title={t('Team heatmap')}
      description={t('Baseline load against capacity per person and ISO week. Click a cell to see its assignments.')}
      className="h-full min-h-0"
    >
      <HeatmapFilters />
      {data && rows.length > 0 && <HeatmapGrid data={{ ...data, rows }} expertiseId={expertiseId} />}
      {isPending && (
        <Table>
          <TableBody>
            <TableLoading columns={weeks + 1} rows={HEATMAP_LOADING_ROWS} />
          </TableBody>
        </Table>
      )}
      {error && !data && <LoadError framed error={error} onRetry={handleRetry} />}
      {data && rows.length === 0 && (
        <EmptyState
          framed
          icon={SearchX}
          title={t('Nobody matches these filters.')}
          description={t('Try another name or expertise, or clear the search.')}
        />
      )}
      <HeatmapLegend />
    </Section>
  )
}
