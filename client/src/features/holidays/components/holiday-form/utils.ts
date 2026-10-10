import type { CreateHolidayInput } from '../../types'
import type { HolidayFormValues } from './types'

/** Both dates set, the end not before the start, and positive hours. */
export function canSubmit(values: HolidayFormValues): boolean {
  return values.from_date !== '' && values.to_date !== '' && values.to_date >= values.from_date && Number(values.hours_per_day) > 0
}

export function toInput(values: HolidayFormValues): CreateHolidayInput {
  return { from_date: values.from_date, to_date: values.to_date, hours_per_day: Number(values.hours_per_day) }
}
