import { useParams } from 'react-router'
import { AbsenceList, AddAbsence } from '@/features/absences'
import { CalendarFeed } from '@/features/calendar-feed'
import { PersonCalendar } from '@/features/person-calendar'
import { can } from '@/resources/rbac'
import { FORBIDDEN_STATUS } from '@/shared/api/constants'
import { useCurrentUser } from '@/shared/auth'
import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { PageSkeleton } from '@/shared/components/page-skeleton'
import { Section } from '@/shared/components/section'
import { useTranslation } from '@/shared/i18n'
import { usePerson } from '@/shared/people'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/shared/ui/tabs'

/** One person: their month calendar and, for planners, tabs with their absences and calendar sync. */
export function PersonDetailPage() {
  const { t } = useTranslation()
  const { id = '' } = useParams()
  const { data: user } = useCurrentUser()
  const { data: person, error, isPending } = usePerson(id)
  const primary = person?.skills.find((skill) => skill.primary)
  const canApprove = can(user?.role, 'absences:approve')
  const canSync = can(user?.role, 'calendar-feeds:manage')
  // Waits for the person, so a person the user may not see costs one request, not two.
  const calendar = person && <PersonCalendar personId={id} canOpenProjects={can(user?.role, 'programs:view')} />

  if (isPending) return <PageSkeleton />
  if (error?.status === FORBIDDEN_STATUS)
    return (
      <PagePlaceholder
        title={t('You can’t see this person')}
        description={t('A manager sees the people in the teams of the programs they manage.')}
      />
    )

  return (
    <div className="flex flex-col gap-6">
      <Section
        variant="page"
        title={person?.name}
        description={
          person &&
          t('{expertise}, {hours} h per week, {nonProject} h of them non-project', {
            expertise: primary?.expertise ?? t('No primary expertise'),
            hours: person.weekly_hours,
            nonProject: person.non_project_hours,
          })
        }
      />
      {canApprove || canSync ? (
        <Tabs defaultValue="schedule" className="gap-5">
          <TabsList variant="line">
            <TabsTrigger value="schedule">{t('Schedule')}</TabsTrigger>
            {canApprove && <TabsTrigger value="absences">{t('Absences')}</TabsTrigger>}
            {canSync && <TabsTrigger value="calendar-sync">{t('Calendar sync')}</TabsTrigger>}
          </TabsList>
          <TabsContent value="schedule">{calendar}</TabsContent>
          {canApprove && (
            <TabsContent value="absences">
              <Section
                title={t('Absences')}
                description={t('Approved absence always lowers capacity; planned absence does when the setting is on.')}
                action={<AddAbsence personId={id} canApprove />}
              >
                <AbsenceList personId={id} canApprove />
              </Section>
            </TabsContent>
          )}
          {canSync && (
            <TabsContent value="calendar-sync">
              <CalendarFeed personId={id} />
            </TabsContent>
          )}
        </Tabs>
      ) : (
        calendar
      )}
    </div>
  )
}
