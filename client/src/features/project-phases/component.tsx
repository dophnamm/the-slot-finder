import { Layers, Plus } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'
import { useCurrentUser } from '@/shared/auth'
import { canEditProject, type Phase, useProject, useReplacePhases, useUpdatePhase } from '@/shared/programs'
import { Section } from '@/shared/components/section'
import { TableEmpty } from '@/shared/components/table-empty'
import { TableLoading } from '@/shared/components/table-loading'
import { useTranslation } from '@/shared/i18n'
import { confirm } from '@/shared/stores/confirm-store'
import { Button } from '@/shared/ui/button'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { currentIsoWeek } from '@/utils/functions'
import { PhaseDialog } from './components/phase-dialog'
import { PhaseRow } from './components/phase-row'
import { PhaseTimeline } from './components/phase-timeline'
import { PHASE_LOADING_ROWS, PHASE_TABLE_COLUMNS } from './constants'
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

/**
 * Phase timeline and table. Adding or editing a phase opens a dialog, whose timeline previews how later
 * phases move before saving.
 */
export function ProjectPhases({ projectId }: ProjectPhasesProps) {
  const { t } = useTranslation()
  const { data: user } = useCurrentUser()
  const { data: project, isPending } = useProject(projectId)
  const [draft, setDraft] = useState<PhaseDraft | null>(null)
  const updatePhase = useUpdatePhase()
  const replacePhases = useReplacePhases(projectId)
  const busy = updatePhase.isPending || replacePhases.isPending

  if (isPending)
    return (
      <Section
        title={t('Phases')}
        description={t('A later end week moves every later phase by the same number of weeks.')}
      >
        <Table>
          <TableBody>
            <TableLoading columns={PHASE_TABLE_COLUMNS} rows={PHASE_LOADING_ROWS} />
          </TableBody>
        </Table>
      </Section>
    )
  if (!project) return null

  const { phases } = project
  const editable = canEditProject(user, project)
  const preview = previewPhases(phases, draft)
  const weeks = timelineWeeks(project, preview)
  const rows = timelineRows(phases, preview, weeks)
  // The page shows the saved phases; an unsaved change is previewed in the dialog only.
  const saved = previewPhases(phases, null)
  const savedWeeks = timelineWeeks(project, saved)
  const savedRows = timelineRows(phases, saved, savedWeeks)
  const thisWeek = currentIsoWeek()
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
  const handleRemove = async (phase: Phase) => {
    const confirmed = await confirm({
      title: t('Remove the phase {name}?', { name: phase.name }),
      confirmLabel: t('Remove'),
      destructive: true,
    })
    if (!confirmed) return
    const kept = phases.filter((other) => other.id !== phase.id).map(toDraft)
    replacePhases.mutate(kept.map(toPhaseInput), { onSuccess: handleListSaved })
  }

  return (
    <Section
      title={t('Phases')}
      description={t('A later end week moves every later phase by the same number of weeks.')}
      action={
        editable && (
          <Button variant="outline" disabled={draft !== null || busy} onClick={handleAdd}>
            <Plus />
            {t('Add phase')}
          </Button>
        )
      }
      contentClassName="flex flex-col gap-4"
    >
      {savedRows.length > 0 && (
        <PhaseTimeline
          weeks={savedWeeks}
          deadlineWeek={project.deadline_week}
          currentWeek={thisWeek}
          rows={savedRows}
        />
      )}
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
          {phases.map((phase) => (
            <PhaseRow
              key={phase.id}
              phase={phase}
              editable={editable}
              disabled={busy || draft !== null}
              onEdit={handleEdit}
              onRemove={handleRemove}
            />
          ))}
          {phases.length === 0 && (
            <TableEmpty
              columns={PHASE_TABLE_COLUMNS}
              icon={Layers}
              title={t('No phases yet')}
              description={t('Split the project into phases such as Discovery, Build and Test.')}
            />
          )}
        </TableBody>
      </Table>
      {draft && (
        <PhaseDialog
          draft={draft}
          canSave={isValidDraft(draft)}
          busy={busy}
          preview={
            <div className="flex flex-col gap-3">
              {rows.length > 0 && (
                <PhaseTimeline weeks={weeks} deadlineWeek={project.deadline_week} currentWeek={thisWeek} rows={rows} />
              )}
              {slip.phases > 0 && (
                <p className="text-sm">
                  {slip.weeks === 1
                    ? t('Saving moves every later phase by 1 week.')
                    : t('Saving moves every later phase by {weeks} weeks.', { weeks: slip.weeks })}
                </p>
              )}
              {late && (
                <p className="text-sm text-destructive">{t('The last phase would end after the deadline week.')}</p>
              )}
            </div>
          }
          onChange={setDraft}
          onSave={handleSave}
          onCancel={handleCancel}
        />
      )}
    </Section>
  )
}
