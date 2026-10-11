/** Limits of one demand row; they match the server's DemandRequest validation. */
export const DEMAND_LIMITS = {
  minHeadcount: 1,
  maxHeadcount: 20,
  maxHours: 100000,
} as const

/** Values a newly added row starts with. */
export const NEW_DEMAND = {
  minLevel: 3,
  headcount: '1',
  totalHours: '40',
} as const
