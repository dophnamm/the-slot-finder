import { useTranslation } from '@/shared/i18n'
import type { WorkedHoursProps } from './types'

/** "Worked N h" under a week's load, once a timesheet import recorded actual hours. */
export function WorkedHours({ hours }: WorkedHoursProps) {
  const { t } = useTranslation()
  if (hours === null) return null

  return <span className="text-xs text-muted-foreground tabular-nums">{t('Worked {hours} h', { hours })}</span>
}
