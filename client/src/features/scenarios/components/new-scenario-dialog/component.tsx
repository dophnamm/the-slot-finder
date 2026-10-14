import { Plus } from 'lucide-react'
import { type ChangeEvent, type FormEvent, useState } from 'react'
import { generatePath, useNavigate } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { useTranslation } from '@/shared/i18n'
import { useOpenProjects } from '@/shared/programs'
import { type ScenarioDetail, useCreateScenario } from '@/shared/scenarios'
import { Button } from '@/shared/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/shared/ui/dialog'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { usePeopleOptions } from '../../use-people-options'
import { AssumptionRow } from './components/assumption-row'
import type { AssumptionDraft, NewScenarioDialogProps } from './types'
import { canCreate, newAssumption, toInput } from './utils'

/** Names a what-if and lists its assumptions; Create computes it and opens the result. */
export function NewScenarioDialog({ open, onOpenChange }: NewScenarioDialogProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const create = useCreateScenario()
  const { data: projects = [], isPending: projectsLoading } = useOpenProjects()
  const { data: people = [], isPending: peopleLoading } = usePeopleOptions()
  const [name, setName] = useState('')
  const [drafts, setDrafts] = useState<AssumptionDraft[]>(() => [newAssumption()])

  const handleNameChange = (event: ChangeEvent<HTMLInputElement>) => setName(event.target.value)
  const handleAdd = () => setDrafts((current) => [...current, newAssumption()])
  const handleChange = (next: AssumptionDraft) =>
    setDrafts((current) => current.map((draft) => (draft.key === next.key ? next : draft)))
  const handleRemove = (key: string) => setDrafts((current) => current.filter((draft) => draft.key !== key))
  const handleCreated = (scenario: ScenarioDetail) => {
    onOpenChange(false)
    navigate(generatePath(ROUTES.scenarioDetail, { id: scenario.id }))
  }
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    create.mutate(toInput(name, drafts), { onSuccess: handleCreated })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-3xl">
        <form className="flex flex-col gap-4" onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>{t('New scenario')}</DialogTitle>
            <DialogDescription>
              {t('Each assumption is applied to a copy of the baseline. Nothing real changes.')}
            </DialogDescription>
          </DialogHeader>
          <div className="flex flex-col gap-2">
            <Label htmlFor="scenario-name">{t('Name')}</Label>
            <Input
              id="scenario-name"
              value={name}
              placeholder={t('e.g. Checkout slips two weeks')}
              onChange={handleNameChange}
            />
          </div>
          <div className="flex flex-col gap-3">
            <span className="text-sm font-medium">{t('Assumptions')}</span>
            {drafts.map((draft) => (
              <AssumptionRow
                key={draft.key}
                draft={draft}
                projects={projects}
                people={people}
                projectsLoading={projectsLoading}
                peopleLoading={peopleLoading}
                removable={drafts.length > 1}
                onChange={handleChange}
                onRemove={handleRemove}
              />
            ))}
            <Button type="button" variant="outline" size="sm" className="self-start" onClick={handleAdd}>
              <Plus />
              {t('Add assumption')}
            </Button>
          </div>
          <DialogFooter>
            <Button type="submit" disabled={!canCreate(name, drafts)} isLoading={create.isPending}>
              {t('Create scenario')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
