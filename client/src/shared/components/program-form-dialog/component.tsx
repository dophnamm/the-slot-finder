import { type ChangeEvent, type FormEvent, useState } from 'react'
import { toast } from 'sonner'
import { can } from '@/resources/rbac'
import { useCurrentUser } from '@/shared/auth'
import { LeadSelect } from '@/shared/components/lead-select'
import { type Program, useSaveProgram } from '@/shared/programs'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/shared/ui/dialog'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { Textarea } from '@/shared/ui/textarea'
import type { ProgramFormDialogProps, ProgramFormValues, ProgramTextField } from './types'
import { canSave, initialValues, toChanges, toInput } from './utils'

/** Creates a program, or edits one. Planners choose the lead; a lead who creates one leads it. */
export function ProgramFormDialog({ open, program, onOpenChange, onSaved }: ProgramFormDialogProps) {
  const { t } = useTranslation()
  const { data: user } = useCurrentUser()
  const [values, setValues] = useState<ProgramFormValues>(() => initialValues(program))
  const save = useSaveProgram(program?.id)
  const canChooseLead = can(user?.role, 'programs:manage-all')

  const handleChange = (event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const field = event.target.name as ProgramTextField
    const { value } = event.target
    setValues((current) => ({ ...current, [field]: value }))
  }
  const handleLeadChange = (leadId: string) => setValues((current) => ({ ...current, lead_id: leadId }))
  const handleSaved = (saved: Program) => {
    toast.success(t('Program saved.'))
    onOpenChange(false)
    onSaved?.(saved)
  }
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const input = program ? toChanges(program, values, canChooseLead) : toInput(values, canChooseLead)
    save.mutate(input, { onSuccess: handleSaved })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-2xl">
        <form className="flex flex-col gap-4" onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>{program ? t('Edit program') : t('New program')}</DialogTitle>
            <DialogDescription>{t('A customer engagement that groups projects under one timeline.')}</DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="flex flex-col gap-2">
              <Label htmlFor="program-name">{t('Name')}</Label>
              <Input id="program-name" name="name" value={values.name} onChange={handleChange} required />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="program-client">{t('Client')}</Label>
              <Input id="program-client" name="client" value={values.client} onChange={handleChange} required />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="program-start">{t('Start date')}</Label>
              <Input id="program-start" name="start_date" type="date" value={values.start_date} onChange={handleChange} required />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="program-end">{t('End date')}</Label>
              <Input id="program-end" name="end_date" type="date" min={values.start_date} value={values.end_date}
                onChange={handleChange} required />
            </div>
            {canChooseLead && (
              <div className="flex flex-col gap-2">
                <Label htmlFor="program-lead">{t('Lead')}</Label>
                <LeadSelect id="program-lead" value={values.lead_id} onChange={handleLeadChange} />
              </div>
            )}
            <div className="flex flex-col gap-2">
              <Label htmlFor="program-charter">{t('Charter link')}</Label>
              <Input id="program-charter" name="charter_url" type="url" value={values.charter_url} onChange={handleChange} />
            </div>
          </div>
          <div className="flex flex-col gap-2">
            <Label htmlFor="program-goal">{t('Goal')}</Label>
            <Textarea id="program-goal" name="goal" value={values.goal} onChange={handleChange} />
          </div>
          <div className="flex flex-col gap-2">
            <Label htmlFor="program-objectives">{t('Objectives, one per line')}</Label>
            <Textarea id="program-objectives" name="objectives" value={values.objectives} onChange={handleChange} />
          </div>
          <DialogFooter>
            <Button type="submit" disabled={!canSave(values) || save.isPending}>
              {t('Save')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
