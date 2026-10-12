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

/** Columns of the demand table, read-only and with the remove column while editing. */
export const DEMAND_TABLE_COLUMNS = 6
export const DEMAND_EDIT_COLUMNS = 7
