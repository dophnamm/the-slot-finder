import type { HolidaySource } from '../../types'

/** The import form as its inputs hold it. */
export interface ImportHolidaysValues {
  source: HolidaySource
  file: File | null
  url: string
  from_date: string
  to_date: string
  hours_per_day: string
}

/** Text fields of the form, matched by their name attribute. */
export type ImportTextField = 'url' | 'from_date' | 'to_date' | 'hours_per_day'
