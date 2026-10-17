import { generatePath } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { TextLink } from '@/shared/components/text-link'
import { useTranslation } from '@/shared/i18n'
import { BAND_CLASSES, BAND_LABELS } from '@/shared/utilization'
import { cn, formatIsoWeekRange } from '@/utils/functions'
import { percent, roundHours } from '../../../../utils'
import type { CellPeopleProps } from './types'

/** Who is in an expertise in a week, highest load first, and how each person stands. */
export function CellPeople({ expertise, cell }: CellPeopleProps) {
  const { t } = useTranslation()

  return (
    <div className="flex flex-col gap-3 text-sm">
      <div className="flex flex-col">
        <span className="font-medium">{expertise}</span>
        <span className="text-xs text-muted-foreground">
          {t('{week} ({dates})', { week: cell.week, dates: formatIsoWeekRange(cell.week) })}
        </span>
      </div>
      {cell.people.length === 0 && (
        <span className="text-muted-foreground">{t('No one has capacity or load in this group this week.')}</span>
      )}
      <ul className="flex flex-col gap-2">
        {cell.people.map((person) => (
          <li key={person.person_id} className="flex items-start justify-between gap-2">
            <div className="flex min-w-0 flex-col">
              <TextLink
                to={generatePath(ROUTES.personDetail, { id: person.person_id })}
                className="truncate font-medium"
              >
                {person.name}
              </TextLink>
              <span className="text-xs text-muted-foreground">
                {t('{load}/{capacity} hrs', { load: roundHours(person.load), capacity: roundHours(person.capacity) })}
                {' · '}
                {t('{hours} hrs free', { hours: roundHours(person.free_hours) })}
              </span>
            </div>
            <span
              title={t(BAND_LABELS[person.band])}
              className={cn('shrink-0 rounded-md px-1.5 py-0.5 text-xs font-medium', BAND_CLASSES[person.band])}
            >
              {person.utilization === null ? '—' : t('{percent}%', { percent: percent(person.utilization) })}
            </span>
          </li>
        ))}
      </ul>
    </div>
  )
}
