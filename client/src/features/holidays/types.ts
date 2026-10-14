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

/** Where an imported holiday calendar comes from. */
export type HolidaySource = 'file' | 'link'

/** A holiday calendar to import: an ICS file or a link, the dates to keep and the hours a holiday takes off. */
export interface ImportHolidaysInput {
  source: HolidaySource
  file: File | null
  url: string
  from_date: string
  to_date: string
  hours_per_day: number
}

/** How many holidays an import added, and how many were there already. */
export interface HolidayImportResult {
  added: number
  skipped: number
}
