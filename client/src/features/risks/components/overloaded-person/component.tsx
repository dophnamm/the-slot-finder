import { generatePath } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { TextLink } from '@/shared/components/text-link'
import { useTranslation } from '@/shared/i18n'
import { shortWeekLabel } from '@/utils/functions'
import type { OverloadedPersonProps } from './types'

export function OverloadedPerson({ person }: OverloadedPersonProps) {
  const { t } = useTranslation()

  return (
    <li className="flex flex-col gap-1 py-3 first:pt-0 last:pb-0">
      <TextLink to={generatePath(ROUTES.personDetail, { id: person.person.id })} className="font-medium">
        {person.person.name}
      </TextLink>
      <div className="flex flex-wrap gap-x-3 gap-y-1 text-xs">
        {person.weeks.map((week) => (
          <span key={week.week} className="text-destructive tabular-nums">
            {t('{week}: {load}/{capacity} hrs', {
              week: shortWeekLabel(week.week),
              load: Math.round(week.load),
              capacity: Math.round(week.capacity),
            })}
          </span>
        ))}
      </div>
    </li>
  )
}
