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
import { Select } from '@/shared/ui/select'
import { MAX_HOURS_PER_DAY } from '@/utils/constants'
import { ASSUMPTION_TYPE_OPTIONS } from './constants'
import type { AssumptionRowProps } from './types'

/** One assumption: move a project by some weeks, or a person away for some dates. */
export function AssumptionRow({
  draft,
  projects,
  people,
  projectsLoading,
  peopleLoading,
  removable,
  onChange,
  onRemove,
}: AssumptionRowProps) {
  const { t } = useTranslation()

  const typeOptions = ASSUMPTION_TYPE_OPTIONS.map((type) => ({ value: type, label: t(ASSUMPTION_TYPE_LABELS[type]) }))
  const projectOptions = [
    { value: '', label: t('Choose a project') },
    ...projects.map((project) => ({
      value: project.id,
      label: t('{project} ({program})', { project: project.name, program: project.program_name ?? '' }),
    })),
  ]
  const personOptions = [
    { value: '', label: t('Choose a person') },
    ...people.map((person) => ({ value: person.id, label: person.name })),
  ]
  const handleTypeChange = (value: string) => onChange({ ...draft, type: value as AssumptionType })
  const handleProjectChange = (value: string) => onChange({ ...draft, project_id: value })
  const handlePersonChange = (value: string) => onChange({ ...draft, person_id: value })
  const handleTextChange = (event: ChangeEvent<HTMLInputElement>) =>
    onChange({ ...draft, [event.target.name]: event.target.value })
  const handleDateChange = (value: string, name: string) => onChange({ ...draft, [name]: value })
  const handleRemove = () => onRemove(draft.key)

  return (
    <div className="flex flex-wrap items-end gap-3 rounded-lg border p-3">
      <div className="flex flex-col gap-1">
        <span className="text-xs text-muted-foreground">{t('Change')}</span>
        <Select aria-label={t('Change')} options={typeOptions} value={draft.type} onValueChange={handleTypeChange} />
      </div>
      {draft.type === SHIFT_PROJECT ? (
        <>
          <div className="flex min-w-56 flex-1 flex-col gap-1">
            <span className="text-xs text-muted-foreground">{t('Project')}</span>
            <Select
              className="w-full"
              aria-label={t('Project')}
              options={projectOptions}
              value={draft.project_id}
              isLoading={projectsLoading}
              onValueChange={handleProjectChange}
            />
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
            <Select
              className="w-full"
              aria-label={t('Person')}
              options={personOptions}
              value={draft.person_id}
              isLoading={peopleLoading}
              onValueChange={handlePersonChange}
            />
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
