import { ClipboardList, Plus } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'
import { useReplaceDemands } from '@/shared/programs'
import { Section } from '@/shared/components/section'
import { TableEmpty } from '@/shared/components/table-empty'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { DEMAND_EDIT_COLUMNS, DEMAND_TABLE_COLUMNS } from './constants'
import { DemandDialog } from './components/demand-dialog'
import { DemandViewRow } from './components/demand-view-row'
import type { DemandDraft, DemandEdit, DemandEditorProps } from './types'
import { newDraft, phaseName, toDemandInput, toDrafts } from './utils'

/**
 * The project's demand as a table. Adding or editing one opens a dialog, and saving it saves the whole
 * list at once, so every change shows in the table right away.
 */
export function DemandEditor({ project, expertises, expertisesLoading, editable }: DemandEditorProps) {
  const { t } = useTranslation()
  const [editing, setEditing] = useState<DemandEdit | null>(null)
  const save = useReplaceDemands(project.id)
  const saved = toDrafts(project.demands)

  const handleAdd = () => setEditing({ index: null, draft: newDraft(expertises) })
  const handleEdit = (index: number) => setEditing({ index, draft: saved[index] })
  const handleOpenChange = (open: boolean) => {
    if (!open) setEditing(null)
  }
  const handleSaved = () => {
    toast.success(t('Demand saved.'))
    setEditing(null)
  }
  const handleSubmit = (draft: DemandDraft) => {
    const index = editing?.index ?? null
    const next = index === null ? [...saved, draft] : saved.map((row, i) => (i === index ? draft : row))
    save.mutate(next.map(toDemandInput), { onSuccess: handleSaved })
  }
  const handleRemoved = () => toast.success(t('Demand removed.'))
  const handleRemove = (index: number) => {
    if (!window.confirm(t('Remove this demand?'))) return
    save.mutate(saved.filter((_, i) => i !== index).map(toDemandInput), { onSuccess: handleRemoved })
  }

  return (
    <Section
      title={t('Demand')}
      description={t(
        'Budget hours per expertise and minimum level. Without a phase, a demand spans the whole project.',
      )}
      action={
        editable && (
          <Button variant="outline" disabled={expertises.length === 0} onClick={handleAdd}>
            <Plus />
            {t('Add demand')}
          </Button>
        )
      }
    >
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t('Expertise')}</TableHead>
            <TableHead>{t('Phase')}</TableHead>
            <TableHead>{t('Minimum level')}</TableHead>
            <TableHead>{t('Headcount')}</TableHead>
            <TableHead>{t('Budget hours')}</TableHead>
            <TableHead>{t('Weeks')}</TableHead>
            {editable && <TableHead />}
          </TableRow>
        </TableHeader>
        <TableBody>
          {project.demands.map((demand, index) => (
            <DemandViewRow
              key={demand.id}
              index={index}
              demand={demand}
              phase={phaseName(project, demand.phase_id)}
              onEdit={editable ? handleEdit : undefined}
              onRemove={editable ? handleRemove : undefined}
            />
          ))}
          {project.demands.length === 0 && (
            <TableEmpty
              columns={editable ? DEMAND_EDIT_COLUMNS : DEMAND_TABLE_COLUMNS}
              icon={ClipboardList}
              title={t('No demand yet.')}
              description={t('Add the hours each expertise needs, in a phase or across the whole project.')}
            />
          )}
        </TableBody>
      </Table>
      {editing && (
        <DemandDialog
          key={editing.draft.key}
          title={editing.index === null ? t('Add demand') : t('Edit demand')}
          initial={editing.draft}
          expertises={expertises}
          expertisesLoading={expertisesLoading}
          project={project}
          saving={save.isPending}
          onSubmit={handleSubmit}
          onOpenChange={handleOpenChange}
        />
      )}
    </Section>
  )
}
