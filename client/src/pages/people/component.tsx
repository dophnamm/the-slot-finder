import { useState } from 'react'
import { HolidaysCard } from '@/features/holidays'
import { PeopleTable, type Person } from '@/features/people'
import { TimesheetImport } from '@/features/timesheet-import'
import { can } from '@/resources/rbac'
import { useCurrentUser } from '@/shared/auth'
import { PageFill } from '@/shared/components/page-fill'
import { useTranslation } from '@/shared/i18n'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/shared/ui/tabs'
import { PersonAbsencesDialog } from './components/person-absences-dialog'

/** People and company holidays, one tab each; each tab's table scrolls inside the page. */
export function PeoplePage() {
  const { t } = useTranslation()
  const { data: user } = useCurrentUser()
  const [absencesOf, setAbsencesOf] = useState<Person | null>(null)

  const handleCloseAbsences = () => setAbsencesOf(null)

  return (
    <PageFill>
      <Tabs defaultValue="people" className="min-h-0 flex-1 gap-5">
        <TabsList variant="line">
          <TabsTrigger value="people">{t('People')}</TabsTrigger>
          <TabsTrigger value="holidays">{t('Holidays')}</TabsTrigger>
        </TabsList>
        <TabsContent value="people" className="flex min-h-0 flex-col">
          <PeopleTable
            onShowAbsences={setAbsencesOf}
            actions={can(user?.role, 'timesheets:import') && <TimesheetImport />}
          />
        </TabsContent>
        <TabsContent value="holidays" className="flex min-h-0 flex-col">
          <HolidaysCard />
        </TabsContent>
      </Tabs>
      <PersonAbsencesDialog
        person={absencesOf}
        canApprove={can(user?.role, 'absences:approve')}
        onClose={handleCloseAbsences}
      />
    </PageFill>
  )
}
