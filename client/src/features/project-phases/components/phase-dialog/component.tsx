import type { ChangeEvent, FormEvent } from 'react'
import { WeekPicker } from '@/shared/components/week-picker'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/shared/ui/dialog'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { formatIsoWeekSpan } from '@/utils/functions'
import type { PhaseDraftField } from '../../types'
import type { PhaseDialogProps } from './types'

/** Adds or edits one phase; the timeline in the dialog previews how later phases move before saving. */
export function PhaseDialog({ draft, canSave, busy, preview, onChange, onSave, onCancel }: PhaseDialogProps) {
  const { t } = useTranslation()

  const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
    const field = event.target.name as PhaseDraftField
    onChange({ ...draft, [field]: event.target.value })
  }
  const handleWeekChange = (value: string, name: string) => onChange({ ...draft, [name as PhaseDraftField]: value })
  const handleOpenChange = (open: boolean) => {
    if (!open) onCancel()
  }
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    onSave()
  }

  return (
    <Dialog open onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-3xl">
        <form className="flex flex-col gap-5" onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>{draft.id === null ? t('Add phase') : t('Edit phase')}</DialogTitle>
            <DialogDescription>
              {t('A later end week moves every later phase by the same number of weeks.')}
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 sm:grid-cols-3">
            <div className="flex flex-col gap-2">
              <Label htmlFor="phase-name">{t('Phase name')}</Label>
              <Input id="phase-name" name="name" value={draft.name} onChange={handleChange} required autoFocus />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="phase-start">{t('Start week')}</Label>
              <WeekPicker id="phase-start" name="start_week" value={draft.start_week} onChange={handleWeekChange} />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="phase-end">{t('End week')}</Label>
              <WeekPicker
                id="phase-end"
                name="end_week"
                min={draft.start_week}
                value={draft.end_week}
                onChange={handleWeekChange}
              />
            </div>
          </div>
          {canSave && (
            <p className="text-sm text-muted-foreground">{formatIsoWeekSpan(draft.start_week, draft.end_week)}</p>
          )}
          {preview}
          <DialogFooter>
            <Button type="button" variant="ghost" disabled={busy} onClick={onCancel}>
              {t('Cancel')}
            </Button>
            <Button type="submit" disabled={!canSave || busy}>
              {t('Save phase')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
