import { AbsenceList, AddAbsence } from '@/features/absences'
import { CalendarFeed } from '@/features/calendar-feed'
import { PersonCalendar } from '@/features/person-calendar'
import { can } from '@/resources/rbac'
import { useCurrentUser } from '@/shared/auth'
import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { Section } from '@/shared/components/section'
import { useTranslation } from '@/shared/i18n'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/shared/ui/tabs'

/** The signed-in person's month, their absences and their calendar sync, one tab each. */
export function MySchedulePage() {
  const { t } = useTranslation()
  const { data: user, isPending } = useCurrentUser()

  if (isPending) return null
  if (!user?.person_id)
    return (
      <PagePlaceholder
        title={t('No person record')}
        description={t('Your account is not linked to a person yet. Ask a planner to add you with the same email.')}
      />
    )

  return (
    <Tabs defaultValue="schedule" className="gap-5">
      <TabsList variant="line">
        <TabsTrigger value="schedule">{t('Schedule')}</TabsTrigger>
        <TabsTrigger value="absences">{t('My absences')}</TabsTrigger>
        <TabsTrigger value="calendar-sync">{t('Calendar sync')}</TabsTrigger>
      </TabsList>
      <TabsContent value="schedule">
        <PersonCalendar personId={user.person_id} canOpenProjects={can(user.role, 'programs:view')} />
      </TabsContent>
      <TabsContent value="absences">
        <Section
          title={t('My absences')}
          description={t('Report leave or training. A planner approves it.')}
          action={<AddAbsence personId={user.person_id} canApprove={false} />}
        >
          <AbsenceList personId={user.person_id} canApprove={false} />
        </Section>
      </TabsContent>
      <TabsContent value="calendar-sync">
        <CalendarFeed personId={user.person_id} />
      </TabsContent>
    </Tabs>
  )
}
