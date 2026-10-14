import { Trash2 } from 'lucide-react'
import type { ChangeEvent } from 'react'
import { DatePicker } from '@/shared/components/date-picker'
import { useTranslation } from '@/shared/i18n'
import {
  ASSUMPTION_TYPE_LABELS,
  type AssumptionType,
  SCENARIO_MAX_SHIFT_WEEKS,
  SHIFT_PROJECT,
} from '@/shared/scenarios'
import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { NativeSelect, NativeSelectOption } from '@/shared/ui/native-select'
import { MAX_HOURS_PER_DAY } from '@/utils/constants'
import { ASSUMPTION_TYPE_OPTIONS } from './constants'
import type { AssumptionRowProps } from './types'

/** One assumption: move a project by some weeks, or a person away for some dates. */
export function AssumptionRow({ draft, projects, people, removable, onChange, onRemove }: AssumptionRowProps) {
  const { t } = useTranslation()

  const handleTypeChange = (event: ChangeEvent<HTMLSelectElement>) =>
    onChange({ ...draft, type: event.target.value as AssumptionType })
  const handleTextChange = (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
    onChange({ ...draft, [event.target.name]: event.target.value })
  const handleDateChange = (value: string, name: string) => onChange({ ...draft, [name]: value })
  const handleRemove = () => onRemove(draft.key)

  return (
    <div className="flex flex-wrap items-end gap-3 rounded-lg border p-3">
      <div className="flex flex-col gap-1">
        <span className="text-xs text-muted-foreground">{t('Change')}</span>
        <NativeSelect aria-label={t('Change')} value={draft.type} onChange={handleTypeChange}>
          {ASSUMPTION_TYPE_OPTIONS.map((type) => (
            <NativeSelectOption key={type} value={type}>
              {t(ASSUMPTION_TYPE_LABELS[type])}
            </NativeSelectOption>
          ))}
        </NativeSelect>
      </div>
      {draft.type === SHIFT_PROJECT ? (
        <>
          <div className="flex min-w-56 flex-1 flex-col gap-1">
            <span className="text-xs text-muted-foreground">{t('Project')}</span>
            <NativeSelect
              className="w-full"
              name="project_id"
              aria-label={t('Project')}
              value={draft.project_id}
              onChange={handleTextChange}
            >
              <NativeSelectOption value="">{t('Choose a project')}</NativeSelectOption>
              {projects.map((project) => (
                <NativeSelectOption key={project.id} value={project.id}>
                  {t('{project} ({program})', { project: project.name, program: project.program_name ?? '' })}
                </NativeSelectOption>
              ))}
            </NativeSelect>
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-xs text-muted-foreground">{t('Weeks (minus is earlier)')}</span>
            <Input
              placeholder={t('e.g. 2')}
              name="weeks"
              type="number"
              className="w-28"
              min={-SCENARIO_MAX_SHIFT_WEEKS}
              max={SCENARIO_MAX_SHIFT_WEEKS}
              step={1}
              aria-label={t('Weeks')}
              value={draft.weeks}
              onChange={handleTextChange}
            />
          </div>
        </>
      ) : (
        <>
          <div className="flex min-w-44 flex-1 flex-col gap-1">
            <span className="text-xs text-muted-foreground">{t('Person')}</span>
            <NativeSelect
              className="w-full"
              name="person_id"
              aria-label={t('Person')}
              value={draft.person_id}
              onChange={handleTextChange}
            >
              <NativeSelectOption value="">{t('Choose a person')}</NativeSelectOption>
              {people.map((person) => (
                <NativeSelectOption key={person.id} value={person.id}>
                  {person.name}
                </NativeSelectOption>
              ))}
            </NativeSelect>
          </div>
          <div className="flex w-40 flex-col gap-1">
            <span className="text-xs text-muted-foreground">{t('From')}</span>
            <DatePicker name="from_date" aria-label={t('From')} value={draft.from_date} onChange={handleDateChange} />
          </div>
          <div className="flex w-40 flex-col gap-1">
            <span className="text-xs text-muted-foreground">{t('To')}</span>
            <DatePicker
              name="to_date"
              aria-label={t('To')}
              min={draft.from_date}
              value={draft.to_date}
              onChange={handleDateChange}
            />
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-xs text-muted-foreground">{t('Hours per day')}</span>
            <Input
              placeholder={t('e.g. 8')}
              name="hours_per_day"
              type="number"
              className="w-24"
              min={0.5}
              max={MAX_HOURS_PER_DAY}
              step="0.5"
              aria-label={t('Hours per day')}
              value={draft.hours_per_day}
              onChange={handleTextChange}
            />
          </div>
        </>
      )}
      <Button
        type="button"
        variant="ghost"
        size="icon-sm"
        aria-label={t('Remove assumption')}
        disabled={!removable}
        onClick={handleRemove}
      >
        <Trash2 />
      </Button>
    </div>
  )
}
