import { can } from '@/resources/rbac'
import { useCurrentUser } from '@/shared/auth'
import { Section } from '@/shared/components/section'
import { useTranslation } from '@/shared/i18n'
import { Table, TableBody } from '@/shared/ui/table'
import { HolidayForm } from './components/holiday-form'
import { HolidayRow } from './components/holiday-row'
import { useDeleteHoliday, useHolidays } from './use-holidays'

/** Company-wide holidays; planners add and remove them. */
export function HolidaysCard() {
  const { t } = useTranslation()
  const { data: user } = useCurrentUser()
  const { data: holidays = [], isPending } = useHolidays()
  const remove = useDeleteHoliday()
  const canManage = can(user?.role, 'holidays:manage')

  const handleDelete = (id: string) => remove.mutate(id)

  return (
    <Section
      title={t('Company holidays')}
      description={t('Holidays lower everyone’s capacity on those days.')}
    >
      <Table>
        <TableBody>
          {holidays.map((holiday) => (
            <HolidayRow key={holiday.id} holiday={holiday} canDelete={canManage} onDelete={handleDelete} />
          ))}
        </TableBody>
      </Table>
      {isPending && <p className="text-center text-sm text-muted-foreground">{t('Loading…')}</p>}
      {!isPending && holidays.length === 0 && (
        <p className="text-center text-sm text-muted-foreground">{t('No holidays yet.')}</p>
      )}
      {canManage && <HolidayForm />}
    </Section>
  )
}
