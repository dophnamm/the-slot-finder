import { useTranslation } from '@/shared/i18n'
import { formatDate, formatDateRange } from '@/utils/functions'
import type { EventSummaryProps } from './types'

/** One sentence for a change: who is away, or which phase or deadline moved. */
export function EventSummary({ event }: EventSummaryProps) {
  const { t } = useTranslation()
  const { payload } = event
  const dates = payload.from_date && payload.to_date ? formatDateRange(payload.from_date, payload.to_date) : ''
  const person = payload.person_name ?? ''

  switch (event.type) {
    case 'AbsenceAdded':
      return <span>{payload.person_id ? t('{person} is away {dates}', { person, dates }) : t('Company holiday {dates}', { dates })}</span>
    case 'AbsenceRemoved':
      return (
        <span>
          {payload.person_id ? t('{person} is no longer away {dates}', { person, dates }) : t('Company holiday {dates} removed', { dates })}
        </span>
      )
    case 'PhaseSlipped':
      return (
        <span>
          {t('{project}: {phase} now ends {week}, {weeks} weeks later', {
            project: payload.project ?? '',
            phase: payload.phase ?? '',
            week: payload.new_end_week ?? '',
            weeks: payload.weeks ?? 0,
          })}
        </span>
      )
    case 'ProjectSlipped':
      return (
        <span>
          {t('{project}: deadline moved from {from} to {to}', {
            project: payload.project ?? '',
            from: payload.old_deadline ? formatDate(payload.old_deadline) : '',
            to: payload.new_deadline ? formatDate(payload.new_deadline) : '',
          })}
        </span>
      )
  }
}
