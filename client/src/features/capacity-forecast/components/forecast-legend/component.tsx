import { useTranslation } from '@/shared/i18n'
import { BAND_CLASSES } from '@/shared/utilization'
import { cn } from '@/utils/functions'
import { FORECAST_TONE_BANDS, FORECAST_TONE_LABELS, FORECAST_TONE_ORDER } from '../../constants'

/** What each cell colour means. */
export function ForecastLegend() {
  const { t } = useTranslation()

  return (
    <div className="flex flex-wrap gap-4 text-xs text-muted-foreground">
      {FORECAST_TONE_ORDER.map((tone) => (
        <span key={tone} className="flex items-center gap-1.5">
          <span className={cn('size-3 rounded-sm', BAND_CLASSES[FORECAST_TONE_BANDS[tone]])} />
          {t(FORECAST_TONE_LABELS[tone])}
        </span>
      ))}
    </div>
  )
}
