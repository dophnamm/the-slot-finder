import { ChevronLeft, ChevronRight } from 'lucide-react'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { addMonths, currentMonth } from '@/utils/functions'
import type { MonthNavProps } from './types'

/** Previous month, back to this month, next month. */
export function MonthNav({ month, onChange }: MonthNavProps) {
  const { t } = useTranslation()

  const handlePrevious = () => onChange(addMonths(month, -1))
  const handleToday = () => onChange(currentMonth())
  const handleNext = () => onChange(addMonths(month, 1))

  return (
    <div className="flex items-center gap-2">
      <Button variant="outline" size="icon-sm" aria-label={t('Previous month')} onClick={handlePrevious}>
        <ChevronLeft />
      </Button>
      <Button variant="outline" size="sm" onClick={handleToday}>
        {t('Today')}
      </Button>
      <Button variant="outline" size="icon-sm" aria-label={t('Next month')} onClick={handleNext}>
        <ChevronRight />
      </Button>
    </div>
  )
}
