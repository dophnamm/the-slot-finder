import type { PlainKey } from '@/shared/i18n'
import type { AssumptionType } from './types'

export const SCENARIOS_QUERY_KEY = ['scenarios'] as const
export const scenarioQueryKey = (id: string) => ['scenarios', id] as const

export const SHIFT_PROJECT: AssumptionType = 'shift_project'
export const ADD_ABSENCE: AssumptionType = 'add_absence'

export const ASSUMPTION_TYPE_LABELS: Record<AssumptionType, PlainKey> = {
  shift_project: 'Move a project',
  add_absence: 'Add leave',
}

/** The most weeks a project may move; matches the server's SCENARIO_MAX_SHIFT_WEEKS default. */
export const SCENARIO_MAX_SHIFT_WEEKS = 12

/** Hours a day a new leave assumption starts with. */
export const DEFAULT_LEAVE_HOURS_PER_DAY = 8
