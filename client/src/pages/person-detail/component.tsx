import { useParams } from 'react-router'
import { AbsenceList } from '@/features/absences'
import { CalendarFeed } from '@/features/calendar-feed'
import { PersonCalendar } from '@/features/person-calendar'
import { can } from '@/resources/rbac'
import { useCurrentUser } from '@/shared/auth'
import { Section } from '@/shared/components/section'
import { useTranslation } from '@/shared/i18n'
import { usePerson } from '@/shared/people'

/** One person: their month calendar and, for planners, their absences and calendar. */
export function PersonDetailPage() {
  const { t } = useTranslation()
  const { id = '' } = useParams()
  const { data: user } = useCurrentUser()
  const { data: person } = usePerson(id)
  const primary = person?.skills.find((skill) => skill.primary)

  return (
    <div className="flex flex-col gap-8">
      <Section
        variant="page"
        title={person?.name ?? t('Loading…')}
        description={
          person &&
          t('{expertise}, {hours} h per week, {nonProject} h of them non-project', {
            expertise: primary?.expertise ?? t('No primary expertise'),
            hours: person.weekly_hours,
            nonProject: person.non_project_hours,
          })
        }
      />
      <PersonCalendar personId={id} canOpenProjects={can(user?.role, 'programs:view')} />
      {can(user?.role, 'absences:approve') && (
        <Section
          title={t('Absences')}
          description={t('Approved absence always lowers capacity; planned absence does when the setting is on.')}
        >
          <AbsenceList personId={id} canApprove />
        </Section>
      )}
      {can(user?.role, 'calendar-feeds:manage') && <CalendarFeed personId={id} />}
    </div>
  )
}
