import {
  ADD_ABSENCE,
  type AssumptionInput,
  type AssumptionType,
  DEFAULT_LEAVE_HOURS_PER_DAY,
  SCENARIO_MAX_SHIFT_WEEKS,
  type ScenarioInput,
  SHIFT_PROJECT,
} from '@/shared/scenarios'
import type { AssumptionDraft } from './types'

/** An empty assumption of the given type. */
export function newAssumption(type: AssumptionType = SHIFT_PROJECT): AssumptionDraft {
  return {
    key: crypto.randomUUID(),
    type,
    project_id: '',
    weeks: '',
    person_id: '',
    from_date: '',
    to_date: '',
    hours_per_day: String(DEFAULT_LEAVE_HOURS_PER_DAY),
  }
}

/** Whether one assumption has everything the server needs. */
export function isComplete(draft: AssumptionDraft): boolean {
  if (draft.type === SHIFT_PROJECT) {
    const weeks = Number(draft.weeks)
    return (
      draft.project_id !== '' && Number.isInteger(weeks) && weeks !== 0 && Math.abs(weeks) <= SCENARIO_MAX_SHIFT_WEEKS
    )
  }
  const hours = Number(draft.hours_per_day)
  return draft.person_id !== '' && draft.from_date !== '' && draft.to_date >= draft.from_date && hours > 0
}

export function canCreate(name: string, drafts: AssumptionDraft[]): boolean {
  return name.trim() !== '' && drafts.length > 0 && drafts.every(isComplete)
}

/** The request body: only the fields each type uses. */
export function toInput(name: string, drafts: AssumptionDraft[]): ScenarioInput {
  const assumptions = drafts.map((draft): AssumptionInput =>
    draft.type === ADD_ABSENCE
      ? {
          type: draft.type,
          person_id: draft.person_id,
          from_date: draft.from_date,
          to_date: draft.to_date,
          hours_per_day: Number(draft.hours_per_day),
        }
      : { type: draft.type, project_id: draft.project_id, weeks: Number(draft.weeks) },
  )
  return { name: name.trim(), assumptions }
}
