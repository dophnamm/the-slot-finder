import { Term } from '@/shared/components/term'
import { useTranslation } from '@/shared/i18n'
import { cn } from '@/utils/functions'
import { BAND_CLASSES, BAND_LABELS, BAND_ORDER, WORKED_CELL_CLASS } from '../../constants'

/** What each cell colour and the stripes mean. */
export function HeatmapLegend() {
  const { t } = useTranslation()

  return (
    <div className="flex flex-wrap gap-4 text-xs text-muted-foreground">
      <span className="flex items-center gap-1.5 border-r pr-4 font-medium text-foreground">
        <Term term="utilization">{t('Utilization')}</Term>
      </span>
      {BAND_ORDER.map((band) => (
        <span key={band} className="flex items-center gap-1.5">
          <span className={cn('size-3 rounded-sm', BAND_CLASSES[band])} />
          {t(BAND_LABELS[band])}
        </span>
      ))}
      <span className="flex items-center gap-1.5">
        <span className="size-3 rounded-sm border bg-hatched" />
        {t('Planned leave, not approved yet')}
      </span>
      <span className="flex items-center gap-1.5">
        <span className={cn('size-3 rounded-sm bg-muted', WORKED_CELL_CLASS)} />
        {t('Past weeks show worked hours when a timesheet was imported.')}
      </span>
    </div>
  )
}
