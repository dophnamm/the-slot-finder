import { UserPlus } from 'lucide-react'
import { type FormEvent, useState } from 'react'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/shared/ui/dialog'
import { Label } from '@/shared/ui/label'
import { Select } from '@/shared/ui/select'
import { TableCell, TableRow } from '@/shared/ui/table'
import { cn } from '@/utils/functions'
import { inWindow } from '../../utils'
import {
  GROUP_LABEL_CELL_CLASS,
  GROUP_ROW_CLASS,
  PERSON_COLUMN_CLASS,
  WEEK_COLUMN_CLASS,
  WINDOW_BAR_CLASS,
} from '../../constants'
import type { GroupHeaderRowProps } from './types'
import { personOptions } from './utils'

/**
 * The demand above its assignments: what it needs, its weeks as a bar, and an "Add person" button for
 * planners that opens a dialog to pick the person.
 */
export function GroupHeaderRow({ group, weeks, people, peopleLoading, editable, busy, onAdd }: GroupHeaderRowProps) {
  const { t } = useTranslation()
  const [personId, setPersonId] = useState('')
  const [open, setOpen] = useState(false)
  const { demand, assignments } = group
  const label = t('{project} · {expertise} · {phase}', {
    project: demand.project,
    expertise: demand.expertise,
    phase: demand.phase ?? t('Whole project'),
  })
  const detail = t('{start} – {end} · level {level}+ · {assigned} of {hours} hrs', {
    start: demand.from_week,
    end: demand.to_week,
    level: demand.min_level,
    assigned: demand.assigned_hours,
    hours: demand.hours,
  })
  const options = personOptions(
    people,
    demand,
    assignments.map((assignment) => assignment.person.id),
  )

  const personSelectOptions = [
    { value: '', label: t('Choose a person') },
    ...options.map((option) => ({
      value: option.id,
      label: option.stretch ? t('{name} (stretch)', { name: option.name }) : option.name,
    })),
  ]
  const handleOpen = () => setOpen(true)
  const handleOpenChange = (next: boolean) => {
    setOpen(next)
    if (!next) setPersonId('')
  }
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    onAdd(demand.id, personId)
    handleOpenChange(false)
  }

  return (
    <TableRow className={GROUP_ROW_CLASS}>
      <TableCell className={cn(PERSON_COLUMN_CLASS, GROUP_LABEL_CELL_CLASS)}>
        <div className="flex items-center gap-2">
          <div className="flex min-w-0 flex-1 flex-col">
            <span className="truncate font-medium" title={label}>
              {label}
            </span>
            <span className="truncate text-xs text-muted-foreground" title={detail}>
              {detail}
            </span>
          </div>
          {editable && (
            <Button variant="outline" size="sm" className="shrink-0" disabled={busy} onClick={handleOpen}>
              <UserPlus />
              {t('Add person')}
            </Button>
          )}
          <Dialog open={open} onOpenChange={handleOpenChange}>
            <DialogContent className="sm:max-w-md">
              <form className="flex flex-col gap-5" onSubmit={handleSubmit}>
                <DialogHeader>
                  <DialogTitle>{t('Add person')}</DialogTitle>
                  <DialogDescription>
                    {t('{project} · {expertise} · {phase}', {
                      project: demand.project,
                      expertise: demand.expertise,
                      phase: demand.phase ?? t('Whole project'),
                    })}
                  </DialogDescription>
                </DialogHeader>
                <div className="flex flex-col gap-2">
                  <Label htmlFor={`add-person-${demand.id}`}>{t('Person')}</Label>
                  <Select
                    id={`add-person-${demand.id}`}
                    className="w-full"
                    options={personSelectOptions}
                    value={personId}
                    isLoading={peopleLoading}
                    onValueChange={setPersonId}
                    required
                  />
                  <p className="text-sm text-muted-foreground">
                    {t(
                      'Only the program team is listed. People with the expertise at the minimum level come first; the others would be stretch.',
                    )}
                  </p>
                </div>
                <DialogFooter>
                  <Button type="submit" disabled={!personId} isLoading={busy}>
                    {t('Add person')}
                  </Button>
                </DialogFooter>
              </form>
            </DialogContent>
          </Dialog>
        </div>
      </TableCell>
      {weeks.map((week) => (
        <TableCell key={week} className={cn(WEEK_COLUMN_CLASS, 'px-0')}>
          {inWindow(week, demand.from_week, demand.to_week) && <div className={WINDOW_BAR_CLASS} />}
        </TableCell>
      ))}
    </TableRow>
  )
}
