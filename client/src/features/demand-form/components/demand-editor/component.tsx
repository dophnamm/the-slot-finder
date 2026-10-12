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
import { DemandRow } from './components/demand-row'
import { DemandViewRow } from './components/demand-view-row'
import type { DemandDraft, DemandEditorProps } from './types'
import { demandWindow, draftsChanged, isValidDraft, newDraft, phaseName, toDemandInput, toDrafts } from './utils'

/** Edits the full demand list and saves it at once; read-only for people who cannot edit the project. */
export function DemandEditor({ project, expertises, editable }: DemandEditorProps) {
  const { t } = useTranslation()
  const [drafts, setDrafts] = useState(() => toDrafts(project.demands))
  const save = useReplaceDemands(project.id)
  const changed = draftsChanged(project.demands, drafts)
  const valid = drafts.every(isValidDraft)
  const empty = editable ? drafts.length === 0 : project.demands.length === 0

  const handleChange = (index: number, draft: DemandDraft) =>
    setDrafts((current) => current.map((row, i) => (i === index ? draft : row)))
  const handleRemove = (index: number) => setDrafts((current) => current.filter((_, i) => i !== index))
  const handleAdd = () => setDrafts((current) => [...current, newDraft(expertises)])
  const handleReset = () => setDrafts(toDrafts(project.demands))
  const handleSaved = () => toast.success(t('Demand saved.'))
  const handleSave = () => save.mutate(drafts.map(toDemandInput), { onSuccess: handleSaved })

  return (
    <Section
      title={t('Demand')}
      description={t(
        'Budget hours per expertise and minimum level. Without a phase, a demand spans the whole project.',
      )}
      action={
        editable && (
          <Button variant="outline" size="sm" disabled={expertises.length === 0} onClick={handleAdd}>
            <Plus />
            {t('Add demand')}
          </Button>
        )
      }
      footer={
        editable && (
          <>
            <Button variant="ghost" disabled={!changed || save.isPending} onClick={handleReset}>
              {t('Reset')}
            </Button>
            <Button disabled={!changed || !valid || save.isPending} onClick={handleSave}>
              {t('Save demand')}
            </Button>
          </>
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
          {editable
            ? drafts.map((draft, index) => (
                <DemandRow
                  key={draft.key}
                  index={index}
                  draft={draft}
                  expertises={expertises}
                  phases={project.phases}
                  window={demandWindow(project, draft.phase_id || null)}
                  onChange={handleChange}
                  onRemove={handleRemove}
                />
              ))
            : project.demands.map((demand) => (
                <DemandViewRow key={demand.id} demand={demand} phase={phaseName(project, demand.phase_id)} />
              ))}
          {empty && (
            <TableEmpty
              columns={editable ? DEMAND_EDIT_COLUMNS : DEMAND_TABLE_COLUMNS}
              icon={ClipboardList}
              title={t('No demand yet.')}
              description={t('Add the hours each expertise needs, in a phase or across the whole project.')}
            />
          )}
        </TableBody>
      </Table>
    </Section>
  )
}
