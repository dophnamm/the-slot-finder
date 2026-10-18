import { type ChangeEvent, type FormEvent, useState } from 'react'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/shared/ui/dialog'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { Select } from '@/shared/ui/select'
import { formatIsoWeekRange } from '@/utils/functions'
import { MAX_WEEK_HOURS } from '../../constants'
import { ALLOCATE_MODE } from './constants'
import type { AllocateHoursDialogProps, AllocateMode } from './types'
import { averageHours, selectWeeks, weekHours } from './utils'

/** Sets one assignment's hours for a range of weeks at once; locked weeks keep their hours. */
export function AllocateHoursDialog({ assignment, weeks, busy, onSave, onCancel }: AllocateHoursDialogProps) {
  const { t } = useTranslation()
  const [from, setFrom] = useState(weeks[0] ?? '')
  const [to, setTo] = useState(weeks.at(-1) ?? '')
  const [mode, setMode] = useState<AllocateMode>(ALLOCATE_MODE.perWeek)
  const [value, setValue] = useState('')

  const { open, skipped } = selectWeeks(weeks, from, to, assignment)
  const hours = value.trim() === '' ? Number.NaN : Number(value)
  const perWeek = Number.isNaN(hours) || hours < 0 ? [] : weekHours(mode, hours, open.length)
  const canSave = perWeek.length > 0 && perWeek.every((week) => week >= 0 && week <= MAX_WEEK_HOURS)
  const hoursLabel = mode === ALLOCATE_MODE.total ? t('Total hours') : t('Hours per week')

  const preview = () => {
    const count = open.length
    if (mode === ALLOCATE_MODE.perWeek) {
      const params = { count, hours, total: hours * count }
      return count === 1
        ? t('{count} week × {hours} hrs = {total} hrs', params)
        : t('{count} weeks × {hours} hrs = {total} hrs', params)
    }
    if (count === 1) return t('{total} hrs over {count} week', { total: hours, count })
    return t('{total} hrs over {count} weeks ≈ {hours} hrs/week', {
      total: hours,
      count,
      hours: averageHours(hours, count),
    })
  }
  const weekOptions = weeks.map((week) => ({
    value: week,
    label: t('{week} · {range}', { week, range: formatIsoWeekRange(week) }),
  }))
  const modeOptions = [
    { value: ALLOCATE_MODE.perWeek, label: t('Hours per week') },
    { value: ALLOCATE_MODE.total, label: t('Total hours') },
  ]
  const handleMode = (value: string) => setMode(value as AllocateMode)
  const handleValue = (event: ChangeEvent<HTMLInputElement>) => setValue(event.target.value)
  const handleOpenChange = (isOpen: boolean) => {
    if (!isOpen) onCancel()
  }
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    onSave(open.map((week, index) => ({ assignment_id: assignment.id, week, hours: perWeek[index] })))
  }

  return (
    <Dialog open onOpenChange={handleOpenChange}>
      <DialogContent>
        <form className="flex flex-col gap-5" onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>{t('Allocate hours')}</DialogTitle>
            <DialogDescription>
              {t('Sets the hours of {name} for many weeks at once. Locked weeks keep their hours.', {
                name: assignment.person.name,
              })}
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="flex flex-col gap-2">
              <Label htmlFor="allocate-from">{t('From week')}</Label>
              <Select
                id="allocate-from"
                className="w-full"
                options={weekOptions}
                value={from}
                onValueChange={setFrom}
              />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="allocate-to">{t('To week')}</Label>
              <Select id="allocate-to" className="w-full" options={weekOptions} value={to} onValueChange={setTo} />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="allocate-mode">{t('Mode')}</Label>
              <Select
                id="allocate-mode"
                className="w-full"
                options={modeOptions}
                value={mode}
                onValueChange={handleMode}
              />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="allocate-hours">{hoursLabel}</Label>
              <Input
                id="allocate-hours"
                placeholder={t('e.g. 16')}
                type="number"
                inputMode="decimal"
                min={0}
                step="any"
                value={value}
                onChange={handleValue}
                required
                autoFocus
              />
            </div>
          </div>
          <div className="flex flex-col gap-1 text-sm text-muted-foreground">
            {canSave && <p>{preview()}</p>}
            {open.length === 0 && <p>{t('No open week in this range.')}</p>}
            {skipped === 1 && <p>{t('{count} locked week is skipped.', { count: skipped })}</p>}
            {skipped > 1 && <p>{t('{count} locked weeks are skipped.', { count: skipped })}</p>}
            {perWeek.length > 0 && !canSave && (
              <p className="text-destructive">{t('A week can hold at most {hours} hrs.', { hours: MAX_WEEK_HOURS })}</p>
            )}
          </div>
          <DialogFooter>
            <Button type="button" variant="ghost" disabled={busy} onClick={onCancel}>
              {t('Cancel')}
            </Button>
            <Button type="submit" disabled={!canSave} isLoading={busy}>
              {t('Allocate hours')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
