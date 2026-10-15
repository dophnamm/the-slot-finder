/** How the hours in the dialog are read: the same hours every week, or a total spread over the weeks. */
export const ALLOCATE_MODE = {
  perWeek: 'per-week',
  total: 'total',
} as const

/** A spread total is rounded down to half hours; the remainder goes to the last week. */
export const SPREAD_STEP_HOURS = 0.5
