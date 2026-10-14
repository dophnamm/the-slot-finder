import { Pencil, Trash2 } from 'lucide-react'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { TableCell, TableRow } from '@/shared/ui/table'
import type { DemandViewRowProps } from './types'

/** One demand as text; people who may edit the project get edit and remove buttons at the end. */
export function DemandViewRow({ index, demand, phase, onEdit, onRemove }: DemandViewRowProps) {
  const { t } = useTranslation()

  const handleEdit = () => onEdit?.(index)
  const handleRemove = () => onRemove?.(index)

  return (
    <TableRow>
      <TableCell className="font-medium">{demand.expertise}</TableCell>
      <TableCell>{phase ?? t('Whole project')}</TableCell>
      <TableCell>{t('Level {level}', { level: demand.min_level })}</TableCell>
      <TableCell>{demand.headcount}</TableCell>
      <TableCell>{demand.total_hours}</TableCell>
      <TableCell className="text-sm text-muted-foreground">
        {t('{start} – {end}', { start: demand.from_week, end: demand.to_week })}
      </TableCell>
      {(onEdit || onRemove) && (
        <TableCell className="w-px text-right">
          <div className="flex justify-end gap-1">
            {onEdit && (
              <Button variant="ghost" size="icon-sm" aria-label={t('Edit demand')} onClick={handleEdit}>
                <Pencil />
              </Button>
            )}
            {onRemove && (
              <Button variant="ghost" size="icon-sm" aria-label={t('Remove demand')} onClick={handleRemove}>
                <Trash2 />
              </Button>
            )}
          </div>
        </TableCell>
      )}
    </TableRow>
  )
}
