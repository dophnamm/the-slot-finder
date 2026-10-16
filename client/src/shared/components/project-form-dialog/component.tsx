import { type ChangeEvent, type FormEvent, useState } from 'react'
import { toast } from 'sonner'
import { can } from '@/resources/rbac'
import { useCurrentUser } from '@/shared/auth'
import { LeadSelect } from '@/shared/components/lead-select'
import { DatePicker } from '@/shared/components/date-picker'
import { useSaveProject } from '@/shared/programs'
import { useTranslation } from '@/shared/i18n'
import { useReplanToast } from '@/shared/staffing'
import { Button } from '@/shared/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/shared/ui/dialog'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import type { ProjectFormDialogProps, ProjectFormValues, ProjectTextField } from './types'
import { canSave, initialValues, toChanges, toInput } from './utils'

/** Adds a project to a program, or edits one, including its scope points. */
export function ProjectFormDialog({ open, programId, project, onOpenChange }: ProjectFormDialogProps) {
  const { t } = useTranslation()
  const replanToast = useReplanToast()
  const { data: user } = useCurrentUser()
  const [values, setValues] = useState<ProjectFormValues>(() => initialValues(project))
  const save = useSaveProject(programId, project?.id)
  const canChooseLead = can(user?.role, 'programs:manage-all')

  const handleChange = (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const field = event.target.name as ProjectTextField
    const { value } = event.target
    setValues((current) => ({ ...current, [field]: value }))
  }
  const handleDateChange = (value: string, name: string) =>
    setValues((current) => ({ ...current, [name as ProjectTextField]: value }))
  const handleLeadChange = (leadId: string) => setValues((current) => ({ ...current, lead_id: leadId }))
  const handleSaved = () => {
    // A moved deadline triggers a re-plan; other edits do not.
    if (project && values.deadline !== project.deadline) replanToast(t('Project saved.'))
    else toast.success(t('Project saved.'))
    onOpenChange(false)
  }
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const input = project ? toChanges(project, values, canChooseLead) : toInput(values, canChooseLead)
    save.mutate(input, { onSuccess: handleSaved })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-xl">
        <form className="flex flex-col gap-4" onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>{project ? t('Edit project') : t('Add project')}</DialogTitle>
            <DialogDescription>
              {t('A new project starts as incoming until a staffing draft is accepted.')}
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="flex flex-col gap-2 sm:col-span-2">
              <Label htmlFor="project-name">{t('Name')}</Label>
              <Input
                placeholder={t('e.g. Checkout Revamp')}
                id="project-name"
                name="name"
                value={values.name}
                onChange={handleChange}
                required
              />
            </div>
            {canChooseLead && (
              <div className="flex flex-col gap-2">
                <Label htmlFor="project-lead">{t('Lead')}</Label>
                <LeadSelect id="project-lead" value={values.lead_id} onChange={handleLeadChange} />
              </div>
            )}
            <div className="flex flex-col gap-2">
              <Label htmlFor="project-start">{t('Earliest start')}</Label>
              <DatePicker
                id="project-start"
                name="start_earliest"
                value={values.start_earliest}
                onChange={handleDateChange}
              />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="project-deadline">{t('Deadline')}</Label>
              <DatePicker
                id="project-deadline"
                name="deadline"
                min={values.start_earliest}
                value={values.deadline}
                onChange={handleDateChange}
              />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="project-scope">{t('Scope points')}</Label>
              <Input
                placeholder={t('e.g. 120')}
                id="project-scope"
                name="scope_points"
                type="number"
                min={0}
                step="1"
                value={values.scope_points}
                onChange={handleChange}
              />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="project-done">{t('Done points')}</Label>
              <Input
                placeholder={t('e.g. 0')}
                id="project-done"
                name="done_points"
                type="number"
                min={0}
                step="1"
                value={values.done_points}
                onChange={handleChange}
              />
            </div>
          </div>
          <DialogFooter>
            <Button type="submit" disabled={!canSave(values)} isLoading={save.isPending}>
              {t('Save')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
