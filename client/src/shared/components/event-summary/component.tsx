import { TEAM_CHANGE_ADDED } from '@/shared/events'
import { useTranslation } from '@/shared/i18n'
import { formatDate, formatDateRange, toIsoWeek } from '@/utils/functions'
import type { EventSummaryProps } from './types'
import { capacityFieldLabels, statusKey } from './utils'

/** One sentence for a change: who is away, or what moved or changed. */
export function EventSummary({ event }: EventSummaryProps) {
  const { t } = useTranslation()
  const { payload } = event
  const dates = payload.from_date && payload.to_date ? formatDateRange(payload.from_date, payload.to_date) : ''
  const person = payload.person_name ?? ''
  const program = payload.program ?? ''
  const added = payload.change === TEAM_CHANGE_ADDED
  const oldStatus = payload.old_status ?? ''
  const newStatus = payload.new_status ?? ''
  const oldStatusKey = statusKey(payload.kind, oldStatus)
  const newStatusKey = statusKey(payload.kind, newStatus)

  switch (event.type) {
    case 'AbsenceAdded':
      return (
        <span>
          {payload.person_id
            ? t('{person} is away {dates}', { person, dates })
            : t('Company holiday {dates}', { dates })}
        </span>
      )
    case 'AbsenceRemoved':
      return (
        <span>
          {payload.person_id
            ? t('{person} is no longer away {dates}', { person, dates })
            : t('Company holiday {dates} removed', { dates })}
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
    case 'SkillChanged':
      return <span>{t('{person}: skills changed', { person })}</span>
    case 'ProgramSlipped':
      return (
        <span>
          {t('{program}: end moved from {from} to {to}', {
            program,
            from: payload.old_end_date ? toIsoWeek(payload.old_end_date) : '',
            to: payload.new_end_date ? toIsoWeek(payload.new_end_date) : '',
          })}
        </span>
      )
    case 'DemandChanged':
      return <span>{t('{project}: demand changed', { project: payload.project ?? '' })}</span>
    case 'PersonCapacityChanged':
      return (
        <span>
          {t('{person}: {fields} changed', {
            person,
            fields: capacityFieldLabels(payload.fields ?? [], t),
          })}
        </span>
      )
    case 'ProjectStatusChanged':
      return (
        <span>
          {t('{name}: {from} → {to}', {
            name: payload.name ?? '',
            from: oldStatusKey ? t(oldStatusKey) : oldStatus,
            to: newStatusKey ? t(newStatusKey) : newStatus,
          })}
        </span>
      )
    case 'TeamChanged':
      if (!person || !program) return <span>{added ? t('Team member added') : t('Team member removed')}</span>
      return (
        <span>
          {added
            ? t('{person} added to the {program} team', { person, program })
            : t('{person} removed from the {program} team', { person, program })}
        </span>
      )
  }
}
