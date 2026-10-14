import { DEFAULT_HOURS_PER_DAY } from '@/utils/constants'
import type { HolidayFormValues } from './types'

export const EMPTY_HOLIDAY: HolidayFormValues = {
  name: '',
  from_date: '',
  to_date: '',
  hours_per_day: String(DEFAULT_HOURS_PER_DAY),
}
