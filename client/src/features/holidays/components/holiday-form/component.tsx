import { type ChangeEvent, type FormEvent, useState } from 'react'
import { toast } from 'sonner'
import { DatePicker } from '@/shared/components/date-picker'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { DialogFooter } from '@/shared/ui/dialog'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { MAX_HOURS_PER_DAY } from '@/utils/constants'
import { useCreateHoliday } from '../../use-holidays'
import { EMPTY_HOLIDAY } from './constants'
import type { HolidayFormProps, HolidayFormValues } from './types'
import { canSubmit, toInput } from './utils'

/** Adds a company holiday; `onCreated` runs once it is saved, e.g. to close the dialog. */
export function HolidayForm({ onCreated }: HolidayFormProps) {
  const { t } = useTranslation()
  const [values, setValues] = useState<HolidayFormValues>(EMPTY_HOLIDAY)
  const create = useCreateHoliday()

  const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
    const field = event.target.name as keyof HolidayFormValues
    const { value } = event.target
    setValues((current) => ({ ...current, [field]: value }))
  }
  const handleDateChange = (value: string, name: string) =>
    setValues((current) => ({ ...current, [name as keyof HolidayFormValues]: value }))
  const handleCreated = () => {
    toast.success(t('Holiday added.'))
    setValues(EMPTY_HOLIDAY)
    onCreated?.()
  }
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    create.mutate(toInput(values), { onSuccess: handleCreated })
  }

  return (
    <form className="flex flex-col gap-5" onSubmit={handleSubmit}>
      <div className="grid gap-4 sm:grid-cols-2">
        <div className="flex flex-col gap-2">
          <Label htmlFor="holiday-from">{t('From')}</Label>
          <DatePicker id="holiday-from" name="from_date" value={values.from_date} onChange={handleDateChange} />
        </div>
        <div className="flex flex-col gap-2">
          <Label htmlFor="holiday-to">{t('To')}</Label>
          <DatePicker
            id="holiday-to"
            name="to_date"
            min={values.from_date}
            value={values.to_date}
            onChange={handleDateChange}
          />
        </div>
        <div className="flex flex-col gap-2">
          <Label htmlFor="holiday-hours">{t('Hours per day')}</Label>
          <Input
            placeholder={t('e.g. 8')}
            id="holiday-hours"
            name="hours_per_day"
            type="number"
            min={0.5}
            max={MAX_HOURS_PER_DAY}
            step="0.5"
            value={values.hours_per_day}
            onChange={handleChange}
            required
          />
        </div>
      </div>
      <DialogFooter>
        <Button type="submit" disabled={!canSubmit(values) || create.isPending}>
          {t('Add holiday')}
        </Button>
      </DialogFooter>
    </form>
  )
}
