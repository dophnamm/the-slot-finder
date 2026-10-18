import { type ChangeEvent, type FormEvent, useState } from 'react'
import { toast } from 'sonner'
import { DatePicker } from '@/shared/components/date-picker'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { DialogFooter } from '@/shared/ui/dialog'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { Select } from '@/shared/ui/select'
import { MAX_HOURS_PER_DAY } from '@/utils/constants'
import { ABSENCE_STATUS_LABELS, ABSENCE_STATUSES, ABSENCE_TYPE_LABELS, PERSONAL_ABSENCE_TYPES } from '../../constants'
import { useCreateAbsence } from '../../use-absences'
import { EMPTY_ABSENCE } from './constants'
import type { AbsenceFormProps, AbsenceFormValues, AbsenceTextField } from './types'
import { canSubmit, isAbsenceStatus, isAbsenceType, toInput } from './utils'

/** Adds leave or training. Planners also choose whether it is planned or approved. */
export function AbsenceForm({ personId, canApprove, onCreated }: AbsenceFormProps) {
  const { t } = useTranslation()
  const [values, setValues] = useState<AbsenceFormValues>(EMPTY_ABSENCE)
  const create = useCreateAbsence(personId)

  const handleTextChange = (event: ChangeEvent<HTMLInputElement>) => {
    const field = event.target.name as AbsenceTextField
    const { value } = event.target
    setValues((current) => ({ ...current, [field]: value }))
  }
  const handleDateChange = (value: string, name: string) =>
    setValues((current) => ({ ...current, [name as AbsenceTextField]: value }))
  const handleTypeChange = (value: string) => {
    if (isAbsenceType(value)) setValues((current) => ({ ...current, type: value }))
  }
  const handleStatusChange = (value: string) => {
    if (isAbsenceStatus(value)) setValues((current) => ({ ...current, status: value }))
  }
  const typeOptions = PERSONAL_ABSENCE_TYPES.map((type) => ({ value: type, label: t(ABSENCE_TYPE_LABELS[type]) }))
  const statusOptions = ABSENCE_STATUSES.map((status) => ({ value: status, label: t(ABSENCE_STATUS_LABELS[status]) }))
  const handleCreated = () => {
    toast.success(t('Absence added.'))
    setValues(EMPTY_ABSENCE)
    onCreated?.()
  }
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    create.mutate(toInput(values, canApprove), { onSuccess: handleCreated })
  }

  return (
    <form className="flex flex-col gap-5" onSubmit={handleSubmit}>
      <div className="grid gap-4 sm:grid-cols-2">
        <div className="flex flex-col gap-2">
          <Label htmlFor={`absence-from-${personId}`}>{t('From')}</Label>
          <DatePicker
            id={`absence-from-${personId}`}
            name="from_date"
            value={values.from_date}
            onChange={handleDateChange}
          />
        </div>
        <div className="flex flex-col gap-2">
          <Label htmlFor={`absence-to-${personId}`}>{t('To')}</Label>
          <DatePicker
            id={`absence-to-${personId}`}
            name="to_date"
            min={values.from_date}
            value={values.to_date}
            onChange={handleDateChange}
          />
        </div>
        <div className="flex flex-col gap-2">
          <Label htmlFor={`absence-hours-${personId}`}>{t('Hours per day')}</Label>
          <Input
            placeholder={t('e.g. 8')}
            id={`absence-hours-${personId}`}
            name="hours_per_day"
            type="number"
            min={0.5}
            max={MAX_HOURS_PER_DAY}
            step="0.5"
            value={values.hours_per_day}
            onChange={handleTextChange}
            required
          />
        </div>
        <div className="flex flex-col gap-2">
          <Label htmlFor={`absence-type-${personId}`}>{t('Type')}</Label>
          <Select
            id={`absence-type-${personId}`}
            className="w-full"
            options={typeOptions}
            value={values.type}
            onValueChange={handleTypeChange}
          />
        </div>
        {canApprove && (
          <div className="flex flex-col gap-2">
            <Label htmlFor={`absence-status-${personId}`}>{t('Status')}</Label>
            <Select
              id={`absence-status-${personId}`}
              className="w-full"
              options={statusOptions}
              value={values.status}
              onValueChange={handleStatusChange}
            />
          </div>
        )}
      </div>
      <DialogFooter>
        <Button type="submit" disabled={!canSubmit(values)} isLoading={create.isPending}>
          {t('Add absence')}
        </Button>
      </DialogFooter>
    </form>
  )
}
