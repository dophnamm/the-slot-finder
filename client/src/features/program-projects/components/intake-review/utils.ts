import type { IntakeSuggestion } from '@/shared/ai'
import type { Expertise } from '@/shared/expertises'
import type { DemandInput, ProjectInput } from '@/shared/programs'
import { SKILL_LEVELS } from '@/utils/constants'
import { isDateInWindow } from '@/utils/functions'
import { DEFAULT_MIN_LEVEL, INTAKE_HOURS_MAX } from './constants'
import type { ReviewDemand, ReviewValues } from './types'

/** The form a suggestion fills in; a field the AI did not find stays empty. */
export function initialValues(suggestion: IntakeSuggestion): ReviewValues {
  return {
    name: suggestion.name ?? '',
    start_earliest: suggestion.start_earliest ?? '',
    deadline: suggestion.deadline ?? '',
    scope_points: suggestion.scope_points === null ? '' : String(suggestion.scope_points),
    demands: suggestion.demands.map((demand, index) => ({
      key: String(index),
      expertise_id: demand.expertise_id,
      total_hours: String(demand.total_hours),
      min_level: demand.min_level ? String(demand.min_level) : '',
      quote: demand.quote,
    })),
  }
}

/** Empty is fine (the default level applies); otherwise a whole level of the scale. */
function levelOk(level: string): boolean {
  return level === '' || (SKILL_LEVELS as readonly number[]).includes(Number(level))
}

function demandOk(demand: ReviewDemand, expertises: Expertise[]): boolean {
  const hours = Number(demand.total_hours)
  return (
    expertises.some((expertise) => expertise.id === demand.expertise_id) &&
    Number.isInteger(hours) &&
    hours > 0 &&
    hours <= INTAKE_HOURS_MAX &&
    levelOk(demand.min_level)
  )
}

/** A name, both dates inside the planning window and not reversed, scope points, and every demand complete. */
export function canCreate(values: ReviewValues, expertises: Expertise[]): boolean {
  const scope = Number(values.scope_points)
  return (
    values.name.trim() !== '' &&
    isDateInWindow(values.start_earliest) &&
    isDateInWindow(values.deadline) &&
    values.deadline >= values.start_earliest &&
    values.scope_points.trim() !== '' &&
    Number.isInteger(scope) &&
    scope >= 0 &&
    values.demands.every((demand) => demandOk(demand, expertises))
  )
}

export function toProjectInput(values: ReviewValues): ProjectInput {
  return {
    name: values.name.trim(),
    start_earliest: values.start_earliest,
    deadline: values.deadline,
    scope_points: Number(values.scope_points),
    done_points: 0,
  }
}

/** One person, the whole project, no phase; the profile is the server's default (flat). */
export function toDemandInputs(demands: ReviewDemand[]): DemandInput[] {
  return demands.map((demand) => ({
    expertise_id: demand.expertise_id,
    phase_id: null,
    min_level: demand.min_level === '' ? DEFAULT_MIN_LEVEL : Number(demand.min_level),
    headcount: 1,
    total_hours: Number(demand.total_hours),
  }))
}
