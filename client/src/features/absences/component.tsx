import { CalendarOff } from 'lucide-react'
import { TableEmpty } from '@/shared/components/table-empty'
import { TableLoading } from '@/shared/components/table-loading'
import { useTranslation } from '@/shared/i18n'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { ABSENCE_TABLE_COLUMNS } from './constants'
import { AbsenceRow } from './components/absence-row'
import type { AbsenceListProps } from './types'
import { useAbsences, useApproveAbsence, useDeleteAbsence } from './use-absences'

/** A person's leave and training; AddAbsence adds more. */
export function AbsenceList({ personId, canApprove }: AbsenceListProps) {
  const { t } = useTranslation()
  const { data: absences = [], isPending } = useAbsences(personId)
  const remove = useDeleteAbsence(personId)
  const approve = useApproveAbsence(personId)

  const handleDelete = (id: string) => remove.mutate(id)
  const handleApprove = (id: string) => approve.mutate(id)

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>{t('Dates')}</TableHead>
          <TableHead>{t('Type')}</TableHead>
          <TableHead>{t('Hours per day')}</TableHead>
          <TableHead>{t('Status')}</TableHead>
          <TableHead />
        </TableRow>
      </TableHeader>
      <TableBody>
        {absences.map((absence) => (
          <AbsenceRow
            key={absence.id}
            absence={absence}
            onDelete={handleDelete}
            onApprove={canApprove ? handleApprove : undefined}
            approving={approve.isPending && approve.variables === absence.id}
          />
        ))}
        {isPending && <TableLoading columns={ABSENCE_TABLE_COLUMNS} />}
        {!isPending && absences.length === 0 && (
          <TableEmpty columns={ABSENCE_TABLE_COLUMNS} icon={CalendarOff} title={t('No absences yet.')} />
        )}
      </TableBody>
    </Table>
  )
}
