import { type ChangeEvent, type FormEvent, useState } from 'react'
import { toast } from 'sonner'
import { DatePicker } from '@/shared/components/date-picker'
import { ROLE_LABELS, ROLES } from '@/resources/rbac'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/shared/ui/dialog'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { NativeSelect, NativeSelectOption } from '@/shared/ui/native-select'
import { useExpertises } from '@/shared/expertises'
import { useSavePerson } from '../../use-people'
import { SkillsEditor } from '../skills-editor'
import { ACTIVE_OPTION } from './constants'
import type { PersonFormDialogProps, PersonFormValues, PersonTextField } from './types'
import { canSave, initialSkills, initialValues, isRole, skillsChanged, toChanges, toCreateInput } from './utils'

/** Adds a person, or edits one: details, dates and skills. The email of a saved person is fixed. */
export function PersonFormDialog({ open, person, onOpenChange }: PersonFormDialogProps) {
  const { t } = useTranslation()
  const [values, setValues] = useState<PersonFormValues>(() => initialValues(person))
  const [skills, setSkills] = useState(() => initialSkills(person))
  const { data: expertises = [] } = useExpertises()
  const save = useSavePerson()
  const roles = Object.values(ROLES)

  const handleTextChange = (event: ChangeEvent<HTMLInputElement>) => {
    const field = event.target.name as PersonTextField
    const { value } = event.target
    setValues((current) => ({ ...current, [field]: value }))
  }
  const handleDateChange = (value: string, name: string) => setValues((current) => ({ ...current, [name as PersonTextField]: value }))
  const handleRoleChange = (event: ChangeEvent<HTMLSelectElement>) => {
    const { value } = event.target
    if (isRole(value, roles)) setValues((current) => ({ ...current, role: value }))
  }
  const handleActiveChange = (event: ChangeEvent<HTMLSelectElement>) => {
    const active = event.target.value === ACTIVE_OPTION.yes
    setValues((current) => ({ ...current, active }))
  }
  const handleSaved = () => {
    toast.success(t('Person saved.'))
    onOpenChange(false)
  }
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (person) {
      const changedSkills = skillsChanged(initialSkills(person), skills) ? skills : null
      save.mutate({ mode: 'update', id: person.id, changes: toChanges(person, values), skills: changedSkills }, { onSuccess: handleSaved })
    } else {
      save.mutate({ mode: 'create', person: toCreateInput(values, skills) }, { onSuccess: handleSaved })
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-2xl">
        <form className="flex flex-col gap-4" onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>{person ? t('Edit person') : t('Add person')}</DialogTitle>
            <DialogDescription>{t('Exactly one skill is primary: the expertise this person is staffed for first.')}</DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="flex flex-col gap-2">
              <Label htmlFor="person-name">{t('Name')}</Label>
              <Input id="person-name" name="name" value={values.name} onChange={handleTextChange} required />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="person-email">{t('Email')}</Label>
              <Input id="person-email" name="email" type="email" value={values.email} onChange={handleTextChange}
                disabled={person !== null} required />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="person-role">{t('Role')}</Label>
              <NativeSelect id="person-role" className="w-full" value={values.role} onChange={handleRoleChange}>
                {roles.map((role) => (
                  <NativeSelectOption key={role} value={role}>
                    {t(ROLE_LABELS[role])}
                  </NativeSelectOption>
                ))}
              </NativeSelect>
            </div>
            {person && (
              <div className="flex flex-col gap-2">
                <Label htmlFor="person-active">{t('Status')}</Label>
                <NativeSelect id="person-active" className="w-full" value={values.active ? ACTIVE_OPTION.yes : ACTIVE_OPTION.no}
                  onChange={handleActiveChange}>
                  <NativeSelectOption value={ACTIVE_OPTION.yes}>{t('Active')}</NativeSelectOption>
                  <NativeSelectOption value={ACTIVE_OPTION.no}>{t('Inactive')}</NativeSelectOption>
                </NativeSelect>
              </div>
            )}
            <div className="flex flex-col gap-2">
              <Label htmlFor="person-weekly-hours">{t('Weekly hours')}</Label>
              <Input id="person-weekly-hours" name="weekly_hours" type="number" min={1} step="0.5"
                value={values.weekly_hours} onChange={handleTextChange} required />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="person-non-project-hours">{t('Non-project hours')}</Label>
              <Input id="person-non-project-hours" name="non_project_hours" type="number" min={0} step="0.5"
                placeholder={t('Default')} value={values.non_project_hours} onChange={handleTextChange} />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="person-start-date">{t('Start date')}</Label>
              <DatePicker id="person-start-date" name="start_date" value={values.start_date} onChange={handleDateChange} />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="person-end-date">{t('End date')}</Label>
              <DatePicker id="person-end-date" name="end_date" min={values.start_date} value={values.end_date} onChange={handleDateChange} />
            </div>
          </div>
          <SkillsEditor skills={skills} expertises={expertises} onChange={setSkills} />
          <DialogFooter>
            <Button type="submit" disabled={!canSave(values, skills) || save.isPending}>
              {t('Save')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
