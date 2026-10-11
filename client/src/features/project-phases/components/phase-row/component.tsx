import { Pencil, Trash2 } from 'lucide-react'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { TableCell, TableRow } from '@/shared/ui/table'
import { formatIsoWeekSpan } from '@/utils/functions'
import type { PhaseRowProps } from './types'

export function PhaseRow({ phase, editable, disabled, onEdit, onRemove }: PhaseRowProps) {
  const { t } = useTranslation()

  const handleEdit = () => onEdit(phase)
  const handleRemove = () => onRemove(phase)

  return (
    <TableRow>
      <TableCell className="font-medium">{phase.name}</TableCell>
      <TableCell>{phase.start_week}</TableCell>
      <TableCell>{phase.end_week}</TableCell>
      <TableCell className="text-muted-foreground">{formatIsoWeekSpan(phase.start_week, phase.end_week)}</TableCell>
      <TableCell>
        {editable && (
          <div className="flex justify-end gap-1">
            <Button type="button" variant="ghost" size="icon-sm" aria-label={t('Edit phase')} disabled={disabled} onClick={handleEdit}>
              <Pencil />
            </Button>
            <Button type="button" variant="ghost" size="icon-sm" aria-label={t('Remove phase')} disabled={disabled} onClick={handleRemove}>
              <Trash2 />
            </Button>
          </div>
        )}
      </TableCell>
    </TableRow>
  )
}
