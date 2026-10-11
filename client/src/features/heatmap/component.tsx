import { Section } from '@/shared/components/section'
import { useTranslation } from '@/shared/i18n'
import { addIsoWeeks } from '@/utils/functions'
import { HeatmapFilters } from './components/heatmap-filters'
import { HeatmapGrid } from './components/heatmap-grid'
import { HeatmapLegend } from './components/heatmap-legend'
import { useHeatmap } from './use-heatmap'
import { useHeatmapFilters } from './use-heatmap-filters'

/** Who is overloaded or underused in the weeks ahead, from the baseline. */
export function Heatmap() {
  const { t } = useTranslation()
  const from = useHeatmapFilters((state) => state.from)
  const weeks = useHeatmapFilters((state) => state.weeks)
  const expertiseId = useHeatmapFilters((state) => state.expertiseId)
  const { data, isPending } = useHeatmap({ from, to: addIsoWeeks(from, weeks - 1), expertise: expertiseId })

  return (
    <Section
      title={t('Team heatmap')}
      description={t('Baseline load against capacity per person and ISO week. Click a cell to see its assignments.')}
      className="h-full min-h-0"
    >
      <HeatmapFilters />
      {data && data.rows.length > 0 && <HeatmapGrid data={data} expertiseId={expertiseId} />}
      {isPending && <p className="py-6 text-center text-sm text-muted-foreground">{t('Loading…')}</p>}
      {data && data.rows.length === 0 && (
        <p className="py-6 text-center text-sm text-muted-foreground">{t('Nobody matches these filters.')}</p>
      )}
      <HeatmapLegend />
    </Section>
  )
}
