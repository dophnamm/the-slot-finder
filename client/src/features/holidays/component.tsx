import { CalendarOff } from 'lucide-react'
import { can } from '@/resources/rbac'
import { useCurrentUser } from '@/shared/auth'
import { Section } from '@/shared/components/section'
import { TableEmpty } from '@/shared/components/table-empty'
import { TableLoading } from '@/shared/components/table-loading'
import { useTranslation } from '@/shared/i18n'
import { Table, TableBody } from '@/shared/ui/table'
import { HOLIDAY_TABLE_COLUMNS } from './constants'
import { AddHoliday } from './components/add-holiday'
import { ImportHolidays } from './components/import-holidays'
import { HolidayRow } from './components/holiday-row'
import { useDeleteHoliday, useHolidays } from './use-holidays'

/** Company-wide holidays; planners add them by hand or import a holiday calendar, and remove them. */
export function HolidaysCard() {
  const { t } = useTranslation()
  const { data: user } = useCurrentUser()
  const { data: holidays = [], isPending } = useHolidays()
  const remove = useDeleteHoliday()
  const canManage = can(user?.role, 'holidays:manage')

  const handleDelete = (id: string) => remove.mutate(id)

  return (
    <Section
      className="min-h-0 flex-1"
      title={t('Company holidays')}
      description={t('Holidays lower everyone’s capacity on those days.')}
      action={
        canManage && (
          <>
            <ImportHolidays />
            <AddHoliday />
          </>
        )
      }
    >
      <Table scroll="both">
        <TableBody>
          {holidays.map((holiday) => (
            <HolidayRow key={holiday.id} holiday={holiday} canDelete={canManage} onDelete={handleDelete} />
          ))}
          {isPending && <TableLoading columns={HOLIDAY_TABLE_COLUMNS} />}
          {!isPending && holidays.length === 0 && (
            <TableEmpty columns={HOLIDAY_TABLE_COLUMNS} icon={CalendarOff} title={t('No holidays yet.')} />
          )}
        </TableBody>
      </Table>
    </Section>
  )
}
