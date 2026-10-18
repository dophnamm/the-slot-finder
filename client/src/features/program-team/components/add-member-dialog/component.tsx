import { type FormEvent, useState } from 'react'
import { toast } from 'sonner'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/shared/ui/dialog'
import { Label } from '@/shared/ui/label'
import { Select } from '@/shared/ui/select'
import { useAddTeamMember, useTeamCandidates } from '../../use-program-team'
import type { AddMemberDialogProps } from './types'
import { candidateOptions } from './utils'

/** Picks a person to put in the program's team (planners only). */
export function AddMemberDialog({ open, programId, teamIds, onOpenChange }: AddMemberDialogProps) {
  const { t } = useTranslation()
  const [personId, setPersonId] = useState('')
  const { data: people = [], isPending } = useTeamCandidates(open)
  const add = useAddTeamMember(programId)
  const options = candidateOptions(people, teamIds)
  const selectOptions = [
    { value: '', label: t('Choose a person') },
    ...options.map((option) => ({
      value: option.id,
      label: option.expertise
        ? t('{name} · {expertise}', { name: option.name, expertise: option.expertise })
        : option.name,
    })),
  ]

  const handleOpenChange = (next: boolean) => {
    onOpenChange(next)
    if (!next) setPersonId('')
  }
  const handleAdded = () => {
    toast.success(t('Member added.'))
    handleOpenChange(false)
  }
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    add.mutate(personId, { onSuccess: handleAdded })
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <form className="flex flex-col gap-5" onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>{t('Add member')}</DialogTitle>
            <DialogDescription>
              {t('Drafts for this program’s projects staff only people in its team.')}
            </DialogDescription>
          </DialogHeader>
          <div className="flex flex-col gap-2">
            <Label htmlFor="team-member">{t('Person')}</Label>
            <Select
              id="team-member"
              className="w-full"
              options={selectOptions}
              value={personId}
              isLoading={isPending}
              onValueChange={setPersonId}
              required
            />
            {!isPending && options.length === 0 && (
              <p className="text-sm text-muted-foreground">{t('Everyone is already in this team.')}</p>
            )}
          </div>
          <DialogFooter>
            <Button type="submit" disabled={!personId} isLoading={add.isPending}>
              {t('Add member')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
