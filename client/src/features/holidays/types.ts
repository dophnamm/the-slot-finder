/** A company-wide holiday; dates are inclusive, e.g. `2026-12-24`. */
export interface Holiday {
  id: string
  /** What the day is, e.g. National Day; may be empty. */
  name: string
  from_date: string
  to_date: string
  hours_per_day: number
}

export interface CreateHolidayInput {
  name: string
  from_date: string
  to_date: string
  hours_per_day: number
}
