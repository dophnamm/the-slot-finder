import { Fragment } from 'react'
import { useTranslation } from '@/shared/i18n'
import { type Assignment, useEditDraft } from '@/shared/staffing'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/ui/card'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { formatIsoWeekRange, shortWeekLabel } from '@/utils/functions'
import { useStaffablePeople } from '../../use-staffable-people'
import { AssignmentRow } from './components/assignment-row'
import { GroupHeaderRow } from './components/group-header-row'
import { PERSON_COLUMN_CLASS, WEEK_COLUMN_CLASS } from './constants'
import type { AssignmentGridProps } from './types'
import { groupByDemand, lockChanges } from './utils'

/** Person × week hours, grouped by demand; planners edit hours, lock weeks, add and remove people. */
export function AssignmentGrid({ draft, editable }: AssignmentGridProps) {
  const { t } = useTranslation()
  const edit = useEditDraft(draft.id)
  const { data: people = [] } = useStaffablePeople(editable)
  const groups = groupByDemand(draft)

  const handleSetHours = (assignment: Assignment, week: string, hours: number) =>
    edit.mutate({ set: [{ assignment_id: assignment.id, week, hours }] })
  const handleToggleLock = (assignment: Assignment) => edit.mutate({ set: lockChanges(assignment) })
  const handleRemove = (assignment: Assignment) => edit.mutate({ remove: [assignment.id] })
  const handleAdd = (demandId: string, personId: string) =>
    edit.mutate({ add: [{ person_id: personId, demand_id: demandId, allocations: [] }] })

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('Assignments')}</CardTitle>
        <CardDescription>
          {t('Hours per ISO week. Locked hours stay as they are when a draft is generated again; stretch is always locked.')}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className={PERSON_COLUMN_CLASS}>{t('Person')}</TableHead>
              {draft.weeks.map((week) => (
                <TableHead key={week} className={WEEK_COLUMN_CLASS}>
                  <div className="flex flex-col leading-tight">
                    <span>{shortWeekLabel(week)}</span>
                    <span className="text-xs font-normal text-muted-foreground">{formatIsoWeekRange(week)}</span>
                  </div>
                </TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {groups.map((group) => (
              <Fragment key={group.demand.id}>
                <GroupHeaderRow group={group} weeks={draft.weeks} people={people} editable={editable}
                  busy={edit.isPending} onAdd={handleAdd} />
                {group.assignments.map((assignment) => (
                  <AssignmentRow key={assignment.id} assignment={assignment} weeks={draft.weeks} editable={editable}
                    busy={edit.isPending} onSetHours={handleSetHours} onToggleLock={handleToggleLock} onRemove={handleRemove} />
                ))}
              </Fragment>
            ))}
          </TableBody>
        </Table>
        {draft.assignments.length === 0 && (
          <p className="py-6 text-center text-sm text-muted-foreground">{t('Nobody is assigned yet.')}</p>
        )}
      </CardContent>
    </Card>
  )
}
