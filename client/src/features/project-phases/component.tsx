import { Layers, Plus } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'
import { useCurrentUser } from '@/shared/auth'
import { canEditProject, type Phase, useProject, useReplacePhases, useUpdatePhase } from '@/shared/programs'
import { Section } from '@/shared/components/section'
import { TableEmpty } from '@/shared/components/table-empty'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { PhaseEditRow } from './components/phase-edit-row'
import { PhaseRow } from './components/phase-row'
import { PhaseTimeline } from './components/phase-timeline'
import { NEW_PHASE_KEY, PHASE_TABLE_COLUMNS } from './constants'
import type { PhaseDraft, ProjectPhasesProps } from './types'
import {
  endsLate,
  isValidDraft,
  newPhaseDraft,
  phaseChanges,
  previewPhases,
  slipOf,
  timelineRows,
  timelineWeeks,
  toDraft,
  toPhaseInput,
} from './utils'

/** Phase timeline and editor. Editing one phase at a time previews how later phases move before saving. */
export function ProjectPhases({ projectId }: ProjectPhasesProps) {
  const { t } = useTranslation()
  const { data: user } = useCurrentUser()
  const { data: project } = useProject(projectId)
  const [draft, setDraft] = useState<PhaseDraft | null>(null)
  const updatePhase = useUpdatePhase()
  const replacePhases = useReplacePhases(projectId)
  const busy = updatePhase.isPending || replacePhases.isPending

  if (!project) return null

  const { phases } = project
  const editable = canEditProject(user, project)
  const preview = previewPhases(phases, draft)
  const weeks = timelineWeeks(project, preview)
  const rows = timelineRows(phases, preview, weeks)
  const slip = slipOf(phases, draft)
  const late = draft !== null && endsLate(preview, project.deadline_week)

  const handleEdit = (phase: Phase) => setDraft(toDraft(phase))
  const handleAdd = () => setDraft(newPhaseDraft(project))
  const handleCancel = () => setDraft(null)
  const handlePhaseSaved = (changed: Phase[]) => {
    const moved = changed.length - 1
    if (moved === 0) toast.success(t('Phase saved.'))
    else if (moved === 1) toast.success(t('Phase saved. {count} later phase moved.', { count: moved }))
    else toast.success(t('Phase saved. {count} later phases moved.', { count: moved }))
    setDraft(null)
  }
  const handleListSaved = () => {
    toast.success(t('Phases saved.'))
    setDraft(null)
  }
  const handleSave = () => {
    if (!draft) return
    if (draft.id === null) {
      replacePhases.mutate([...phases.map(toDraft), draft].map(toPhaseInput), { onSuccess: handleListSaved })
    } else {
      updatePhase.mutate({ phaseId: draft.id, changes: phaseChanges(phases, draft) }, { onSuccess: handlePhaseSaved })
    }
  }
  const handleRemove = (phase: Phase) => {
    if (!window.confirm(t('Remove the phase {name}?', { name: phase.name }))) return
    const kept = phases.filter((other) => other.id !== phase.id).map(toDraft)
    replacePhases.mutate(kept.map(toPhaseInput), { onSuccess: handleListSaved })
  }

  return (
    <Section
      title={t('Phases')}
      description={t('A later end week moves every later phase by the same number of weeks.')}
      action={
        editable && (
          <Button variant="outline" size="sm" disabled={draft !== null || busy} onClick={handleAdd}>
            <Plus />
            {t('Add phase')}
          </Button>
        )
      }
      contentClassName="flex flex-col gap-4"
    >
      {rows.length > 0 && <PhaseTimeline weeks={weeks} deadlineWeek={project.deadline_week} rows={rows} />}
      {slip.phases > 0 && (
        <p className="text-sm">
          {slip.weeks === 1
            ? t('Saving moves every later phase by 1 week.')
            : t('Saving moves every later phase by {weeks} weeks.', { weeks: slip.weeks })}
        </p>
      )}
      {late && <p className="text-sm text-destructive">{t('The last phase would end after the deadline week.')}</p>}
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t('Phase')}</TableHead>
            <TableHead>{t('Start week')}</TableHead>
            <TableHead>{t('End week')}</TableHead>
            <TableHead>{t('Dates')}</TableHead>
            <TableHead />
          </TableRow>
        </TableHeader>
        <TableBody>
          {phases.map((phase) =>
            draft?.id === phase.id ? (
              <PhaseEditRow
                key={phase.id}
                draft={draft}
                canSave={isValidDraft(draft)}
                busy={busy}
                onChange={setDraft}
                onSave={handleSave}
                onCancel={handleCancel}
              />
            ) : (
              <PhaseRow
                key={phase.id}
                phase={phase}
                editable={editable}
                disabled={busy || draft !== null}
                onEdit={handleEdit}
                onRemove={handleRemove}
              />
            ),
          )}
          {draft?.id === null && (
            <PhaseEditRow
              key={NEW_PHASE_KEY}
              draft={draft}
              canSave={isValidDraft(draft)}
              busy={busy}
              onChange={setDraft}
              onSave={handleSave}
              onCancel={handleCancel}
            />
          )}
          {phases.length === 0 && draft === null && (
            <TableEmpty
              columns={PHASE_TABLE_COLUMNS}
              icon={Layers}
              title={t('No phases yet')}
              description={t('Split the project into phases such as Discovery, Build and Test.')}
            />
          )}
        </TableBody>
      </Table>
    </Section>
  )
}
