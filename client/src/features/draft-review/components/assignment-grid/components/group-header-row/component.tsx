import { UserPlus } from 'lucide-react'
import { type ChangeEvent, useState } from 'react'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { NativeSelect, NativeSelectOption } from '@/shared/ui/native-select'
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

/** The demand above its assignments: what it needs, its weeks as a bar, and "add person" for planners. */
export function GroupHeaderRow({ group, weeks, people, editable, busy, onAdd }: GroupHeaderRowProps) {
  const { t } = useTranslation()
  const [personId, setPersonId] = useState('')
  const { demand, assignments } = group
  const options = personOptions(
    people,
    demand,
    assignments.map((assignment) => assignment.person.id),
  )

  const handleChange = (event: ChangeEvent<HTMLSelectElement>) => setPersonId(event.target.value)
  const handleAdd = () => {
    onAdd(demand.id, personId)
    setPersonId('')
  }

  return (
    <TableRow className={GROUP_ROW_CLASS}>
      <TableCell className={cn(PERSON_COLUMN_CLASS, GROUP_LABEL_CELL_CLASS)}>
        <div className="flex flex-col gap-2">
          <div className="flex flex-col">
            <span className="font-medium">
              {t('{project} · {expertise} · {phase}', {
                project: demand.project,
                expertise: demand.expertise,
                phase: demand.phase ?? t('Whole project'),
              })}
            </span>
            <span className="text-xs text-muted-foreground">
              {t('{start} – {end} · level {level}+ · {assigned} of {hours} h', {
                start: demand.from_week,
                end: demand.to_week,
                level: demand.min_level,
                assigned: demand.assigned_hours,
                hours: demand.hours,
              })}
            </span>
          </div>
          {editable && (
            <div className="flex items-center gap-2">
              <NativeSelect
                className="min-w-0 flex-1 bg-card"
                aria-label={t('Person to add')}
                value={personId}
                onChange={handleChange}
              >
                <NativeSelectOption value="">{t('Choose a person')}</NativeSelectOption>
                {options.map((option) => (
                  <NativeSelectOption key={option.id} value={option.id}>
                    {option.stretch ? t('{name} (stretch)', { name: option.name }) : option.name}
                  </NativeSelectOption>
                ))}
              </NativeSelect>
              <Button variant="outline" size="sm" disabled={!personId || busy} onClick={handleAdd}>
                <UserPlus />
                {t('Add person')}
              </Button>
            </div>
          )}
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
