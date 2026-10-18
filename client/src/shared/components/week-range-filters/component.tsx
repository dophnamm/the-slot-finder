import { ChevronLeft, ChevronRight } from 'lucide-react'
import { WeekPicker } from '@/shared/components/week-picker'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Label } from '@/shared/ui/label'
import { Select } from '@/shared/ui/select'
import { useWeekRange } from '@/shared/week-range'
import { WEEK_COUNT_OPTIONS } from '@/utils/constants'
import { addIsoWeeks } from '@/utils/functions'

/** First week and number of weeks, shared by the heatmap and the capacity forecast. */
export function WeekRangeFilters() {
  const { t } = useTranslation()
  const { from, weeks, setFrom, setWeeks } = useWeekRange()

  const handlePrevious = () => setFrom(addIsoWeeks(from, -weeks))
  const handleNext = () => setFrom(addIsoWeeks(from, weeks))
  const weekOptions = WEEK_COUNT_OPTIONS.map((count) => ({
    value: String(count),
    label: t('{count} weeks', { count }),
  }))
  const handleWeeksChange = (value: string) => setWeeks(Number(value))

  return (
    <>
      <div className="flex flex-col gap-2">
        <Label htmlFor="week-range-from">{t('From week')}</Label>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="icon" aria-label={t('Previous weeks')} onClick={handlePrevious}>
            <ChevronLeft />
          </Button>
          <WeekPicker id="week-range-from" name="from" value={from} onChange={setFrom} />
          <Button variant="outline" size="icon" aria-label={t('Next weeks')} onClick={handleNext}>
            <ChevronRight />
          </Button>
        </div>
      </div>
      <div className="flex flex-col gap-2">
        <Label htmlFor="week-range-weeks">{t('Weeks')}</Label>
        <Select id="week-range-weeks" options={weekOptions} value={String(weeks)} onValueChange={handleWeeksChange} />
      </div>
    </>
  )
}
