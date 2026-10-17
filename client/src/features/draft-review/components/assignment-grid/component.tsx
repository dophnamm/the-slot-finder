import { UserX } from 'lucide-react'
import { Fragment, useState } from 'react'
import { toast } from 'sonner'
import { Section } from '@/shared/components/section'
import { TableEmpty } from '@/shared/components/table-empty'
import { useTranslation } from '@/shared/i18n'
import { type Assignment, type DraftEdit, useEditDraft } from '@/shared/staffing'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { cn, formatIsoWeekRange, shortWeekLabel } from '@/utils/functions'
import { useStaffablePeople } from '../../use-staffable-people'
import { AllocateHoursDialog } from './components/allocate-hours-dialog'
import { AssignmentRow } from './components/assignment-row'
import { GroupHeaderRow } from './components/group-header-row'
import { SwapPersonDialog } from './components/swap-person-dialog'
import { LABEL_COLUMNS, PERSON_COLUMN_CLASS, WEEK_COLUMN_CLASS } from './constants'
import type { AssignmentGridProps, DragSource } from './types'
import { groupByDemand, inWindow, lockChanges, moveHours, weekLoads } from './utils'

/** Person × week hours, grouped by demand; planners edit hours, lock weeks, add and remove people. */
export function AssignmentGrid({ draft, editable }: AssignmentGridProps) {
  const { t } = useTranslation()
  const edit = useEditDraft(draft.id)
  const { data: people = [], isPending: peopleLoading } = useStaffablePeople(editable)
  const groups = groupByDemand(draft)
  const [allocating, setAllocating] = useState<Assignment | null>(null)
  const [drag, setDrag] = useState<DragSource | null>(null)
  const [swapping, setSwapping] = useState<Assignment | null>(null)
  const allocateWeeks = allocating
    ? draft.weeks.filter((week) => inWindow(week, allocating.demand.from_week, allocating.demand.to_week))
    : []

  const handleSetHours = (assignment: Assignment, week: string, hours: number) =>
    edit.mutate({ set: [{ assignment_id: assignment.id, week, hours }] })
  const handleToggleLock = (assignment: Assignment) => edit.mutate({ set: lockChanges(assignment) })
  const handleRemove = (assignment: Assignment) => edit.mutate({ remove: [assignment.id] })
  const handleAllocate = (assignment: Assignment) => setAllocating(assignment)
  const handleAllocateCancel = () => setAllocating(null)
  const handleAllocateSave = (changes: NonNullable<DraftEdit['set']>) =>
    edit.mutate(
      { set: changes },
      {
        onSuccess: () => {
          setAllocating(null)
          toast.success(t('Hours allocated.'))
        },
      },
    )
  const handleDragStart = (source: DragSource) => setDrag(source)
  const handleDragEnd = () => setDrag(null)
  const handleDropHours = (target: DragSource) => {
    const changes = drag ? moveHours(drag, target) : null
    setDrag(null)
    if (changes) edit.mutate({ set: changes })
  }
  const handleSwap = (assignment: Assignment) => setSwapping(assignment)
  const handleSwapCancel = () => setSwapping(null)
  const handleSwapChoose = (personId: string) => {
    if (!swapping) return
    const from = swapping.person.name
    edit.mutate(
      {
        remove: [swapping.id],
        add: [
          {
            person_id: personId,
            demand_id: swapping.demand.id,
            allocations: swapping.allocations.map(({ week, hours }) => ({ week, hours })),
          },
        ],
      },
      {
        onSuccess: (updated) => {
          const to = updated.assignments.find((a) => a.demand.id === swapping.demand.id && a.person.id === personId)
          setSwapping(null)
          toast.success(t('Swapped {from} for {to}.', { from, to: to?.person.name ?? '' }))
        },
      },
    )
  }
  const handleAdd = (demandId: string, personId: string) =>
    edit.mutate({ add: [{ person_id: personId, demand_id: demandId, allocations: [] }] })

  return (
    <Section
      title={t('Assignments')}
      description={t(
        'Hours per week. Locked hours stay as they are when a draft is generated again; stretch is always locked. A week is coloured by the whole load of the person, as on the heatmap.',
      )}
    >
      {editable && <p className="text-xs text-muted-foreground">{t('Drag hours to another week or person.')}</p>}
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className={cn(PERSON_COLUMN_CLASS, 'bg-muted')}>{t('Person')}</TableHead>
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
              <GroupHeaderRow
                group={group}
                weeks={draft.weeks}
                people={people}
                peopleLoading={peopleLoading}
                editable={editable}
                busy={edit.isPending}
                onAdd={handleAdd}
              />
              {group.assignments.map((assignment) => (
                <AssignmentRow
                  key={assignment.id}
                  assignment={assignment}
                  weeks={draft.weeks}
                  loads={weekLoads(draft, assignment)}
                  bands={draft.bands}
                  editable={editable}
                  drag={drag}
                  onDragStart={handleDragStart}
                  onDragEnd={handleDragEnd}
                  onDropHours={handleDropHours}
                  busy={edit.isPending}
                  onSetHours={handleSetHours}
                  onToggleLock={handleToggleLock}
                  onRemove={handleRemove}
                  onAllocate={handleAllocate}
                  onSwap={handleSwap}
                />
              ))}
            </Fragment>
          ))}
          {draft.assignments.length === 0 && (
            <TableEmpty
              columns={draft.weeks.length + LABEL_COLUMNS}
              icon={UserX}
              title={t('Nobody is assigned yet.')}
            />
          )}
        </TableBody>
      </Table>
      {allocating && (
        <AllocateHoursDialog
          assignment={allocating}
          weeks={allocateWeeks}
          busy={edit.isPending}
          onSave={handleAllocateSave}
          onCancel={handleAllocateCancel}
        />
      )}
      {swapping && (
        <SwapPersonDialog
          draftId={draft.id}
          assignment={swapping}
          busy={edit.isPending}
          onChoose={handleSwapChoose}
          onCancel={handleSwapCancel}
        />
      )}
    </Section>
  )
}
