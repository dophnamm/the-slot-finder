import { useTranslation } from '@/shared/i18n'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { AbsenceForm } from './components/absence-form'
import { AbsenceRow } from './components/absence-row'
import type { AbsenceListProps } from './types'
import { useAbsences, useDeleteAbsence } from './use-absences'

/** A person's leave and training, with a form to add more. */
export function AbsenceList({ personId, canApprove }: AbsenceListProps) {
  const { t } = useTranslation()
  const { data: absences = [], isPending } = useAbsences(personId)
  const remove = useDeleteAbsence(personId)

  const handleDelete = (id: string) => remove.mutate(id)

  return (
    <div className="flex flex-col gap-4">
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
            <AbsenceRow key={absence.id} absence={absence} onDelete={handleDelete} />
          ))}
        </TableBody>
      </Table>
      {isPending && <p className="text-center text-sm text-muted-foreground">{t('Loading…')}</p>}
      {!isPending && absences.length === 0 && (
        <p className="text-center text-sm text-muted-foreground">{t('No absences yet.')}</p>
      )}
      <AbsenceForm personId={personId} canApprove={canApprove} />
    </div>
  )
}
