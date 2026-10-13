import { CalendarSync } from 'lucide-react'
import { useTranslation } from '@/shared/i18n'
import { ABSENCE_SOURCE_BADGE, CALENDAR_SOURCE } from '@/shared/people'
import { Badge } from '@/shared/ui/badge'
import type { AbsenceSourceBadgeProps } from './types'

/** Marks an absence read from the person's calendar; one entered by hand shows nothing. */
export function AbsenceSourceBadge({ source, compact }: AbsenceSourceBadgeProps) {
  const { t } = useTranslation()
  if (source !== CALENDAR_SOURCE) return null
  if (compact) return <CalendarSync data-icon="inline-start" role="img" aria-label={t('From calendar')} />

  return (
    <Badge variant={ABSENCE_SOURCE_BADGE[source]}>
      <CalendarSync data-icon="inline-start" aria-hidden />
      {t('From calendar')}
    </Badge>
  )
}
