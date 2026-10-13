import { useState } from 'react'
import { HolidaysCard } from '@/features/holidays'
import { PeopleTable, type Person } from '@/features/people'
import { TimesheetImport } from '@/features/timesheet-import'
import { can } from '@/resources/rbac'
import { useCurrentUser } from '@/shared/auth'
import { PersonAbsencesDialog } from './components/person-absences-dialog'

export function PeoplePage() {
  const { data: user } = useCurrentUser()
  const [absencesOf, setAbsencesOf] = useState<Person | null>(null)

  const handleCloseAbsences = () => setAbsencesOf(null)

  return (
    <div className="flex flex-col gap-8">
      <PeopleTable
        onShowAbsences={setAbsencesOf}
        actions={can(user?.role, 'timesheets:import') && <TimesheetImport />}
      />
      <HolidaysCard />
      <PersonAbsencesDialog
        person={absencesOf}
        canApprove={can(user?.role, 'absences:approve')}
        onClose={handleCloseAbsences}
      />
    </div>
  )
}
