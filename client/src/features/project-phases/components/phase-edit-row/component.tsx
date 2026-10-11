import type { ChangeEvent } from 'react'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { TableCell, TableRow } from '@/shared/ui/table'
import { formatIsoWeekSpan } from '@/utils/functions'
import type { PhaseDraftField } from '../../types'
import type { PhaseEditRowProps } from './types'

/** A phase row in edit mode: name and week inputs; the timeline above previews the change. */
export function PhaseEditRow({ draft, canSave, busy, onChange, onSave, onCancel }: PhaseEditRowProps) {
  const { t } = useTranslation()

  const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
    const field = event.target.name as PhaseDraftField
    onChange({ ...draft, [field]: event.target.value })
  }

  return (
    <TableRow>
      <TableCell>
        <Input name="name" aria-label={t('Phase name')} value={draft.name} onChange={handleChange} autoFocus />
      </TableCell>
      <TableCell>
        <Input name="start_week" type="week" aria-label={t('Start week')} value={draft.start_week} onChange={handleChange} />
      </TableCell>
      <TableCell>
        <Input name="end_week" type="week" min={draft.start_week} aria-label={t('End week')} value={draft.end_week}
          onChange={handleChange} />
      </TableCell>
      <TableCell className="text-muted-foreground">{canSave && formatIsoWeekSpan(draft.start_week, draft.end_week)}</TableCell>
      <TableCell>
        <div className="flex justify-end gap-1">
          <Button type="button" size="sm" disabled={!canSave || busy} onClick={onSave}>
            {t('Save')}
          </Button>
          <Button type="button" size="sm" variant="ghost" disabled={busy} onClick={onCancel}>
            {t('Cancel')}
          </Button>
        </div>
      </TableCell>
    </TableRow>
  )
}
