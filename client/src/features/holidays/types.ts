/** A company-wide holiday; dates are inclusive, e.g. `2026-12-24`. */
export interface Holiday {
  id: string
  from_date: string
  to_date: string
  hours_per_day: number
}

export interface CreateHolidayInput {
  from_date: string
  to_date: string
  hours_per_day: number
}
