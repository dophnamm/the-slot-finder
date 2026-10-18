import { type ChangeEvent, type FormEvent, useState } from 'react'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/shared/ui/dialog'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { Select } from '@/shared/ui/select'
import { SKILL_LEVELS } from '@/utils/constants'
import { DEMAND_LIMITS } from '../../constants'
import { demandWindow, isHoursStep, isValidDraft } from '../../utils'
import type { DemandDialogProps, DemandTextField } from './types'

/** Adds or edits one demand; saving it saves the project's whole demand list. */
export function DemandDialog({
  title,
  initial,
  expertises,
  expertisesLoading,
  project,
  saving,
  onSubmit,
  onOpenChange,
}: DemandDialogProps) {
  const { t } = useTranslation()
  const [draft, setDraft] = useState(initial)
  const hoursOff = !isHoursStep(draft.total_hours)
  const weeks = demandWindow(project, draft.phase_id || null)

  const expertiseOptions = expertises.map((expertise) => ({ value: expertise.id, label: expertise.name }))
  const phaseOptions = [
    { value: '', label: t('Whole project') },
    ...project.phases.map((phase) => ({ value: phase.id, label: phase.name })),
  ]
  const levelOptions = SKILL_LEVELS.map((level) => ({ value: String(level), label: t('Level {level}', { level }) }))
  const handleExpertiseChange = (value: string) => setDraft((current) => ({ ...current, expertise_id: value }))
  const handlePhaseChange = (value: string) => setDraft((current) => ({ ...current, phase_id: value }))
  const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
    const field = event.target.name as DemandTextField
    const { value } = event.target
    setDraft((current) => ({ ...current, [field]: value }))
  }
  const handleLevelChange = (value: string) => setDraft((current) => ({ ...current, min_level: Number(value) }))
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    onSubmit(draft)
  }

  return (
    <Dialog open onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <form className="flex flex-col gap-5" onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>{title}</DialogTitle>
            <DialogDescription>
              {t('Budget hours per expertise and minimum level. Without a phase, a demand spans the whole project.')}
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="flex flex-col gap-2">
              <Label htmlFor="demand-expertise">{t('Expertise')}</Label>
              <Select
                id="demand-expertise"
                className="w-full"
                options={expertiseOptions}
                value={draft.expertise_id}
                isLoading={expertisesLoading}
                onValueChange={handleExpertiseChange}
              />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="demand-phase">{t('Phase')}</Label>
              <Select
                id="demand-phase"
                className="w-full"
                options={phaseOptions}
                value={draft.phase_id}
                onValueChange={handlePhaseChange}
              />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="demand-level">{t('Minimum level')}</Label>
              <Select
                id="demand-level"
                className="w-full"
                options={levelOptions}
                value={String(draft.min_level)}
                onValueChange={handleLevelChange}
              />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="demand-headcount">{t('Headcount')}</Label>
              <Input
                placeholder={t('e.g. 2')}
                id="demand-headcount"
                name="headcount"
                type="number"
                min={DEMAND_LIMITS.minHeadcount}
                max={DEMAND_LIMITS.maxHeadcount}
                step="1"
                value={draft.headcount}
                onChange={handleChange}
                required
              />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="demand-hours">{t('Budget hours')}</Label>
              <Input
                placeholder={t('e.g. 320')}
                id="demand-hours"
                name="total_hours"
                type="number"
                min={0}
                max={DEMAND_LIMITS.maxHours}
                step={DEMAND_LIMITS.hoursStep}
                value={draft.total_hours}
                aria-invalid={hoursOff}
                aria-describedby={hoursOff ? 'demand-hours-hint' : undefined}
                onChange={handleChange}
                required
              />
              {hoursOff && (
                <p id="demand-hours-hint" className="text-sm text-destructive">
                  {t('Use whole hours')}
                </p>
              )}
            </div>
            <div className="flex flex-col gap-2">
              <span className="text-sm font-medium">{t('Weeks')}</span>
              <span className="flex h-8 items-center text-sm text-muted-foreground">
                {t('{start} – {end}', { start: weeks.from, end: weeks.to })}
              </span>
            </div>
          </div>
          <DialogFooter>
            <Button type="submit" disabled={!isValidDraft(draft)} isLoading={saving}>
              {t('Save demand')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
