import { useTranslation } from '@/shared/i18n'
import type { WorkedHoursProps } from './types'

/** "Worked N" next to a week's planned hours; a dash until a timesheet import recorded actual hours. */
export function WorkedHours({ hours }: WorkedHoursProps) {
  const { t } = useTranslation()

  return (
    <span className="text-xs text-muted-foreground tabular-nums">
      {t('Worked')} <strong className="font-semibold text-foreground">{hours ?? t('—')}</strong>
    </span>
  )
}
