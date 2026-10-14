import { can } from '@/resources/rbac'
import { ROUTES } from '@/resources/routes'
import { useCurrentUser } from '@/shared/auth'
import { useTranslation } from '@/shared/i18n'
import { usePerson } from '@/shared/people'
import { Trail } from '../trail'
import type { PersonTrailProps } from './types'

/** People › the person's name. "People" is a link only for those who may open the People list. */
export function PersonTrail({ personId }: PersonTrailProps) {
  const { t } = useTranslation()
  const { data: user } = useCurrentUser()
  const { data: person, error } = usePerson(personId)
  const peopleLink = can(user?.role, 'people:view') ? ROUTES.people : undefined
  const fallback = error ? t('Person') : t('Loading…')

  return <Trail crumbs={[{ label: t('People'), to: peopleLink }, { label: person?.name ?? fallback }]} />
}
