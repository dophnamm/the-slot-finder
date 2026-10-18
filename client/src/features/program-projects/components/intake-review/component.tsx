import { type ChangeEvent, type FormEvent, useState } from 'react'
import { generatePath } from 'react-router'
import { toast } from 'sonner'
import { ROUTES } from '@/resources/routes'
import { DatePicker } from '@/shared/components/date-picker'
import { TextLink } from '@/shared/components/text-link'
import { useExpertises } from '@/shared/expertises'
import { useTranslation } from '@/shared/i18n'
import { useCreateProject } from '@/shared/programs'
import type { ProjectDetail } from '@/shared/programs'
import { Button } from '@/shared/ui/button'
import { DialogFooter } from '@/shared/ui/dialog'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { planningWindow } from '@/utils/functions'
import { maskedSummary } from '../intake-dialog/utils'
import { DemandRow } from './components/demand-row'
import type { DemandRowField } from './components/demand-row/types'
import type { CreateResult, IntakeReviewProps, ReviewTextField, ReviewValues } from './types'
import { canCreate, initialValues, toDemandInputs, toProjectInput } from './utils'

/** The suggestion as an editable form; nothing is saved until Create project. */
export function IntakeReview({ programId, suggestion, onClose }: IntakeReviewProps) {
  const { t } = useTranslation()
  const { data: expertises = [], isPending: expertisesLoading } = useExpertises()
  const create = useCreateProject(programId)
  const [values, setValues] = useState<ReviewValues>(() => initialValues(suggestion))
  // Set when the project exists but its demands could not be saved.
  const [orphan, setOrphan] = useState<ProjectDetail | null>(null)
  const bounds = planningWindow()

  const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
    const field = event.target.name as ReviewTextField
    const { value } = event.target
    setValues((current) => ({ ...current, [field]: value }))
  }
  const handleDateChange = (value: string, name: string) => setValues((current) => ({ ...current, [name]: value }))
  const handleDemandChange = (key: string, field: DemandRowField, value: string) =>
    setValues((current) => ({
      ...current,
      demands: current.demands.map((demand) => (demand.key === key ? { ...demand, [field]: value } : demand)),
    }))
  const handleRemove = (key: string) =>
    setValues((current) => ({ ...current, demands: current.demands.filter((demand) => demand.key !== key) }))
  const handleCreated = ({ project, demandsSaved }: CreateResult) => {
    if (demandsSaved) {
      toast.success(t('Project created.'))
      onClose()
    } else setOrphan(project)
  }
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (!canCreate(values, expertises) || create.isPending || orphan) return
    create.mutate(
      { project: toProjectInput(values), demands: toDemandInputs(values.demands) },
      { onSuccess: handleCreated },
    )
  }

  return (
    <form className="flex flex-col gap-4" onSubmit={handleSubmit}>
      <div className="flex flex-col gap-2 text-sm">
        <p>{maskedSummary(suggestion.masked)}</p>
        {suggestion.warnings.length > 0 && (
          <ul className="list-disc pl-5 text-muted-foreground">
            {suggestion.warnings.map((warning) => (
              <li key={warning}>{warning}</li>
            ))}
          </ul>
        )}
      </div>
      <div className="grid gap-4 sm:grid-cols-2">
        <div className="flex flex-col gap-2 sm:col-span-2">
          <Label htmlFor="intake-name">{t('Name')}</Label>
          <Input
            id="intake-name"
            name="name"
            placeholder={t('e.g. Checkout Revamp')}
            value={values.name}
            onChange={handleChange}
          />
        </div>
        <div className="flex flex-col gap-2">
          <Label htmlFor="intake-start">{t('Earliest start')}</Label>
          <DatePicker
            id="intake-start"
            name="start_earliest"
            min={bounds.minDate}
            max={bounds.maxDate}
            value={values.start_earliest}
            onChange={handleDateChange}
          />
        </div>
        <div className="flex flex-col gap-2">
          <Label htmlFor="intake-deadline">{t('Deadline')}</Label>
          <DatePicker
            id="intake-deadline"
            name="deadline"
            min={values.start_earliest > bounds.minDate ? values.start_earliest : bounds.minDate}
            max={bounds.maxDate}
            value={values.deadline}
            onChange={handleDateChange}
          />
        </div>
        <div className="flex flex-col gap-2">
          <Label htmlFor="intake-scope">{t('Scope points')}</Label>
          <Input
            id="intake-scope"
            placeholder={t('e.g. 120')}
            name="scope_points"
            type="number"
            min={0}
            step="1"
            value={values.scope_points}
            onChange={handleChange}
          />
        </div>
      </div>
      {values.demands.length > 0 && (
        <ul className="flex flex-col gap-3">
          {values.demands.map((demand) => (
            <DemandRow
              key={demand.key}
              demand={demand}
              expertises={expertises}
              expertisesLoading={expertisesLoading}
              onChange={handleDemandChange}
              onRemove={handleRemove}
            />
          ))}
        </ul>
      )}
      {orphan && (
        <p role="alert" className="text-sm text-destructive">
          {t('The project was created without its demand. Add the demand on the project page.')}{' '}
          <TextLink to={generatePath(ROUTES.projectDetail, { programId: orphan.program_id, id: orphan.id })}>
            {t('Open project')}
          </TextLink>
        </p>
      )}
      <DialogFooter>
        <Button type="submit" disabled={!canCreate(values, expertises) || orphan !== null} isLoading={create.isPending}>
          {t('Create project')}
        </Button>
      </DialogFooter>
    </form>
  )
}
