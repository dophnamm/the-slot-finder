import type { ChangeEvent } from 'react'
import { useExpertises } from '@/shared/expertises'
import { useTranslation } from '@/shared/i18n'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { NativeSelect, NativeSelectOption } from '@/shared/ui/native-select'
import { isIsoWeek } from '@/utils/functions'
import { WEEK_COUNT_OPTIONS } from '../../constants'
import { useHeatmapFilters } from '../../use-heatmap-filters'

/** First week, number of weeks and expertise of the heatmap. */
export function HeatmapFilters() {
  const { t } = useTranslation()
  const { data: expertises = [] } = useExpertises()
  const { from, weeks, expertiseId, setFrom, setWeeks, setExpertiseId } = useHeatmapFilters()

  const handleFromChange = (event: ChangeEvent<HTMLInputElement>) => {
    if (isIsoWeek(event.target.value)) setFrom(event.target.value)
  }
  const handleWeeksChange = (event: ChangeEvent<HTMLSelectElement>) => setWeeks(Number(event.target.value))
  const handleExpertiseChange = (event: ChangeEvent<HTMLSelectElement>) => setExpertiseId(event.target.value)

  return (
    <div className="flex flex-wrap items-end gap-4">
      <div className="flex flex-col gap-2">
        <Label htmlFor="heatmap-from">{t('From week')}</Label>
        <Input id="heatmap-from" type="week" className="w-44" value={from} onChange={handleFromChange} />
      </div>
      <div className="flex flex-col gap-2">
        <Label htmlFor="heatmap-weeks">{t('Weeks')}</Label>
        <NativeSelect id="heatmap-weeks" value={String(weeks)} onChange={handleWeeksChange}>
          {WEEK_COUNT_OPTIONS.map((count) => (
            <NativeSelectOption key={count} value={String(count)}>
              {t('{count} weeks', { count })}
            </NativeSelectOption>
          ))}
        </NativeSelect>
      </div>
      <div className="flex flex-col gap-2">
        <Label htmlFor="heatmap-expertise">{t('Expertise')}</Label>
        <NativeSelect id="heatmap-expertise" value={expertiseId} onChange={handleExpertiseChange}>
          <NativeSelectOption value="">{t('All expertises')}</NativeSelectOption>
          {expertises.map((expertise) => (
            <NativeSelectOption key={expertise.id} value={expertise.id}>
              {expertise.name}
            </NativeSelectOption>
          ))}
        </NativeSelect>
      </div>
    </div>
  )
}
