import type { Phase, PhaseChanges, PhaseInput, ProjectDetail } from '@/shared/programs'
import { addIsoWeeks, isIsoWeek, isoWeekList, isoWeeksBetween } from '@/utils/functions'
import type { CellState, TimelineRowData } from './components/phase-timeline'
import { MAX_TIMELINE_WEEKS, NEW_PHASE_KEY, NEW_PHASE_WEEKS } from './constants'
import type { PhaseDraft, Slip } from './types'

export function toDraft(phase: Phase): PhaseDraft {
  return { id: phase.id, name: phase.name, start_week: phase.start_week, end_week: phase.end_week }
}

/** A phase starting the week after the last one, or at the project's start. */
export function newPhaseDraft(project: ProjectDetail): PhaseDraft {
  const last = project.phases.at(-1)
  const start = last ? addIsoWeeks(last.end_week, 1) : project.start_week
  return { id: null, name: '', start_week: start, end_week: addIsoWeeks(start, NEW_PHASE_WEEKS - 1) }
}

function hasWeeks(span: PhaseDraft): boolean {
  return isIsoWeek(span.start_week) && isIsoWeek(span.end_week)
}

/** A name and both weeks, the end not before the start. */
export function isValidDraft(draft: PhaseDraft): boolean {
  return draft.name.trim() !== '' && hasWeeks(draft) && isoWeeksBetween(draft.start_week, draft.end_week) >= 0
}

/** The server's rule: a later end moves every later phase by the same weeks; an earlier end moves nothing. */
export function slipOf(phases: Phase[], draft: PhaseDraft | null): Slip {
  const edited = phases.find((phase) => phase.id === draft?.id)
  if (!draft || !edited || !isIsoWeek(draft.end_week)) return { weeks: 0, phases: 0 }
  const weeks = isoWeeksBetween(edited.end_week, draft.end_week)
  if (weeks <= 0) return { weeks: 0, phases: 0 }
  return { weeks, phases: phases.filter((phase) => phase.sequence > edited.sequence).length }
}

/** The phases as they would be after saving the draft. */
export function previewPhases(phases: Phase[], draft: PhaseDraft | null): PhaseDraft[] {
  const saved = phases.map(toDraft)
  if (!draft) return saved
  if (draft.id === null) return [...saved, draft]
  const edited = phases.find((phase) => phase.id === draft.id)
  const slip = slipOf(phases, draft)
  return phases.map((phase, index) => {
    if (phase.id === draft.id) return draft
    if (slip.weeks > 0 && edited && phase.sequence > edited.sequence) {
      return {
        ...saved[index],
        start_week: addIsoWeeks(phase.start_week, slip.weeks),
        end_week: addIsoWeeks(phase.end_week, slip.weeks),
      }
    }
    return saved[index]
  })
}

/** ISO week labels have a fixed width, so text order is week order. */
function earliest(weeks: string[]): string {
  return [...weeks].sort()[0]
}

function latest(weeks: string[]): string {
  return [...weeks].sort().at(-1) ?? weeks[0]
}

/** The project's start through its deadline, widened to fit every phase. */
export function timelineWeeks(project: ProjectDetail, spans: PhaseDraft[]): string[] {
  const shown = spans.filter(hasWeeks)
  const from = earliest([project.start_week, ...shown.map((span) => span.start_week)])
  let to = latest([project.deadline_week, ...shown.map((span) => span.end_week)])
  if (isoWeeksBetween(from, to) >= MAX_TIMELINE_WEEKS) to = addIsoWeeks(from, MAX_TIMELINE_WEEKS - 1)
  return isoWeekList(from, to)
}

function covers(span: PhaseDraft | Phase, week: string): boolean {
  return isIsoWeek(span.start_week) && isIsoWeek(span.end_week) && week >= span.start_week && week <= span.end_week
}

function cellState(week: string, span: PhaseDraft, before: Phase | undefined, moved: boolean): CellState {
  if (covers(span, week)) return moved ? 'moved' : 'planned'
  if (moved && before && covers(before, week)) return 'vacated'
  return 'empty'
}

export function timelineRows(phases: Phase[], spans: PhaseDraft[], weeks: string[]): TimelineRowData[] {
  return spans.map((span) => {
    const before = phases.find((phase) => phase.id === span.id)
    const moved = !before || before.start_week !== span.start_week || before.end_week !== span.end_week
    return {
      key: span.id ?? NEW_PHASE_KEY,
      name: span.name,
      cells: weeks.map((week) => cellState(week, span, before, moved)),
    }
  })
}

/** Whether the last phase would end after the deadline week. */
export function endsLate(spans: PhaseDraft[], deadlineWeek: string): boolean {
  const ends = spans.filter(hasWeeks).map((span) => span.end_week)
  return ends.length > 0 && latest(ends) > deadlineWeek
}

export function toPhaseInput(span: PhaseDraft): PhaseInput {
  return { id: span.id ?? undefined, name: span.name.trim(), start_week: span.start_week, end_week: span.end_week }
}

/** Only the fields of the draft that differ from the saved phase. */
export function phaseChanges(phases: Phase[], draft: PhaseDraft): PhaseChanges {
  const saved = phases.find((phase) => phase.id === draft.id)
  const name = draft.name.trim()
  const changes: PhaseChanges = {}
  if (name !== saved?.name) changes.name = name
  if (draft.start_week !== saved?.start_week) changes.start_week = draft.start_week
  if (draft.end_week !== saved?.end_week) changes.end_week = draft.end_week
  return changes
}
