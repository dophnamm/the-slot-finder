import { type ChangeEvent, type FormEvent, useState } from 'react'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/shared/ui/dialog'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { NativeSelect, NativeSelectOption } from '@/shared/ui/native-select'
import { SKILL_LEVELS } from '@/utils/constants'
import { DEMAND_LIMITS } from '../../constants'
import { demandWindow, isValidDraft } from '../../utils'
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
  const weeks = demandWindow(project, draft.phase_id || null)

  const handleChange = (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const field = event.target.name as DemandTextField
    const { value } = event.target
    setDraft((current) => ({ ...current, [field]: value }))
  }
  const handleLevelChange = (event: ChangeEvent<HTMLSelectElement>) => {
    const level = Number(event.target.value)
    setDraft((current) => ({ ...current, min_level: level }))
  }
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
              <NativeSelect
                id="demand-expertise"
                name="expertise_id"
                className="w-full"
                value={draft.expertise_id}
                isLoading={expertisesLoading}
                onChange={handleChange}
              >
                {expertises.map((expertise) => (
                  <NativeSelectOption key={expertise.id} value={expertise.id}>
                    {expertise.name}
                  </NativeSelectOption>
                ))}
              </NativeSelect>
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="demand-phase">{t('Phase')}</Label>
              <NativeSelect
                id="demand-phase"
                name="phase_id"
                className="w-full"
                value={draft.phase_id}
                onChange={handleChange}
              >
                <NativeSelectOption value="">{t('Whole project')}</NativeSelectOption>
                {project.phases.map((phase) => (
                  <NativeSelectOption key={phase.id} value={phase.id}>
                    {phase.name}
                  </NativeSelectOption>
                ))}
              </NativeSelect>
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="demand-level">{t('Minimum level')}</Label>
              <NativeSelect
                id="demand-level"
                className="w-full"
                value={String(draft.min_level)}
                onChange={handleLevelChange}
              >
                {SKILL_LEVELS.map((level) => (
                  <NativeSelectOption key={level} value={String(level)}>
                    {t('Level {level}', { level })}
                  </NativeSelectOption>
                ))}
              </NativeSelect>
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
                step="any"
                value={draft.total_hours}
                onChange={handleChange}
                required
              />
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
